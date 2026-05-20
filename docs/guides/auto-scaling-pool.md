---
page_title: "Auto-scaling pool"
description: |-
  Create an auto-scaling pool with a launch template, CPU-driven scaling rules, and an integrated load balancer.
---

# Auto-scaling pool

This guide creates a CPU-driven auto-scaling pool fronted by a load balancer. The deployment provisions instances on demand: when average CPU across the pool exceeds the high threshold, the controller adds nodes (up to `max_capacity`); when CPU drops below the low threshold, it removes nodes (down to `desired_capacity`).

Components:

- `acecloud_auto_scaling_template` — the "blueprint" describing what each scaled instance looks like (flavor, image, network, security groups, user-data)
- `acecloud_auto_scaling_deployment` — the controller: capacity bounds, scaling parameters, optional embedded load balancer

The load balancer can be created and managed by the auto-scaling deployment itself (set `is_integrated_with_lb = true` and provide `lb_data`), or referenced as an existing LB. This guide uses the integrated variant — simpler and the common case.

## Complete example

```hcl
terraform {
  required_providers {
    acecloud = {
      source  = "AceCloudAI/acecloud"
      version = "~> 0.1"
    }
  }
}

provider "acecloud" {
  api_key_id           = var.acecloud_api_key_id
  api_key_secret       = var.acecloud_api_key_secret
  api_key_service_name = "terraform"
  region               = "ap-south-noi-1"
  project_id           = var.acecloud_project_id
}

data "acecloud_flavors" "all" {}
data "acecloud_images" "all" {}

locals {
  flavor_id = [for f in data.acecloud_flavors.all.flavors : f.id if f.name == "C4i.medium"][0]
  image_id  = [for i in data.acecloud_images.all.images : i.id if i.name == "Ubuntu-24.04-LTS"][0]
}

# Network
resource "acecloud_vpc" "main" {
  name              = "asg-vpc"
  admin_state_up    = true
  subnet_name       = "asg-subnet"
  subnet_cidr       = "10.30.0.0/24"
  subnet_ip_version = 4
}

# Security — allow HTTP from the LB subnet, SSH from anywhere for ops
resource "acecloud_security_group" "web" {
  name        = "asg-web-sg"
  description = "Auto-scaling pool members"

  rules {
    direction        = "ingress"
    protocol         = "tcp"
    port_range_min   = 22
    port_range_max   = 22
    remote_ip_prefix = "0.0.0.0/0"
    ethertype        = "IPv4"
  }

  rules {
    direction        = "ingress"
    protocol         = "tcp"
    port_range_min   = 8080
    port_range_max   = 8080
    remote_ip_prefix = "10.30.0.0/24"
    ethertype        = "IPv4"
  }
}

# SSH key for ops access to scaled instances
resource "acecloud_key_pair" "ops" {
  name = "asg-ops"
}

# Launch template — describes what each scaled instance looks like
resource "acecloud_auto_scaling_template" "web" {
  name                   = "web-tier-template"
  type                   = "linux"
  description            = "Web tier launch template"
  volume_type            = "NVMe based High IOPS Storage"
  volume_size            = 40
  vol_del_on_termination = true
  flavor_id              = local.flavor_id
  image_id               = local.image_id
  is_instance_snapshot   = false
  key_name               = acecloud_key_pair.ops.name
  network_id             = acecloud_vpc.main.id
  subnet_id              = acecloud_vpc.main.subnet_id
  security_groups        = [acecloud_security_group.web.id]

  # user_data is cloud-init / bootstrap data, base64-encoded. Runs once at
  # provision time on each scaled instance. Keep it idempotent.
  user_data = base64encode(<<-EOT
    #!/bin/bash
    apt-get update && apt-get install -y nginx
    systemctl enable --now nginx
  EOT
  )
}

# Auto-scaling deployment — the controller + integrated load balancer
resource "acecloud_auto_scaling_deployment" "web" {
  name                  = "web-tier-asg"
  description           = "CPU-driven web tier"
  template_id           = acecloud_auto_scaling_template.web.id
  desired_capacity      = 2
  max_capacity          = 5
  nodes_scale_count     = 1
  scaling_parameter     = "cpu"
  min_threshold         = 30  # scale in when avg CPU < 30%
  max_threshold         = 70  # scale out when avg CPU > 70%
  cool_down_time        = 180 # seconds between consecutive scale actions
  user_email            = ["ops@example.com"]
  is_integrated_with_lb = true

  lb_data {
    lb_name          = "web-tier-asg-lb"
    tags             = ["ALB"]
    assign_public_ip = true
    is_existing_lb   = false

    listener {
      listener_name          = "http"
      listener_protocol      = "HTTP"
      listener_protocol_port = 80
    }

    pool {
      pool_protocol      = "HTTP"
      pool_protocol_port = 80
      lb_algorithm       = "ROUND_ROBIN"
    }

    health_monitor {
      monitor_protocol    = "HTTP"
      monitor_url_path    = "/"
      monitor_http_method = "GET"
    }
  }

  depends_on = [acecloud_auto_scaling_template.web]
}

output "deployment_id" { value = acecloud_auto_scaling_deployment.web.id }

variable "acecloud_api_key_id" {
  type      = string
  sensitive = true
}
variable "acecloud_api_key_secret" {
  type      = string
  sensitive = true
}
variable "acecloud_project_id" {
  type = string
}
```

## How scaling decisions are made

| Setting | Meaning |
|---|---|
| `desired_capacity` | The pool starts here and never goes below this number |
| `max_capacity` | Hard upper bound; the controller will not provision past this |
| `nodes_scale_count` | How many nodes to add (or remove) in a single scale action |
| `scaling_parameter` | Metric to monitor (`cpu` is the common case) |
| `max_threshold` | When the average across the pool exceeds this %, scale **out** by `nodes_scale_count` |
| `min_threshold` | When the average drops below this %, scale **in** by `nodes_scale_count` (never below `desired_capacity`) |
| `cool_down_time` | Minimum seconds between two consecutive scale actions, prevents oscillation |

