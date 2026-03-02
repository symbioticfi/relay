# `utils operator info` Command Reference

## utils operator info

Print operator information

```
utils operator info [flags]
```

### Options

```
  -e, --epoch uint                                   Network epoch to fetch info
      --external-voting-power-provider stringArray   External voting power provider config in format 'id=<id>,url=<url>[,secure=<bool>][,ca-cert-file=<path>][,server-name=<name>][,timeout=<duration>][,headers=<k:v|k2:v2>]'
  -h, --help                                         help for info
      --key-tag uint8                                key tag (default 255)
      --password string                              Keystore password
  -p, --path string                                  Path to keystore (default "./keystore.jks")
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

