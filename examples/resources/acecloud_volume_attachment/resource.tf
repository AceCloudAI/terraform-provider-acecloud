# Attach a data volume to an instance with independent lifecycles.
# Use acecloud_volume_attachment (instead of an inline `volumes` block on the
# instance) when:
#   - the volume should outlive the instance, or
#   - you're attaching a pre-existing volume not created by this config.

data "acecloud_flavors" "all" {}
data "acecloud_images" "all" {}

resource "acecloud_vpc" "main" {
  name              = "data-vpc"
  description       = "VPC for instance + data volume example"
  admin_state_up    = true
  subnet_name       = "data-subnet"
  subnet_cidr       = "10.10.0.0/24"
  subnet_ip_version = 4
}

resource "acecloud_security_group" "default" {
  name        = "data-sg"
  description = "Allow SSH"

  rules {
    direction        = "ingress"
    protocol         = "tcp"
    port_range_min   = 22
    port_range_max   = 22
    remote_ip_prefix = "0.0.0.0/0"
    ethertype        = "IPv4"
  }
}

resource "acecloud_instance" "web" {
  name        = "web-01"
  flavor_id   = [for f in data.acecloud_flavors.all.flavors : f.id if f.name == "C4i.medium"][0]
  boot_uuid   = [for i in data.acecloud_images.all.images : i.id if i.name == "Ubuntu-24.04-LTS"][0]
  source_type = "image"

  delete_on_termination = true
  network_ids           = [acecloud_vpc.main.id]
  security_group_ids    = [acecloud_security_group.default.id]
  billing_type          = "hourly"

  volumes {
    size         = 20
    boot         = true
    volume_type  = "NVMe based High IOPS Storage"
    billing_type = "hourly"
  }
}

# Data volume created independently of the instance.
resource "acecloud_volume" "data" {
  name         = "shared-data"
  description  = "Persistent data volume"
  size         = 200
  volume_type  = "NVMe based High IOPS Storage"
  billing_type = "hourly"
}

# Attach. delete_on_termination = false keeps the volume around when
# the instance is destroyed.
resource "acecloud_volume_attachment" "data" {
  instance_id           = acecloud_instance.web.id
  volume_id             = acecloud_volume.data.id
  delete_on_termination = false
}
