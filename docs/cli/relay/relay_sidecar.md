# `relay sidecar` Command Reference

## relay_sidecar

Relay sidecar for signature aggregation

### Synopsis

A P2P service for collecting and aggregating signatures for Ethereum contracts.

```
relay_sidecar [flags]
```

### Options

```
      --aggregation-policy-max-unsigners uint   Max unsigners for low cost agg policy (default 50)
      --api.listen string                       API Server listener address
      --api.max-allowed-streams uint            Max allowed streams count API Server (default 100)
      --api.verbose-logging                     Enable verbose logging for the API Server
      --cache.network-config-size int           Network config cache size (default 10)
      --cache.validator-set-size int            Validator set cache size (default 10)
      --circuits-dir string                     Directory path to load zk circuits from, if empty then zp prover is disabled
      --config string                           Path to config file (default "config.yaml")
      --driver.address string                   Driver contract address
      --driver.chain-id uint                    Driver contract chain id
      --evm.chains strings                      Chains, comma separated rpc-url,..
      --evm.max-calls int                       Max calls in multicall
  -h, --help                                    help for relay_sidecar
      --key-cache.enabled                       Enable key cache (default true)
      --key-cache.size int                      Key cache size (default 100)
      --keystore.password string                Password for the keystore file, if provided will be used to decrypt the keystore file
      --keystore.path string                    Path to optional keystore file, if provided will be used instead of secret-keys flag
      --log.level string                        Log level (debug, info, warn, error) (default "info")
      --log.mode string                         Log mode (text, pretty, json) (default "json")
      --metrics.listen string                   Http listener address for metrics endpoint
      --metrics.pprof                           Enable pprof debug endpoints
      --p2p.bootnodes strings                   List of bootnodes in multiaddr format
      --p2p.dht-mode string                     DHT mode: auto, server, client, disabled (default "server")
      --p2p.listen string                       P2P listen address
      --p2p.mdns                                Enable mDNS discovery for P2P
      --secret-keys secret-key-slice            Secret keys, comma separated {namespace}/{type}/{id}/{key},..
      --signal.buffer-size int                  Signal buffer size (default 20)
      --signal.worker-count int                 Signal worker count (default 10)
      --storage-dir string                      Dir to store data (default ".data")
      --sync.enabled                            Enable signature syncer (default true)
      --sync.epochs uint                        Epochs to sync (default 5)
      --sync.period duration                    Signature sync period (default 5s)
      --sync.timeout duration                   Signature sync timeout (default 1m0s)
```

