package dedibox_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/acctest"
)

// TestAccDediboxReverseDNS_Basic tests reverse DNS creation
// WARNING: This test requires an existing IP ID.
// Run with: TF_ACC=1 DEDIBOX_IP_ID=12345 go test -v -run TestAccDediboxReverseDNS_Basic
func TestAccDediboxReverseDNS_Basic(t *testing.T) {
	ipID := getenv("DEDIBOX_IP_ID", "")
	if ipID == "" {
		t.Skip("DEDIBOX_IP_ID not set, skipping reverse DNS test")
	}

	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
					resource "scaleway_dedibox_reverse_dns" "test" {
						zone    = "%s"
						ip_id   = %s
						reverse = "test.example.com"
					}
				`, DediboxZone, ipID),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("scaleway_dedibox_reverse_dns.test", "reverse", "test.example.com"),
					resource.TestCheckResourceAttrSet("scaleway_dedibox_reverse_dns.test", "address"),
				),
			},
			{
				Config: fmt.Sprintf(`
					resource "scaleway_dedibox_reverse_dns" "test" {
						zone    = "%s"
						ip_id   = %s
						reverse = "updated.example.com"
					}
				`, DediboxZone, ipID),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("scaleway_dedibox_reverse_dns.test", "reverse", "updated.example.com"),
				),
			},
		},
	})
}
