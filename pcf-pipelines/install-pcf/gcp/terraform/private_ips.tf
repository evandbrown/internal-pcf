// Private static IP address for ILB that fronts HAProxy
resource "google_compute_address" "internal_haproxy" {
  name         = "${var.prefix}-haproxy"
  subnetwork   = "${google_compute_subnetwork.subnet-ops-manager.self_link}"
  address_type = "INTERNAL"
}

// Private static IP address for ILB that fronts ssh-proxy
resource "google_compute_address" "internal_ssh_proxy" {
  name         = "${var.prefix}-ssh-haproxy"
  subnetwork   = "${google_compute_subnetwork.subnet-ops-manager.self_link}"
  address_type = "INTERNAL"
}

// Private static IP address for ILB that fronts loggregator/doppler
resource "google_compute_address" "internal_wss_logs" {
  name         = "${var.prefix}-wss-logs"
  subnetwork   = "${google_compute_subnetwork.subnet-ops-manager.self_link}"
  address_type = "INTERNAL"
}
