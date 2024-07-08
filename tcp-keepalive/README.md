tcp-keepalive container
===============

This directory provides the container image `tcp-keepalive` and its source code.

`tcp-keepalive` is a simple TCP server and client that exchange keepalive messages with each other.

Usage
-----

Server
-----

```console
$ tcp-keepalive server -h
Run the tcp-keepalive server

Usage:
  tcp-keepalive server [flags]

Flags:
  -h, --help            help for server
  -l, --listen string   Listen address and port (default "127.0.0.1:8000")
```

### Metrics

```
receive_total{role="server",result="error"}
receive_total{role="server",result="success"}

send_total{role="server",result="error"}
send_total{role="server",result="success"}
```

Client
-----

```console
$ tcp-keepalive client -h
Run the tcp-keepalive client

Usage:
  tcp-keepalive client [flags]

Flags:
  -h, --help                      help for client
  -i, --interval duration         Interval to send a keepalive message (default 5s)
  -y, --retry int                 Number of retries (-1 means infinite)
  -r, --retry-interval duration   Connect retry interval (default 1s)
  -s, --server string             server address (default "127.0.0.1:8000")
  -t, --timeout duration          Deadline to receive a keepalive message (default 15s)
```

### Metrics

```
connection{role="client",state="closed"}
connection{role="client",state="established"}
connection{role="client",state="unestablished"}

receive_total{role="client",result="error"}
receive_total{role="client",result="success"}
receive_total{role="client",result="timeout"}

retry_count{role="client"}
retry_total{role="client"}

send_total{role="client",result="error"}
send_total{role="client",result="success"}
send_total{role="client",result="timeout"}
```
