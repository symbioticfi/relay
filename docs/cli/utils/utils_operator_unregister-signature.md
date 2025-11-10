# `utils operator unregister-signature` Command Reference

## utils operator unregister-signature

Generate EIP-712 signature for operator unregistration

```
utils operator unregister-signature [flags]
```

### Options

```
  -h, --help                       help for unregister-signature
      --secret-keys secretKeyMap   Secret key for signing in format 'chainId:key' (e.g. '1:0xabc')
```

### Options inherited from parent commands

```
  -c, --chains strings          Chains rpc url, comma separated
      --driver.address string   Driver contract address
      --driver.chainid uint     Driver contract chain id
      --log.level string        log level(info, debug, warn, error) (default "info")
      --log.mode string         log mode(pretty, text, json) (default "text")
```

### SEE ALSO

* [utils operator](utils_operator.md)	 - Operator tool

