package dedibox

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/scaleway/scaleway-sdk-go/api/dedibox/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/datasource"
)

func DataSourceFailoverIP() *schema.Resource {
	// Generate datasource schema from resource schema
	dsSchema := datasource.SchemaFromResourceSchema(ResourceFailoverIP().Schema)

	// Add failover_ip_id as the primary key
	dsSchema["failover_ip_id"] = &schema.Schema{
		Type:          schema.TypeInt,
		Optional:      true,
		Description:   "The ID of the failover IP",
		ConflictsWith: []string{"address"},
	}

	// Add address as an alternative lookup key
	dsSchema["address"] = &schema.Schema{
		Type:          schema.TypeString,
		Optional:      true,
		Computed:      true,
		Description:   "The IP address of the failover IP",
		ConflictsWith: []string{"failover_ip_id"},
	}

	// Make non-computed fields optional in data source
	dsSchema["offer_id"].Required = false
	dsSchema["offer_id"].Optional = true
	dsSchema["offer_id"].Computed = true

	dsSchema["project_id"].Required = false
	dsSchema["project_id"].Optional = true
	dsSchema["project_id"].Computed = true

	return &schema.Resource{
		ReadContext: dataSourceFailoverIPRead,
		Schema:      dsSchema,
	}
}

func dataSourceFailoverIPRead(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	api, zone, err := newAPIWithZone(d, m)
	if err != nil {
		return diag.FromErr(err)
	}

	ipID, ipIDSet := d.GetOk("failover_ip_id")
	address, addressSet := d.GetOk("address")

	if !ipIDSet && !addressSet {
		return diag.FromErr(fmt.Errorf("either failover_ip_id or address must be set"))
	}

	var failoverIP *dedibox.FailoverIP

	if ipIDSet {
		failoverIP, err = api.GetFailoverIP(&dedibox.GetFailoverIPRequest{
			Zone: zone,
			IPID: uint64(ipID.(int)),
		}, scw.WithContext(ctx))
		if err != nil {
			return diag.FromErr(err)
		}
	} else {
		// Search by address
		res, err := api.ListFailoverIPs(&dedibox.ListFailoverIPsRequest{
			Zone: zone,
		}, scw.WithAllPages(), scw.WithContext(ctx))
		if err != nil {
			return diag.FromErr(err)
		}

		addressStr := address.(string)
		for _, ip := range res.FailoverIPs {
			if ip.Address.String() == addressStr {
				failoverIP = ip
				break
			}
		}

		if failoverIP == nil {
			return diag.FromErr(fmt.Errorf("failover IP with address %s not found in zone %s", addressStr, zone))
		}
	}

	d.SetId(datasource.NewZonedID(fmt.Sprintf("%d", failoverIP.ID), zone))

	_ = d.Set("zone", zone)
	_ = d.Set("failover_ip_id", failoverIP.ID)
	_ = d.Set("address", failoverIP.Address.String())
	_ = d.Set("reverse", failoverIP.Reverse)
	_ = d.Set("ip_version", failoverIP.IPVersion.String())
	_ = d.Set("cidr", failoverIP.Cidr)
	_ = d.Set("netmask", failoverIP.Netmask.String())
	_ = d.Set("gateway_ip", failoverIP.GatewayIP.String())
	_ = d.Set("status", failoverIP.Status.String())
	_ = d.Set("type", failoverIP.Type.String())

	if failoverIP.Mac != nil {
		_ = d.Set("mac", *failoverIP.Mac)
	}

	if failoverIP.ServerID != nil {
		_ = d.Set("server_id", int(*failoverIP.ServerID))
	}

	if failoverIP.Block != nil {
		_ = d.Set("block", []map[string]any{flattenFailoverBlock(failoverIP.Block)})
	}

	return nil
}

// findFailoverIPByAddress helper function to find a failover IP by address
func findFailoverIPByAddress(ctx context.Context, api *dedibox.API, zone scw.Zone, address string) (*dedibox.FailoverIP, error) {
	res, err := api.ListFailoverIPs(&dedibox.ListFailoverIPsRequest{
		Zone: zone,
	}, scw.WithAllPages(), scw.WithContext(ctx))
	if err != nil {
		return nil, err
	}

	for _, ip := range res.FailoverIPs {
		if ip.Address.String() == address {
			return ip, nil
		}
	}

	return nil, fmt.Errorf("failover IP %s not found in zone %s", address, zone)
}
