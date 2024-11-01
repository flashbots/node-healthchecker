# node-healthchecker

Composite health (sync-status) checker for blockchain nodes.

Supported nodes:

- [x] Geth
- [x] Lighthouse
- [x] Op-node
- [x] Reth

## TL;DR

```shell
./node-healthchecker serve \
  --healthcheck-geth-base-url http://127.0.0.1:8545 \
  --healthcheck-lighthouse-base-url http://127.0.0.1:3500 \
  --server-listen-address 127.0.0.1:8080
```

```shell
curl -isS http://127.0.0.1:8080
```

- Unhappy path:

    ```text
    HTTP/1.1 500 Internal Server Error
    Content-Type: application/text
    Date: Mon, 30 Oct 2023 08:06:25 GMT
    Content-Length: 354

    0: error while checking sync-status of lighthouse at 'http://127.0.0.1:3500/lighthouse/syncing': Get "http://127.0.0.1:3500/lighthouse/syncing": dial tcp 127.0.0.1:3500: connect: connection refused
    1: error while checking sync-status of geth at 'http://127.0.0.1:8545/': Get "http://127.0.0.1:8545/": dial tcp 127.0.0.1:8545: connect: connection refused
    ```

- Happy path:

    ```text
    HTTP/1.1 200 OK
    Date: Mon, 30 Oct 2023 08:08:18 GMT
    Content-Length: 0
    ```

## CLI

```haskell
NAME:
   node-healthchecker serve - run node-healthchecker server

USAGE:
   node-healthchecker serve [command options]

GLOBAL OPTIONS:
   --log-level value  logging level (default: "info") [$NH_LOG_LEVEL]
   --log-mode value   logging mode (default: "prod") [$NH_LOG_MODE]

OPTIONS:
   HEALTHCHECK

   --healthcheck-cache-timeout duration  re-use healthcheck results for the specified duration (default: disabled) [$NH_HEALTHCHECK_CACHE_TIMEOUT]
   --healthcheck-timeout duration        maximum duration of a single healthcheck (default: 1s) [$NH_HEALTHCHECK_TIMEOUT]

   --healthcheck-timeout duration  maximum duration of a single healthcheck (default: 1s) [$NH_HEALTHCHECK_TIMEOUT]

   HEALTHCHECK GETH

   --healthcheck-geth-base-url url  base url of geth's HTTP-RPC endpoint [$NH_HEALTHCHECK_GETH_BASE_URL]

   HEALTHCHECK LIGHTHOUSE

   --healthcheck-lighthouse-base-url url  base url of lighthouse's HTTP-API endpoint [$NH_HEALTHCHECK_LIGHTHOUSE_BASE_URL]

   HEALTHCHECK OP-NODE

   --healthcheck-op-node-base-url url         base url of op-node's RPC endpoint [$NH_HEALTHCHECK_OP_NODE_BASE_URL]
   --healthcheck-op-node-conf-distance value  number of l1 blocks that verifier keeps distance from the l1 head before deriving l2 data from (default: 0) [$NH_HEALTHCHECK_OP_NODE_CONF_DISTANCE]

   HEALTHCHECK RETH

   --healthcheck-reth-base-url url  base url of reth's HTTP-RPC endpoint [$NH_HEALTHCHECK_RETH_BASE_URL]

   HTTP STATUS

   --http-status-error status    http status to report on healthchecks with errors (default: 500) [$NH_HTTP_STATUS_ERROR]
   --http-status-ok status       http status to report on good healthchecks (default: 200) [$NH_HTTP_STATUS_OK]
   --http-status-warning status  http status to report on healthchecks with warnings (default: 202) [$NH_HTTP_STATUS_WARNING]

   SERVER

   --server-listen-address host:port  host:port for the server to listen on (default: "xxx.xxx.xxx.xxx:8080") [$NH_SERVER_LISTEN_ADDRESS]
```
