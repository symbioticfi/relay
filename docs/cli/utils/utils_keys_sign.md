# `utils keys sign` Command Reference

## utils keys sign

Sign a message with a relay key

```
utils keys sign [flags]
```

### Options

```
  -h, --help                 help for sign
      --key-tag uint8        key tag for relay keys (default 255)
      --message-hex string   raw message bytes to sign in hex
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

