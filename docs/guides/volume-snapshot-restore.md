---
page_title: "Volume snapshot and restore"
subcategory: "Storage"
description: |-
  Take a point-in-time snapshot of a volume, then restore from it by creating a new volume from the snapshot ID.
---

# Volume snapshot and restore

Snapshots are fast, in-region copies of a volume at a single point in time. Common use cases:

- Pre-change safety net (snapshot before a risky change, drop the snapshot once verified)
- Cloning a volume to a new instance
- Branching from a base volume

This guide covers the full cycle: snapshot a volume, then create a new volume from the snapshot ID and attach it to an instance.

This guide assumes you already have an instance with an attached data volume. If you need to set one up first, see the **Single instance with SSH** guide.

## Example

```hcl
# Existing data volume attached to an existing instance.
# (See the Single instance with SSH guide for instance + volume setup.)

# 1. Snapshot the data volume.
#    Take the snapshot after the volume is attached and any in-flight writes
#    are fsynced. For application-consistent snapshots, freeze the filesystem
#    in the guest OS first.
resource "acecloud_snapshot" "data" {
  name        = "source-data-snapshot"
  description = "Point-in-time snapshot before risky change"
  volume_id   = acecloud_volume.source_data.id

  depends_on = [acecloud_volume_attachment.source_data]
}

# 2. Restore: create a new volume from the snapshot.
#    Size must be at least the source volume's size.
resource "acecloud_volume" "restored_data" {
  name         = "restored-data"
  description  = "Volume restored from snapshot"
  size         = 50
  volume_type  = "ssd"
  billing_type = "hourly"
  snapshot_id  = acecloud_snapshot.data.id
}

# 3. Optionally attach the restored volume to an instance.
resource "acecloud_volume_attachment" "restored_data" {
  instance_id           = acecloud_instance.target.id
  volume_id             = acecloud_volume.restored_data.id
  delete_on_termination = false
}

output "snapshot_id"        { value = acecloud_snapshot.data.id }
output "restored_volume_id" { value = acecloud_volume.restored_data.id }
```
