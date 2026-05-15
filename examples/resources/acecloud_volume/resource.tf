resource "acecloud_volume" "data" {
  name        = "data-volume-01"
  size        = 100
  volume_type = "NVMe based High IOPS Storage"
  description = "Data volume for application storage"

  metadata = {
    environment = "dev"
    managed_by  = "terraform"
  }
}

output "volume_id" {
  value = acecloud_volume.data.id
}
