---
page_title: "Web tier behind a load balancer"
subcategory: "Load Balancer"
description: |-
  Distribute traffic across two compute instances using a load balancer, listener, pool, members, and HTTP health monitor.
---

# Web tier behind a load balancer

This guide stands up a two-instance web tier behind a Layer 7 load balancer with an HTTP health check. Traffic flow:

```
internet → floating IP → load balancer → listener (port 80)
                                             ↓
                                         pool (round-robin)
                                          ↓        ↓
                                       web-01    web-02
                                       :8080     :8080
```

The two instances run identical user-data scripts that bring up a tiny HTTP server on port 8080. The load balancer listens on port 80 and proxies to the pool. The health monitor probes `/health` on each member and removes any that stop responding.

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
data "acecloud_vpcs" "all" {}

locals {
  flavor_id           = [for f in data.acecloud_flavors.all.flavors : f.id if f.name == "C4i.medium"][0]
  image_id            = [for i in data.acecloud_images.all.images : i.id if i.name == "Ubuntu-24.04-LTS"][0]
  external_network_id = [for v in data.acecloud_vpcs.all.vpcs : v.id if v.router_external && v.shared][0]
}

# Network
resource "acecloud_vpc" "main" {
  name              = "web-tier-vpc"
  admin_state_up    = true
  subnet_name       = "web-tier-subnet"
  subnet_cidr       = "10.10.0.0/24"
  subnet_ip_version = 4
}

resource "acecloud_router" "main" {
  name                        = "web-tier-router"
  admin_state_up              = true
  external_gateway_network_id = local.external_network_id
}

resource "acecloud_router_interface" "main" {
  router_id = acecloud_router.main.id
  subnet_id = acecloud_vpc.main.subnet_id
}

# Security — LB ingress on port 80, allow LB to reach members on 8080
resource "acecloud_security_group" "web" {
  name        = "web-tier-sg"
  description = "Allow HTTP from LB, SSH for ops"

  # SSH for operators
  rules {
    direction        = "ingress"
    protocol         = "tcp"
    port_range_min   = 22
    port_range_max   = 22
    remote_ip_prefix = "0.0.0.0/0"
    ethertype        = "IPv4"
  }

  # HTTP from the LB subnet (same VPC)
  rules {
    direction        = "ingress"
    protocol         = "tcp"
    port_range_min   = 8080
    port_range_max   = 8080
    remote_ip_prefix = "10.10.0.0/24"
    ethertype        = "IPv4"
  }
}

# Two web servers. Provision your application via a baked image or
# configuration-management tool of your choice. Each instance must run
# the web service on port 8080 (the port the load balancer pool targets).
resource "acecloud_instance" "web" {
  count       = 2
  name        = "web-0${count.index + 1}"
  flavor_id   = local.flavor_id
  boot_uuid   = local.image_id
  source_type = "image"

  delete_on_termination = true
  network_ids           = [acecloud_vpc.main.id]
  security_group_ids    = [acecloud_security_group.web.id]
  billing_type          = "monthly"

  volumes {
    size         = 20
    boot         = true
    volume_type  = "ssd"
    billing_type = "hourly"
  }
}

# Load balancer
resource "acecloud_load_balancer" "main" {
  name        = "web-tier-lb"
  description = "Public LB fronting the web tier"
  subnet_id   = acecloud_vpc.main.subnet_id
  tags        = ["ALB"]
}

resource "acecloud_lb_listener" "http" {
  name            = "http"
  loadbalancer_id = acecloud_load_balancer.main.id
  protocol        = "HTTP"
  protocol_port   = 80
}

resource "acecloud_lb_pool" "http" {
  name         = "web-tier-pool"
  listener_id  = acecloud_lb_listener.http.id
  protocol     = "HTTP"
  lb_algorithm = "ROUND_ROBIN"
}

resource "acecloud_lb_pool_member" "web" {
  count         = 2
  pool_id       = acecloud_lb_pool.http.id
  address       = acecloud_instance.web[count.index].private_ip
  protocol_port = 8080
  name          = "web-0${count.index + 1}"
}

resource "acecloud_lb_health_monitor" "http" {
  name        = "http-health"
  pool_id     = acecloud_lb_pool.http.id
  type        = "HTTP"
  url_path    = "/health"
  http_method = "GET"
  delay       = 10
  timeout     = 5
  max_retries = 3
}

# Public IP for the LB
resource "acecloud_floating_ip" "lb" {
  floating_network_id = local.external_network_id
}

# Note: LB FIPs attach to the LB's VIP port (provider exposes this via the
# load balancer's vip_port_id if available; otherwise associate manually
# via the console).

output "lb_address" {
  value = acecloud_floating_ip.lb.floating_ip_address
}

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

