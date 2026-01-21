package dedibox

import (
	"context"
	"fmt"
	"strconv"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/scaleway/scaleway-sdk-go/api/dedibox/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/httperrors"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/zonal"
)

func ResourceReverseDNS() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceReverseDNSCreate,
		ReadContext:   resourceReverseDNSRead,
		UpdateContext: resourceReverseDNSUpdate,
		DeleteContext: resourceReverseDNSDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: &schema.ResourceTimeout{
			Create:  schema.DefaultTimeout(defaultServerTimeout),
			Read:    schema.DefaultTimeout(defaultServerTimeout),
			Update:  schema.DefaultTimeout(defaultServerTimeout),
			Delete:  schema.DefaultTimeout(defaultServerTimeout),
			Default: schema.DefaultTimeout(defaultServerTimeout),
		},
		Schema: map[string]*schema.Schema{
			"ip_id": {
				Type:        schema.TypeInt,
				Required:    true,
				ForceNew:    true,
				Description: "ID of the IP to set reverse DNS for",
			},
			"reverse": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Reverse DNS value (PTR record)",
			},
			"zone": zonal.Schema(),

			// Computed attributes
			"address": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "IP address",
			},
			"ip_version": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "IP version (IPv4 or IPv6)",
			},
			"cidr": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "CIDR notation",
			},
			"netmask": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Network mask",
			},
			"gateway": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Gateway IP",
			},
			"status": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "IP status",
			},
		},
	}
}

func resourceReverseDNSCreate(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	api, zone, err := newAPIWithZone(d, m)
	if err != nil {
		return diag.FromErr(err)
	}

	ipID := uint64(d.Get("ip_id").(int))
	reverse := d.Get("reverse").(string)

	ip, err := api.UpdateReverse(&dedibox.UpdateReverseRequest{
		Zone:    zone,
		IPID:    ipID,
		Reverse: reverse,
	}, scw.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(zonal.NewIDString(zone, fmt.Sprintf("%d", ip.IPID)))

	return resourceReverseDNSRead(ctx, d, m)
}

func resourceReverseDNSRead(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	api, zonedID, err := NewAPIWithZoneAndID(m, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	ipID, err := strconv.ParseUint(zonedID.ID, 10, 64)
	if err != nil {
		return diag.FromErr(fmt.Errorf("invalid IP ID: %s", zonedID.ID))
	}

	// Get the IP info via failover IP API (works for both server and failover IPs)
	ip, err := api.GetFailoverIP(&dedibox.GetFailoverIPRequest{
		Zone: zonedID.Zone,
		IPID: ipID,
	}, scw.WithContext(ctx))
	if err != nil {
		if httperrors.Is404(err) {
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}

	_ = d.Set("zone", zonedID.Zone)
	_ = d.Set("ip_id", ip.ID)
	_ = d.Set("reverse", ip.Reverse)
	_ = d.Set("address", ip.Address.String())
	_ = d.Set("ip_version", ip.IPVersion.String())
	_ = d.Set("cidr", ip.Cidr)
	_ = d.Set("netmask", ip.Netmask.String())
	_ = d.Set("gateway_ip", ip.GatewayIP.String())
	_ = d.Set("status", ip.Status.String())

	return nil
}

func resourceReverseDNSUpdate(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	api, zonedID, err := NewAPIWithZoneAndID(m, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	ipID, err := strconv.ParseUint(zonedID.ID, 10, 64)
	if err != nil {
		return diag.FromErr(fmt.Errorf("invalid IP ID: %s", zonedID.ID))
	}

	if d.HasChange("reverse") {
		_, err = api.UpdateReverse(&dedibox.UpdateReverseRequest{
			Zone:    zonedID.Zone,
			IPID:    ipID,
			Reverse: d.Get("reverse").(string),
		}, scw.WithContext(ctx))
		if err != nil {
			return diag.FromErr(err)
		}
	}

	return resourceReverseDNSRead(ctx, d, m)
}

func resourceReverseDNSDelete(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	api, zonedID, err := NewAPIWithZoneAndID(m, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	ipID, err := strconv.ParseUint(zonedID.ID, 10, 64)
	if err != nil {
		return diag.FromErr(fmt.Errorf("invalid IP ID: %s", zonedID.ID))
	}

	// Reset reverse DNS to empty
	_, err = api.UpdateReverse(&dedibox.UpdateReverseRequest{
		Zone:    zonedID.Zone,
		IPID:    ipID,
		Reverse: "",
	}, scw.WithContext(ctx))
	if err != nil && !httperrors.Is404(err) {
		return diag.FromErr(err)
	}

	return nil
}
