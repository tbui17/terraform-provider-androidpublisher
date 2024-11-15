// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccUserDataSource(t *testing.T) {
	env := NewEnvironmentVariables()
	config := fmt.Sprintf(`
data "androidpublisher_user" "test" {
  developer_id = %q
}
`, env.TestDeveloperId)
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Read testing
			{
				Config: config,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.androidpublisher_user.test", "developer_id", env.TestDeveloperId),
					resource.TestCheckResourceAttrWith("data.androidpublisher_user.test", "value.#", testCheckResourceCountNotEmpty),
				),
			},
		},
	})
}

func testCheckResourceCountNotEmpty(inp string) error {

	i, err := strconv.Atoi(inp)
	if err != nil {
		return err
	}
	if i == 0 {
		return fmt.Errorf("expected at least one resource, got zero")
	}
	return nil
}
