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
      --aggregation-policy-max-unsigners uint    Max unsigners for low cost agg policy (default 50)
      --api.http-gateway                         Enable HTTP/JSON REST API gateway on /api/v1/* path
      --api.listen string                        API Server listener address
      --api.max-allowed-streams uint             Max allowed streams count API Server (default 100)
      --api.verbose-logging                      Enable verbose logging for the API Server
      --badger.block-cache-size int              BadgerDB block cache size in bytes, 0 = disabled, -1 = badger default (256 MB), default: 128 MB (default 134217728)
      --badger.compact-l0-on-close               BadgerDB compact L0 on graceful shutdown (badger default: false) (default true)
      --badger.mem-table-size int                BadgerDB memtable size in bytes, 32 MB (badger default: 64 MB) (default 33554432)
      --badger.num-level-zero-tables int         BadgerDB L0 tables before compaction triggers (badger default: 5) (default 3)
      --badger.num-level-zero-tables-stall int   BadgerDB L0 tables before writes stall (badger default: 15) (default 8)
      --badger.num-memtables int                 BadgerDB number of memtables (badger default: 5) (default 3)
      --badger.value-log-file-size int           BadgerDB value log file size in bytes, 512 MB (badger default: ~1 GB) (default 536870912)
      --cache.network-config-size int            Network config cache size (default 10)
      --cache.validator-set-size int             Validator set cache size (default 10)
      --circuits-dir string                      Directory path to load zk circuits from, if empty then zp prover is disabled
      --config string                            Path to config file (default "config.yaml")
      --driver.address string                    Driver contract address
      --driver.chain-id uint                     Driver contract chain id
      --evm.chains strings                       Chains, comma separated rpc-url,..
      --evm.fallback-gas-prices gas-price-map    Per-chain fallback gas prices in wei when eth_maxPriorityFeePerGas is not supported (e.g., --evm.fallback-gas-prices 1=2000000000)
      --evm.max-calls int                        Max calls in multicall
      --force-role.aggregator                    Force node to act as aggregator regardless of deterministic scheduling
      --force-role.committer                     Force node to act as committer regardless of deterministic scheduling
  -h, --help                                     help for relay_sidecar
      --key-cache.enabled                        Enable key cache (default true)
      --key-cache.size int                       Key cache size (default 100)
      --keystore.password string                 Password for the keystore file, if provided will be used to decrypt the keystore file
      --keystore.path string                     Path to optional keystore file, if provided will be used instead of secret-keys flag
      --log.level string                         Log level (debug, info, warn, error) (default "info")
      --log.mode string                          Log mode (text, pretty, json) (default "json")
      --metrics.listen string                    Http listener address for metrics endpoint
      --metrics.pprof                            Enable pprof debug endpoints
      --p2p.bootnodes strings                    List of bootnodes in multiaddr format
      --p2p.dht-mode string                      DHT mode: auto, server, client, disabled (default "server")
      --p2p.listen string                        P2P listen address
      --p2p.mdns                                 Enable mDNS discovery for P2P
      --pruner.enabled                           Enable automatic pruning of old epoch data (default: false)
      --pruner.interval duration                 How often to run pruning (default: 1h) (default 1h0m0s)
      --retention.proof-epochs uint              Number of historical proof epochs to retain (0 = unlimited)
      --retention.signature-epochs uint          Number of historical signature epochs to retain (0 = unlimited)
      --retention.valset-epochs uint             Number of historical validator set epochs to retain (0 = unlimited)
      --secret-keys secret-key-slice             Secret keys, comma separated {namespace}/{type}/{id}/{key},..
      --signal.buffer-size int                   Signal buffer size (default 20)
      --signal.worker-count int                  Signal worker count (default 10)
      --storage-dir string                       Dir to store data (default ".data")
      --sync.enabled                             Enable signature syncer (default true)
      --sync.epochs uint                         Epochs to sync (default 5)
      --sync.period duration                     Signature sync period (default 5s)
      --sync.timeout duration                    Signature sync timeout (default 1m0s)
      --tracing.enabled                          Enable distributed tracing
      --tracing.endpoint string                  OTLP endpoint for tracing (e.g., Jaeger) (default "localhost:4317")
      --tracing.sample-rate float                Trace sampling rate (0.0 to 1.0) (default 1)
```

