variable "domain" {
  type = string
}

data "fastly_package_hash" "service" {
  filename = "package.tar.gz"
}

resource "fastly_service_compute" "service" {
  name = var.domain

  domain {
    name = var.domain
  }

  backend {
    name              = "Host 1"
    address           = "httpbin.org"
    port              = 443
    use_ssl           = true
    ssl_cert_hostname = "httpbin.org"
    ssl_sni_hostname  = "httpbin.org"
  }

  package {
    source_code_hash = data.fastly_package_hash.service.hash
    filename         = data.fastly_package_hash.service.filename
  }

  resource_link {
    name        = fastly_configstore.config_store.name
    resource_id = fastly_configstore.config_store.id
  }
  resource_link {
    name        = fastly_secretstore.secret_store.name
    resource_id = fastly_secretstore.secret_store.id
  }
  resource_link {
    name        = fastly_kvstore.kv_store.name
    resource_id = fastly_kvstore.kv_store.id
  }

  comment = ""
}

resource "fastly_configstore" "config_store" {
  name = "config_store"
}

resource "fastly_configstore_entries" "config_store" {
  entries = {
    "item0" = "0"
    "item1" = "1"
    "item2" = "2"
    "item3" = "3"
    "item4" = "4"
    "item5" = "5"
    "item6" = "6"
    "item7" = "7"
    "item8" = "8"
    "item9" = "9"
  }
  store_id       = fastly_configstore.config_store.id
  manage_entries = true
}

resource "fastly_secretstore" "secret_store" {
  name = "secret_store"
}

resource "fastly_kvstore" "kv_store" {
  name = "kv_store"
}

output "id" {
  value = fastly_service_compute.service.id
}
