# Uptime calculator from prometheus

A tool to calculate Individual API Endpoint's uptime using Prometheus server's HTTP API.

## Building

    go build

## Usage

    -Configure uptime.yml
     
     Example:
        server: http://localhost:9090
        query:
              - sum(increase(2xx_request_count{request_endpoint='GET_/v1/example/a/b'}[1d]))
              - sum(increase(5xx_request_count{request_endpoint='GET_/v1/example/a/b'}[1d]))
        email: john_doe@example.com
    
    -Run the program ./uptime_from_prometheus

