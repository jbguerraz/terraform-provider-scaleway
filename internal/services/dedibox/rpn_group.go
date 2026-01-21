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
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/meta"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/verify"
)

func ResourceRpnGroup() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceRpnGroupCreate,
		ReadContext:   resourceRpnGroupRead,
		UpdateContext: resourceRpnGroupUpdate,
		DeleteContext: resourceRpnGroupDelete,
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
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Name of the RPN group",
			},
			"project_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "Project ID the RPN group belongs to",
			},
			"type": {
				Type:             schema.TypeString,
				Optional:         true,
				ForceNew:         true,
				Default:          "standard",
				Description:      "RPN group type (standard)",
				ValidateDiagFunc: verify.ValidateEnum[dedibox.RpnV2GroupType](),
			},
			"server_ids": {
				Type:        schema.TypeSet,
				Optional:    true,
				Description: "List of server IDs that are members of this RPN group",
				Elem: &schema.Schema{
					Type: schema.TypeInt,
				},
			},
			"rpnv1_compatible": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "Enable RPN V1 compatibility",
			},

			// Computed attributes
			"organization_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Organization ID the RPN group belongs to",
			},
			"status": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Status of the RPN group",
			},
			"owner": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Owner of the RPN group",
			},
			"members_count": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "Number of members in the RPN group",
			},
			"subnet": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "RPN group subnet",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"address": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Subnet IP address",
						},
						"cidr": {
							Type:        schema.TypeInt,
							Computed:    true,
							Description: "Subnet CIDR",
						},
					},
				},
			},
			"gateway": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "RPN group gateway",
			},
		},
	}
}

func newRpnV2API(m any) *dedibox.RpnV2API {
	return dedibox.NewRpnV2API(meta.ExtractScwClient(m))
}

func resourceRpnGroupCreate(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	api := newRpnV2API(m)

	serverIDs := expandServerIDs(d.Get("server_ids").(*schema.Set).List())

	createReq := &dedibox.RpnV2ApiCreateRpnV2GroupRequest{
		ProjectID: d.Get("project_id").(string),
		Name:      d.Get("name").(string),
		Type:      dedibox.RpnV2GroupType(d.Get("type").(string)),
		Servers:   serverIDs,
	}

	group, err := api.CreateRpnV2Group(createReq, scw.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(fmt.Sprintf("%d", group.ID))

	// Wait for the group to be ready
	_, err = api.WaitForRpnV2Group(&dedibox.WaitForRpnV2GroupRequest{
		GroupID:       group.ID,
		Timeout:       scw.TimeDurationPtr(d.Timeout(schema.TimeoutCreate)),
		RetryInterval: scw.TimeDurationPtr(retryInterval),
	}, scw.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}

	// Enable RPN V1 compatibility if requested
	if d.Get("rpnv1_compatible").(bool) {
		err = api.EnableRpnV2GroupCompatibility(&dedibox.RpnV2ApiEnableRpnV2GroupCompatibilityRequest{
			GroupID: group.ID,
		}, scw.WithContext(ctx))
		if err != nil {
			return diag.FromErr(err)
		}
	}

	return resourceRpnGroupRead(ctx, d, m)
}

func resourceRpnGroupRead(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	api := newRpnV2API(m)

	groupID, err := strconv.ParseUint(d.Id(), 10, 64)
	if err != nil {
		return diag.FromErr(fmt.Errorf("invalid RPN group ID: %s", d.Id()))
	}

	group, err := api.GetRpnV2Group(&dedibox.RpnV2ApiGetRpnV2GroupRequest{
		GroupID: groupID,
	}, scw.WithContext(ctx))
	if err != nil {
		if httperrors.Is404(err) {
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}

	_ = d.Set("name", group.Name)
	_ = d.Set("organization_id", group.OrganizationID)
	_ = d.Set("project_id", group.ProjectID)
	_ = d.Set("type", group.Type.String())
	_ = d.Set("status", group.Status.String())
	_ = d.Set("owner", group.Owner)
	_ = d.Set("members_count", group.MembersCount)
	_ = d.Set("rpnv1_compatible", group.CompatibleRpnv1)

	if group.Subnet != nil {
		_ = d.Set("subnet", []map[string]any{{
			"address": group.Subnet.Address.String(),
			"cidr":    group.Subnet.Cidr,
		}})
	}

	if group.Gateway != nil {
		_ = d.Set("gateway", group.Gateway.String())
	}

	// Get members to set server_ids
	membersResp, err := api.ListRpnV2Members(&dedibox.RpnV2ApiListRpnV2MembersRequest{
		GroupID: groupID,
	}, scw.WithAllPages(), scw.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}

	serverIDs := make([]int, 0)
	for _, member := range membersResp.Members {
		if member.Server != nil {
			serverIDs = append(serverIDs, int(member.Server.ID))
		}
	}
	_ = d.Set("server_ids", serverIDs)

	return nil
}

func resourceRpnGroupUpdate(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	api := newRpnV2API(m)

	groupID, err := strconv.ParseUint(d.Id(), 10, 64)
	if err != nil {
		return diag.FromErr(fmt.Errorf("invalid RPN group ID: %s", d.Id()))
	}

	if d.HasChange("name") {
		name := d.Get("name").(string)
		_, err = api.UpdateRpnV2GroupName(&dedibox.RpnV2ApiUpdateRpnV2GroupNameRequest{
			GroupID: groupID,
			Name:    &name,
		}, scw.WithContext(ctx))
		if err != nil {
			return diag.FromErr(err)
		}
	}

	if d.HasChange("rpnv1_compatible") {
		if d.Get("rpnv1_compatible").(bool) {
			err = api.EnableRpnV2GroupCompatibility(&dedibox.RpnV2ApiEnableRpnV2GroupCompatibilityRequest{
				GroupID: groupID,
			}, scw.WithContext(ctx))
		} else {
			err = api.DisableRpnV2GroupCompatibility(&dedibox.RpnV2ApiDisableRpnV2GroupCompatibilityRequest{
				GroupID: groupID,
			}, scw.WithContext(ctx))
		}
		if err != nil {
			return diag.FromErr(err)
		}
	}

	if d.HasChange("server_ids") {
		old, new := d.GetChange("server_ids")
		oldSet := old.(*schema.Set)
		newSet := new.(*schema.Set)

		// Servers to add
		toAdd := newSet.Difference(oldSet).List()
		if len(toAdd) > 0 {
			err = api.AddRpnV2Members(&dedibox.RpnV2ApiAddRpnV2MembersRequest{
				GroupID: groupID,
				Servers: expandServerIDs(toAdd),
			}, scw.WithContext(ctx))
			if err != nil {
				return diag.FromErr(err)
			}
		}

		// Servers to remove - need to find member IDs
		toRemove := oldSet.Difference(newSet).List()
		if len(toRemove) > 0 {
			// Get current members to find member IDs for servers to remove
			membersResp, err := api.ListRpnV2Members(&dedibox.RpnV2ApiListRpnV2MembersRequest{
				GroupID: groupID,
			}, scw.WithAllPages(), scw.WithContext(ctx))
			if err != nil {
				return diag.FromErr(err)
			}

			serverIDsToRemove := make(map[uint64]bool)
			for _, id := range toRemove {
				serverIDsToRemove[uint64(id.(int))] = true
			}

			memberIDsToRemove := make([]uint64, 0)
			for _, member := range membersResp.Members {
				if member.Server != nil && serverIDsToRemove[member.Server.ID] {
					memberIDsToRemove = append(memberIDsToRemove, member.ID)
				}
			}

			if len(memberIDsToRemove) > 0 {
				err = api.DeleteRpnV2Members(&dedibox.RpnV2ApiDeleteRpnV2MembersRequest{
					GroupID:   groupID,
					MemberIDs: memberIDsToRemove,
				}, scw.WithContext(ctx))
				if err != nil {
					return diag.FromErr(err)
				}
			}
		}

		// Wait for group to be ready after member changes
		_, err = api.WaitForRpnV2Group(&dedibox.WaitForRpnV2GroupRequest{
			GroupID:       groupID,
			Timeout:       scw.TimeDurationPtr(d.Timeout(schema.TimeoutUpdate)),
			RetryInterval: scw.TimeDurationPtr(retryInterval),
		}, scw.WithContext(ctx))
		if err != nil {
			return diag.FromErr(err)
		}
	}

	return resourceRpnGroupRead(ctx, d, m)
}

func resourceRpnGroupDelete(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	api := newRpnV2API(m)

	groupID, err := strconv.ParseUint(d.Id(), 10, 64)
	if err != nil {
		return diag.FromErr(fmt.Errorf("invalid RPN group ID: %s", d.Id()))
	}

	err = api.DeleteRpnV2Group(&dedibox.RpnV2ApiDeleteRpnV2GroupRequest{
		GroupID: groupID,
	}, scw.WithContext(ctx))
	if err != nil && !httperrors.Is404(err) {
		return diag.FromErr(err)
	}

	return nil
}

func expandServerIDs(ids []any) []uint64 {
	serverIDs := make([]uint64, len(ids))
	for i, id := range ids {
		serverIDs[i] = uint64(id.(int))
	}
	return serverIDs
}
