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

var (
	DediboxOfferName = getenv("DEDIBOX_OFFER_NAME", "Start-1-S-SATA")
	DediboxZone      = getenv("DEDIBOX_ZONE", "fr-par-1")
)

func TestAccDataSourceDediboxOffer_Basic(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
					data "scaleway_dedibox_offer" "test1" {
						zone = "%s"
						name = "%s"
					}
				`, DediboxZone, DediboxOfferName),
				Check: resource.ComposeTestCheckFunc(
					isDediboxOfferPresent(tt, "data.scaleway_dedibox_offer.test1"),
					resource.TestCheckResourceAttr("data.scaleway_dedibox_offer.test1", "name", DediboxOfferName),
					resource.TestCheckResourceAttrSet("data.scaleway_dedibox_offer.test1", "offer_id"),
					resource.TestCheckResourceAttrSet("data.scaleway_dedibox_offer.test1", "commercial_range"),
				),
			},
		},
	})
}

func TestAccDataSourceDediboxOffer_ByID(t *testing.T) {
	tt := acctest.NewTestTools(t)
	defer tt.Cleanup()

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: tt.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
					data "scaleway_dedibox_offer" "test1" {
						zone = "%s"
						name = "%s"
					}

					data "scaleway_dedibox_offer" "test2" {
						zone     = "%s"
						offer_id = data.scaleway_dedibox_offer.test1.offer_id
					}
				`, DediboxZone, DediboxOfferName, DediboxZone),
				Check: resource.ComposeTestCheckFunc(
					isDediboxOfferPresent(tt, "data.scaleway_dedibox_offer.test1"),
					isDediboxOfferPresent(tt, "data.scaleway_dedibox_offer.test2"),
					resource.TestCheckResourceAttrPair("data.scaleway_dedibox_offer.test2", "offer_id", "data.scaleway_dedibox_offer.test1", "offer_id"),
					resource.TestCheckResourceAttr("data.scaleway_dedibox_offer.test2", "name", DediboxOfferName),
				),
			},
		},
	})
}

func isDediboxOfferPresent(tt *acctest.TestTools, n string) resource.TestCheckFunc {
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

		offers, err := api.ListOffers(&dedibox.ListOffersRequest{
			Zone: zone,
		}, scw.WithAllPages(), scw.WithContext(context.Background()))
		if err != nil {
			return err
		}

		for _, offer := range offers.Offers {
			if fmt.Sprintf("%d", offer.ID) == id {
				return nil
			}
		}

		return fmt.Errorf("offer %s not found", id)
	}
}
