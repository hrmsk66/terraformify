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

  dynamicsnippet {
    name     = "My Dynamic Snippet One"
    priority = 100
    type     = "recv"
  }

  dynamicsnippet {
    name     = "My Dynamic Snippet Two"
    priority = 110
    type     = "recv"
  }

  force_destroy = true
}

resource "fastly_service_dynamic_snippet_content" "my_dynamic_snippet_one" {
  for_each = {
    for d in fastly_service_vcl.service.dynamicsnippet : d.name => d if d.name == "My Dynamic Snippet One"
  }
  service_id = fastly_service_vcl.service.id
  snippet_id = each.value.snippet_id
  content    = <<-EOT
    if ( req.url ) {
        set req.http.my-snippet-test-header-one = "true";
    }
    EOT
}

resource "fastly_service_dynamic_snippet_content" "my_dynamic_snippet_two" {
  for_each = {
    for d in fastly_service_vcl.service.dynamicsnippet : d.name => d if d.name == "My Dynamic Snippet Two"
  }
  service_id = fastly_service_vcl.service.id
  snippet_id = each.value.snippet_id
  content    = <<-EOT
    if ( req.url ) {
        set req.http.my-snippet-test-header-two = "true";
    }
    EOT
}

output "id" {
  value = fastly_service_vcl.service.id
}
