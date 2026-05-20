# Upgrading the AceCloud Terraform Provider

## Upgrading to v0.1.1

Drop-in replacement for v0.1.0. No schema changes, no state migration.

### Highlights

- `terraform destroy` now succeeds in a single invocation on multi-resource stacks
- `acecloud_volume_backup` create wait raised from 5 min to 30 min
- Resources deleted outside Terraform are now detected as drift on next plan
- GPU flavor + hourly/spot billing rejected at `terraform plan` time
- `acecloud_floating_ip_association` surfaces actionable errors on out-of-band drift
- `acecloud_key_pair.public_key` validated as at least 100 characters at plan time
- Terraform Registry sidebar now groups all guides under a single "Guides" parent

### Upgrade

```hcl
terraform {
  required_providers {
    acecloud = {
      source  = "AceCloudAI/acecloud"
      version = "~> 0.1.1"   # was "~> 0.1.0"
    }
  }
}
```

Then `terraform init -upgrade`. No state migration required.

## Upgrading to v0.1.0

This is the first release of the AceCloud Terraform Provider under the official `AceCloudAI` Registry namespace.

```hcl
terraform {
  required_providers {
    acecloud = {
      source  = "AceCloudAI/acecloud"
      version = "~> 0.1"
    }
  }
}
```

### If you used an earlier preview

A community-maintained preview of this provider was previously published under a personal Registry namespace. That namespace has been retired. To migrate:

1. Update the `source` line in your `required_providers` block to `AceCloudAI/acecloud`.
2. Update the `version` pin to `~> 0.1`.
3. Run `terraform init -upgrade`.
4. Run `terraform plan` and confirm no spurious changes.

No state migration is required — resource and data-source schemas match the previous preview at its final release, and new optional fields default to the previous implicit behaviour.

### New optional fields

- `acecloud_volume_backup.force` — optional, defaults to `false`. Set to `true` to back up a volume that is still attached and in use; leave unset (or `false`) when the source volume is unattached. Existing configurations that backed up unattached volumes do not need to be changed.
