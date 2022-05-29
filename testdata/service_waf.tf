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

  logging_papertrail {
    address = "xxx.example.com"
    format = jsonencode(
      {
        anomaly_score          = "%%{waf.anomaly_score}V"
        client_ip              = "%a"
        datacenter             = "%%{server.datacenter}V"
        fastly_info            = "%%{fastly_info.state}V"
        http_violation_score   = "%%{waf.http_violation_score}V"
        lfi_score              = "%%{waf.lfi_score}V"
        php_injection_score    = "%%{waf.php_injection_score}V"
        rce_score              = "%%{waf.rce_score}V"
        req_body_bytes         = "%%{req.body_bytes_read}V"
        req_h_accept_encoding  = "%%{cstr_escape(req.http.Accept-Encoding)}V"
        req_h_host             = "%%{cstr_escape(req.http.Host)}V"
        req_h_referer          = "%%{cstr_escape(req.http.referer)}V"
        req_h_user_agent       = "%%{cstr_escape(req.http.User-Agent)}V"
        req_header_bytes       = "%%{req.header_bytes_read}V"
        req_method             = "%m"
        req_uri                = "%%{cstr_escape(req.url)}V"
        request_id             = "%%{req.http.fastly-soc-x-request-id}V"
        resp_body_bytes        = "%%{resp.body_bytes_written}V"
        resp_bytes             = "%%{resp.bytes_written}V"
        resp_header_bytes      = "%%{resp.header_bytes_written}V"
        resp_status            = "%%{resp.status}V"
        rfi_score              = "%%{waf.rfi_score}V"
        service_id             = "%%{req.service_id}V"
        session_fixation_score = "%%{waf.session_fixation_score}V"
        sql_injection_score    = "%%{waf.sql_injection_score}V"
        start_time             = "%%{time.start.sec}V"
        type                   = "req"
        waf_blocked            = "%%{waf.blocked}V"
        waf_executed           = "%%{waf.executed}V"
        waf_failures           = "%%{waf.failures}V"
        waf_logged             = "%%{waf.logged}V"
        xss_score              = "%%{waf.xss_score}V"
      }
    )
    format_version     = 2
    name               = "weblogs"
    port               = 12345
    response_condition = "waf-soc-logging"
  }
  logging_papertrail {
    address = "xxx.example.com"
    format = jsonencode(
      {
        anomaly_score = "%%{waf.anomaly_score}V"
        logdata       = "%%{cstr_escape(waf.logdata)}V"
        request_id    = "%%{req.http.fastly-soc-x-request-id}V"
        rule_id       = "%%{waf.rule_id}V"
        severity      = "%%{waf.severity}V"
        type          = "waf"
        waf_message   = "%%{waf.message}V"
      }
    )
    format_version = 2
    name           = "waflogs"
    placement      = "waf_debug"
    port           = 12345
  }

  response_object {
    content           = "{ \"Access Denied\" : \"\"} req.http.fastly-soc-x-request-id {\"\" }"
    content_type      = "application/json"
    name              = "WAF_Response"
    request_condition = "false"
    response          = "Forbidden"
    status            = 403
  }

  snippet {
    content  = <<-EOT
          if (!req.http.fastly-csi-request-id) {
            set req.http.fastly-csi-request-id = now.sec substr(digest.hash_sha256(randomstr(64) req.http.host req.url req.http.Fastly-Client-IP server.identity), 0, 21);
            set req.http.fastly-soc-x-request-id = req.http.fastly-csi-request-id;
          }
      EOT
    name     = "fastly_csi_init"
    priority = 5
    type     = "recv"
  }
  snippet {
    content  = <<-EOT
          unset req.http.rqpass;
          if (!req.http.fastly-soc-x-request-id) {
            set req.http.fastly-soc-x-request-id = digest.hash_sha256(now randomstr(64) req.http.host req.url req.http.Fastly-Client-IP server.identity);
          }
      EOT
    name     = "Fastly_WAF_Snippet"
    priority = 10
    type     = "recv"
  }

  waf {
    prefetch_condition = "WAF_Prefetch"
    response_object    = "WAF_Response"
  }

  force_destroy = true
}

# WAF resource settings
data "fastly_waf_rules" "default" {
  tags                    = ["owasp", "application-multi"]
  exclude_modsec_rule_ids = [4112031, 4112011, 4112012]
}

variable "type_status" {
  type = map(string)
  default = {
    score     = "score"
    threshold = "log"
    strict    = "log"
  }
}

resource "fastly_service_waf_configuration" "waf" {
  waf_id                           = fastly_service_vcl.service.waf[0].waf_id
  http_violation_score_threshold   = 5
  inbound_anomaly_score_threshold  = 15
  lfi_score_threshold              = 5
  php_injection_score_threshold    = 5
  rce_score_threshold              = 5
  rfi_score_threshold              = 5
  session_fixation_score_threshold = 5
  sql_injection_score_threshold    = 15
  xss_score_threshold              = 15
  allowed_request_content_type     = "application/x-www-form-urlencoded|multipart/form-data|text/xml|application/xml|application/x-amf|application/json|text/plain"
  arg_name_length                  = 800
  arg_length                       = 2000
  paranoia_level                   = 3

  dynamic "rule" {
    for_each = data.fastly_waf_rules.default.rules
    content {
      modsec_rule_id = rule.value.modsec_rule_id
      status         = lookup(var.type_status, rule.value.type, "log")
    }
  }
}

output "id" {
  value = fastly_service_vcl.service.id
}
