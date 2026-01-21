---
subcategory: "Dedibox"
page_title: "Scaleway: scaleway_dedibox_failover_ip"
---

# Resource: scaleway_dedibox_failover_ip

Creates and manages Scaleway Dedibox failover IPs. Failover IPs can be moved between servers and provide high availability for your services. For more information, see the [API documentation](https://www.scaleway.com/en/developers/api/dedibox/).

## Example Usage

### Basic

```terraform
resource "scaleway_dedibox_failover_ip" "my_ip" {
  zone       = "fr-par-1"
  offer_id   = 1  # Failover IP offer ID
  project_id = "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx"
}
```

### Attached to a server

```terraform
resource "scaleway_dedibox_server" "my_server" {
  zone       = "fr-par-1"
  offer_id   = data.scaleway_dedibox_offer.my_offer.offer_id
  project_id = "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx"
}

resource "scaleway_dedibox_failover_ip" "my_ip" {
  zone       = "fr-par-1"
  offer_id   = 1
  project_id = "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx"
  server_id  = scaleway_dedibox_server.my_server.id
}
```

## Argument Reference

The following arguments are supported:

- `offer_id` - (Required) The offer ID for the failover IP.

~> **Important:** Updates to `offer_id` will recreate the failover IP.

- `project_id` - (Required) The ID of the project the failover IP is associated with.

- `server_id` - (Optional) The ID of the server to attach the failover IP to. If not specified, the IP will be unattached.

- `zone` - (Defaults to [provider](../index.md#zone) `zone`) The [zone](../guides/regions_and_zones.md#zones) in which the failover IP should be created.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

- `id` - The ID of the failover IP.

~> **Important:** Dedibox failover IP IDs are [zoned](../guides/regions_and_zones.md#resource-ids), which means they are of the form `{zone}/{id}`, e.g. `fr-par-1/12345`

- `address` - The IP address.
- `reverse` - The reverse DNS value.
- `ip_version` - The IP version (IPv4 or IPv6).
- `cidr` - The CIDR notation.
- `netmask` - The network mask.
- `gateway` - The gateway IP.
- `status` - The status of the failover IP.

## Import

Dedibox failover IPs can be imported using the `{zone}/{id}`, e.g.

```bash
terraform import scaleway_dedibox_failover_ip.my_ip fr-par-1/12345
```

## Timeouts

The `timeouts` block allows you to specify [timeouts](https://www.terraform.io/docs/configuration/blocks/resources/syntax.html#operation-timeouts) for certain actions:

- `create` - (Defaults to 60 minutes) Used when creating the failover IP.
- `update` - (Defaults to 10 minutes) Used when attaching/detaching the IP.
- `delete` - (Defaults to 10 minutes) Used when deleting the failover IP.
