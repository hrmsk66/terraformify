variable "domain" {
  type = string
}

resource "fastly_service_compute" "service" {
  name               = var.domain

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

  backend {
    name              = "Host 2"
    address           = "httpbin.org"
    port              = 443
    use_ssl           = true
    ssl_cert_hostname = "httpbin.org"
    ssl_sni_hostname  = "httpbin.org"
  }

  dictionary {
    name          = "dict1"
  }
  dictionary {
    name          = "dict2"
  }

  logging_s3 {
    bucket_name      = "s3_bucket"
    domain           = "s3.amazonaws.com"
    gzip_level       = 0
    message_type     = "blank"
    name             = "log to s3"
    path             = "/"
    period           = 3600
    redundancy       = "standard"
    s3_access_key    = "XXXXXXXX123456789123"
    s3_secret_key    = "XXXXXXXXX1234567891234567891234567891234"
    timestamp_format = "%Y-%m-%dT%H:%M:%S.000"
  }

  package {
    source_code_hash = filesha512("package.tar.gz")
    filename         = "package.tar.gz"
  }

  product_enablement {
    fanout     = false
    websockets = false
  }
  comment = ""
}

resource "fastly_service_dictionary_items" "dict1" {
  dictionary_id = each.value.dictionary_id
  items = {
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
  service_id = fastly_service_compute.service.id

  for_each = {
    for d in fastly_service_compute.service.dictionary : d.name => d if d.name == "dict1"
  }
}

resource "fastly_service_dictionary_items" "dict2" {
  dictionary_id = each.value.dictionary_id
  items         = {}
  service_id    = fastly_service_compute.service.id

  for_each = {
    for d in fastly_service_compute.service.dictionary : d.name => d if d.name == "dict2"
  }
}

output "id" {
  value = fastly_service_compute.service.id
}
