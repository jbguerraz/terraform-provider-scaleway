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
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/types"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/verify"
)

func DataSourceOS() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceOSRead,
		Schema: map[string]*schema.Schema{
			"os_id": {
				Type:          schema.TypeInt,
				Optional:      true,
				Description:   "ID of the OS",
				ConflictsWith: []string{"name"},
			},
			"name": {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				Description:   "Name of the OS (partial match supported)",
				ConflictsWith: []string{"os_id"},
			},
			"server_id": {
				Type:        schema.TypeInt,
				Optional:    true,
				Description: "Filter OS by compatible server ID",
			},
			"type": {
				Type:             schema.TypeString,
				Optional:         true,
				Description:      "Type of OS (server, virtualization, panel, desktop, custom)",
				ValidateDiagFunc: verify.ValidateEnum[dedibox.OSType](),
			},
			"zone": zonal.Schema(),

			// Computed attributes
			"version": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Version of the OS",
			},
			"arch": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Architecture of the OS",
			},
			"display_name": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Display name of the OS",
			},
			"allow_custom_partitioning": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Whether the OS allows custom partitioning",
			},
			"allow_ssh_keys": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Whether the OS allows SSH keys",
			},
			"requires_user": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Whether the OS requires a user",
			},
			"requires_admin_password": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Whether the OS requires an admin password",
			},
			"requires_panel_password": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Whether the OS requires a panel password",
			},
			"requires_license": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Whether the OS requires a license",
			},
			"max_partitions": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "Maximum number of partitions",
			},
			"released_at": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "OS release date",
			},
		},
	}
}

func dataSourceOSRead(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	api, zone, err := newAPIWithZone(d, m)
	if err != nil {
		return diag.FromErr(err)
	}

	osID, osIDSet := d.GetOk("os_id")
	name, nameSet := d.GetOk("name")

	if !osIDSet && !nameSet {
		return diag.FromErr(fmt.Errorf("either os_id or name must be set"))
	}

	var os *dedibox.OS

	if osIDSet {
		os, err = api.GetOS(&dedibox.GetOSRequest{
			Zone: zone,
			OsID: uint64(osID.(int)),
		}, scw.WithContext(ctx))
		if err != nil {
			return diag.FromErr(err)
		}
	} else {
		// Search by name
		listReq := &dedibox.ListOSRequest{
			Zone: zone,
		}

		if serverID, ok := d.GetOk("server_id"); ok {
			listReq.ServerID = uint64(serverID.(int))
		}

		if osType, ok := d.GetOk("type"); ok {
			listReq.Type = dedibox.OSType(osType.(string))
		}

		res, err := api.ListOS(listReq, scw.WithAllPages(), scw.WithContext(ctx))
		if err != nil {
			return diag.FromErr(err)
		}

		nameStr := strings.ToLower(name.(string))
		var matches []*dedibox.OS

		for _, o := range res.Os {
			if strings.Contains(strings.ToLower(o.Name), nameStr) ||
				strings.Contains(strings.ToLower(o.DisplayName), nameStr) {
				matches = append(matches, o)
			}
		}

		if len(matches) == 0 {
			return diag.FromErr(fmt.Errorf("no OS found with name containing %q in zone %s", name, zone))
		}

		if len(matches) > 1 {
			// Try exact match
			for _, o := range matches {
				if strings.EqualFold(o.Name, nameStr) || strings.EqualFold(o.DisplayName, nameStr) {
					os = o
					break
				}
			}
			if os == nil {
				return diag.FromErr(fmt.Errorf("multiple OS found with name containing %q in zone %s, please use os_id to be more specific", name, zone))
			}
		} else {
			os = matches[0]
		}
	}

	d.SetId(datasource.NewZonedID(fmt.Sprintf("%d", os.ID), zone))

	_ = d.Set("os_id", os.ID)
	_ = d.Set("zone", zone)
	_ = d.Set("name", os.Name)
	_ = d.Set("version", os.Version)
	_ = d.Set("type", os.Type.String())
	_ = d.Set("arch", os.Arch.String())
	_ = d.Set("display_name", os.DisplayName)
	_ = d.Set("allow_custom_partitioning", os.AllowCustomPartitioning)
	_ = d.Set("allow_ssh_keys", os.AllowSSHKeys)
	_ = d.Set("requires_user", os.RequiresUser)
	_ = d.Set("requires_admin_password", os.RequiresAdminPassword)
	_ = d.Set("requires_panel_password", os.RequiresPanelPassword)
	_ = d.Set("requires_license", os.RequiresLicense)

	if os.MaxPartitions != nil {
		_ = d.Set("max_partitions", *os.MaxPartitions)
	}

	_ = d.Set("released_at", types.FlattenTime(os.ReleasedAt))

	return nil
}
