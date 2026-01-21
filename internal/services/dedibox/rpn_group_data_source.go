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

func DataSourceRpnGroup() *schema.Resource {
	// Generate datasource schema from resource schema
	dsSchema := datasource.SchemaFromResourceSchema(ResourceRpnGroup().Schema)

	// Add group_id as the primary key
	dsSchema["group_id"] = &schema.Schema{
		Type:        schema.TypeInt,
		Required:    true,
		Description: "The ID of the RPN group",
	}

	// Make non-computed fields optional in data source
	dsSchema["name"].Required = false
	dsSchema["name"].Optional = true
	dsSchema["name"].Computed = true

	dsSchema["project_id"].Required = false
	dsSchema["project_id"].Optional = true
	dsSchema["project_id"].Computed = true

	dsSchema["type"].Required = false
	dsSchema["type"].Optional = true
	dsSchema["type"].Computed = true
	dsSchema["type"].Default = nil

	dsSchema["server_ids"].Optional = true
	dsSchema["server_ids"].Computed = true

	dsSchema["rpnv1_compatible"].Optional = true
	dsSchema["rpnv1_compatible"].Computed = true
	dsSchema["rpnv1_compatible"].Default = nil

	return &schema.Resource{
		ReadContext: dataSourceRpnGroupRead,
		Schema:      dsSchema,
	}
}

func dataSourceRpnGroupRead(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	api := newRpnV2API(m)

	groupID := uint64(d.Get("group_id").(int))

	group, err := api.GetRpnV2Group(&dedibox.RpnV2ApiGetRpnV2GroupRequest{
		GroupID: groupID,
	}, scw.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(fmt.Sprintf("%d", group.ID))

	_ = d.Set("group_id", group.ID)
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
