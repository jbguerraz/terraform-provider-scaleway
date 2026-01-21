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

func ResourceFailoverIP() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceFailoverIPCreate,
		ReadContext:   resourceFailoverIPRead,
		UpdateContext: resourceFailoverIPUpdate,
		DeleteContext: resourceFailoverIPDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: &schema.ResourceTimeout{
			Create:  schema.DefaultTimeout(defaultServiceTimeout),
			Read:    schema.DefaultTimeout(defaultServerTimeout),
			Update:  schema.DefaultTimeout(defaultServerTimeout),
			Delete:  schema.DefaultTimeout(defaultServerTimeout),
			Default: schema.DefaultTimeout(defaultServerTimeout),
		},
		Schema: map[string]*schema.Schema{
			"offer_id": {
				Type:        schema.TypeInt,
				Required:    true,
				ForceNew:    true,
				Description: "ID of the failover IP offer",
			},
			"project_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "Project ID the failover IP belongs to",
			},
			"server_id": {
				Type:        schema.TypeInt,
				Optional:    true,
				Description: "ID of the server to attach the failover IP to",
			},
			"zone": zonal.Schema(),

			// Computed attributes
			"address": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "IP address of the failover IP",
			},
			"reverse": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Reverse DNS of the failover IP",
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
				Description: "Netmask of the failover IP",
			},
			"gateway_ip": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Gateway IP",
			},
			"mac": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "MAC address of the failover IP",
			},
			"status": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Status of the failover IP",
			},
			"type": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Interface type of the failover IP",
			},
			"block": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "Block information if part of a failover block",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:        schema.TypeInt,
							Computed:    true,
							Description: "Block ID",
						},
						"address": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Block IP address",
						},
						"cidr": {
							Type:        schema.TypeInt,
							Computed:    true,
							Description: "Block CIDR",
						},
						"netmask": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Block netmask",
						},
						"gateway_ip": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Block gateway IP",
						},
						"ip_version": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Block IP version",
						},
						"nameservers": {
							Type:        schema.TypeList,
							Computed:    true,
							Description: "Block nameservers",
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
					},
				},
			},
		},
	}
}

func resourceFailoverIPCreate(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	api, zone, err := newAPIWithZone(d, m)
	if err != nil {
		return diag.FromErr(err)
	}

	createReq := &dedibox.CreateFailoverIPsRequest{
		Zone:      zone,
		OfferID:   uint64(d.Get("offer_id").(int)),
		ProjectID: d.Get("project_id").(string),
		Quantity:  1, // Create one IP per resource
	}

	resp, err := api.CreateFailoverIPs(createReq, scw.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}

	if len(resp.Services) == 0 {
		return diag.FromErr(fmt.Errorf("no failover IP service was created"))
	}

	service := resp.Services[0]

	// Wait for the service to be delivered
	service, err = api.WaitForService(&dedibox.WaitForServiceRequest{
		ServiceID:     service.ID,
		Zone:          zone,
		Timeout:       scw.TimeDurationPtr(d.Timeout(schema.TimeoutCreate)),
		RetryInterval: scw.TimeDurationPtr(retryInterval),
	}, scw.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}

	if service.ResourceID == nil {
		return diag.FromErr(fmt.Errorf("failover IP was not provisioned, service status: %s", service.ProvisioningStatus))
	}

	ipID := *service.ResourceID
	d.SetId(zonal.NewIDString(zone, fmt.Sprintf("%d", ipID)))

	// Attach to server if specified
	if serverID, ok := d.GetOk("server_id"); ok {
		err = api.AttachFailoverIPs(&dedibox.AttachFailoverIPsRequest{
			Zone:     zone,
			ServerID: uint64(serverID.(int)),
			FipsIDs:  []uint64{ipID},
		}, scw.WithContext(ctx))
		if err != nil {
			return diag.FromErr(err)
		}
	}

	return resourceFailoverIPRead(ctx, d, m)
}

func resourceFailoverIPRead(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	api, zonedID, err := NewAPIWithZoneAndID(m, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	ipID, err := strconv.ParseUint(zonedID.ID, 10, 64)
	if err != nil {
		return diag.FromErr(fmt.Errorf("invalid failover IP ID: %s", zonedID.ID))
	}

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
	_ = d.Set("address", ip.Address.String())
	_ = d.Set("reverse", ip.Reverse)
	_ = d.Set("ip_version", ip.IPVersion.String())
	_ = d.Set("cidr", ip.Cidr)
	_ = d.Set("netmask", ip.Netmask.String())
	_ = d.Set("gateway_ip", ip.GatewayIP.String())
	_ = d.Set("status", ip.Status.String())
	_ = d.Set("type", ip.Type.String())

	if ip.Mac != nil {
		_ = d.Set("mac", *ip.Mac)
	}

	if ip.ServerID != nil {
		_ = d.Set("server_id", int(*ip.ServerID))
	} else {
		_ = d.Set("server_id", nil)
	}

	if ip.Block != nil {
		_ = d.Set("block", []map[string]any{flattenFailoverBlock(ip.Block)})
	}

	return nil
}

func resourceFailoverIPUpdate(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	api, zonedID, err := NewAPIWithZoneAndID(m, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	ipID, err := strconv.ParseUint(zonedID.ID, 10, 64)
	if err != nil {
		return diag.FromErr(fmt.Errorf("invalid failover IP ID: %s", zonedID.ID))
	}

	if d.HasChange("server_id") {
		old, new := d.GetChange("server_id")

		// Detach from old server if was attached
		if old != nil && old.(int) != 0 {
			err = api.DetachFailoverIPs(&dedibox.DetachFailoverIPsRequest{
				Zone:    zonedID.Zone,
				FipsIDs: []uint64{ipID},
			}, scw.WithContext(ctx))
			if err != nil {
				return diag.FromErr(err)
			}
		}

		// Attach to new server if specified
		if new != nil && new.(int) != 0 {
			err = api.AttachFailoverIPs(&dedibox.AttachFailoverIPsRequest{
				Zone:     zonedID.Zone,
				ServerID: uint64(new.(int)),
				FipsIDs:  []uint64{ipID},
			}, scw.WithContext(ctx))
			if err != nil {
				return diag.FromErr(err)
			}
		}
	}

	return resourceFailoverIPRead(ctx, d, m)
}

func resourceFailoverIPDelete(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	api, zonedID, err := NewAPIWithZoneAndID(m, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	ipID, err := strconv.ParseUint(zonedID.ID, 10, 64)
	if err != nil {
		return diag.FromErr(fmt.Errorf("invalid failover IP ID: %s", zonedID.ID))
	}

	// Detach from server if attached
	if serverID, ok := d.GetOk("server_id"); ok && serverID.(int) != 0 {
		err = api.DetachFailoverIPs(&dedibox.DetachFailoverIPsRequest{
			Zone:    zonedID.Zone,
			FipsIDs: []uint64{ipID},
		}, scw.WithContext(ctx))
		if err != nil && !httperrors.Is404(err) {
			return diag.FromErr(err)
		}
	}

	err = api.DeleteFailoverIP(&dedibox.DeleteFailoverIPRequest{
		Zone: zonedID.Zone,
		IPID: ipID,
	}, scw.WithContext(ctx))
	if err != nil && !httperrors.Is404(err) {
		return diag.FromErr(err)
	}

	return nil
}

func flattenFailoverBlock(block *dedibox.FailoverBlock) map[string]any {
	if block == nil {
		return nil
	}

	return map[string]any{
		"id":          block.ID,
		"address":     block.Address.String(),
		"cidr":        block.Cidr,
		"netmask":     block.Netmask.String(),
		"gateway_ip":  block.GatewayIP.String(),
		"ip_version":  block.IPVersion.String(),
		"nameservers": block.Nameservers,
	}
}
