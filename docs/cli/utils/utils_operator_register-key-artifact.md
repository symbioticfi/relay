# `utils operator register-key-artifact` Command Reference

## utils operator register-key-artifact

Build operator key registration artifact without submitting a transaction

```
utils operator register-key-artifact [flags]
```

### Options

```
  -h, --help                       help for register-key-artifact
      --key-tag uint8              key tag (default 255)
      --operator-address string    operator address used for artifact generation
      --password string            Keystore password
  -p, --path string                Path to keystore (default "./keystore.jks")
```

### Options inherited from parent commands

```
  -c, --chains strings                  Chains rpc url, comma separated
      --driver.address string           Driver contract address
      --driver.chainid uint             Driver contract chain id
      --log.level string                log level(info, debug, warn, error) (default "info")
      --log.mode string                 log mode(pretty, text, json) (default "text")
      --voting-provider-chain-id uint   Voting power provider chain id
```

### SEE ALSO

* [utils operator](utils_operator.md)	 - Operator tool
