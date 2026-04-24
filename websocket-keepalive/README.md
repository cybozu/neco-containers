# Websocket keepalive

websocket-keepalive は websocket サーバ・クライアントの疎通の継続性を検証するためのプログラムである。

TCP 版の [tcp-keepalive](../tcp-keepalive) に対応する WebSocket 版であり、L7 ロードバランサや Ingress Controller など HTTP レイヤで中継される経路の疎通監視に用いる。

## Server

サーバは websocket-keepalive クライアントからの接続を待ち受け、接続したクライアントからの Ping メッセージに Pong メッセージで応答を返す。複数クライアントを同時に受け付ける。

Ping が `--ping-interval` の 2 倍の時間届かなかった接続は、read deadline を超過したものとして閉じる。

```console
$ ./websocket-keepalive server -h
Start a WebSocket server that handles client connections

Usage:
  websocket-keepalive server [flags]

Flags:
  -h, --help                     help for server
  -l, --listen string            Host to listen on (default "0.0.0.0")
  -m, --metrics                  Enable metrics (default true)
  -a, --metrics-server string    Metrics server address and port (default "0.0.0.0:8081")
  -i, --ping-interval duration   Expected client ping interval (used to compute read deadline) (default 10s)
  -p, --port int                 Port to listen on (default 9000)

Global Flags:
      --log-level string   Log level (debug, info, warn, error) (default "info")
```

### Metrics

```
established{role="server",local="<local_addr>",remote="<remote_addr>"}
received_ping_total{role="server",local="<local_addr>",remote="<remote_addr>"}
sent_pong_total{role="server",local="<local_addr>",remote="<remote_addr>"}
```

## Client

クライアントは `--host` と `--port` で指定された宛先のサーバに対して Websocket のコネクションを接続する。
接続を確立すると、`--ping-interval` で指定された間隔でサーバに対して Ping メッセージを送信し、サーバからの Pong メッセージを待ち受ける。
Pong メッセージの待ち受けは `--ping-interval` で指定された値の 2 倍の時間である。この期間 Pong メッセージが返ってこない場合は、クライアントから Ping メッセージを再送する。
`--max-retry-limit` で指定された回数再送を繰り返しても Pong メッセージが返ってこない場合は、コネクションが破棄されたものとして接続を切り、exit code 1 でプログラムを終了する。

再接続ループは持たない。Kubernetes の Deployment など上位のリスタート機構によって再起動される前提。これは tcp-keepalive とは異なる挙動である。

`--ping-interval` は経路上の最短 NAT / firewall idle timeout より十分短く設定すること。長すぎるとアイドルタイムアウトを越えて false positive で切断されうる。

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

### Metrics

```
established{role="client",local="<local_addr>",remote="<remote_addr>"}
sent_ping_total{role="client",local="<local_addr>",remote="<remote_addr>"}
received_pong_total{role="client",local="<local_addr>",remote="<remote_addr>"}
ping_retry_count_total{role="client",local="<local_addr>",remote="<remote_addr>"}
```
