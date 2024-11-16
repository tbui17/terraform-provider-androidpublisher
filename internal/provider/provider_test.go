// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
)

// testAccProtoV6ProviderFactories are used to instantiate a provider during
// acceptance testing. The factory function will be invoked for every Terraform
// CLI command executed to create a provider server to which the CLI can
// reattach.
var testAccProtoV6ProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){
	"androidpublisher": providerserver.NewProtocol6WithError(New("test")()),
}

type EnvironmentVariables struct {
	TestEmail             string
	TestDeveloperId       string
	GoogleCredentialsJson string
}

func NewEnvironmentVariables() EnvironmentVariables {
	res := EnvironmentVariables{
		TestEmail:       os.Getenv("TEST_EMAIL"),
		TestDeveloperId: os.Getenv("TEST_DEVELOPER_ID"),
	}
	return res
}

func testAccPreCheck(t *testing.T) {
	// You can add code here to run prior to any test case execution, for example assertions
	// about the appropriate environment variables being set are common to see in a pre-check
	// function.

	env := NewEnvironmentVariables()
	var missingVariables []string
	if env.TestEmail == "" {
		missingVariables = append(missingVariables, "TEST_EMAIL")
	}
	if env.TestDeveloperId == "" {
		missingVariables = append(missingVariables, "TEST_DEVELOPER_ID")
	}
	if len(missingVariables) > 0 {
		t.Fatalf("Environment variables missing: %v", missingVariables)
	}

}
