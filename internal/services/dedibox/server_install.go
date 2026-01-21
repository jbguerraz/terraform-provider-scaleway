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
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/verify"
)

func ResourceServerInstall() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceServerInstallCreate,
		ReadContext:   resourceServerInstallRead,
		DeleteContext: resourceServerInstallDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: &schema.ResourceTimeout{
			Create:  schema.DefaultTimeout(defaultServiceTimeout),
			Read:    schema.DefaultTimeout(defaultServerTimeout),
			Delete:  schema.DefaultTimeout(defaultServerTimeout),
			Default: schema.DefaultTimeout(defaultServerTimeout),
		},
		Schema: map[string]*schema.Schema{
			"server_id": {
				Type:        schema.TypeInt,
				Required:    true,
				ForceNew:    true,
				Description: "ID of the server to install",
			},
			"os_id": {
				Type:        schema.TypeInt,
				Required:    true,
				ForceNew:    true,
				Description: "ID of the OS to install",
			},
			"hostname": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "Hostname to set on the server",
			},
			"user_login": {
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
				Description: "User login to create (if OS requires it)",
			},
			"user_password": {
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
				Sensitive:   true,
				Description: "User password (if OS requires it)",
			},
			"root_password": {
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
				Sensitive:   true,
				Description: "Root password (if OS requires admin password)",
			},
			"panel_password": {
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
				Sensitive:   true,
				Description: "Panel password (if OS requires it)",
			},
			"ssh_key_ids": {
				Type:        schema.TypeList,
				Optional:    true,
				ForceNew:    true,
				Description: "SSH key IDs to authorize on the server",
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"partition": {
				Type:        schema.TypeList,
				Optional:    true,
				ForceNew:    true,
				Description: "Custom partitions (if OS allows custom partitioning)",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"file_system": {
							Type:             schema.TypeString,
							Required:         true,
							Description:      "File system type (ext4, xfs, swap, etc.)",
							ValidateDiagFunc: verify.ValidateEnum[dedibox.PartitionFileSystem](),
						},
						"mount_point": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "Mount point (e.g., /, /home, /var)",
						},
						"raid_level": {
							Type:             schema.TypeString,
							Optional:         true,
							Default:          "no_raid",
							Description:      "RAID level (no_raid, raid0, raid1, raid5, raid6, raid10)",
							ValidateDiagFunc: verify.ValidateEnum[dedibox.RaidArrayRaidLevel](),
						},
						"capacity": {
							Type:        schema.TypeInt,
							Required:    true,
							Description: "Capacity in bytes",
						},
						"connectors": {
							Type:        schema.TypeList,
							Optional:    true,
							Description: "Disk connectors to use",
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
					},
				},
			},
			"license_offer_id": {
				Type:        schema.TypeInt,
				Optional:    true,
				ForceNew:    true,
				Description: "License offer ID (if OS requires a license)",
			},
			"zone": zonal.Schema(),

			// Computed attributes
			"status": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Installation status",
			},
			"panel_url": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Panel URL (if OS has a panel)",
			},
		},
	}
}

func resourceServerInstallCreate(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	api, zone, err := newAPIWithZone(d, m)
	if err != nil {
		return diag.FromErr(err)
	}

	serverID := uint64(d.Get("server_id").(int))

	installReq := &dedibox.InstallServerRequest{
		Zone:     zone,
		ServerID: serverID,
		OsID:     uint64(d.Get("os_id").(int)),
		Hostname: d.Get("hostname").(string),
	}

	if v, ok := d.GetOk("user_login"); ok {
		login := v.(string)
		installReq.UserLogin = &login
	}

	if v, ok := d.GetOk("user_password"); ok {
		password := v.(string)
		installReq.UserPassword = &password
	}

	if v, ok := d.GetOk("root_password"); ok {
		password := v.(string)
		installReq.RootPassword = &password
	}

	if v, ok := d.GetOk("panel_password"); ok {
		password := v.(string)
		installReq.PanelPassword = &password
	}

	if v, ok := d.GetOk("ssh_key_ids"); ok {
		sshKeys := v.([]interface{})
		installReq.SSHKeyIDs = make([]string, len(sshKeys))
		for i, key := range sshKeys {
			installReq.SSHKeyIDs[i] = key.(string)
		}
	}

	if v, ok := d.GetOk("partition"); ok {
		partitions := v.([]interface{})
		installReq.Partitions = make([]*dedibox.InstallPartition, len(partitions))
		for i, p := range partitions {
			partition := p.(map[string]interface{})
			installPartition := &dedibox.InstallPartition{
				FileSystem: dedibox.PartitionFileSystem(partition["file_system"].(string)),
				Capacity:   scw.Size(partition["capacity"].(int)),
				RaidLevel:  dedibox.RaidArrayRaidLevel(partition["raid_level"].(string)),
			}

			if mp, ok := partition["mount_point"].(string); ok && mp != "" {
				installPartition.MountPoint = &mp
			}

			if connectors, ok := partition["connectors"].([]interface{}); ok {
				installPartition.Connectors = make([]string, len(connectors))
				for j, c := range connectors {
					installPartition.Connectors[j] = c.(string)
				}
			}

			installReq.Partitions[i] = installPartition
		}
	}

	if v, ok := d.GetOk("license_offer_id"); ok {
		licenseID := uint64(v.(int))
		installReq.LicenseOfferID = &licenseID
	}

	_, err = api.InstallServer(installReq, scw.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(zonal.NewIDString(zone, fmt.Sprintf("%d", serverID)))

	// Wait for installation to complete
	_, err = api.WaitForServerInstall(&dedibox.WaitForServerInstallRequest{
		ServerID:      serverID,
		Zone:          zone,
		Timeout:       scw.TimeDurationPtr(d.Timeout(schema.TimeoutCreate)),
		RetryInterval: scw.TimeDurationPtr(retryInterval),
	}, scw.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}

	return resourceServerInstallRead(ctx, d, m)
}

func resourceServerInstallRead(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	api, zonedID, err := NewAPIWithZoneAndID(m, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	serverID, err := strconv.ParseUint(zonedID.ID, 10, 64)
	if err != nil {
		return diag.FromErr(fmt.Errorf("invalid server ID: %s", zonedID.ID))
	}

	install, err := api.GetServerInstall(&dedibox.GetServerInstallRequest{
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

	_ = d.Set("zone", zonedID.Zone)
	_ = d.Set("server_id", serverID)
	_ = d.Set("os_id", install.OsID)
	_ = d.Set("hostname", install.Hostname)
	_ = d.Set("status", install.Status.String())

	if install.UserLogin != nil {
		_ = d.Set("user_login", *install.UserLogin)
	}

	if install.SSHKeyIDs != nil {
		_ = d.Set("ssh_key_ids", install.SSHKeyIDs)
	}

	if install.PanelURL != nil {
		_ = d.Set("panel_url", *install.PanelURL)
	}

	// Flatten partitions
	if install.Partitions != nil {
		partitions := make([]map[string]interface{}, len(install.Partitions))
		for i, p := range install.Partitions {
			partition := map[string]interface{}{
				"file_system": p.FileSystem.String(),
				"raid_level":  p.RaidLevel.String(),
				"capacity":    int(p.Capacity),
				"connectors":  p.Connectors,
			}
			if p.MountPoint != nil {
				partition["mount_point"] = *p.MountPoint
			}
			partitions[i] = partition
		}
		_ = d.Set("partition", partitions)
	}

	return nil
}

func resourceServerInstallDelete(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	// Server installations cannot be deleted, only the server itself can be deleted
	// We simply remove from state
	return nil
}
