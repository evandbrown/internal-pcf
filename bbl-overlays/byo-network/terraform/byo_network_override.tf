variable "subnet_cidr" {
  type = "string"
}

variable "cidr_newbits" {
  type = "string"
}

variable "existing-bbl-network" {
  type = "string"
}

variable "existing-bbl-subnet" {
  type = "string"
}

variable "existing-host-project" {
  type = "string"
}

resource "google_compute_network" "bbl-network" {
  count = 0
}

resource "google_compute_subnetwork" "bbl-subnet" {
  count = 0
}

data "google_compute_network" "bbl-network" {
  name    = "${var.existing-bbl-network}"
  project = "${var.existing-host-project}"
}

data "google_compute_subnetwork" "bbl-subnet" {
  name    = "${var.existing-bbl-subnet}"
  project = "${var.existing-host-project}"
}

resource "google_compute_firewall" "external" {
  network = "${data.google_compute_network.bbl-network.name}"
  project = "${var.existing-host-project}"
}

resource "google_compute_firewall" "bosh-director" {
  network = "${data.google_compute_network.bbl-network.name}"
  project = "${var.existing-host-project}"
}

resource "google_compute_firewall" "internal-to-director" {
  network = "${data.google_compute_network.bbl-network.name}"
  project = "${var.existing-host-project}"
}

resource "google_compute_firewall" "jumpbox-to-all" {
  network = "${data.google_compute_network.bbl-network.name}"
  project = "${var.existing-host-project}"
}

resource "google_compute_firewall" "internal" {
  network = "${data.google_compute_network.bbl-network.name}"
  project = "${var.existing-host-project}"
}

resource "google_compute_firewall" "bosh-open" {
  network       = "${data.google_compute_network.bbl-network.name}"
  source_ranges = ["${google_compute_address.jumpbox-ip.address}/32"]
  project       = "${var.existing-host-project}"

  allow {
    ports    = ["22", "6868", "8443", "8844", "25555"]
    protocol = "tcp"
  }

  target_tags = ["${var.env_id}-bosh-director"]
}

output "network" {
  value = "${data.google_compute_network.bbl-network.name}"
}

output "subnetwork" {
  value = "${data.google_compute_subnetwork.bbl-subnet.name}"
}

output "network_host_project" {
  value = "${var.existing-host-project}"
}
