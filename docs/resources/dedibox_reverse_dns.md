---
subcategory: "Dedibox"
page_title: "Scaleway: scaleway_dedibox_reverse_dns"
---

# Resource: scaleway_dedibox_reverse_dns

Manages reverse DNS (PTR records) for Scaleway Dedibox IPs. For more information, see the [API documentation](https://www.scaleway.com/en/developers/api/dedibox/).

## Example Usage

### Basic

```terraform
resource "scaleway_dedibox_reverse_dns" "my_reverse" {
  zone    = "fr-par-1"
  ip_id   = 12345
  reverse = "my-server.example.com"
}
```

### With failover IP

```terraform
resource "scaleway_dedibox_failover_ip" "my_ip" {
  zone       = "fr-par-1"
  offer_id   = 1
  project_id = "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx"
}

resource "scaleway_dedibox_reverse_dns" "my_reverse" {
  zone    = "fr-par-1"
  ip_id   = scaleway_dedibox_failover_ip.my_ip.id
  reverse = "failover.example.com"
}
```

## Argument Reference

The following arguments are supported:

- `ip_id` - (Required) The ID of the IP to set reverse DNS for.

~> **Important:** Updates to `ip_id` will recreate the reverse DNS record.

- `reverse` - (Required) The reverse DNS value (PTR record). Must be a valid FQDN.

- `zone` - (Defaults to [provider](../index.md#zone) `zone`) The [zone](../guides/regions_and_zones.md#zones) in which the IP exists.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

- `id` - The ID of the reverse DNS resource.

~> **Important:** Dedibox reverse DNS IDs are [zoned](../guides/regions_and_zones.md#resource-ids), which means they are of the form `{zone}/{ip_id}`, e.g. `fr-par-1/12345`

- `address` - The IP address.
- `ip_version` - The IP version (IPv4 or IPv6).
- `cidr` - The CIDR notation.
- `netmask` - The network mask.
- `gateway` - The gateway IP.
- `status` - The IP status.

## Import

Dedibox reverse DNS can be imported using the `{zone}/{ip_id}`, e.g.

```bash
terraform import scaleway_dedibox_reverse_dns.my_reverse fr-par-1/12345
```

## Notes

- Deleting this resource will reset the reverse DNS to an empty value.
- The reverse DNS value must resolve to the IP address for it to be effective.
