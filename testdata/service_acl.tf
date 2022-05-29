variable "domain" {
  type = string
}

resource "fastly_service_vcl" "service" {
  name = var.domain

  domain {
    name = var.domain
  }

  backend {
    name              = "httpbin"
    address           = "httpbin.org"
    port              = 443
    use_ssl           = true
    ssl_cert_hostname = "httpbin.org"
    ssl_sni_hostname  = "httpbin.org"
  }

  snippet {
    name     = "block list in recv"
    type     = "recv"
    content  = <<-EOT
        if (fastly.ff.visits_this_service == 0 && req.http.Fastly-Client-IP ~ Generated_by_IP_block_list) {
            error 403 "Forbidden";
        }
        EOT
    priority = 90
  }

  snippet {
    name     = "allow list in recv"
    type     = "recv"
    content  = <<-EOT
        if (fastly.ff.visits_this_service == 0 && req.http.Fastly-Client-IP !~ allow_list) {
            error 403 "Forbidden";
        }
        EOT
    priority = 100
  }

  acl {
    name = "Generated_by_IP_block_list"
  }

  acl {
    name = "allow list"
  }

  force_destroy = true
}

# ACL entries
resource "fastly_service_acl_entries" "generated_by_ip_block_list" {
  for_each = {
    for a in fastly_service_vcl.service.acl : a.name => a if a.name == "Generated_by_IP_block_list"
  }

  acl_id     = each.value.acl_id
  service_id = fastly_service_vcl.service.id

  entry {
    ip      = "192.168.1.0"
    subnet  = "24"
    negated = false
  }
  entry {
    ip      = "192.168.2.0"
    subnet  = "28"
    negated = false
  }
}

resource "fastly_service_acl_entries" "allow_list" {
  for_each = {
    for a in fastly_service_vcl.service.acl : a.name => a if a.name == "allow list"
  }

  acl_id     = each.value.acl_id
  service_id = fastly_service_vcl.service.id

  entry {
    ip      = "192.168.3.0"
    subnet  = "24"
    negated = false
  }
  entry {
    ip      = "192.168.4.0"
    subnet  = "28"
    negated = false
  }
}

output "id" {
  value = fastly_service_vcl.service.id
}
