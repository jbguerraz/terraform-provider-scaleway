---
subcategory: "Dedibox"
page_title: "Scaleway: scaleway_dedibox_server"
---

# scaleway_dedibox_server

Gets information about a Dedibox server. For more information, see the [API documentation](https://www.scaleway.com/en/developers/api/dedibox/).

## Example Usage

```hcl
# Get server by ID
data "scaleway_dedibox_server" "my_server" {
  zone      = "fr-par-1"
  server_id = 12345
}
```

## Argument Reference

- `server_id` - (Required) The ID of the server.

- `zone` - (Defaults to [provider](../index.md#zone) `zone`) The [zone](../guides/regions_and_zones.md#zones) in which the server exists.

## Attributes Reference

In addition to all above arguments, the following attributes are exported:

- `id` - The ID of the server.

~> **Important:** Dedibox servers' IDs are [zoned](../guides/regions_and_zones.md#resource-ids), which means they are of the form `{zone}/{id}`, e.g. `fr-par-1/12345`

- `hostname` - The hostname of the server.
- `status` - The status of the server.
- `offer_id` - The offer ID.
- `offer_name` - The name of the offer.
- `os_id` - The ID of the installed operating system.
- `power_status` - The power status of the server.
- `hardware_comment` - Hardware-related comments.
- `rescue_enabled` - Whether rescue mode is enabled.
- `organization_id` - The organization ID.
- `project_id` - The project ID.

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
