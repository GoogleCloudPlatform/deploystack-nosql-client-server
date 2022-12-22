package main

import (
	"fmt"
	"os"
	"testing"

	"github.com/GoogleCloudPlatform/deploystack"
	"github.com/GoogleCloudPlatform/deploystack/dstester"
)

var (
	ops              = dstester.NewOperationsSet()
	project, _       = deploystack.ProjectID()
	projectNumber, _ = deploystack.ProjectNumber(project)
	basename         = "nosql-client-server"
	debug            = false

	tf = dstester.Terraform{
		Dir: "../terraform",
		Vars: map[string]string{
			"project_id":     project,
			"project_number": projectNumber,
			"region":         "us-central1",
			"zone":           "us-central1-a",
			"basename":       basename,
		},
	}

	resources = dstester.Resources{
		Project: project,
		Items: []dstester.Resource{
			{
				Product: "compute instances",
				Name:    "client",
			},
			{
				Product: "compute instances",
				Name:    "server",
			},
			{
				Product: "compute networks",
				Name:    fmt.Sprintf("%s-network", basename),
			},
			{
				Product: "compute firewall-rules",
				Name:    "deploystack-allow-http",
			},
			{
				Product: "compute firewall-rules",
				Name:    "deploystack-allow-ssh",
			},
			{
				Product: "compute firewall-rules",
				Name:    "deploystack-allow-internal",
			},
		},
	}
)

func init() {
	if os.Getenv("debug") != "" {
		debug = true
	}

	ops.Add("postApply", dstester.Operation{Output: "client_url", Type: "httpPoll"})
}

func TestListCommands(t *testing.T) {
	resources.Init()
	dstester.DebugCommands(t, tf, resources)
}

func TestStack(t *testing.T) {
	dstester.TestStack(t, tf, resources, ops, debug)
}

func TestClean(t *testing.T) {
	if os.Getenv("clean") == "" {
		t.Skip("Clean must be very intentionally called")
	}

	resources.Init()
	dstester.Clean(t, tf, resources)
}
