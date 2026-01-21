---
subcategory: "Dedibox"
page_title: "Scaleway: scaleway_dedibox_server"
---

# Resource: scaleway_dedibox_server

Creates and manages Scaleway Dedibox dedicated servers. For more information, see the [API documentation](https://www.scaleway.com/en/developers/api/dedibox/).

## Example Usage

### Basic

```terraform
data "scaleway_dedibox_offer" "my_offer" {
  zone         = "fr-par-1"
  name         = "Start-1-S-SATA"
  commercial_range = "start"
}

resource "scaleway_dedibox_server" "my_server" {
  zone       = "fr-par-1"
  offer_id   = data.scaleway_dedibox_offer.my_offer.offer_id
  project_id = "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx"
}
```

### With options

```terraform
data "scaleway_dedibox_offer" "my_offer" {
  zone = "fr-par-1"
  name = "Start-1-S-SATA"
}

resource "scaleway_dedibox_server" "my_server" {
  zone       = "fr-par-1"
  offer_id   = data.scaleway_dedibox_offer.my_offer.offer_id
  project_id = "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx"

  option_ids = [123, 456]  # Option IDs for additional features
}
```

## Argument Reference

The following arguments are supported:

- `offer_id` - (Required) The offer ID for the Dedibox server. Use the `scaleway_dedibox_offer` data source to find available offers.

~> **Important:** Updates to `offer_id` will recreate the server.

- `project_id` - (Required) The ID of the project the server is associated with.

- `option_ids` - (Optional) List of option IDs to enable on the server.

- `zone` - (Defaults to [provider](../index.md#zone) `zone`) The [zone](../guides/regions_and_zones.md#zones) in which the server should be created.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

- `id` - The ID of the server.

~> **Important:** Dedibox servers' IDs are [zoned](../guides/regions_and_zones.md#resource-ids), which means they are of the form `{zone}/{id}`, e.g. `fr-par-1/12345`

- `hostname` - The hostname of the server.
- `status` - The status of the server (e.g., `ready`, `delivering`, `error`).
- `offer_name` - The name of the offer.
- `os_id` - The ID of the installed operating system.
- `power_status` - The power status of the server.
- `hardware_comment` - Hardware-related comments.
- `rescue_enabled` - Whether rescue mode is enabled.
- `organization_id` - The organization ID the server is associated with.

- `options` - The options enabled on the server.
    - `id` - The option offer ID.
    - `name` - The option name.
    - `expires_at` - The expiration date of the option.

- `location` - The location of the server.
    - `rack` - The rack identifier.
    - `room` - The room identifier.
    - `datacenter_name` - The datacenter name.

- `ip` - The primary IP addresses of the server.
    - `ip_id` - The IP ID.
    - `address` - The IP address.
    - `reverse` - The reverse DNS.
    - `version` - The IP version (IPv4 or IPv6).
    - `cidr` - The CIDR notation.
    - `netmask` - The network mask.
    - `gateway` - The gateway IP.

## Import

Dedibox servers can be imported using the `{zone}/{id}`, e.g.

```bash
terraform import scaleway_dedibox_server.my_server fr-par-1/12345
```

## Timeouts

The `timeouts` block allows you to specify [timeouts](https://www.terraform.io/docs/configuration/blocks/resources/syntax.html#operation-timeouts) for certain actions:

- `create` - (Defaults to 60 minutes) Used when creating the server (waiting for delivery).
- `delete` - (Defaults to 10 minutes) Used when deleting the server.
