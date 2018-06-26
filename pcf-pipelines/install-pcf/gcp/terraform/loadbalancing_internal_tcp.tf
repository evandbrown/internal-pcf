//// Internal TCP load balancer that fronts PCF's HAProxy
resource "google_compute_forwarding_rule" "cf-haproxy" {
  name                  = "${var.prefix}-haproxy-lb"
  backend_service       = "${google_compute_region_backend_service.cf-haproxy.self_link}"
  ports                 = ["80", "443"]
  network               = "${google_compute_network.pcf-virt-net.self_link}"
  subnetwork            = "${google_compute_subnetwork.subnet-ops-manager.self_link}"
  load_balancing_scheme = "INTERNAL"
  ip_address            = "${google_compute_address.internal_haproxy.address}"
}

resource "google_compute_instance_group" "cf-haproxy" {
  count       = 3
  name        = "${var.prefix}-haproxy-lb"
  description = "terraform generated pcf instance group that is multi-zone for tcp load balancing to haproxy"

  zone    = "${element(list(var.gcp_zone_1,var.gcp_zone_2,var.gcp_zone_3), count.index)}"
  network = "${google_compute_network.pcf-virt-net.self_link}"
}

resource "google_compute_region_backend_service" "cf-haproxy" {
  name        = "${var.prefix}-haproxy-lb-backend"
  protocol    = "TCP"
  timeout_sec = 10

  backend {
    group = "${google_compute_instance_group.cf-haproxy.0.self_link}"
  }

  backend {
    group = "${google_compute_instance_group.cf-haproxy.1.self_link}"
  }

  backend {
    group = "${google_compute_instance_group.cf-haproxy.2.self_link}"
  }

  health_checks = ["${google_compute_health_check.cf-haproxy.self_link}"]
}

resource "google_compute_health_check" "cf-haproxy" {
  name               = "${var.prefix}-haproxy-lb-healthcheck"
  check_interval_sec = 5
  timeout_sec        = 5

  ssl_health_check {}
}

//// Internal TCP load balancer that fronts the SSH proxy
resource "google_compute_forwarding_rule" "cf-ssh-proxy" {
  name                  = "${var.prefix}-ssh-proxy-lb"
  backend_service       = "${google_compute_region_backend_service.cf-ssh-proxy.self_link}"
  ports                 = ["2222"]
  network               = "${google_compute_network.pcf-virt-net.self_link}"
  subnetwork            = "${google_compute_subnetwork.subnet-ops-manager.self_link}"
  load_balancing_scheme = "INTERNAL"
  ip_address            = "${google_compute_address.internal_ssh_proxy.address}"
}

resource "google_compute_instance_group" "cf-ssh-proxy" {
  count       = 3
  name        = "${var.prefix}-ssh-proxy-lb"
  description = "terraform generated pcf instance group that is multi-zone for tcp load balancing to the SSH proxy"

  zone    = "${element(list(var.gcp_zone_1,var.gcp_zone_2,var.gcp_zone_3), count.index)}"
  network = "${google_compute_network.pcf-virt-net.self_link}"
}

resource "google_compute_region_backend_service" "cf-ssh-proxy" {
  name        = "${var.prefix}-ssh-proxy-lb-backend"
  protocol    = "TCP"
  timeout_sec = 10

  backend {
    group = "${google_compute_instance_group.cf-ssh-proxy.0.self_link}"
  }

  backend {
    group = "${google_compute_instance_group.cf-ssh-proxy.1.self_link}"
  }

  backend {
    group = "${google_compute_instance_group.cf-ssh-proxy.2.self_link}"
  }

  health_checks = ["${google_compute_health_check.cf-ssh-proxy.self_link}"]
}

resource "google_compute_health_check" "cf-ssh-proxy" {
  name               = "${var.prefix}-ssh-proxy-lb-healthcheck"
  check_interval_sec = 5
  timeout_sec        = 5

  tcp_health_check {
    port = 2222
  }
}

//// Internal TCP load balancer that fronts wss-logs
resource "google_compute_forwarding_rule" "cf-wss-logs" {
  name                  = "${var.prefix}-wss-logs-lb"
  backend_service       = "${google_compute_region_backend_service.cf-wss-logs.self_link}"
  ports                 = ["443"]
  network               = "${google_compute_network.pcf-virt-net.self_link}"
  subnetwork            = "${google_compute_subnetwork.subnet-ops-manager.self_link}"
  load_balancing_scheme = "INTERNAL"
  ip_address            = "${google_compute_address.internal_wss_logs.address}"
}

resource "google_compute_instance_group" "cf-wss-logs" {
  count       = 3
  name        = "${var.prefix}-wss-logs-lb"
  description = "terraform generated pcf instance group that is multi-zone for tcp load balancing to the SSH proxy"
  zone        = "${element(list(var.gcp_zone_1,var.gcp_zone_2,var.gcp_zone_3), count.index)}"
  network     = "${google_compute_network.pcf-virt-net.self_link}"
}

resource "google_compute_region_backend_service" "cf-wss-logs" {
  name        = "${var.prefix}-wss-logs-lb-backend"
  protocol    = "TCP"
  timeout_sec = 10

  backend {
    group = "${google_compute_instance_group.cf-wss-logs.0.self_link}"
  }

  backend {
    group = "${google_compute_instance_group.cf-wss-logs.1.self_link}"
  }

  backend {
    group = "${google_compute_instance_group.cf-wss-logs.2.self_link}"
  }

  health_checks = ["${google_compute_health_check.cf-wss-logs.self_link}"]
}

resource "google_compute_health_check" "cf-wss-logs" {
  name               = "${var.prefix}-wss-logs-lb-healthcheck"
  check_interval_sec = 5
  timeout_sec        = 5

  ssl_health_check {}
}
