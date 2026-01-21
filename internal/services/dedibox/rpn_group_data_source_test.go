package dedibox_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/acctest"
)

func TestAccDataSourceDediboxRpnGroup_Basic(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
					resource "scaleway_dedibox_rpn_group" "test" {
						name       = "tf-test-rpn-group-ds"
						type       = "standard"
						project_id = "` + tt.Meta.ProjectID + `"
					}

					data "scaleway_dedibox_rpn_group" "test" {
						group_id = scaleway_dedibox_rpn_group.test.id
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair("data.scaleway_dedibox_rpn_group.test", "name", "scaleway_dedibox_rpn_group.test", "name"),
					resource.TestCheckResourceAttrPair("data.scaleway_dedibox_rpn_group.test", "type", "scaleway_dedibox_rpn_group.test", "type"),
					resource.TestCheckResourceAttrSet("data.scaleway_dedibox_rpn_group.test", "status"),
				),
			},
		},
	})
}
