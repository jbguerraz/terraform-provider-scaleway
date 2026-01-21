package dedibox

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/scaleway/scaleway-sdk-go/api/dedibox/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/httperrors"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/zonal"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/types"
)

func ResourceServer() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceServerCreate,
		ReadContext:   resourceServerRead,
		UpdateContext: resourceServerUpdate,
		DeleteContext: resourceServerDelete,
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
				Description: "ID of the server offer",
			},
			"project_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "Project ID the server belongs to",
			},
			"datacenter_name": {
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
				Description: "Name of the datacenter where the server should be provisioned",
			},
			"hostname": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "Hostname of the server",
			},
			"option_ids": {
				Type:        schema.TypeList,
				Optional:    true,
				ForceNew:    true,
				Description: "List of server option IDs to subscribe",
				Elem: &schema.Schema{
					Type: schema.TypeInt,
				},
			},
			"tags": {
				Type:        schema.TypeList,
				Optional:    true,
				Description: "Tags associated with the server",
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"zone": zonal.Schema(),

			// Computed attributes
			"organization_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Organization ID the server belongs to",
			},
			"offer_name": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Name of the server offer",
			},
			"status": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Status of the server",
			},
			"has_bmc": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Whether the server has BMC access",
			},
			"abuse_contact": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Abuse contact email",
			},
			"created_at": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Date of creation of the server",
			},
			"updated_at": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Date of last modification of the server",
			},
			"expired_at": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Expiration date of the server",
			},
			"location": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "Location of the server",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"rack": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Rack location",
						},
						"room": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Room location",
						},
						"datacenter": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Datacenter location",
						},
					},
				},
			},
			"os": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "OS installed on the server",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:        schema.TypeInt,
							Computed:    true,
							Description: "OS ID",
						},
						"name": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "OS name",
						},
						"version": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "OS version",
						},
					},
				},
			},
			"interfaces": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "Network interfaces of the server",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"type": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Interface type",
						},
						"mac": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "MAC address",
						},
						"ips": {
							Type:        schema.TypeList,
							Computed:    true,
							Description: "IP addresses",
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"ip_id": {
										Type:        schema.TypeInt,
										Computed:    true,
										Description: "IP ID",
									},
									"address": {
										Type:        schema.TypeString,
										Computed:    true,
										Description: "IP address",
									},
									"reverse": {
										Type:        schema.TypeString,
										Computed:    true,
										Description: "Reverse DNS",
									},
									"version": {
										Type:        schema.TypeString,
										Computed:    true,
										Description: "IP version",
									},
									"cidr": {
										Type:        schema.TypeInt,
										Computed:    true,
										Description: "CIDR notation",
									},
									"netmask": {
										Type:        schema.TypeString,
										Computed:    true,
										Description: "Netmask",
									},
									"gateway": {
										Type:        schema.TypeString,
										Computed:    true,
										Description: "Gateway",
									},
									"semantic": {
										Type:        schema.TypeString,
										Computed:    true,
										Description: "IP semantic (public, private, etc.)",
									},
									"status": {
										Type:        schema.TypeString,
										Computed:    true,
										Description: "IP status",
									},
								},
							},
						},
					},
				},
			},
			"options": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "Options subscribed on the server",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:        schema.TypeInt,
							Computed:    true,
							Description: "Option ID",
						},
						"name": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Option name",
						},
						"expires_at": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Option expiration date",
						},
					},
				},
			},
		},
	}
}

func resourceServerCreate(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	api, zone, err := newAPIWithZone(d, m)
	if err != nil {
		return diag.FromErr(err)
	}

	createReq := &dedibox.CreateServerRequest{
		Zone:      zone,
		OfferID:   uint64(d.Get("offer_id").(int)),
		ProjectID: d.Get("project_id").(string),
	}

	if datacenter, ok := d.GetOk("datacenter_name"); ok {
		dc := datacenter.(string)
		createReq.DatacenterName = &dc
	}

	if optionIDs, ok := d.GetOk("option_ids"); ok {
		ids := optionIDs.([]interface{})
		createReq.ServerOptionIDs = make([]uint64, len(ids))
		for i, id := range ids {
			createReq.ServerOptionIDs[i] = uint64(id.(int))
		}
	}

	// Create the server order
	service, err := api.CreateServer(createReq, scw.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}

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
		return diag.FromErr(fmt.Errorf("server was not provisioned, service status: %s", service.ProvisioningStatus))
	}

	serverID := *service.ResourceID
	d.SetId(zonal.NewIDString(zone, fmt.Sprintf("%d", serverID)))

	// Update hostname and tags if specified
	if hostname, ok := d.GetOk("hostname"); ok {
		_, err = api.UpdateServer(&dedibox.UpdateServerRequest{
			Zone:     zone,
			ServerID: serverID,
			Hostname: scw.StringPtr(hostname.(string)),
		}, scw.WithContext(ctx))
		if err != nil {
			return diag.FromErr(err)
		}
	}

	if tags, ok := d.GetOk("tags"); ok {
		tagList := make([]string, len(tags.([]interface{})))
		for i, tag := range tags.([]interface{}) {
			tagList[i] = tag.(string)
		}
		_, err = api.UpdateServerTags(&dedibox.UpdateServerTagsRequest{
			Zone:     zone,
			ServerID: serverID,
			Tags:     tagList,
		}, scw.WithContext(ctx))
		if err != nil {
			return diag.FromErr(err)
		}
	}

	return resourceServerRead(ctx, d, m)
}

func resourceServerRead(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	api, zonedID, err := NewAPIWithZoneAndID(m, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	serverID, err := strconv.ParseUint(zonedID.ID, 10, 64)
	if err != nil {
		return diag.FromErr(fmt.Errorf("invalid server ID: %s", zonedID.ID))
	}

	server, err := api.GetServer(&dedibox.GetServerRequest{
		Zone:     zonedID.Zone,
		ServerID: serverID,
	}, scw.WithContext(ctx))
	if err != nil {
		if httperrors.Is404(err) {
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}

	_ = d.Set("zone", server.Zone)
	_ = d.Set("hostname", server.Hostname)
	_ = d.Set("organization_id", server.OrganizationID)
	_ = d.Set("project_id", server.ProjectID)
	_ = d.Set("status", server.Status.String())
	_ = d.Set("has_bmc", server.HasBmc)
	_ = d.Set("abuse_contact", server.AbuseContact)
	_ = d.Set("created_at", types.FlattenTime(server.CreatedAt))
	_ = d.Set("updated_at", types.FlattenTime(server.UpdatedAt))
	_ = d.Set("expired_at", types.FlattenTime(server.ExpiredAt))
	_ = d.Set("tags", server.Tags)

	if server.Offer != nil {
		_ = d.Set("offer_id", server.Offer.ID)
		_ = d.Set("offer_name", server.Offer.Name)
	}

	if server.Location != nil {
		_ = d.Set("location", []map[string]any{flattenServerLocation(server.Location)})
	}

	if server.Os != nil {
		_ = d.Set("os", []map[string]any{flattenOS(server.Os)})
	}

	_ = d.Set("interfaces", flattenNetworkInterfaces(server.Interfaces))
	_ = d.Set("options", flattenServerOptions(server.Options))

	return nil
}

func resourceServerUpdate(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	api, zonedID, err := NewAPIWithZoneAndID(m, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	serverID, err := strconv.ParseUint(zonedID.ID, 10, 64)
	if err != nil {
		return diag.FromErr(fmt.Errorf("invalid server ID: %s", zonedID.ID))
	}

	if d.HasChange("hostname") {
		_, err = api.UpdateServer(&dedibox.UpdateServerRequest{
			Zone:     zonedID.Zone,
			ServerID: serverID,
			Hostname: scw.StringPtr(d.Get("hostname").(string)),
		}, scw.WithContext(ctx))
		if err != nil {
			return diag.FromErr(err)
		}
	}

	if d.HasChange("tags") {
		tags := d.Get("tags").([]interface{})
		tagList := make([]string, len(tags))
		for i, tag := range tags {
			tagList[i] = tag.(string)
		}
		_, err = api.UpdateServerTags(&dedibox.UpdateServerTagsRequest{
			Zone:     zonedID.Zone,
			ServerID: serverID,
			Tags:     tagList,
		}, scw.WithContext(ctx))
		if err != nil {
			return diag.FromErr(err)
		}
	}

	return resourceServerRead(ctx, d, m)
}

func resourceServerDelete(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	api, zonedID, err := NewAPIWithZoneAndID(m, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	serverID, err := strconv.ParseUint(zonedID.ID, 10, 64)
	if err != nil {
		return diag.FromErr(fmt.Errorf("invalid server ID: %s", zonedID.ID))
	}

	err = api.DeleteServer(&dedibox.DeleteServerRequest{
		Zone:     zonedID.Zone,
		ServerID: serverID,
	}, scw.WithContext(ctx))
	if err != nil && !httperrors.Is404(err) {
		return diag.FromErr(err)
	}

	// Wait for the server to be deleted
	err = waitForServerDeletion(ctx, api, zonedID.Zone, serverID, d.Timeout(schema.TimeoutDelete))
	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func waitForServerDeletion(ctx context.Context, api *dedibox.API, zone scw.Zone, serverID uint64, timeout time.Duration) error {
	retryInterval := 15 * time.Second
	timeoutTimer := time.NewTimer(timeout)
	ticker := time.NewTicker(retryInterval)
	defer timeoutTimer.Stop()
	defer ticker.Stop()

	for {
		select {
		case <-timeoutTimer.C:
			return fmt.Errorf("timeout waiting for server deletion")
		case <-ticker.C:
			_, err := api.GetServer(&dedibox.GetServerRequest{
				Zone:     zone,
				ServerID: serverID,
			}, scw.WithContext(ctx))
			if err != nil {
				if httperrors.Is404(err) {
					return nil
				}
				return err
			}
			// Server still exists, continue waiting
		}
	}
}
