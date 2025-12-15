terraform {
  required_providers {
    acecloud = {
      source  = "acecloud/acecloud"
      version = "0.1.0" 
    }
  }
}


provider "acecloud" {
  api_endpoint = "http://localhost:3001"
  api_key      = "7c57b3ec-58a6-4428-b2c6-e9a6ce330682.f5e83544566d861f879b03e1f8b6ca894cc50c9c8636a6e7079fa7bc0d9efddf" 
  region       = "ap-south-mum-1"
  project_id   = "251b42b560eb415db84cdc285fb125f4"
}
resource "acecloud_vm" "example" {
  name                  = "instance2"
  flavor                = "71eddda0-7b6b-4873-a5f5-ada8a2031059"
  boot_uuid             = "ce0d036e-b8f7-411f-8be1-e03e47ce5fcd"
  delete_on_termination = true
  key = "k8s-sujal"
  network        = ["c92f8cb2-cf4d-4614-be27-e9ad192fc3e8"]
  security_group = ["36bbdcb7-facc-4346-9ce4-c15c33977c72"]

  source_type       = "image"
  availability_zone = "nova"
  billing_type      = "hourly"


  volumes {
    boot         = true
    volume_type  = "NVMe based High IOPS Storage"
    size         = 20
    billing_type = "hourly"
  }

  vm_count = 1
}