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
      --silent       Server doesn't send keepalive message
      --retry-limit int     The limit to retry, 0 is no limit
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
  -h, --help                      help for client
      --ignore-server-msg         Ignore whether receiving the message from server or not
  -i, --interval duration         Interval to send a keepalive message (default 5s)
  -y, --retry                     Try to connect after a previous connection is closed
  -r, --retry-interval duration   Connect retry interval (default 1s)
  -s, --server string             Server running host (default "127.0.0.1:8000")
  -t, --timeout duration          Deadline to receive a keepalive message (default 15s)
```
