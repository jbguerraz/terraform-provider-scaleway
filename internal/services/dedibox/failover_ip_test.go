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

// TestAccDediboxFailoverIP_Basic tests failover IP creation
// WARNING: This test provisions real failover IPs which incur costs.
// Run with: TF_ACC=1 go test -v -run TestAccDediboxFailoverIP_Basic
func TestAccDediboxFailoverIP_Basic(t *testing.T) {
	t.Skip("Skipping failover IP test - provisions real resources and incurs costs")

	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		CheckDestroy:             isDediboxFailoverIPDestroyed(tt),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
					resource "scaleway_dedibox_failover_ip" "test" {
						zone       = "%s"
						offer_id   = 1
						project_id = "`+tt.Meta.ProjectID+`"
					}
				`, DediboxZone),
				Check: resource.ComposeTestCheckFunc(
					isDediboxFailoverIPPresent(tt, "scaleway_dedibox_failover_ip.test"),
					resource.TestCheckResourceAttrSet("scaleway_dedibox_failover_ip.test", "address"),
					resource.TestCheckResourceAttrSet("scaleway_dedibox_failover_ip.test", "status"),
				),
			},
		},
	})
}

func isDediboxFailoverIPPresent(tt *acctest.TestTools, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]

		if !ok {
			return fmt.Errorf("not found: %s", n)
		}

		zone, id, err := zonal.ParseID(rs.Primary.ID)
		if err != nil {
			return err
		}

		ipID, err := strconv.ParseUint(id, 10, 64)
		if err != nil {
			return fmt.Errorf("invalid IP ID: %s", id)
		}

		api := dedibox.NewAPI(tt.Meta.ScwClient())

		_, err = api.GetFailoverIP(&dedibox.GetFailoverIPRequest{
			Zone: zone,
			IPID: ipID,
		}, scw.WithContext(context.Background()))
		if err != nil {
			return err
		}

		return nil
	}
}

func isDediboxFailoverIPDestroyed(tt *acctest.TestTools) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "scaleway_dedibox_failover_ip" {
				continue
			}

			zone, id, err := zonal.ParseID(rs.Primary.ID)
			if err != nil {
				return err
			}

			ipID, err := strconv.ParseUint(id, 10, 64)
			if err != nil {
				return fmt.Errorf("invalid IP ID: %s", id)
			}

			api := dedibox.NewAPI(tt.Meta.ScwClient())

			_, err = api.GetFailoverIP(&dedibox.GetFailoverIPRequest{
				Zone: zone,
				IPID: ipID,
			}, scw.WithContext(context.Background()))

			// If there's no error, the IP still exists
			if err == nil {
				return fmt.Errorf("failover IP %s still exists", rs.Primary.ID)
			}
		}

		return nil
	}
}
