resource "acecloud_vpc" "main" {
  name              = "lb-vpc"
  description       = "VPC for load balancer example"
  admin_state_up    = true
  subnet_name       = "lb-subnet"
  subnet_cidr       = "10.20.0.0/24"
  subnet_ip_version = 4
}

resource "acecloud_load_balancer" "main" {
  name        = "app-lb"
  description = "Application load balancer for web tier"
  subnet_id   = acecloud_vpc.main.subnet_id
  tags        = ["ALB"]
}
