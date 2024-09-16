# node-healthchecker

Composite health (sync status) checker for blockchain nodes.

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
