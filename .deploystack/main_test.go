package main

import (
	"fmt"
	"testing"

	"github.com/GoogleCloudPlatform/deploystack"
	"github.com/GoogleCloudPlatform/deploystack/dstester"
)

var (
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

	resources = dstester.GCPResources{
		Project: project,
		Items: []dstester.GCPResource{
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

	checks = []dstester.Check{
		{
			Output: "client_url",
			Type:   "httpPoll",
		},
	}
)

func TestCreateDestroy(t *testing.T) {
	resources.Init()
	tf.InitApplyForTest(t, debug)
	dstester.TextExistence(t, resources.Items)

	dstester.TestChecks(t, checks, tf)

	tf.DestroyForTest(t, debug)
	dstester.TextNonExistence(t, resources.Items)
}

// func TestCreation(t *testing.T) {
// 	resources.Init()
// 	tf.InitApplyForTest(t, debug)
// 	dstester.TextExistence(t, resources.Items)
// }

// func TestPolls(t *testing.T) {
// 	dstester.TestChecks(t, checks, tf)
// }

// func TestCreateAndPoll(t *testing.T) {
// 	TestCreation(t)
// 	TestPolls(t)
// }

// func TestDestruction(t *testing.T) {
// 	tf.DestroyForTest(t, debug)
// 	dstester.TextNonExistence(t, resources.Items)
// }
