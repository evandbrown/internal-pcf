resource "google_compute_address" "jumpbox-ip" {
  name         = "${var.env_id}-jumpbox-ip"
  address_type = "INTERNAL"
  subnetwork   = "${data.google_compute_subnetwork.bbl-subnet.self_link}"
}

output "jumpbox_url" {
  value = "${google_compute_address.jumpbox-ip.address}:22"
}

output "director_address" {
  value = "https://${google_compute_address.jumpbox-ip.address}:25555"
}

output "jumpbox__internal_ip" {
  value = "${google_compute_address.jumpbox-ip.address}"
}
