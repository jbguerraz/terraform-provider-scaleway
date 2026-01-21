---
subcategory: "Dedibox"
page_title: "Scaleway: scaleway_dedibox_offer"
---

# scaleway_dedibox_offer

Gets information about a Dedibox offer. For more information, see the [API documentation](https://www.scaleway.com/en/developers/api/dedibox/).

## Example Usage

```hcl
# Get info by offer name
data "scaleway_dedibox_offer" "my_offer" {
  zone = "fr-par-1"
  name = "Start-1-S-SATA"
}

# Get info by offer ID
data "scaleway_dedibox_offer" "my_offer" {
  zone     = "fr-par-1"
  offer_id = 12345
}

# Filter by commercial range
data "scaleway_dedibox_offer" "my_offer" {
  zone             = "fr-par-1"
  name             = "Start-1-S-SATA"
  commercial_range = "start"
}
```

## Argument Reference

- `name` - (Optional) The offer name. Only one of `name` and `offer_id` should be specified.

- `offer_id` - (Optional) The offer ID. Only one of `name` and `offer_id` should be specified.

- `commercial_range` - (Optional) Filter by commercial range (e.g., `start`, `pro`, `store`, `core`).

- `zone` - (Defaults to [provider](../index.md#zone) `zone`) The [zone](../guides/regions_and_zones.md#zones) in which to search for offers.

## Attributes Reference

In addition to all above arguments, the following attributes are exported:

- `id` - The ID of the offer.

~> **Important:** Dedibox offer IDs are [zoned](../guides/regions_and_zones.md#resource-ids), which means they are of the form `{zone}/{id}`, e.g. `fr-par-1/12345`

- `payment_frequency` - The payment frequency (e.g., `monthly`).

- `pricing` - The pricing information.
    - `currency` - The currency.
    - `unit_price` - The unit price in the smallest currency unit.
    - `recurring_price` - The recurring price.

- `server_info` - Server hardware specifications.
    - `core_count` - Number of CPU cores.
    - `threads_per_core` - Threads per CPU core.
    - `memory` - Memory in bytes.
    - `stock` - Stock status (`available`, `low`, `empty`).
    - `bandwidth` - Network bandwidth in bps.
    - `rpm` - Disk RPM (for HDD).
    - `disks` - List of disk specifications.
        - `capacity` - Disk capacity in bytes.
        - `type` - Disk type (e.g., `sata`, `ssd`, `nvme`).

- `available_options` - List of available options for this offer.
    - `id` - Option ID.
    - `name` - Option name.

- `max_bandwidth` - Maximum bandwidth in bps.
- `quota` - Quota limits for this offer.
