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
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/zonal"
)

// TestAccDediboxServer_Basic tests server creation
// WARNING: This test provisions real servers which incur costs.
// Run with: TF_ACC=1 go test -v -run TestAccDediboxServer_Basic
func TestAccDediboxServer_Basic(t *testing.T) {
	t.Skip("Skipping server test - provisions real hardware and incurs costs")

	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		CheckDestroy:             isDediboxServerDestroyed(tt),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
					data "scaleway_dedibox_offer" "test" {
						zone = "%s"
						name = "%s"
					}

					resource "scaleway_dedibox_server" "test" {
						zone       = "%s"
						offer_id   = data.scaleway_dedibox_offer.test.offer_id
						project_id = "`+tt.Meta.ProjectID+`"
					}
				`, DediboxZone, DediboxOfferName, DediboxZone),
				Check: resource.ComposeTestCheckFunc(
					isDediboxServerPresent(tt, "scaleway_dedibox_server.test"),
					resource.TestCheckResourceAttrSet("scaleway_dedibox_server.test", "hostname"),
					resource.TestCheckResourceAttrSet("scaleway_dedibox_server.test", "status"),
				),
			},
		},
	})
}

func isDediboxServerPresent(tt *acctest.TestTools, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]

		if !ok {
			return fmt.Errorf("not found: %s", n)
		}

		zone, id, err := zonal.ParseID(rs.Primary.ID)
		if err != nil {
			return err
		}

		serverID, err := strconv.ParseUint(id, 10, 64)
		if err != nil {
			return fmt.Errorf("invalid server ID: %s", id)
		}

		api := dedibox.NewAPI(tt.Meta.ScwClient())

		_, err = api.GetServer(&dedibox.GetServerRequest{
			Zone:     zone,
			ServerID: serverID,
		}, scw.WithContext(context.Background()))
		if err != nil {
			return err
		}

		return nil
	}
}

func isDediboxServerDestroyed(tt *acctest.TestTools) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "scaleway_dedibox_server" {
				continue
			}

			zone, id, err := zonal.ParseID(rs.Primary.ID)
			if err != nil {
				return err
			}

			serverID, err := strconv.ParseUint(id, 10, 64)
			if err != nil {
				return fmt.Errorf("invalid server ID: %s", id)
			}

			api := dedibox.NewAPI(tt.Meta.ScwClient())

			_, err = api.GetServer(&dedibox.GetServerRequest{
				Zone:     zone,
				ServerID: serverID,
			}, scw.WithContext(context.Background()))

			// If there's no error, the server still exists
			if err == nil {
				return fmt.Errorf("server %s still exists", rs.Primary.ID)
			}
		}

		return nil
	}
}
