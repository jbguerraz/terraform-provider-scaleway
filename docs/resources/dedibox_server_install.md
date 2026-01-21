---
subcategory: "Dedibox"
page_title: "Scaleway: scaleway_dedibox_server_install"
---

# Resource: scaleway_dedibox_server_install

Installs an operating system on a Scaleway Dedibox server. This resource allows you to configure OS installation including partitioning, SSH keys, and licenses. For more information, see the [API documentation](https://www.scaleway.com/en/developers/api/dedibox/).

## Example Usage

### Basic

```terraform
data "scaleway_dedibox_os" "ubuntu" {
  zone = "fr-par-1"
  name = "Ubuntu"
  type = "server"
}

resource "scaleway_dedibox_server_install" "my_install" {
  zone      = "fr-par-1"
  server_id = scaleway_dedibox_server.my_server.id
  os_id     = data.scaleway_dedibox_os.ubuntu.os_id
  hostname  = "my-server.example.com"
}
```

### With SSH keys

```terraform
data "scaleway_dedibox_os" "ubuntu" {
  zone = "fr-par-1"
  name = "Ubuntu"
  type = "server"
}

resource "scaleway_dedibox_server_install" "my_install" {
  zone      = "fr-par-1"
  server_id = scaleway_dedibox_server.my_server.id
  os_id     = data.scaleway_dedibox_os.ubuntu.os_id
  hostname  = "my-server.example.com"

  ssh_key_ids = ["key-id-1", "key-id-2"]
}
```

### With user and password

```terraform
data "scaleway_dedibox_os" "proxmox" {
  zone = "fr-par-1"
  name = "Proxmox"
  type = "virtualization"
}

resource "scaleway_dedibox_server_install" "my_install" {
  zone          = "fr-par-1"
  server_id     = scaleway_dedibox_server.my_server.id
  os_id         = data.scaleway_dedibox_os.proxmox.os_id
  hostname      = "proxmox.example.com"
  user_login    = "admin"
  user_password = "secure-password"
  root_password = "root-secure-password"
}
```

### With custom partitioning

```terraform
data "scaleway_dedibox_os" "ubuntu" {
  zone = "fr-par-1"
  name = "Ubuntu"
  type = "server"
}

resource "scaleway_dedibox_server_install" "my_install" {
  zone      = "fr-par-1"
  server_id = scaleway_dedibox_server.my_server.id
  os_id     = data.scaleway_dedibox_os.ubuntu.os_id
  hostname  = "my-server.example.com"

  partition {
    file_system = "ext4"
    mount_point = "/"
    raid_level  = "raid1"
    capacity    = 50000000000  # 50GB in bytes
  }

  partition {
    file_system = "ext4"
    mount_point = "/home"
    raid_level  = "raid1"
    capacity    = 100000000000  # 100GB in bytes
  }

  partition {
    file_system = "swap"
    raid_level  = "no_raid"
    capacity    = 8000000000  # 8GB in bytes
  }
}
```

## Argument Reference

The following arguments are supported:

- `server_id` - (Required) The ID of the server to install.

~> **Important:** Updates to `server_id` will recreate the installation.

- `os_id` - (Required) The ID of the operating system to install. Use the `scaleway_dedibox_os` data source to find available OS IDs.

~> **Important:** Updates to `os_id` will recreate the installation.

- `hostname` - (Required) The hostname to set on the server. Must be a valid hostname.

~> **Important:** Updates to `hostname` will recreate the installation.

- `user_login` - (Optional) User login to create on the server. Required by some operating systems.

- `user_password` - (Optional, Sensitive) User password. Required by some operating systems.

- `root_password` - (Optional, Sensitive) Root/admin password. Required by some operating systems.

- `panel_password` - (Optional, Sensitive) Panel password for operating systems with web panels (e.g., Plesk, cPanel).

- `ssh_key_ids` - (Optional) List of SSH key IDs to authorize on the server.

- `partition` - (Optional) Custom partition configuration. Only available for operating systems that support custom partitioning.
    - `file_system` - (Required) File system type. Possible values: `ext4`, `xfs`, `swap`, `fat32`, `ntfs`, `ufs`, `zfs`.
    - `mount_point` - (Optional) Mount point for the partition (e.g., `/`, `/home`, `/var`). Not required for swap.
    - `raid_level` - (Optional, default `no_raid`) RAID level. Possible values: `no_raid`, `raid0`, `raid1`, `raid5`, `raid6`, `raid10`.
    - `capacity` - (Required) Capacity in bytes.
    - `connectors` - (Optional) List of disk connectors to use for this partition.

- `license_offer_id` - (Optional) License offer ID for operating systems that require a license (e.g., Windows, cPanel).

- `zone` - (Defaults to [provider](../index.md#zone) `zone`) The [zone](../guides/regions_and_zones.md#zones) in which the server exists.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

- `id` - The ID of the installation resource.

~> **Important:** Dedibox server install IDs are [zoned](../guides/regions_and_zones.md#resource-ids), which means they are of the form `{zone}/{server_id}`, e.g. `fr-par-1/12345`

- `status` - The installation status.
- `panel_url` - The panel URL (if the OS has a web panel).

## Import

Dedibox server installations can be imported using the `{zone}/{server_id}`, e.g.

```bash
terraform import scaleway_dedibox_server_install.my_install fr-par-1/12345
```

## Timeouts

The `timeouts` block allows you to specify [timeouts](https://www.terraform.io/docs/configuration/blocks/resources/syntax.html#operation-timeouts) for certain actions:

- `create` - (Defaults to 60 minutes) Used when installing the OS (waiting for installation to complete).

## Notes

- Server installations cannot be truly "deleted" - removing this resource from Terraform only removes it from state.
- To reinstall a server, you must destroy and recreate this resource.
- Check the OS requirements using the `scaleway_dedibox_os` data source to see which fields are required (user, password, panel password, etc.).
