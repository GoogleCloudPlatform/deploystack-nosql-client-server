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

variable "project_id" {
  type = string
}

variable "project_number" {
  type = string
}

variable "zone" {
  type = string
}

variable "region" {
  type = string
}

variable "basename" {
  type = string
}

# Enabling services in your GCP project
variable "gcp_service_list" {
  description = "The list of apis necessary for the project"
  type        = list(string)
  default = [
    "compute.googleapis.com",
  ]
}

resource "google_project_service" "all" {
  for_each                   = toset(var.gcp_service_list)
  project                    = var.project_number
  service                    = each.key
  disable_dependent_services = false
  disable_on_destroy         = false
}


resource "google_compute_firewall" "default-allow-http" {
  name    = "default-allow-http"
  project = var.project_number
  network = "projects/${var.project_id}/global/networks/default"

  allow {
    protocol = "tcp"
    ports    = ["80"]
  }

  source_ranges = ["0.0.0.0/0"]

  target_tags = ["http-server"]
}

# Create Instances
resource "google_compute_instance" "server" {
  name         = "server"
  zone         = var.zone
  project      = var.project_id
  machine_type = "e2-standard-2"
  tags         = ["http-server"]


  boot_disk {
    auto_delete = true
    device_name = "server"
    initialize_params {
      image = "family/ubuntu-1804-lts"
      size  = 10
      type  = "pd-standard"
    }
  }

  network_interface {
    network = "default"
    access_config {
      // Ephemeral public IP
    }
  }

  metadata_startup_script = <<SCRIPT
    apt-get update
    apt-get install -y mongodb
    service mongodb stop
    sed -i 's/bind_ip = 127.0.0.1/bind_ip = 0.0.0.0/' /etc/mongodb.conf
    iptables -t nat -A PREROUTING -p tcp --dport 80 -j REDIRECT --to-port 27017
    service mongodb start
  SCRIPT
  depends_on              = [google_project_service.all]
}

resource "google_compute_instance" "client" {
  name         = "client"
  zone         = var.zone
  project      = var.project_id
  machine_type = "e2-standard-2"
  tags         = ["http-server", "https-server"]

  boot_disk {
    auto_delete = true
    device_name = "client"
    initialize_params {
      image = "family/ubuntu-1804-lts"
      size  = 10
      type  = "pd-standard"
    }
  }

  network_interface {
    network = "default"

    access_config {
      // Ephemeral public IP
    }
  }

  metadata_startup_script = <<SCRIPT
    add-apt-repository ppa:longsleep/golang-backports -y && \
    apt update -y && \
    apt install golang-go -y
    mkdir /modcache
    mkdir /go
    mkdir /app && cd /app
    curl https://raw.githubusercontent.com/GoogleCloudPlatform/golang-samples/main/compute/quickstart/compute_quickstart_sample.go --output main.go
    go mod init exec
    GOPATH=/go GOMODCACHE=/modcache GOCACHE=/modcache go mod tidy
    GOPATH=/go GOMODCACHE=/modcache GOCACHE=/modcache go get -u 
    sed -i 's/mongoport = "80"/mongoport = "27017"/' /app/main.go
    echo "GOPATH=/go GOMODCACHE=/modcache GOCACHE=/modcache HOST=${google_compute_instance.server.network_interface.0.network_ip} go run main.go"
    GOPATH=/go GOMODCACHE=/modcache GOCACHE=/modcache HOST=${google_compute_instance.server.network_interface.0.network_ip} go run main.go & 
  SCRIPT

  depends_on = [google_project_service.all]
}

output "client_url" {
  value = "http://${google_compute_instance.client.network_interface.0.access_config.0.nat_ip}"
}
