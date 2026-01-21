package dedibox

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/scaleway/scaleway-sdk-go/api/dedibox/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/datasource"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/zonal"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/verify"
)

func DataSourceOffer() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceOfferRead,
		Schema: map[string]*schema.Schema{
			"name": {
				Type:          schema.TypeString,
				Optional:      true,
				Description:   "Exact name of the desired offer",
				ConflictsWith: []string{"offer_id"},
			},
			"offer_id": {
				Type:          schema.TypeInt,
				Optional:      true,
				Description:   "ID of the desired offer",
				ConflictsWith: []string{"name"},
			},
			"commercial_range": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Filter on commercial range",
			},
			"catalog": {
				Type:             schema.TypeString,
				Optional:         true,
				Default:          "all",
				ValidateDiagFunc: verify.ValidateEnum[dedibox.OfferCatalog](),
				Description:      "Filter on catalog (all, default, beta, reseller, premium, volume, admin, inactive)",
			},
			"available_only": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     true,
				Description: "Only return available offers",
			},
			"zone": zonal.Schema(),

			// Computed attributes
			"payment_frequency": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Payment frequency of the offer (monthly or hourly)",
			},
			"pricing_cents": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "Price of the offer in cents",
			},
			"pricing_currency": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Currency of the price",
			},
			"bandwidth": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "Available bandwidth with the offer in bits/s",
			},
			"stock": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Stock status for this offer",
			},
			"cpu": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "CPU specifications of the offer",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "CPU name",
						},
						"core_count": {
							Type:        schema.TypeInt,
							Computed:    true,
							Description: "Number of cores",
						},
						"frequency": {
							Type:        schema.TypeInt,
							Computed:    true,
							Description: "Frequency of the CPU in MHz",
						},
						"thread_count": {
							Type:        schema.TypeInt,
							Computed:    true,
							Description: "Number of threads",
						},
					},
				},
			},
			"disk": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "Disk specifications of the offer",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"type": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Type of disk",
						},
						"capacity": {
							Type:        schema.TypeInt,
							Computed:    true,
							Description: "Capacity of the disk in bytes",
						},
					},
				},
			},
			"memory": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "Memory specifications of the offer",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"type": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Type of memory",
						},
						"capacity": {
							Type:        schema.TypeInt,
							Computed:    true,
							Description: "Capacity of the memory in bytes",
						},
						"frequency": {
							Type:        schema.TypeInt,
							Computed:    true,
							Description: "Frequency of the memory in MHz",
						},
						"is_ecc": {
							Type:        schema.TypeBool,
							Computed:    true,
							Description: "True if error-correcting code is available",
						},
					},
				},
			},
		},
	}
}

func dataSourceOfferRead(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	api, zone, err := newAPIWithZone(d, m)
	if err != nil {
		return diag.FromErr(err)
	}

	offerID, offerIDSet := d.GetOk("offer_id")
	name, nameSet := d.GetOk("name")

	if !offerIDSet && !nameSet {
		return diag.FromErr(fmt.Errorf("either offer_id or name must be set"))
	}

	var offer *dedibox.Offer

	if offerIDSet {
		// Get offer by ID - need to list and filter
		offer, err = findOfferByID(ctx, api, zone, uint64(offerID.(int)))
		if err != nil {
			return diag.FromErr(err)
		}
	} else {
		// Search by name
		listRequest := &dedibox.ListOffersRequest{
			Zone: zone,
		}

		if commercialRange, ok := d.GetOk("commercial_range"); ok {
			cr := commercialRange.(string)
			listRequest.CommercialRange = &cr
		}

		if catalog, ok := d.GetOk("catalog"); ok {
			listRequest.Catalog = dedibox.OfferCatalog(catalog.(string))
		}

		if availableOnly, ok := d.GetOk("available_only"); ok && availableOnly.(bool) {
			available := true
			listRequest.AvailableOnly = &available
		}

		res, err := api.ListOffers(listRequest, scw.WithAllPages(), scw.WithContext(ctx))
		if err != nil {
			return diag.FromErr(err)
		}

		var matches []*dedibox.Offer
		nameToUpper := strings.ToUpper(name.(string))

		for _, o := range res.Offers {
			if strings.ToUpper(o.Name) == nameToUpper {
				matches = append(matches, o)
			}
		}

		if len(matches) == 0 {
			return diag.FromErr(fmt.Errorf("no offer found with the name %s in zone %s", name, zone))
		}

		if len(matches) > 1 {
			return diag.FromErr(fmt.Errorf("multiple offers found with the name %s in zone %s, please use offer_id to be more specific", name, zone))
		}

		offer = matches[0]
	}

	// Set the resource data
	zonedID := datasource.NewZonedID(fmt.Sprintf("%d", offer.ID), zone)
	d.SetId(zonedID)

	_ = d.Set("offer_id", offer.ID)
	_ = d.Set("zone", zone)
	_ = d.Set("name", offer.Name)
	_ = d.Set("catalog", offer.Catalog.String())
	_ = d.Set("payment_frequency", offer.PaymentFrequency.String())

	if offer.Pricing != nil {
		_ = d.Set("pricing_cents", offer.Pricing.Units*100+int64(offer.Pricing.Nanos/10000000))
		_ = d.Set("pricing_currency", offer.Pricing.CurrencyCode)
	}

	// Set server-specific info if this is a server offer
	if offer.ServerInfo != nil {
		_ = d.Set("bandwidth", offer.ServerInfo.Bandwidth)
		_ = d.Set("stock", offer.ServerInfo.Stock.String())
		_ = d.Set("cpu", flattenCPUs(offer.ServerInfo.CPUs))
		_ = d.Set("disk", flattenDisks(offer.ServerInfo.Disks))
		_ = d.Set("memory", flattenMemories(offer.ServerInfo.Memories))
	}

	return nil
}

func findOfferByID(ctx context.Context, api *dedibox.API, zone scw.Zone, offerID uint64) (*dedibox.Offer, error) {
	res, err := api.ListOffers(&dedibox.ListOffersRequest{
		Zone: zone,
	}, scw.WithAllPages(), scw.WithContext(ctx))
	if err != nil {
		return nil, err
	}

	for _, offer := range res.Offers {
		if offer.ID == offerID {
			return offer, nil
		}
	}

	return nil, fmt.Errorf("offer %d not found in zone %s", offerID, zone)
}
