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
