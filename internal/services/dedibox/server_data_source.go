package dedibox

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/datasource"
)

func DataSourceServer() *schema.Resource {
	// Generate datasource schema from resource schema
	dsSchema := datasource.SchemaFromResourceSchema(ResourceServer().Schema)

	// Add server_id as the primary key
	dsSchema["server_id"] = &schema.Schema{
		Type:        schema.TypeString,
		Required:    true,
		Description: "The ID of the server",
	}

	// Make non-computed fields optional in data source
	dsSchema["offer_id"].Required = false
	dsSchema["offer_id"].Optional = true
	dsSchema["offer_id"].Computed = true

	dsSchema["project_id"].Required = false
	dsSchema["project_id"].Optional = true
	dsSchema["project_id"].Computed = true

	return &schema.Resource{
		ReadContext: dataSourceServerRead,
		Schema:      dsSchema,
	}
}

func dataSourceServerRead(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	_, zone, err := newAPIWithZone(d, m)
	if err != nil {
		return diag.FromErr(err)
	}

	serverID, _ := d.GetOk("server_id")
	d.SetId(datasource.NewZonedID(serverID.(string), zone))

	return resourceServerRead(ctx, d, m)
}
