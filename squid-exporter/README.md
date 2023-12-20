# squid-exporter
Prometheus exporter for squid

## Description
squid-exporter converts squid counters and service_times to prometheus metrics.

## Usage
```
    ./squid-exporter -squid-host localhost -squid-port 3128 -metrics-port 8080
```

## Option
| option        | default   | description         |
| ----          | ----      | ----                |
| -squid-host   | localhost | squid host          |
| -squid-port   | 3128      | squid port          |
| -metrics-port | 9100      | metrics expose port |

