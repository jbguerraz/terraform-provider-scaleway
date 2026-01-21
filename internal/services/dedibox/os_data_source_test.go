package dedibox_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/scaleway/scaleway-sdk-go/api/dedibox/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/acctest"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/zonal"
)

func TestAccDataSourceDediboxOS_Basic(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
					data "scaleway_dedibox_os" "test1" {
						zone = "%s"
						name = "Ubuntu"
					}
				`, DediboxZone),
				Check: resource.ComposeTestCheckFunc(
					isDediboxOSPresent(tt, "data.scaleway_dedibox_os.test1"),
					resource.TestCheckResourceAttrSet("data.scaleway_dedibox_os.test1", "os_id"),
					resource.TestCheckResourceAttrSet("data.scaleway_dedibox_os.test1", "name"),
					resource.TestCheckResourceAttrSet("data.scaleway_dedibox_os.test1", "version"),
					resource.TestCheckResourceAttrSet("data.scaleway_dedibox_os.test1", "arch"),
				),
			},
		},
	})
}

func TestAccDataSourceDediboxOS_ByType(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
					data "scaleway_dedibox_os" "test1" {
						zone = "%s"
						name = "Ubuntu"
						type = "server"
					}
				`, DediboxZone),
				Check: resource.ComposeTestCheckFunc(
					isDediboxOSPresent(tt, "data.scaleway_dedibox_os.test1"),
					resource.TestCheckResourceAttr("data.scaleway_dedibox_os.test1", "type", "server"),
				),
			},
		},
	})
}

func isDediboxOSPresent(tt *acctest.TestTools, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]

		if !ok {
			return fmt.Errorf("not found: %s", n)
		}

		zone, id, err := zonal.ParseID(rs.Primary.ID)
		if err != nil {
			return err
		}

		api := dedibox.NewAPI(tt.Meta.ScwClient())

		osList, err := api.ListOS(&dedibox.ListOSRequest{
			Zone: zone,
		}, scw.WithAllPages(), scw.WithContext(context.Background()))
		if err != nil {
			return err
		}

		for _, os := range osList.Os {
			if fmt.Sprintf("%d", os.ID) == id {
				return nil
			}
		}

		return fmt.Errorf("os %s not found", id)
	}
}
