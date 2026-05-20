---
page_title: "Multi-region and multi-project setups"
description: |-
  Patterns for managing AceCloud infrastructure across multiple regions and projects using Terraform provider aliases.
---

# Multi-region and multi-project setups

For configurations that span more than one region or more than one project, use **Terraform provider aliases**.

> **About credentials across regions/projects:** AceCloud API keys can be scoped to a specific service, account, or project. Whether a single key works across the regions or projects you target depends on how the key was issued. The examples below cover both scenarios — **same credentials reused** and **separate credentials per provider alias**.

## Pattern 1: one configuration, two regions, **same credentials**

Use this when your API key is account-scoped and authorized across all regions you target.

```hcl
locals {
  # Single set of credentials reused across regions.
  shared_auth = {
    api_key_id           = var.acecloud_api_key_id
    api_key_secret       = var.acecloud_api_key_secret
    api_key_service_name = "terraform"
  }
}

# Default provider — used by resources that don't specify a `provider =` argument.
provider "acecloud" {
  api_key_id           = local.shared_auth.api_key_id
  api_key_secret       = local.shared_auth.api_key_secret
  api_key_service_name = local.shared_auth.api_key_service_name
  region               = "ap-south-noi-1"
  project_id           = var.primary_project_id
}

# Aliased provider — same credentials, different region.
provider "acecloud" {
  alias                = "mumbai"
  api_key_id           = local.shared_auth.api_key_id
  api_key_secret       = local.shared_auth.api_key_secret
  api_key_service_name = local.shared_auth.api_key_service_name
  region               = "ap-south-mum-1"
  project_id           = var.primary_project_id
}

resource "acecloud_instance" "primary" {
  # Uses the default provider → noida
  name = "primary-app"
  # ... other attributes ...
}

resource "acecloud_instance" "dr" {
  provider = acecloud.mumbai
  name     = "dr-app"
  # ... other attributes ...
}
```

## Pattern 2: one configuration, two regions, **separate credentials per region**

Use this when each region requires its own API key (different account, different team-owned key, or region-scoped key). Every provider alias gets its own credential set; nothing is shared.

```hcl
# Noida — primary region, primary team's key.
provider "acecloud" {
  api_key_id           = var.noida_api_key_id
  api_key_secret       = var.noida_api_key_secret
  api_key_service_name = "terraform-primary"
  region               = "ap-south-noi-1"
  project_id           = var.noida_project_id
}

# Mumbai — DR region, separate team or separate account, different key.
provider "acecloud" {
  alias                = "mumbai"
  api_key_id           = var.mumbai_api_key_id
  api_key_secret       = var.mumbai_api_key_secret
  api_key_service_name = "terraform-dr"
  region               = "ap-south-mum-1"
  project_id           = var.mumbai_project_id
}

resource "acecloud_instance" "primary" {
  name = "primary-app"
  # ... other attributes ...
}

resource "acecloud_instance" "dr" {
  provider = acecloud.mumbai
  name     = "dr-app"
  # ... other attributes ...
}
```

Per-alias keys can also represent different scopes within the same region — for example, one key for production workloads and a separate read-only key for an audit account:

```hcl
provider "acecloud" {
  # Primary: full-access key for the workloads team.
  api_key_id           = var.primary_api_key_id
  api_key_secret       = var.primary_api_key_secret
  api_key_service_name = "terraform"
  region               = "ap-south-noi-1"
  project_id           = var.primary_project_id
}

provider "acecloud" {
  alias                = "audit"
  # Audit: read-only key bound to a separate audit project / service identity.
  api_key_id           = var.audit_api_key_id
  api_key_secret       = var.audit_api_key_secret
  api_key_service_name = "terraform-audit"
  region               = "ap-south-noi-1"
  project_id           = var.audit_project_id
}
```

## Pattern 3: one configuration, two projects (same region)

Same idea, applied to projects within one region.

```hcl
provider "acecloud" {
  # ... auth for shared services project ...
  region     = "ap-south-noi-1"
  project_id = var.shared_services_project
}

provider "acecloud" {
  alias      = "apps"
  # ... auth that is authorized for the apps project (may be the same or different key) ...
  region     = "ap-south-noi-1"
  project_id = var.apps_project
}

resource "acecloud_vpc" "shared" {
  # Default provider → shared services project
  name              = "shared-vpc"
  subnet_name       = "shared-sub"
  subnet_cidr       = "10.0.0.0/24"
  subnet_ip_version = 4
}

resource "acecloud_instance" "app" {
  provider = acecloud.apps
  name     = "app-01"
  # ... other attributes ...
}
```

> **Securing per-alias credentials:** when each alias has its own credentials, never hard-code them. Use sensitive Terraform variables sourced from environment variables (`TF_VAR_noida_api_key_secret=…`), a CI/CD secret store, or a remote backend with `terraform_remote_state` pulling encrypted values. Avoid committing more than one credential set to the same `.tfvars` file even in encrypted form — segment per-region credentials into separate stores so a leak in one region doesn't expose the others.

