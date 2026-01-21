package dedibox_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/acctest"
)

// TestAccDataSourceDediboxFailoverIP_Basic tests the failover IP data source
// WARNING: This test requires an existing failover IP ID.
// Run with: TF_ACC=1 DEDIBOX_FAILOVER_IP_ID=12345 go test -v -run TestAccDataSourceDediboxFailoverIP_Basic
func TestAccDataSourceDediboxFailoverIP_Basic(t *testing.T) {
	ipID := getenv("DEDIBOX_FAILOVER_IP_ID", "")
	if ipID == "" {
		t.Skip("DEDIBOX_FAILOVER_IP_ID not set, skipping failover IP data source test")
	}

	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
					data "scaleway_dedibox_failover_ip" "test" {
						zone  = "%s"
						ip_id = %s
					}
				`, DediboxZone, ipID),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.scaleway_dedibox_failover_ip.test", "address"),
					resource.TestCheckResourceAttrSet("data.scaleway_dedibox_failover_ip.test", "status"),
				),
			},
		},
	})
}
