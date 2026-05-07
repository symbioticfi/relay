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
      --aggregation-policy-max-unsigners uint            Max unsigners for low cost agg policy (default 50)
      --aggregation.catchup.enabled                      Enable periodic aggregation catch-up loop (default true)
      --aggregation.catchup.epochs-offset int            Number of epochs back from latest to skip before scanning begins
      --aggregation.catchup.epochs-to-check int          Number of epochs to scan per catch-up cycle (default 20)
      --aggregation.catchup.interval duration            How often to run aggregation catch-up (default 1m0s)
      --aggregation.catchup.max-requests-per-cycle int   Max requests to check per cycle (0 = unlimited)
      --aggregation.cross-epoch-aggregation              Allow latest-epoch aggregators to aggregate proofs for older epochs when original aggregators are offline
      --aggregation.worker-count int                     Max simultaneous proof aggregations, reduce for ZK circuits with high memory and cpu usage (default 10)
      --api.http-gateway                                 Enable HTTP/JSON REST API gateway on /api/v1/* path
      --api.listen string                                API Server listener address
      --api.max-allowed-streams uint                     Max allowed streams count API Server (default 100)
      --api.verbose-logging                              Enable verbose logging for the API Server
      --badger.block-cache-size int                      BadgerDB block cache size in bytes, 0 = disabled (default 134217728)
      --badger.compact-l0-on-close                       BadgerDB compact L0 on graceful shutdown (default true)
      --badger.mem-table-size int                        BadgerDB memtable size in bytes (default 33554432)
      --badger.num-compactors int                        BadgerDB concurrent compaction goroutines (default 2)
      --badger.num-level-zero-tables int                 BadgerDB L0 tables before compaction triggers (default 3)
      --badger.num-level-zero-tables-stall int           BadgerDB L0 tables before writes stall (default 8)
      --badger.num-memtables int                         BadgerDB number of memtables (default 3)
      --badger.value-log-file-size int                   BadgerDB value log file size in bytes, 512 MB (default 536870912)
      --badger.value-log-gc-discard-ratio float          BadgerDB value log GC discard ratio (0.0-1.0) (default 0.5)
      --badger.value-log-gc-interval duration            BadgerDB value log GC interval, 0 = disabled (default 5m0s)
      --bbolt.compact-on-startup                         Compact database on startup to reclaim free pages (default true)
      --bbolt.initial-mmap-size int                      Initial mmap size in bytes (0 = default)
      --bbolt.max-batch-delay duration                   Max delay before flushing a batch write (0 = bbolt default 10ms) (default 2ms)
      --bbolt.max-batch-size int                         Max operations per batch write (0 = bbolt default 1000)
      --bbolt.no-freelist-sync                           Skip writing freelist to disk on every commit (faster writes, slower startup) (default true)
      --bbolt.stats-log-interval duration                Interval for logging bbolt database stats (0 = disabled)
      --cache.network-config-size int                    Network config cache size (default 10)
      --cache.validator-set-size int                     Validator set cache size (default 10)
      --circuits-dir string                              Directory path to load zk circuits from, if empty then zk prover is disabled
      --config string                                    Path to config file (default "config.yaml")
      --driver.address string                            Driver contract address
      --driver.chain-id uint                             Driver contract chain id
      --evm.chains strings                               Chains, comma separated rpc-url,..
      --evm.fallback-gas-prices gas-price-map            Per-chain fallback gas prices in wei when eth_maxPriorityFeePerGas is not supported (e.g., --evm.fallback-gas-prices 1=2000000000)
      --evm.max-calls int                                Max calls in multicall
      --force-role.aggregator                            Force node to act as aggregator regardless of deterministic scheduling
      --force-role.committer                             Force node to act as committer regardless of deterministic scheduling
  -h, --help                                             help for relay_sidecar
      --key-cache.enabled                                Enable key cache (default true)
      --key-cache.size int                               Key cache size (default 100)
      --keystore.password string                         Password for the keystore file, if provided will be used to decrypt the keystore file
      --keystore.path string                             Path to optional keystore file, if provided will be used instead of secret-keys flag
      --log.level string                                 Log level (debug, info, warn, error) (default "info")
      --log.mode string                                  Log mode (text, pretty, json) (default "json")
      --metrics.listen string                            Http listener address for metrics endpoint
      --metrics.pprof                                    Enable pprof debug endpoints
      --p2p.bootnodes strings                            List of bootnodes in multiaddr format
      --p2p.dht-mode string                              DHT mode: auto, server, client, disabled (default "server")
      --p2p.listen string                                P2P listen address
      --p2p.mdns                                         Enable mDNS discovery for P2P
      --p2p.publish-timeout duration                     Maximum time a single pubsub publish may block (default 10s)
      --pruner.batch-pause duration                      Pause between prune batches to yield to live writers (bbolt only — badger has no batching) (0 = no pause) (default 100ms)
      --pruner.batch-size int                            Number of request IDs to delete per database transaction during pruning (0 = unbatched) (default 100)
      --pruner.enabled                                   Enable automatic pruning of old epoch data
      --pruner.interval duration                         How often to run pruning (default 1h0m0s)
      --retention.proof-epochs uint                      Number of historical proof epochs to retain (0 = unlimited)
      --retention.signature-epochs uint                  Number of historical signature epochs to retain (0 = unlimited)
      --retention.valset-epochs uint                     Number of historical validator set epochs to retain (0 = unlimited)
      --secret-keys secret-key-slice                     Secret keys, comma separated {namespace}/{type}/{id}/{secret},..
      --signal.buffer-size int                           Signal buffer size (default 20)
      --signal.worker-count int                          Signal worker count (default 10)
      --storage-dir string                               Dir to store data (default ".data")
      --storage-type string                              Storage backend type (badger, bbolt) (default "bbolt")
      --sync.enabled                                     Enable signature syncer (default true)
      --sync.epochs uint                                 Number of recent epochs to sync from peers on startup (default 5)
      --sync.period duration                             Signature sync period (default 5s)
      --sync.timeout duration                            Signature sync timeout (default 1m0s)
      --tracing.enabled                                  Enable distributed tracing
      --tracing.endpoint string                          OTLP endpoint for tracing (e.g., Jaeger) (default "localhost:4317")
      --tracing.sample-rate float                        Trace sampling rate (0.0 to 1.0) (default 1)
```

