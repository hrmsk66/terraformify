# fastly_service_acl_entries.allow_list:
resource "fastly_service_acl_entries" "allow_list" {
  acl_id     = each.value.acl_id
  service_id = fastly_service_vcl.service.id

  entry {
    comment = "ACL Entry 1"
    ip      = "192.168.0.0"
    negated = false
    subnet  = "24"
  }
  entry {
    comment = "ACL Entry 2"
    ip      = "192.168.1.0"
    negated = false
    subnet  = "24"
  }
  for_each = {
    for d in fastly_service_vcl.service.acl : d.name => d if d.name == "allow_list"
  }
}

# fastly_service_acl_entries.generated_by_ip_block_list:
resource "fastly_service_acl_entries" "generated_by_ip_block_list" {
  acl_id     = each.value.acl_id
  service_id = fastly_service_vcl.service.id

  entry {
    ip      = "192.168.3.0"
    negated = false
  }
  entry {
    ip      = "192.168.4.0"
    negated = false
  }
  for_each = {
    for d in fastly_service_vcl.service.acl : d.name => d if d.name == "Generated_by_IP_block_list"
  }
}

# fastly_service_dictionary_items.config_table:
resource "fastly_service_dictionary_items" "config_table" {
  dictionary_id = each.value.dictionary_id
  items = {
    "maintenance" = "true"
    "otherconfig" = "false"
  }
  service_id = fastly_service_vcl.service.id
  for_each = {
    for d in fastly_service_vcl.service.dictionary : d.name => d if d.name == "config_table"
  }
}

# fastly_service_dictionary_items.redirect_table:
resource "fastly_service_dictionary_items" "redirect_table" {
  dictionary_id = each.value.dictionary_id
  items = {
    "/bar" = "/image"
    "/baz" = "/image"
    "/foo" = "/image"
  }
  service_id = fastly_service_vcl.service.id
  for_each = {
    for d in fastly_service_vcl.service.dictionary : d.name => d if d.name == "redirect_table"
  }
}

# fastly_service_dynamic_snippet_content.my_dynamic_snippet_one:
resource "fastly_service_dynamic_snippet_content" "my_dynamic_snippet_one" {
  content    = file("./vcl/dsnippet_my_dynamic_snippet_one.vcl")
  service_id = fastly_service_vcl.service.id
  snippet_id = each.value.snippet_id
  for_each = {
    for d in fastly_service_vcl.service.dynamicsnippet : d.name => d if d.name == "My Dynamic Snippet One"
  }
}

# fastly_service_dynamic_snippet_content.my_dynamic_snippet_two:
resource "fastly_service_dynamic_snippet_content" "my_dynamic_snippet_two" {
  content    = file("./vcl/dsnippet_my_dynamic_snippet_two.vcl")
  service_id = fastly_service_vcl.service.id
  snippet_id = each.value.snippet_id
  for_each = {
    for d in fastly_service_vcl.service.dynamicsnippet : d.name => d if d.name == "My Dynamic Snippet Two"
  }
}

# fastly_service_vcl.service:
resource "fastly_service_vcl" "service" {
  comment            = "terraformify test service"
  default_ttl        = 3600
  name               = "terraformify.hkakehas.tokyo"
  stale_if_error     = true
  stale_if_error_ttl = 43200

  acl {
    force_destroy = false
    name          = "allow_list"
  }
  acl {
    force_destroy = false
    name          = "Generated_by_IP_block_list"
  }

  backend {
    address               = "apps.fastly.com"
    auto_loadbalance      = false
    between_bytes_timeout = 10000
    connect_timeout       = 1000
    error_threshold       = 0
    first_byte_timeout    = 15000
    max_conn              = 200
    name                  = "apps"
    port                  = 80
    ssl_check_cert        = true
    use_ssl               = false
    weight                = 9
  }
  backend {
    address               = "developer.fastly.com"
    auto_loadbalance      = false
    between_bytes_timeout = 10000
    connect_timeout       = 1000
    error_threshold       = 0
    first_byte_timeout    = 15000
    max_conn              = 200
    name                  = "developer_updated"
    port                  = 80
    ssl_check_cert        = true
    use_ssl               = false
    weight                = 100
  }
  backend {
    address               = "httpbin.org"
    auto_loadbalance      = false
    between_bytes_timeout = 10000
    connect_timeout       = 1000
    error_threshold       = 0
    first_byte_timeout    = 15000
    max_conn              = 200
    name                  = "httpbin"
    port                  = 443
    ssl_cert_hostname     = "httpbin.org"
    ssl_check_cert        = true
    ssl_sni_hostname      = "httpbin.org"
    use_ssl               = true
    weight                = 100
  }
  backend {
    address               = "www.fastly.com"
    auto_loadbalance      = false
    between_bytes_timeout = 10000
    connect_timeout       = 1000
    error_threshold       = 0
    first_byte_timeout    = 15000
    max_conn              = 200
    name                  = "www"
    port                  = 80
    ssl_check_cert        = true
    use_ssl               = false
    weight                = 100
  }
  backend {
    address               = "www.fastlydemo.net"
    auto_loadbalance      = false
    between_bytes_timeout = 10000
    connect_timeout       = 1000
    error_threshold       = 0
    first_byte_timeout    = 15000
    max_conn              = 200
    name                  = "demo"
    port                  = 80
    ssl_check_cert        = true
    use_ssl               = false
    weight                = 100
  }

  condition {
    name      = "Generated by IP block list"
    priority  = 0
    statement = "client.ip ~ Generated_by_IP_block_list"
    type      = "REQUEST"
  }
  condition {
    name      = "Generated by synthetic response for 404 page"
    priority  = 0
    statement = "beresp.status == 404"
    type      = "CACHE"
  }
  condition {
    name      = "Generated by synthetic response for 503 page"
    priority  = 0
    statement = "beresp.status == 503"
    type      = "CACHE"
  }
  condition {
    name      = "Generated by synthetic response for robots.txt"
    priority  = 0
    statement = "req.url.path == \"/robots.txt\""
    type      = "REQUEST"
  }
  condition {
    name      = "WAF_Prefetch"
    priority  = 10
    statement = "req.backend.is_origin && !req.http.rqpass"
    type      = "PREFETCH"
  }
  condition {
    name      = "false"
    priority  = 10
    statement = "!req.url"
    type      = "REQUEST"
  }
  condition {
    name      = "waf-soc-logging"
    priority  = 10
    statement = "waf.executed"
    type      = "RESPONSE"
  }

  dictionary {
    force_destroy = false
    name          = "config_table"
    write_only    = false
  }
  dictionary {
    force_destroy = false
    name          = "redirect_table"
    write_only    = false
  }

  director {
    backends = [
      "apps",
    ]
    name    = "director_apps"
    quorum  = 75
    retries = 5
    type    = 3
  }
  director {
    backends = [
      "demo",
      "www",
    ]
    name    = "director_www_demo"
    quorum  = 75
    retries = 5
    type    = 3
  }
  director {
    backends = [
      "developer_updated",
    ]
    name    = "director_developer"
    quorum  = 30
    retries = 10
    type    = 4
  }

  domain {
    name = "hkakehas.tokyo"
  }
  domain {
    name = "terraformify.hkakehas.tokyo"
  }

  dynamicsnippet {
    name     = "My Dynamic Snippet One"
    priority = 110
    type     = "recv"
  }
  dynamicsnippet {
    name     = "My Dynamic Snippet Two"
    priority = 110
    type     = "recv"
  }

  gzip {
    content_types = [
      "text/html",
      "application/x-javascript",
      "text/css",
      "application/javascript",
      "text/javascript",
      "application/json",
      "application/vnd.ms-fontobject",
      "application/x-font-opentype",
      "application/x-font-truetype",
      "application/x-font-ttf",
      "application/xml",
      "font/eot",
      "font/opentype",
      "font/otf",
      "image/svg+xml",
      "image/vnd.microsoft.icon",
      "text/plain",
      "text/xml",
    ]
    extensions = [
      "css",
      "js",
      "html",
      "eot",
      "ico",
      "otf",
      "ttf",
      "json",
      "svg",
    ]
    name = "Generated by default gzip policy"
  }

  header {
    action        = "set"
    destination   = "http.Strict-Transport-Security"
    ignore_if_set = false
    name          = "Generated by force TLS and enable HSTS"
    priority      = 100
    source        = "\"max-age=300\""
    type          = "response"
  }

  healthcheck {
    check_interval    = 60000
    expected_response = 200
    host              = "httpbin.org"
    http_version      = "1.1"
    initial           = 1
    method            = "HEAD"
    name              = "my healthcheck"
    path              = "/200"
    threshold         = 1
    timeout           = 5000
    window            = 2
  }

  logging_papertrail {
    address            = "xxx.papertrail.com"
    format             = file("./logformat/weblogs.json")
    format_version     = 2
    name               = "weblogs"
    port               = 12345
    response_condition = "waf-soc-logging"
  }
  logging_papertrail {
    address        = "xxx.papertrail.com"
    format         = file("./logformat/waflogs.json")
    format_version = 2
    name           = "waflogs"
    placement      = "waf_debug"
    port           = 12345
  }

  logging_s3 {
    bucket_name      = "my_s3_bucket"
    domain           = "s3.amazonaws.com"
    format           = file("./logformat/my_s3_endpoint.txt")
    format_version   = 2
    gzip_level       = 0
    message_type     = "blank"
    name             = "my S3 endpoint"
    path             = "/"
    period           = 3600
    redundancy       = "standard"
    s3_access_key    = "XXXXXXXX123456789123"
    s3_secret_key    = "XXXXXXXXX1234567891234567891234567891234"
    timestamp_format = "%Y-%m-%dT%H:%M:%S.000"
  }

  request_setting {
    bypass_busy_wait = false
    force_miss       = false
    force_ssl        = true
    geo_headers      = false
    max_stale_age    = 0
    name             = "Generated by force TLS and enable HSTS"
    timer_support    = false
    xff              = ""
  }

  response_object {
    content_type      = "text/html"
    name              = "Generated by IP block list"
    request_condition = "Generated by IP block list"
    response          = "Forbidden"
    status            = 403
    content           = file("./content/generated_by_ip_block_list.txt")
  }
  response_object {
    content           = file("./content/generated_by_synthetic_response_for_robots_txt.txt")
    content_type      = "text/plain"
    name              = "Generated by synthetic response for robots.txt"
    request_condition = "Generated by synthetic response for robots.txt"
    response          = "OK"
    status            = 200
  }
  response_object {
    content           = file("./content/waf_response.txt")
    content_type      = "application/json"
    name              = "WAF_Response"
    request_condition = "false"
    response          = "Forbidden"
    status            = 403
  }
  response_object {
    cache_condition = "Generated by synthetic response for 404 page"
    content         = file("./content/generated_by_synthetic_response_for_404_page.txt")
    content_type    = "text/html"
    name            = "Generated by synthetic response for 404 page"
    response        = "Not Found"
    status          = 404
  }
  response_object {
    cache_condition = "Generated by synthetic response for 503 page"
    content         = file("./content/generated_by_synthetic_response_for_503_page.txt")
    content_type    = "text/html"
    name            = "Generated by synthetic response for 503 page"
    response        = "Service Unavailable"
    status          = 503
  }

  snippet {
    content  = file("./vcl/snippet_fastly_csi_init.vcl")
    name     = "fastly_csi_init"
    priority = 5
    type     = "recv"
  }
  snippet {
    content  = file("./vcl/snippet_error_redirects.vcl")
    name     = "error_redirects"
    priority = 100
    type     = "error"
  }
  snippet {
    content  = file("./vcl/snippet_recv_redirects.vcl")
    name     = "recv_redirects"
    priority = 100
    type     = "recv"
  }
  snippet {
    content  = file("./vcl/snippet_recv_allow_list.vcl")
    name     = "recv_allow_list"
    priority = 90
    type     = "recv"
  }
  snippet {
    content  = file("./vcl/snippet_fastly_waf_snippet.vcl")
    name     = "Fastly_WAF_Snippet"
    priority = 10
    type     = "recv"
  }

  vcl {
    content = file("./vcl/main.vcl")
    main    = true
    name    = "main"
  }
  vcl {
    content = file("./vcl/config_check.vcl")
    main    = false
    name    = "config_check"
  }

  waf {
    disabled           = false
    prefetch_condition = "WAF_Prefetch"
    response_object    = "WAF_Response"
  }
}

# fastly_service_waf_configuration.waf:
resource "fastly_service_waf_configuration" "waf" {
  allowed_http_versions                = "HTTP/1.0 HTTP/1.1 HTTP/2 HTTP/3"
  allowed_methods                      = "GET HEAD POST OPTIONS PUT PATCH DELETE"
  allowed_request_content_type         = "application/x-www-form-urlencoded|multipart/form-data|text/xml|application/xml|application/x-amf|application/json|text/plain"
  allowed_request_content_type_charset = "utf-8|iso-8859-1|iso-8859-15|windows-1252"
  arg_length                           = 2000
  arg_name_length                      = 800
  combined_file_sizes                  = 10000000
  critical_anomaly_score               = 5
  crs_validate_utf8_encoding           = false
  error_anomaly_score                  = 4
  http_violation_score_threshold       = 5
  inbound_anomaly_score_threshold      = 15
  lfi_score_threshold                  = 5
  max_file_size                        = 10000000
  max_num_args                         = 255
  notice_anomaly_score                 = 2
  paranoia_level                       = 3
  php_injection_score_threshold        = 5
  rce_score_threshold                  = 5
  restricted_extensions                = ".asa/ .asax/ .ascx/ .backup/ .bak/ .bat/ .cdx/ .cer/ .cfg/ .cmd/ .com/ .config/ .conf/ .cs/ .csproj/ .csr/ .dat/ .db/ .dbf/ .dll/ .dos/ .htr/ .htw/ .ida/ .idc/ .idq/ .inc/ .ini/ .key/ .licx/ .lnk/ .log/ .mdb/ .old/ .pass/ .pdb/ .pol/ .printer/ .pwd/ .rdb/ .resources/ .resx/ .sql/ .swp/ .sys/ .vb/ .vbs/ .vbproj/ .vsdisco/ .webinfo/ .xsd/ .xsx/"
  restricted_headers                   = "/proxy/ /lock-token/ /content-range/ /if/"
  rfi_score_threshold                  = 5
  session_fixation_score_threshold     = 5
  sql_injection_score_threshold        = 15
  total_arg_length                     = 6400
  waf_id                               = fastly_service_vcl.service.waf[0].waf_id
  warning_anomaly_score                = 3
  xss_score_threshold                  = 15

  rule {
    modsec_rule_id = 1010010
    revision       = 1
    status         = "log"
  }
  rule {
    modsec_rule_id = 1010020
    revision       = 1
    status         = "log"
  }
  rule {
    modsec_rule_id = 1010030
    revision       = 1
    status         = "log"
  }
  rule {
    modsec_rule_id = 1010040
    revision       = 1
    status         = "log"
  }
  rule {
    modsec_rule_id = 1010050
    revision       = 1
    status         = "log"
  }
  rule {
    modsec_rule_id = 1010060
    revision       = 1
    status         = "log"
  }
  rule {
    modsec_rule_id = 1010070
    revision       = 1
    status         = "log"
  }
  rule {
    modsec_rule_id = 1010080
    revision       = 1
    status         = "log"
  }
  rule {
    modsec_rule_id = 1010090
    revision       = 1
    status         = "log"
  }
  rule {
    modsec_rule_id = 2100098
    revision       = 1
    status         = "log"
  }
  rule {
    modsec_rule_id = 2100099
    revision       = 1
    status         = "log"
  }
  rule {
    modsec_rule_id = 2100101
    revision       = 1
    status         = "log"
  }
  rule {
    modsec_rule_id = 2100102
    revision       = 1
    status         = "log"
  }
  rule {
    modsec_rule_id = 4100020
    revision       = 1
    status         = "log"
  }
  rule {
    modsec_rule_id = 4112010
    revision       = 1
    status         = "log"
  }
  rule {
    modsec_rule_id = 4112013
    revision       = 1
    status         = "log"
  }
  rule {
    modsec_rule_id = 4112014
    revision       = 1
    status         = "log"
  }
  rule {
    modsec_rule_id = 4112015
    revision       = 1
    status         = "log"
  }
  rule {
    modsec_rule_id = 4112016
    revision       = 1
    status         = "log"
  }
  rule {
    modsec_rule_id = 4112018
    revision       = 1
    status         = "log"
  }
  rule {
    modsec_rule_id = 4112019
    revision       = 1
    status         = "log"
  }
  rule {
    modsec_rule_id = 4112060
    revision       = 1
    status         = "log"
  }
  rule {
    modsec_rule_id = 4113001
    revision       = 1
    status         = "log"
  }
  rule {
    modsec_rule_id = 4113002
    revision       = 1
    status         = "log"
  }
  rule {
    modsec_rule_id = 4113010
    revision       = 1
    status         = "log"
  }
  rule {
    modsec_rule_id = 4113020
    revision       = 1
    status         = "log"
  }
  rule {
    modsec_rule_id = 4113030
    revision       = 1
    status         = "log"
  }
  rule {
    modsec_rule_id = 4113050
    revision       = 1
    status         = "log"
  }
  rule {
    modsec_rule_id = 4114100
    revision       = 1
    status         = "score"
  }
  rule {
    modsec_rule_id = 4114200
    revision       = 1
    status         = "score"
  }
  rule {
    modsec_rule_id = 4114220
    revision       = 1
    status         = "score"
  }
  rule {
    modsec_rule_id = 4114240
    revision       = 1
    status         = "score"
  }
  rule {
    modsec_rule_id = 4114300
    revision       = 1
    status         = "score"
  }
  rule {
    modsec_rule_id = 4120010
    revision       = 1
    status         = "log"
  }
  rule {
    modsec_rule_id = 4120011
    revision       = 2
    status         = "log"
  }
  rule {
    modsec_rule_id = 4134010
    revision       = 1
    status         = "log"
  }
  rule {
    modsec_rule_id = 910100
    revision       = 2
    status         = "score"
  }
  rule {
    modsec_rule_id = 911100
    revision       = 1
    status         = "score"
  }
  rule {
    modsec_rule_id = 913100
    revision       = 2
    status         = "score"
  }
  rule {
    modsec_rule_id = 913101
    revision       = 1
    status         = "score"
  }
  rule {
    modsec_rule_id = 913102
    revision       = 2
    status         = "score"
  }
  rule {
    modsec_rule_id = 913110
    revision       = 1
    status         = "score"
  }
  rule {
    modsec_rule_id = 913120
    revision       = 1
    status         = "score"
  }
  rule {
    modsec_rule_id = 920100
    revision       = 2
    status         = "score"
  }
  rule {
    modsec_rule_id = 920120
    revision       = 1
    status         = "score"
  }
  rule {
    modsec_rule_id = 920121
    revision       = 1
    status         = "score"
  }
  rule {
    modsec_rule_id = 920160
    revision       = 2
    status         = "score"
  }
  rule {
    modsec_rule_id = 920170
    revision       = 2
    status         = "score"
  }
  rule {
    modsec_rule_id = 920171
    revision       = 2
    status         = "score"
  }
  rule {
    modsec_rule_id = 920180
    revision       = 2
    status         = "score"
  }
  rule {
    modsec_rule_id = 920181
    revision       = 1
    status         = "score"
  }
  rule {
    modsec_rule_id = 920190
    revision       = 2
    status         = "score"
  }
  rule {
    modsec_rule_id = 920200
    revision       = 2
    status         = "score"
  }
  rule {
    modsec_rule_id = 920201
    revision       = 2
    status         = "score"
  }
  rule {
    modsec_rule_id = 920202
    revision       = 2
    status         = "score"
  }
  rule {
    modsec_rule_id = 920210
    revision       = 2
    status         = "score"
  }
  rule {
    modsec_rule_id = 920220
    revision       = 1
    status         = "score"
  }
  rule {
    modsec_rule_id = 920230
    revision       = 1
    status         = "score"
  }
  rule {
    modsec_rule_id = 920240
    revision       = 1
    status         = "score"
  }
  rule {
    modsec_rule_id = 920250
    revision       = 1
    status         = "score"
  }
  rule {
    modsec_rule_id = 920260
    revision       = 1
    status         = "score"
  }
  rule {
    modsec_rule_id = 920270
    revision       = 2
    status         = "score"
  }
  rule {
    modsec_rule_id = 920271
    revision       = 1
    status         = "score"
  }
  rule {
    modsec_rule_id = 920272
    revision       = 1
    status         = "score"
  }
  rule {
    modsec_rule_id = 920273
    revision       = 1
    status         = "score"
  }
  rule {
    modsec_rule_id = 920274
    revision       = 2
    status         = "score"
  }
  rule {
    modsec_rule_id = 920275
    revision       = 1
    status         = "score"
  }
  rule {
    modsec_rule_id = 920300
    revision       = 1
    status         = "score"
  }
  rule {
    modsec_rule_id = 920310
    revision       = 1
    status         = "score"
  }
  rule {
    modsec_rule_id = 920311
    revision       = 1
    status         = "score"
  }
  rule {
    modsec_rule_id = 920320
    revision       = 1
    status         = "score"
  }
  rule {
    modsec_rule_id = 920330
    revision       = 1
    status         = "score"
  }
  rule {
    modsec_rule_id = 920340
    revision       = 1
    status         = "score"
  }
  rule {
    modsec_rule_id = 920341
    revision       = 1
    status         = "score"
  }
  rule {
    modsec_rule_id = 920360
    revision       = 1
    status         = "score"
  }
  rule {
    modsec_rule_id = 920370
    revision       = 1
    status         = "score"
  }
  rule {
    modsec_rule_id = 920380
    revision       = 1
    status         = "score"
  }
  rule {
    modsec_rule_id = 920390
    revision       = 1
    status         = "score"
  }
  rule {
    modsec_rule_id = 920400
    revision       = 1
    status         = "score"
  }
  rule {
    modsec_rule_id = 920410
    revision       = 1
    status         = "score"
  }
  rule {
    modsec_rule_id = 920420
    revision       = 2
    status         = "score"
  }
  rule {
    modsec_rule_id = 920430
    revision       = 1
    status         = "score"
  }
  rule {
    modsec_rule_id = 920440
    revision       = 1
    status         = "score"
  }
  rule {
    modsec_rule_id = 920450
    revision       = 2
    status         = "score"
  }
  rule {
    modsec_rule_id = 920460
    revision       = 1
    status         = "score"
  }
  rule {
    modsec_rule_id = 920470
    revision       = 2
    status         = "score"
  }
  rule {
    modsec_rule_id = 920480
    revision       = 2
    status         = "score"
  }
  rule {
    modsec_rule_id = 920490
    revision       = 1
    status         = "score"
  }
  rule {
    modsec_rule_id = 920500
    revision       = 1
    status         = "score"
  }
  rule {
    modsec_rule_id = 920510
    revision       = 1
    status         = "score"
  }
  rule {
    modsec_rule_id = 921110
    revision       = 2
    status         = "score"
  }
  rule {
    modsec_rule_id = 921120
    revision       = 2
    status         = "score"
  }
  rule {
    modsec_rule_id = 921130
    revision       = 2
    status         = "score"
  }
  rule {
    modsec_rule_id = 921140
    revision       = 2
    status         = "score"
  }
  rule {
    modsec_rule_id = 921150
    revision       = 1
    status         = "score"
  }
  rule {
    modsec_rule_id = 921151
    revision       = 1
    status         = "score"
  }
  rule {
    modsec_rule_id = 921160
    revision       = 2
    status         = "score"
  }
  rule {
    modsec_rule_id = 921190
    revision       = 1
    status         = "score"
  }
  rule {
    modsec_rule_id = 921200
    revision       = 1
    status         = "score"
  }
  rule {
    modsec_rule_id = 930100
    revision       = 2
    status         = "score"
  }
  rule {
    modsec_rule_id = 930110
    revision       = 2
    status         = "score"
  }
  rule {
    modsec_rule_id = 930120
    revision       = 2
    status         = "score"
  }
  rule {
    modsec_rule_id = 930130
    revision       = 2
    status         = "score"
  }
  rule {
    modsec_rule_id = 931100
    revision       = 2
    status         = "score"
  }
  rule {
    modsec_rule_id = 931110
    revision       = 2
    status         = "score"
  }
  rule {
    modsec_rule_id = 931120
    revision       = 1
    status         = "score"
  }
  rule {
    modsec_rule_id = 931130
    revision       = 2
    status         = "score"
  }
  rule {
    modsec_rule_id = 932100
    revision       = 2
    status         = "score"
  }
  rule {
    modsec_rule_id = 932101
    revision       = 1
    status         = "score"
  }
  rule {
    modsec_rule_id = 932105
    revision       = 1
    status         = "score"
  }
  rule {
    modsec_rule_id = 932106
    revision       = 1
    status         = "score"
  }
  rule {
    modsec_rule_id = 932110
    revision       = 1
    status         = "score"
  }
  rule {
    modsec_rule_id = 932115
    revision       = 1
    status         = "score"
  }
  rule {
    modsec_rule_id = 932120
    revision       = 1
    status         = "score"
  }
  rule {
    modsec_rule_id = 932130
    revision       = 1
    status         = "score"
  }
  rule {
    modsec_rule_id = 932140
    revision       = 1
    status         = "score"
  }
  rule {
    modsec_rule_id = 932150
    revision       = 1
    status         = "score"
  }
  rule {
    modsec_rule_id = 932160
    revision       = 1
    status         = "score"
  }
  rule {
    modsec_rule_id = 932170
    revision       = 1
    status         = "score"
  }
  rule {
    modsec_rule_id = 932171
    revision       = 1
    status         = "score"
  }
  rule {
    modsec_rule_id = 932180
    revision       = 1
    status         = "score"
  }
  rule {
    modsec_rule_id = 932190
    revision       = 1
    status         = "score"
  }
  rule {
    modsec_rule_id = 932200
    revision       = 1
    status         = "score"
  }
  rule {
    modsec_rule_id = 933100
    revision       = 2
    status         = "score"
  }
  rule {
    modsec_rule_id = 933110
    revision       = 1
    status         = "score"
  }
  rule {
    modsec_rule_id = 933111
    revision       = 1
    status         = "score"
  }
  rule {
    modsec_rule_id = 933120
    revision       = 1
    status         = "score"
  }
  rule {
    modsec_rule_id = 933130
    revision       = 1
    status         = "score"
  }
  rule {
    modsec_rule_id = 933131
    revision       = 1
    status         = "score"
  }
  rule {
    modsec_rule_id = 933140
    revision       = 1
    status         = "score"
  }
  rule {
    modsec_rule_id = 933150
    revision       = 1
    status         = "score"
  }
  rule {
    modsec_rule_id = 933151
    revision       = 1
    status         = "score"
  }
  rule {
    modsec_rule_id = 933160
    revision       = 1
    status         = "score"
  }
  rule {
    modsec_rule_id = 933161
    revision       = 1
    status         = "score"
  }
  rule {
    modsec_rule_id = 933170
    revision       = 1
    status         = "score"
  }
  rule {
    modsec_rule_id = 933180
    revision       = 1
    status         = "score"
  }
  rule {
    modsec_rule_id = 933190
    revision       = 1
    status         = "score"
  }
  rule {
    modsec_rule_id = 933200
    revision       = 1
    status         = "score"
  }
  rule {
    modsec_rule_id = 933210
    revision       = 2
    status         = "score"
  }
  rule {
    modsec_rule_id = 934100
    revision       = 1
    status         = "score"
  }
  rule {
    modsec_rule_id = 941100
    revision       = 1
    status         = "score"
  }
  rule {
    modsec_rule_id = 941101
    revision       = 1
    status         = "score"
  }
  rule {
    modsec_rule_id = 941110
    revision       = 1
    status         = "score"
  }
  rule {
    modsec_rule_id = 941120
    revision       = 2
    status         = "score"
  }
  rule {
    modsec_rule_id = 941130
    revision       = 2
    status         = "score"
  }
  rule {
    modsec_rule_id = 941140
    revision       = 1
    status         = "score"
  }
  rule {
    modsec_rule_id = 941150
    revision       = 1
    status         = "score"
  }
  rule {
    modsec_rule_id = 941160
    revision       = 2
    status         = "score"
  }
  rule {
    modsec_rule_id = 941170
    revision       = 2
    status         = "score"
  }
  rule {
    modsec_rule_id = 941180
    revision       = 1
    status         = "score"
  }
  rule {
    modsec_rule_id = 941190
    revision       = 2
    status         = "score"
  }
  rule {
    modsec_rule_id = 941210
    revision       = 2
    status         = "score"
  }
  rule {
    modsec_rule_id = 941220
    revision       = 2
    status         = "score"
  }
  rule {
    modsec_rule_id = 941230
    revision       = 2
    status         = "score"
  }
  rule {
    modsec_rule_id = 941240
    revision       = 2
    status         = "score"
  }
  rule {
    modsec_rule_id = 941250
    revision       = 2
    status         = "score"
  }
  rule {
    modsec_rule_id = 941260
    revision       = 2
    status         = "score"
  }
  rule {
    modsec_rule_id = 941270
    revision       = 2
    status         = "score"
  }
  rule {
    modsec_rule_id = 941280
    revision       = 2
    status         = "score"
  }
  rule {
    modsec_rule_id = 941290
    revision       = 2
    status         = "score"
  }
  rule {
    modsec_rule_id = 941300
    revision       = 2
    status         = "score"
  }
  rule {
    modsec_rule_id = 941320
    revision       = 1
    status         = "score"
  }
  rule {
    modsec_rule_id = 941330
    revision       = 2
    status         = "score"
  }
  rule {
    modsec_rule_id = 941340
    revision       = 2
    status         = "score"
  }
  rule {
    modsec_rule_id = 941360
    revision       = 1
    status         = "score"
  }
  rule {
    modsec_rule_id = 941370
    revision       = 2
    status         = "score"
  }
  rule {
    modsec_rule_id = 941380
    revision       = 1
    status         = "score"
  }
  rule {
    modsec_rule_id = 942100
    revision       = 1
    status         = "score"
  }
  rule {
    modsec_rule_id = 942101
    revision       = 1
    status         = "score"
  }
  rule {
    modsec_rule_id = 942110
    revision       = 1
    status         = "score"
  }
  rule {
    modsec_rule_id = 942120
    revision       = 1
    status         = "score"
  }
  rule {
    modsec_rule_id = 942130
    revision       = 2
    status         = "score"
  }
  rule {
    modsec_rule_id = 942140
    revision       = 1
    status         = "score"
  }
  rule {
    modsec_rule_id = 942150
    revision       = 1
    status         = "score"
  }
  rule {
    modsec_rule_id = 942160
    revision       = 2
    status         = "score"
  }
  rule {
    modsec_rule_id = 942170
    revision       = 1
    status         = "score"
  }
  rule {
    modsec_rule_id = 942180
    revision       = 2
    status         = "score"
  }
  rule {
    modsec_rule_id = 942190
    revision       = 2
    status         = "score"
  }
  rule {
    modsec_rule_id = 942200
    revision       = 1
    status         = "score"
  }
  rule {
    modsec_rule_id = 942210
    revision       = 2
    status         = "score"
  }
  rule {
    modsec_rule_id = 942220
    revision       = 1
    status         = "score"
  }
  rule {
    modsec_rule_id = 942230
    revision       = 2
    status         = "score"
  }
  rule {
    modsec_rule_id = 942240
    revision       = 1
    status         = "score"
  }
  rule {
    modsec_rule_id = 942250
    revision       = 1
    status         = "score"
  }
  rule {
    modsec_rule_id = 942251
    revision       = 1
    status         = "score"
  }
  rule {
    modsec_rule_id = 942260
    revision       = 2
    status         = "score"
  }
  rule {
    modsec_rule_id = 942270
    revision       = 2
    status         = "score"
  }
  rule {
    modsec_rule_id = 942280
    revision       = 2
    status         = "score"
  }
  rule {
    modsec_rule_id = 942290
    revision       = 1
    status         = "score"
  }
  rule {
    modsec_rule_id = 942300
    revision       = 2
    status         = "score"
  }
  rule {
    modsec_rule_id = 942310
    revision       = 2
    status         = "score"
  }
  rule {
    modsec_rule_id = 942320
    revision       = 1
    status         = "score"
  }
  rule {
    modsec_rule_id = 942330
    revision       = 2
    status         = "score"
  }
  rule {
    modsec_rule_id = 942340
    revision       = 1
    status         = "score"
  }
  rule {
    modsec_rule_id = 942350
    revision       = 2
    status         = "score"
  }
  rule {
    modsec_rule_id = 942360
    revision       = 2
    status         = "score"
  }
  rule {
    modsec_rule_id = 942361
    revision       = 1
    status         = "score"
  }
  rule {
    modsec_rule_id = 942370
    revision       = 1
    status         = "score"
  }
  rule {
    modsec_rule_id = 942380
    revision       = 1
    status         = "score"
  }
  rule {
    modsec_rule_id = 942390
    revision       = 1
    status         = "score"
  }
  rule {
    modsec_rule_id = 942400
    revision       = 1
    status         = "score"
  }
  rule {
    modsec_rule_id = 942410
    revision       = 1
    status         = "score"
  }
  rule {
    modsec_rule_id = 942420
    revision       = 1
    status         = "score"
  }
  rule {
    modsec_rule_id = 942421
    revision       = 1
    status         = "score"
  }
  rule {
    modsec_rule_id = 942430
    revision       = 1
    status         = "score"
  }
  rule {
    modsec_rule_id = 942431
    revision       = 1
    status         = "score"
  }
  rule {
    modsec_rule_id = 942432
    revision       = 1
    status         = "score"
  }
  rule {
    modsec_rule_id = 942440
    revision       = 2
    status         = "score"
  }
  rule {
    modsec_rule_id = 942450
    revision       = 2
    status         = "score"
  }
  rule {
    modsec_rule_id = 942460
    revision       = 1
    status         = "score"
  }
  rule {
    modsec_rule_id = 942470
    revision       = 1
    status         = "score"
  }
  rule {
    modsec_rule_id = 942480
    revision       = 1
    status         = "score"
  }
  rule {
    modsec_rule_id = 942490
    revision       = 1
    status         = "score"
  }
  rule {
    modsec_rule_id = 942500
    revision       = 2
    status         = "score"
  }
  rule {
    modsec_rule_id = 942510
    revision       = 2
    status         = "score"
  }
  rule {
    modsec_rule_id = 942511
    revision       = 1
    status         = "score"
  }
  rule {
    modsec_rule_id = 943100
    revision       = 1
    status         = "score"
  }
  rule {
    modsec_rule_id = 943110
    revision       = 2
    status         = "score"
  }
  rule {
    modsec_rule_id = 943120
    revision       = 1
    status         = "score"
  }
  rule {
    modsec_rule_id = 944100
    revision       = 1
    status         = "score"
  }
  rule {
    modsec_rule_id = 944110
    revision       = 1
    status         = "score"
  }
  rule {
    modsec_rule_id = 944120
    revision       = 2
    status         = "score"
  }
  rule {
    modsec_rule_id = 944130
    revision       = 1
    status         = "score"
  }
  rule {
    modsec_rule_id = 944200
    revision       = 1
    status         = "score"
  }
  rule {
    modsec_rule_id = 944210
    revision       = 1
    status         = "score"
  }
  rule {
    modsec_rule_id = 944240
    revision       = 2
    status         = "score"
  }
  rule {
    modsec_rule_id = 944250
    revision       = 1
    status         = "score"
  }
  rule {
    modsec_rule_id = 944300
    revision       = 1
    status         = "score"
  }
}
