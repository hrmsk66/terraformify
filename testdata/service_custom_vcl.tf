variable "domain" {
  type = string
}

resource "fastly_service_vcl" "service" {
  name               = var.domain
  stale_if_error     = true
  stale_if_error_ttl = 43200

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

  director {
    backends = [
      "Host 1",
      "Host 2",
    ]
    name    = "director"
    quorum  = 75
    retries = 5
    type    = 3
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

  logging_s3 {
    bucket_name      = "s3_bucket"
    domain           = "s3.amazonaws.com"
    format           = <<-EOT
            {
                "timestamp": "%%{strftime(\{"%Y-%m-%dT%H:%M:%S%z"\}, time.start)}V",
                "client_ip": "%%{req.http.Fastly-Client-IP}V",
                "geo_country": "%%{client.geo.country_name}V",
                "geo_city": "%%{client.geo.city}V",
                "host": "%%{if(req.http.Fastly-Orig-Host, req.http.Fastly-Orig-Host, req.http.Host)}V",
                "url": "%%{json.escape(req.url)}V",
                "request_method": "%%{json.escape(req.method)}V",
                "request_protocol": "%%{json.escape(req.proto)}V",
                "request_referer": "%%{json.escape(req.http.referer)}V",
                "request_user_agent": "%%{json.escape(req.http.User-Agent)}V",
                "response_state": "%%{json.escape(fastly_info.state)}V",
                "response_status": %%{resp.status}V,
                "response_reason": %%{if(resp.response, "%22"+json.escape(resp.response)+"%22", "null")}V,
                "response_body_size": %%{resp.body_bytes_written}V,
                "fastly_server": "%%{json.escape(server.identity)}V",
                "fastly_is_edge": %%{if(fastly.ff.visits_this_service == 0, "true", "false")}V
              }
        EOT
    format_version   = 2
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

  request_setting {
    bypass_busy_wait = false
    force_miss       = false
    force_ssl        = true
    geo_headers      = false
    max_stale_age    = 0
    name             = "Generated by force TLS and enable HSTS"
    timer_support    = false
  }

  response_object {
    content           = <<-EOT
            User-Agent: *
            Disallow:
        EOT
    content_type      = "text/plain"
    name              = "Generated by synthetic response for robots.txt"
    request_condition = "Generated by synthetic response for robots.txt"
    response          = "OK"
    status            = 200
  }
  response_object {
    cache_condition = "Generated by synthetic response for 404 page"
    content         = <<-EOT
            <!DOCTYPE html>
            <html>
              <head>
                <meta charset="UTF-8">
                <title>404</title>
              </head>
              <body>
                404
              </body>
            </html>
        EOT
    content_type    = "text/html"
    name            = "Generated by synthetic response for 404 page"
    response        = "Not Found"
    status          = 404
  }
  response_object {
    cache_condition = "Generated by synthetic response for 503 page"
    content         = <<-EOT
            <!DOCTYPE html>
            <html>
              <head>
                <meta charset="UTF-8">
                <title>503</title>
              </head>
              <body>
                503
              </body>
            </html>
        EOT
    content_type    = "text/html"
    name            = "Generated by synthetic response for 503 page"
    response        = "Service Unavailable"
    status          = 503
  }

  vcl {
    main    = true
    name    = "main"
    content = <<-EOT
            sub vcl_recv {
            #FASTLY recv

              # Normally, you should consider requests other than GET and HEAD to be uncacheable
              # (to this we add the special FASTLYPURGE method)
              if (req.method != "HEAD" && req.method != "GET" && req.method != "FASTLYPURGE") {
                return(pass);
              }

              # If you are using image optimization, insert the code to enable it here
              # See https://developer.fastly.com/reference/io/ for more information.

              return(lookup);
            }

            sub vcl_hash {
              set req.hash += req.url;
              set req.hash += req.http.host;
              #FASTLY hash
              return(hash);
            }

            sub vcl_hit {
            #FASTLY hit
              return(deliver);
            }

            sub vcl_miss {
            #FASTLY miss
              return(fetch);
            }

            sub vcl_pass {
            #FASTLY pass
              return(pass);
            }

            sub vcl_fetch {
            #FASTLY fetch

              # Unset headers that reduce cacheability for images processed using the Fastly image optimizer
              if (req.http.X-Fastly-Imageopto-Api) {
                unset beresp.http.Set-Cookie;
                unset beresp.http.Vary;
              }

              # Log the number of restarts for debugging purposes
              if (req.restarts > 0) {
                set beresp.http.Fastly-Restarts = req.restarts;
              }

              # If the response is setting a cookie, make sure it is not cached
              if (beresp.http.Set-Cookie) {
                return(pass);
              }

              # By default we set a TTL based on the `Cache-Control` header but we don't parse additional directives
              # like `private` and `no-store`.  Private in particular should be respected at the edge:
              if (beresp.http.Cache-Control ~ "(private|no-store)") {
                return(pass);
              }

              # If no TTL has been provided in the response headers, set a default
              if (!beresp.http.Expires && !beresp.http.Surrogate-Control ~ "max-age" && !beresp.http.Cache-Control ~ "(s-maxage|max-age)") {
                set beresp.ttl = 3600s;

                # Apply a longer default TTL for images processed using Image Optimizer
                if (req.http.X-Fastly-Imageopto-Api) {
                  set beresp.ttl = 2592000s; # 30 days
                  set beresp.http.Cache-Control = "max-age=2592000, public";
                }
              }

              return(deliver);
            }

            sub vcl_error {
            #FASTLY error
              return(deliver);
            }

            sub vcl_deliver {
            #FASTLY deliver
              return(deliver);
            }

            sub vcl_log {
            #FASTLY log
            }
        EOT
  }
}

output "id" {
  value = fastly_service_vcl.service.id
}
