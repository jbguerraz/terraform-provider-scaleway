package dedibox_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/acctest"
)

// TestAccDataSourceDediboxServer_Basic tests the server data source
// WARNING: This test requires an existing server ID.
// Run with: TF_ACC=1 DEDIBOX_SERVER_ID=12345 go test -v -run TestAccDataSourceDediboxServer_Basic
func TestAccDataSourceDediboxServer_Basic(t *testing.T) {
	serverID := getenv("DEDIBOX_SERVER_ID", "")
	if serverID == "" {
		t.Skip("DEDIBOX_SERVER_ID not set, skipping server data source test")
	}

	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
					data "scaleway_dedibox_server" "test" {
						zone      = "%s"
						server_id = %s
					}
				`, DediboxZone, serverID),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.scaleway_dedibox_server.test", "hostname"),
					resource.TestCheckResourceAttrSet("data.scaleway_dedibox_server.test", "status"),
					resource.TestCheckResourceAttrSet("data.scaleway_dedibox_server.test", "offer_name"),
				),
			},
		},
	})
}
