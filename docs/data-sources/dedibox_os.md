---
subcategory: "Dedibox"
page_title: "Scaleway: scaleway_dedibox_os"
---

# scaleway_dedibox_os

Gets information about an available operating system for Dedibox servers. For more information, see the [API documentation](https://www.scaleway.com/en/developers/api/dedibox/).

## Example Usage

```hcl
# Get OS by name
data "scaleway_dedibox_os" "ubuntu" {
  zone = "fr-par-1"
  name = "Ubuntu"
}

# Get OS by ID
data "scaleway_dedibox_os" "my_os" {
  zone  = "fr-par-1"
  os_id = 12345
}

# Get OS compatible with a specific server
data "scaleway_dedibox_os" "ubuntu_for_server" {
  zone      = "fr-par-1"
  name      = "Ubuntu"
  server_id = 67890
}

# Get OS by type
data "scaleway_dedibox_os" "proxmox" {
  zone = "fr-par-1"
  name = "Proxmox"
  type = "virtualization"
}
```

## Argument Reference

- `name` - (Optional) The OS name (partial match supported). Only one of `name` and `os_id` should be specified.

- `os_id` - (Optional) The OS ID. Only one of `name` and `os_id` should be specified.

- `server_id` - (Optional) Filter OS by compatibility with a specific server ID.

- `type` - (Optional) Filter by OS type. Possible values: `server`, `virtualization`, `panel`, `desktop`, `custom`.

- `zone` - (Defaults to [provider](../index.md#zone) `zone`) The [zone](../guides/regions_and_zones.md#zones) in which to search for OS.

## Attributes Reference

In addition to all above arguments, the following attributes are exported:

- `id` - The ID of the OS.

~> **Important:** Dedibox OS IDs are [zoned](../guides/regions_and_zones.md#resource-ids), which means they are of the form `{zone}/{id}`, e.g. `fr-par-1/12345`

- `version` - The version of the OS.
- `arch` - The architecture of the OS (e.g., `x86_64`, `arm64`).
- `display_name` - The display name of the OS.

- `allow_custom_partitioning` - Whether the OS allows custom partitioning.
- `allow_ssh_keys` - Whether the OS allows SSH keys.
- `requires_user` - Whether the OS requires a user to be created.
- `requires_admin_password` - Whether the OS requires an admin/root password.
- `requires_panel_password` - Whether the OS requires a panel password.
- `requires_license` - Whether the OS requires a license.

- `max_partitions` - Maximum number of partitions allowed.
- `released_at` - The OS release date.

## Notes

Use the computed attributes to determine which fields are required when using `scaleway_dedibox_server_install`:
- If `requires_user` is true, provide `user_login` and `user_password`
- If `requires_admin_password` is true, provide `root_password`
- If `requires_panel_password` is true, provide `panel_password`
- If `requires_license` is true, provide `license_offer_id`
- If `allow_ssh_keys` is true, you can provide `ssh_key_ids`
- If `allow_custom_partitioning` is true, you can provide custom `partition` blocks
