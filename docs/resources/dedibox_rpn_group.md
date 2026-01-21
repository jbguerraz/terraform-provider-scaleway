---
subcategory: "Dedibox"
page_title: "Scaleway: scaleway_dedibox_rpn_group"
---

# Resource: scaleway_dedibox_rpn_group

Creates and manages Scaleway Dedibox RPN (Real Private Network) V2 groups. RPN provides private networking between Dedibox servers. For more information, see the [API documentation](https://www.scaleway.com/en/developers/api/dedibox/).

## Example Usage

### Basic

```terraform
resource "scaleway_dedibox_rpn_group" "my_rpn" {
  name       = "my-rpn-group"
  project_id = "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx"
  type       = "standard"
}
```

### With servers

```terraform
resource "scaleway_dedibox_server" "server1" {
  zone       = "fr-par-1"
  offer_id   = data.scaleway_dedibox_offer.my_offer.offer_id
  project_id = "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx"
}

resource "scaleway_dedibox_server" "server2" {
  zone       = "fr-par-1"
  offer_id   = data.scaleway_dedibox_offer.my_offer.offer_id
  project_id = "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx"
}

resource "scaleway_dedibox_rpn_group" "my_rpn" {
  name       = "my-rpn-group"
  project_id = "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx"
  type       = "standard"

  server_ids = [
    scaleway_dedibox_server.server1.id,
    scaleway_dedibox_server.server2.id,
  ]
}
```

### RPN V1 compatible

```terraform
resource "scaleway_dedibox_rpn_group" "my_rpn" {
  name             = "my-rpn-group"
  project_id       = "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx"
  type             = "standard"
  rpnv1_compatible = true
}
```

## Argument Reference

The following arguments are supported:

- `name` - (Required) The name of the RPN group.

- `project_id` - (Required) The ID of the project the RPN group is associated with.

- `type` - (Required) The type of RPN group. Possible values are `standard` and `qinq`.

~> **Important:** Updates to `type` will recreate the RPN group.

- `server_ids` - (Optional) List of server IDs to add as members of the RPN group.

- `rpnv1_compatible` - (Optional, default `false`) Whether the group should be compatible with RPN V1.

~> **Important:** Updates to `rpnv1_compatible` will recreate the RPN group.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

- `id` - The ID of the RPN group.

- `organization_id` - The organization ID the group is associated with.
- `status` - The status of the RPN group.
- `owner` - The owner of the RPN group.
- `members_count` - The number of members in the group.

- `subnet` - The subnet information for the RPN group.
    - `address` - The subnet address.
    - `cidr` - The CIDR notation.

- `gateway` - The gateway IP for the RPN group.

## Import

Dedibox RPN groups can be imported using the group ID, e.g.

```bash
terraform import scaleway_dedibox_rpn_group.my_rpn 12345
```

## Timeouts

The `timeouts` block allows you to specify [timeouts](https://www.terraform.io/docs/configuration/blocks/resources/syntax.html#operation-timeouts) for certain actions:

- `create` - (Defaults to 10 minutes) Used when creating the RPN group.
- `update` - (Defaults to 10 minutes) Used when updating members.
- `delete` - (Defaults to 10 minutes) Used when deleting the RPN group.
