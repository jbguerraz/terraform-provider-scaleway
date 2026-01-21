---
subcategory: "Dedibox"
page_title: "Scaleway: scaleway_dedibox_failover_ip"
---

# scaleway_dedibox_failover_ip

Gets information about a Dedibox failover IP. For more information, see the [API documentation](https://www.scaleway.com/en/developers/api/dedibox/).

## Example Usage

```hcl
# Get failover IP by ID
data "scaleway_dedibox_failover_ip" "my_ip" {
  zone  = "fr-par-1"
  ip_id = 12345
}
```

## Argument Reference

- `ip_id` - (Required) The ID of the failover IP.

- `zone` - (Defaults to [provider](../index.md#zone) `zone`) The [zone](../guides/regions_and_zones.md#zones) in which the failover IP exists.

## Attributes Reference

In addition to all above arguments, the following attributes are exported:

- `id` - The ID of the failover IP.

~> **Important:** Dedibox failover IP IDs are [zoned](../guides/regions_and_zones.md#resource-ids), which means they are of the form `{zone}/{id}`, e.g. `fr-par-1/12345`

- `address` - The IP address.
- `reverse` - The reverse DNS value.
- `ip_version` - The IP version (IPv4 or IPv6).
- `cidr` - The CIDR notation.
- `netmask` - The network mask.
- `gateway` - The gateway IP.
- `status` - The status of the failover IP.
- `type` - The type of failover IP.
- `mac` - The MAC address (if applicable).
- `server_id` - The ID of the server the IP is attached to (if any).
