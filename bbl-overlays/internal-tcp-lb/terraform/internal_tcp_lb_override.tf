variable "ilb_address" {
  type = "string"
}

resource "google_compute_forwarding_rule" "internal-concourse-web" {
  name                  = "internal-concourse-web"
  backend_service       = "${google_compute_region_backend_service.internal_concourse_web.self_link}"
  ports                 = ["80", "443"]
  network               = "${data.google_compute_network.bbl-network.self_link}"
  subnetwork            = "${data.google_compute_subnetwork.bbl-subnet.self_link}"
  load_balancing_scheme = "INTERNAL"
  ip_address            = "${var.ilb_address}"
}

resource "google_compute_region_backend_service" "internal_concourse_web" {
  name        = "internal-concourse-web"
  protocol    = "TCP"
  timeout_sec = 10

  backend {
    group = "${google_compute_instance_group.internal_concourse_web.self_link}"
  }

  health_checks = ["${google_compute_health_check.internal_concourse_web.self_link}"]
}

resource "google_compute_health_check" "internal_concourse_web" {
  name               = "internal-concourse-web"
  check_interval_sec = 5
  timeout_sec        = 5

  ssl_health_check {}
}

resource "google_compute_instance_group" "internal_concourse_web" {
  name = "internal-concourse-web"

  zone    = "${var.zone}"
  network = "${data.google_compute_network.bbl-network.self_link}"
}

resource "google_compute_firewall" "internal_concourse_web" {
  name    = "internal-concourse-web"
  network = "${data.google_compute_network.bbl-network.self_link}"

  allow {
    protocol = "tcp"
  }

  source_ranges = ["0.0.0.0/0"]
  target_tags   = ["concourse-web"]
}

output "ilb_url" {
  value = "https://${var.ilb_address}"
}

output "ilb_backend_service" {
  value = "${google_compute_region_backend_service.internal_concourse_web.name}"
}
