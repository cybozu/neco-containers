# Websocket keepalive

websocket-keepalive は websocket サーバ・クライアントの疎通の継続性を検証するためのプログラムである。

## Server

サーバは websocket-keepalive クライアントからの接続を待ち受け、接続したクライアントからの Ping メッセージに Pong メッセージで応答を返すのみである。

```console
$ ./websocket-keepalive server -h
Start a WebSocket server that handles client connections

Usage:
  websocket-keepalive server [flags]

Flags:
  -h, --help                     help for server
  -l, --listen string            Host to listen on (default "0.0.0.0")
  -i, --ping-interval duration   Ping interval in seconds (default 5s)
  -p, --port int                 Port to listen on (default 9000)

Global Flags:
      --log-level string   Log level (debug, info, warn, error) (default "info")
```

## Client

クライアントは `--host` と `--port` で指定された宛先のサーバに対して Websocket のコネクションを接続する。
接続を確立すると、`--ping-interval` で指定された間隔でサーバに対して Ping メッセージを送信し、サーバからの Pong メッセージを待ち受ける。
Pong メッセージの待ち受けは `--ping-interval` で指定された値の 2 倍の時間である。この期間 Pong メッセージが返ってこない場合は、クライアントから Ping メッセージを再送する。
`--max-retry-limit` で指定された回数再送を繰り返しても Pong メッセージが返ってこない場合は、コネクションが破棄されたものとして接続を切り、プログラムを終了する。

```console
$ ./websocket-keepalive client -h
Start a WebSocket client that sends periodic ping messages

Usage:
  websocket-keepalive client [flags]

Flags:
  -h, --help                     help for client
  -H, --host string              Server host to connect to (default "localhost")
  -r, --max-retry-limit int      Limit for retrying to send ping (default 3)
  -m, --metrics                  Enable metrics (default true)
  -a, --metrics-server string    Metrics server address and port (default "0.0.0.0:8080")
  -i, --ping-interval duration   Interval for sending ping messages (default 10s)
  -p, --port int                 Server port to connect to (default 9000)

Global Flags:
      --log-level string   Log level (debug, info, warn, error) (default "info")
```

## Metrics

メトリクスはクライアントのみが `--metrics-server` で指定されたポートで出力する。

出力するメトリクスは以下のようになっている。

- `established`
  - Websocket で接続が確立していることを示す。確立していたら 1 を出力する。
- `ping_retry_count_total`
  - Ping メッセージを再送した回数を示すカウンタ。
