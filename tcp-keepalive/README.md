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
run the TCP server

Usage:
  tcp-keepalive server [flags]

Flags:
  -h, --help                help for server
  -i, --interval duration   Interval to send a keepalive message (default 5s)
  -l, --listen string       Listen address and port (default ":8000")
  -t, --timeout duration    Deadline to receive a keepalive message (default 15s)
```

Client
-----

```console
$ tcp-keepalive client -h
run the TCP client

Usage:
  tcp-keepalive client [flags]

Flags:
  -r, --connect-retry duration   Connect retry interval (default 1s)
  -h, --help                     help for client
  -i, --interval duration        Interval to send a keepalive message (default 5s)
  -s, --server string            Server running host (default "127.0.0.1:8000")
  -t, --timeout duration         Deadline to receive a keepalive message (default 15s)
```
