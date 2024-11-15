// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"os"
	"testing"
)

type EnvironmentVariables struct {
	TestEmail       string
	TestDeveloperId string
}

func NewEnvironmentVariables() EnvironmentVariables {
	res := EnvironmentVariables{
		TestEmail:       os.Getenv("TEST_EMAIL"),
		TestDeveloperId: os.Getenv("TEST_DEVELOPER_ID"),
	}

	if res.TestEmail == "" || res.TestDeveloperId == "" {
		panic("Please set TEST_EMAIL and TEST_DEVELOPER_ID environment variables")
	}
	return res
}

func TestAccUserResource(t *testing.T) {
	env := NewEnvironmentVariables()
	createConfig := fmt.Sprintf(`
resource "androidpublisher_user" "test" {
  email = %q
  developer_id = %q
  developer_account_permissions = [ "CAN_VIEW_APP_QUALITY_GLOBAL"]
}
`, env.TestEmail, env.TestDeveloperId)

	updateConfig := fmt.Sprintf(`
	resource "androidpublisher_user" "test" {
	 email = %q
	 developer_id = %q
	 developer_account_permissions = [ "CAN_VIEW_APP_QUALITY_GLOBAL","CAN_VIEW_NON_FINANCIAL_DATA_GLOBAL"]
	}
	`, env.TestEmail, env.TestDeveloperId)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: createConfig,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("androidpublisher_user.test", "email", env.TestEmail),
					resource.TestCheckResourceAttr("androidpublisher_user.test", "developer_id", env.TestDeveloperId),
					resource.TestCheckResourceAttr("androidpublisher_user.test", "developer_account_permissions.#", "1"),
					resource.TestCheckResourceAttr("androidpublisher_user.test", "developer_account_permissions.0", "CAN_VIEW_APP_QUALITY_GLOBAL"),
				),
			},
			// Update and Read testing
			{
				Config: updateConfig,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("androidpublisher_user.test", "email", env.TestEmail),
					resource.TestCheckResourceAttr("androidpublisher_user.test", "developer_id", env.TestDeveloperId),
					resource.TestCheckResourceAttr("androidpublisher_user.test", "developer_account_permissions.#", "2"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}
