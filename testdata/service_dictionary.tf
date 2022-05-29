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

  dictionary {
    name = "some config"
  }

  dictionary {
    name = "some other config"
  }

  force_destroy = true
}

resource "fastly_service_dictionary_items" "some_config" {
  for_each = {
    for d in fastly_service_vcl.service.dictionary : d.name => d if d.name == "some config"
  }

  service_id    = fastly_service_vcl.service.id
  dictionary_id = each.value.dictionary_id
  items = {
    "config 1" = "foo"
    "config 2" = "bar"
    "config 3" = "baz"
  }
}

resource "fastly_service_dictionary_items" "some_other_config" {
  for_each = {
    for d in fastly_service_vcl.service.dictionary : d.name => d if d.name == "some other config"
  }

  service_id    = fastly_service_vcl.service.id
  dictionary_id = each.value.dictionary_id
  items = {
    "config 1" = "foo"
    "config 2" = "bar"
    "config 3" = "baz"
  }
}

output "id" {
  value = fastly_service_vcl.service.id
}
