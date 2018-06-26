resource "google_compute_ssl_certificate" "ssl-cert" {
  name_prefix = "${var.prefix}-lb-cert-"
  description = "user provided ssl private key / ssl certificate pair"
  certificate = "${var.pcf_ert_ssl_cert}"
  private_key = "${var.pcf_ert_ssl_key}"

  lifecycle {
    create_before_destroy = true
  }
}
