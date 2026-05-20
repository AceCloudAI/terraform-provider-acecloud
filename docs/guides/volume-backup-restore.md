---
page_title: "Volume backup and restore"
description: |-
  Create a durable backup of a volume, then restore by creating a new volume from the backup ID.
---

# Volume backup and restore

This guide covers the volume backup lifecycle: take a backup of a volume that is attached to a running instance, then restore by creating a new volume from the backup ID.

Use cases:

- Periodic full backups for compliance or retention
- Pre-migration checkpoint before a large data change
- Restoring a backup into a separate volume for verification or DR

This guide assumes you already have an instance with an attached data volume. If you need to set one up first, see the **Single instance with SSH** guide.

## Example

```hcl
# Existing data volume attached to an existing instance.
# (See the Single instance with SSH guide for instance + volume setup.)

# 1. Back up the data volume.
#    `force = true` lets the backup run while the volume is still attached
#    and in use. Leave it unset for unattached volumes.
resource "acecloud_volume_backup" "weekly" {
  name        = "source-data-weekly"
  description = "Weekly backup of source data volume"
  volume_id   = acecloud_volume.source_data.id
  force       = true
}

# 2. Restore: create a new volume from the backup ID.
#    Same region by default. Size must be at least the source volume's size.
resource "acecloud_volume" "restored" {
  name         = "restored-from-backup"
  description  = "Volume restored from the weekly backup"
  size         = 100
  volume_type  = "NVMe based High IOPS Storage"
  billing_type = "hourly"
  backup_id    = acecloud_volume_backup.weekly.id
}

# 3. Optionally attach the restored volume to an instance.
resource "acecloud_volume_attachment" "restored" {
  instance_id           = acecloud_instance.target.id
  volume_id             = acecloud_volume.restored.id
  delete_on_termination = false
}

output "backup_id"          { value = acecloud_volume_backup.weekly.id }
output "restored_volume_id" { value = acecloud_volume.restored.id }
```
