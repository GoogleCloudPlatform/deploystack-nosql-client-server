/**
 * Copyright 2022 Google LLC
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *      http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

terraform {
  required_providers {
    google = {
      source = "hashicorp/google"
    }
    google-beta = {
      source = "hashicorp/google-beta"
    }
    random = {
      source = "hashicorp/random"
    }
  }
}

locals {
  mongo_user = "trainer_app"
}

resource "random_password" "mongo_password" {
  length  = 32
  special = false
}

# Enabling services in your GCP project


data "google_project" "project" {
  project_id = var.project_id
}

resource "google_project_service" "all" {
  for_each                   = toset(var.gcp_service_list)
  project                    = data.google_project.project.number
  service                    = each.key
  disable_dependent_services = false
  disable_on_destroy         = false
}


data "google_compute_network" "default" {
  project    = var.project_id
  name       = "default"
  depends_on = [google_project_service.all]
}

resource "google_compute_network" "main" {
  provider                = google-beta
  name                    = "${var.basename}-network"
  auto_create_subnetworks = true
  project                 = var.project_id
  depends_on              = [google_project_service.all]
}

resource "google_compute_router" "router" {
  name       = "${var.basename}-router"
  project    = var.project_id
  region     = var.region
  network    = google_compute_network.main.id
  depends_on = [google_project_service.all]
}

resource "google_compute_router_nat" "nat" {
  name                               = "${var.basename}-nat"
  project                            = var.project_id
  router                             = google_compute_router.router.name
  region                             = google_compute_router.router.region
  nat_ip_allocate_option             = "AUTO_ONLY"
  source_subnetwork_ip_ranges_to_nat = "ALL_SUBNETWORKS_ALL_IP_RANGES"
}

resource "google_compute_firewall" "default-allow-http" {
  name    = "deploystack-allow-http"
  project = data.google_project.project.number
  network = google_compute_network.main.name

  allow {
    protocol = "tcp"
    ports    = ["80"]
  }

  source_ranges = ["0.0.0.0/0"]

  target_tags = ["http-server"]
  depends_on  = [google_project_service.all]
}

resource "google_compute_firewall" "default-allow-internal" {
  name    = "deploystack-allow-internal"
  project = data.google_project.project.number
  network = google_compute_network.main.name

  allow {
    protocol = "tcp"
    ports    = ["27017"]
  }

  source_tags = ["mongo-client"]
  target_tags = ["mongo-server"]
  depends_on  = [google_project_service.all]
}

resource "google_compute_firewall" "default-allow-ssh" {
  name    = "deploystack-allow-ssh"
  project = data.google_project.project.number
  network = google_compute_network.main.name

  allow {
    protocol = "tcp"
    ports    = ["22"]
  }

  source_ranges = ["0.0.0.0/0"]

  target_tags = ["ssh-server"]
  depends_on  = [google_project_service.all]
}

# Create Instances
resource "google_compute_instance" "server" {
  name                      = "server"
  zone                      = var.zone
  project                   = var.project_id
  machine_type              = "e2-standard-2"
  tags                      = ["ssh-server", "mongo-server"]
  allow_stopping_for_update = true


  boot_disk {
    auto_delete = true
    device_name = "server"
    initialize_params {
      image = "family/ubuntu-2004-lts"
      size  = 10
      type  = "pd-standard"
    }
  }

  network_interface {
    network = google_compute_network.main.name
  }

  service_account {
    scopes = ["https://www.googleapis.com/auth/logging.write"]
  }

  metadata = {
    mongo-user     = local.mongo_user
    mongo-password = random_password.mongo_password.result
  }

  metadata_startup_script = <<-SCRIPT
    #!/bin/bash
    set -euo pipefail
    trap 'rm -f /run/init_mongo.js' EXIT

    INTERNAL_IP=$(curl -sf -H "Metadata-Flavor: Google" http://metadata.google.internal/computeMetadata/v1/instance/network-interfaces/0/ip)
    MONGO_USER=$(curl -sf -H "Metadata-Flavor: Google" http://metadata.google.internal/computeMetadata/v1/instance/attributes/mongo-user)
    MONGO_PASS=$(curl -sf -H "Metadata-Flavor: Google" http://metadata.google.internal/computeMetadata/v1/instance/attributes/mongo-password)

    if [ ! -f /var/lib/mongodb/.auth_initialized ]; then
      export DEBIAN_FRONTEND=noninteractive
      apt-get update
      apt-get install -y mongodb

      systemctl stop mongodb
      sed -i 's/^bind_ip\s*=.*/bind_ip = 127.0.0.1/' /etc/mongodb.conf
      sed -i 's/^auth\s*=.*/#auth = true/' /etc/mongodb.conf
      systemctl start mongodb

      until mongo --quiet --host 127.0.0.1 --port 27017 --eval "db.adminCommand('ping')" >/dev/null 2>&1; do
        sleep 1
      done

      (
        umask 077
        cat <<EOF > /run/init_mongo.js
    var adminDb = db.getSiblingDB("admin");
    if (adminDb.getUser("$MONGO_USER") === null) {
      adminDb.createUser({
        user: "$MONGO_USER",
        pwd: "$MONGO_PASS",
        roles: [{ role: "readWrite", db: "test" }]
      });
    } else {
      adminDb.updateUser("$MONGO_USER", {
        pwd: "$MONGO_PASS",
        roles: [{ role: "readWrite", db: "test" }]
      });
    }
    EOF
      )
      mongo --quiet --host 127.0.0.1 --port 27017 /run/init_mongo.js >/dev/null
      rm -f /run/init_mongo.js

      touch /var/lib/mongodb/.auth_initialized
      chmod 600 /var/lib/mongodb/.auth_initialized
    fi

    systemctl stop mongodb
    sed -i "s/^bind_ip\s*=.*/bind_ip = 127.0.0.1,$INTERNAL_IP/" /etc/mongodb.conf
    grep -q "^bind_ip = 127.0.0.1,$INTERNAL_IP$" /etc/mongodb.conf || echo "bind_ip = 127.0.0.1,$INTERNAL_IP" >> /etc/mongodb.conf
    sed -i 's/^#*\s*auth\s*=.*/auth = true/' /etc/mongodb.conf
    grep -q '^auth = true$' /etc/mongodb.conf || echo 'auth = true' >> /etc/mongodb.conf
    systemctl start mongodb
  SCRIPT
  depends_on = [
    google_project_service.all,
    google_compute_router_nat.nat,
  ]
}

resource "google_compute_instance" "client" {
  name                      = "client"
  zone                      = var.zone
  project                   = var.project_id
  machine_type              = "e2-standard-2"
  tags                      = ["http-server", "https-server", "ssh-server", "mongo-client"]
  allow_stopping_for_update = true

  boot_disk {
    auto_delete = true
    device_name = "client"
    initialize_params {
      image = "family/ubuntu-2004-lts"
      size  = 10
      type  = "pd-standard"
    }
  }
  service_account {
    scopes = ["https://www.googleapis.com/auth/logging.write"]
  }

  network_interface {
    network = google_compute_network.main.name

    access_config {
      // Ephemeral public IP
    }
  }

  metadata = {
    mongo-user      = local.mongo_user
    mongo-password  = random_password.mongo_password.result
    client-main-go  = file("${path.module}/../code/client/main.go")
    client-model-go = file("${path.module}/../code/client/model.go")
  }

  metadata_startup_script = <<-SCRIPT
    #!/bin/bash
    set -euo pipefail

    MONGO_USER=$(curl -sf -H "Metadata-Flavor: Google" http://metadata.google.internal/computeMetadata/v1/instance/attributes/mongo-user)
    MONGO_PASS=$(curl -sf -H "Metadata-Flavor: Google" http://metadata.google.internal/computeMetadata/v1/instance/attributes/mongo-password)

    if [ ! -f /app/deploystack-nosql-client-server/code/client/trainers ]; then
      export DEBIAN_FRONTEND=noninteractive
      add-apt-repository ppa:longsleep/golang-backports -y
      apt-get update -y
      apt-get install -y golang-go git
      mkdir -p /modcache /go /app
      rm -rf /app/deploystack-nosql-client-server
      git clone https://github.com/GoogleCloudPlatform/deploystack-nosql-client-server.git /app/deploystack-nosql-client-server
      cd /app/deploystack-nosql-client-server/code/client
      curl -sf -H "Metadata-Flavor: Google" http://metadata.google.internal/computeMetadata/v1/instance/attributes/client-main-go > main.go
      curl -sf -H "Metadata-Flavor: Google" http://metadata.google.internal/computeMetadata/v1/instance/attributes/client-model-go > model.go
      GOPATH=/go GOMODCACHE=/modcache GOCACHE=/modcache go build -o trainers main.go model.go
    fi

    (
      umask 077
      cat <<EOF > /etc/trainers.env
    DBHOST=${google_compute_instance.server.network_interface.0.network_ip}
    DBUSER=$MONGO_USER
    DBPASS=$MONGO_PASS
    EOF
    )

    cat <<'EOF' > /etc/systemd/system/trainers.service
    [Unit]
    Description=DeployStack NoSQL Trainer Client
    After=network-online.target
    Wants=network-online.target

    [Service]
    Type=simple
    WorkingDirectory=/app/deploystack-nosql-client-server/code/client
    EnvironmentFile=/etc/trainers.env
    ExecStart=/app/deploystack-nosql-client-server/code/client/trainers
    Restart=always
    RestartSec=3

    [Install]
    WantedBy=multi-user.target
    EOF

    systemctl daemon-reload
    systemctl enable --now trainers.service
  SCRIPT

  depends_on = [google_project_service.all]
}
