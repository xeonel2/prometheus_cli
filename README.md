# Uptime calculator from prometheus

A tool to calculate Individual API Endpoint's uptime using Prometheus server's HTTP API.

## Building

    go build

## Usage

    -Configure uptime.yml
     
     Example:
        server: http://localhost:9090
        emailfrom: monitoring@example.com
        emailto: john_doe@example.com
        smtphost: mail.example.com
        smtpport: 25
        smtpuser: blahblah
        smtppwd: passworff
        endpoints:
                - name: Service A GET_/v1/example/a/b
                  success: sum(max_over_time(logstash_http_2xx_count{request_endpoint='GET_/v1/example/a/b'}[1d]) - min_over_time(logstash_http_2xx_count{request_endpoint='GET_/v1/example/a/b'}[1d]))
                  failed: sum(max_over_time(logstash_http_5xx_count{request_endpoint='GET_/v1/example/a/b'}[1d]) - min_over_time(logstash_http_5xx_count{request_endpoint='GET_/v1/example/a/b'}[1d]))
                - name: Service A POST_/v1/example/c/d
                  success: sum(max_over_time(logstash_http_2xx_count{request_endpoint='POST_/v1/example/c/d'}[1d]) - min_over_time(logstash_http_2xx_count{request_endpoint='POST_/v1/example/c/d'}[1d]))
                  failed: sum(max_over_time(logstash_http_5xx_count{request_endpoint='POST_/v1/example/c/d'}[1d]) - min_over_time(logstash_http_5xx_count{request_endpoint='POST_/v1/example/c/d'}[1d]))
                - name: Service B GET_/areyouthere
                  success: sum(max_over_time(logstash_http_2xx_count{request_endpoint='GET_/areyouthere'}[1d]) - min_over_time(logstash_http_2xx_count{request_endpoint='GET_/areyouthere'}[1d]))
                  failed: sum(max_over_time(logstash_http_5xx_count{request_endpoint='GET_/areyouthere'}[1d]) - min_over_time(logstash_http_5xx_count{request_endpoint='GET_/areyouthere'}[1d]))

    
    -Run the program ./uptime_from_prometheus

