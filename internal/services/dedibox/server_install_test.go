package dedibox_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/acctest"
)

// TestAccDediboxServerInstall_Basic tests server installation
// WARNING: This test requires an existing server and will reinstall its OS.
// Run with: TF_ACC=1 DEDIBOX_SERVER_ID=12345 go test -v -run TestAccDediboxServerInstall_Basic
func TestAccDediboxServerInstall_Basic(t *testing.T) {
	serverID := getenv("DEDIBOX_SERVER_ID", "")
	if serverID == "" {
		t.Skip("DEDIBOX_SERVER_ID not set, skipping server install test")
	}

	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
					data "scaleway_dedibox_os" "ubuntu" {
						zone = "%s"
						name = "Ubuntu"
						type = "server"
					}

					resource "scaleway_dedibox_server_install" "test" {
						zone      = "%s"
						server_id = %s
						os_id     = data.scaleway_dedibox_os.ubuntu.os_id
						hostname  = "test-server.example.com"
					}
				`, DediboxZone, DediboxZone, serverID),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("scaleway_dedibox_server_install.test", "hostname", "test-server.example.com"),
					resource.TestCheckResourceAttrSet("scaleway_dedibox_server_install.test", "status"),
				),
			},
		},
	})
}
