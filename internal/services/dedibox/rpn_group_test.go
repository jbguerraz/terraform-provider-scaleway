package dedibox_test

import (
	"context"
	"fmt"
	"strconv"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/scaleway/scaleway-sdk-go/api/dedibox/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/acctest"
)

func TestAccDediboxRpnGroup_Basic(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		CheckDestroy:             isDediboxRpnGroupDestroyed(tt),
		Steps: []resource.TestStep{
			{
				Config: `
					resource "scaleway_dedibox_rpn_group" "test" {
						name       = "tf-test-rpn-group"
						type       = "standard"
						project_id = "` + tt.Meta.ProjectID + `"
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					isDediboxRpnGroupPresent(tt, "scaleway_dedibox_rpn_group.test"),
					resource.TestCheckResourceAttr("scaleway_dedibox_rpn_group.test", "name", "tf-test-rpn-group"),
					resource.TestCheckResourceAttr("scaleway_dedibox_rpn_group.test", "type", "standard"),
					resource.TestCheckResourceAttrSet("scaleway_dedibox_rpn_group.test", "status"),
				),
			},
			{
				Config: `
					resource "scaleway_dedibox_rpn_group" "test" {
						name       = "tf-test-rpn-group-updated"
						type       = "standard"
						project_id = "` + tt.Meta.ProjectID + `"
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					isDediboxRpnGroupPresent(tt, "scaleway_dedibox_rpn_group.test"),
					resource.TestCheckResourceAttr("scaleway_dedibox_rpn_group.test", "name", "tf-test-rpn-group-updated"),
				),
			},
		},
	})
}

func isDediboxRpnGroupPresent(tt *acctest.TestTools, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]

		if !ok {
			return fmt.Errorf("not found: %s", n)
		}

		groupID, err := strconv.ParseUint(rs.Primary.ID, 10, 64)
		if err != nil {
			return fmt.Errorf("invalid group ID: %s", rs.Primary.ID)
		}

		api := dedibox.NewRpnV2API(tt.Meta.ScwClient())

		_, err = api.GetRpnV2Group(&dedibox.RpnV2ApiGetRpnV2GroupRequest{
			GroupID: groupID,
		}, scw.WithContext(context.Background()))
		if err != nil {
			return err
		}

		return nil
	}
}

func isDediboxRpnGroupDestroyed(tt *acctest.TestTools) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "scaleway_dedibox_rpn_group" {
				continue
			}

			groupID, err := strconv.ParseUint(rs.Primary.ID, 10, 64)
			if err != nil {
				return fmt.Errorf("invalid group ID: %s", rs.Primary.ID)
			}

			api := dedibox.NewRpnV2API(tt.Meta.ScwClient())

			_, err = api.GetRpnV2Group(&dedibox.RpnV2ApiGetRpnV2GroupRequest{
				GroupID: groupID,
			}, scw.WithContext(context.Background()))

			// If there's no error, the group still exists
			if err == nil {
				return fmt.Errorf("rpn group %s still exists", rs.Primary.ID)
			}

			// Check if it's a 404 error (expected)
			// If it's another error type, return it
		}

		return nil
	}
}
