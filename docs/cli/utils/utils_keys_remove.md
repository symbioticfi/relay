# `utils keys remove` Command Reference

## utils keys remove

Remove key

```
utils keys remove [flags]
```

### Options

```
      --chain-id int16   chain id for evm keys, use 0 for default key for all chains (default -1)
      --evm              use evm namespace keys
  -h, --help             help for remove
      --key-tag uint8    key tag for relay keys (default 255)
      --p2p              use p2p key
      --relay            use relay namespace keys
```

### Options inherited from parent commands

```
      --log.level string   log level(info, debug, warn, error) (default "info")
      --log.mode string    log mode(pretty, text, json) (default "text")
      --password string    Keystore password
  -p, --path string        Path to keystore (default "./keystore.jks")
```

### SEE ALSO

* [utils keys](utils_keys.md)	 - Keys tool

