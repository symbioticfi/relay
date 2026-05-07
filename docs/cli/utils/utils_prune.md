# `utils prune` Command Reference

## utils prune

Prune old epoch data from the relay storage (offline)

### Synopsis

Opens the relay storage directory in offline mode (no live writers) and prunes valset / proof / signature entities older than the configured retention. Optionally compacts the underlying database file when --compact is set.

The relay sidecar must be stopped while this command runs. Both bbolt and badger acquire an exclusive file-lock on open, so a still-running sidecar will surface as a clear open-error rather than silently corrupting state.

WARNING: for speed, bbolt is opened with NoSync and NoFreelistSync. A SIGKILL or power loss mid-prune will not corrupt the file (bbolt MVCC is durable on transaction boundaries), but the next open may be slower as the freelist is rebuilt. Let this command finish; do not kill -9.

```
utils prune [flags]
```

### Options

```
      --badger.flatten-workers int        Number of parallel workers for badger Flatten (only when --compact is set) (default 4)
      --compact                           After pruning, compact the database file (bbolt: rewrite; badger: Flatten + value log GC)
  -h, --help                              help for prune
      --prune-batch-size int              Number of request IDs to delete per database transaction (larger = faster but holds writer lock longer) (default 1000)
      --retention.proof-epochs uint       Keep this many most-recent epochs of aggregation proofs (0 = skip)
      --retention.signature-epochs uint   Keep this many most-recent epochs of signatures (0 = skip)
      --retention.valset-epochs uint      Keep this many most-recent epochs of valset data (0 = skip)
      --storage-dir string                Directory containing the relay storage (badger or bbolt) (default ".data")
```

### Options inherited from parent commands

```
      --log.level string   log level(info, debug, warn, error) (default "info")
      --log.mode string    log mode(pretty, text, json) (default "text")
```

### SEE ALSO

* [utils](utils.md)	 - Utils tool

