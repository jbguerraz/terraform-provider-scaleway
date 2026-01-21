---
subcategory: "Dedibox"
page_title: "Scaleway: scaleway_dedibox_rpn_group"
---

# scaleway_dedibox_rpn_group

Gets information about a Dedibox RPN (Real Private Network) V2 group. For more information, see the [API documentation](https://www.scaleway.com/en/developers/api/dedibox/).

## Example Usage

```hcl
# Get RPN group by ID
data "scaleway_dedibox_rpn_group" "my_rpn" {
  group_id = 12345
}
```

## Argument Reference

- `group_id` - (Required) The ID of the RPN group.

## Attributes Reference

In addition to all above arguments, the following attributes are exported:

- `id` - The ID of the RPN group.

- `name` - The name of the RPN group.
- `type` - The type of RPN group (`standard` or `qinq`).
- `status` - The status of the RPN group.
- `owner` - The owner of the RPN group.
- `members_count` - The number of members in the group.
- `organization_id` - The organization ID.
- `project_id` - The project ID.
- `rpnv1_compatible` - Whether the group is compatible with RPN V1.

- `server_ids` - List of server IDs that are members of the group.

- `subnet` - The subnet information for the RPN group.
    - `address` - The subnet address.
    - `cidr` - The CIDR notation.

- `gateway` - The gateway IP for the RPN group.
