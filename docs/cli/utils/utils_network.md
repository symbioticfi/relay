# `utils network` Command Reference

## utils network

Network tool

### Options

```
  -c, --chains strings                                   Chains rpc url, comma separated
      --config string                                    Path to config file with external-voting-power-providers settings
      --driver.address string                            Driver contract address
      --driver.chainid uint                              Driver contract chain id
  -e, --epoch uint                                       Network epoch to fetch info
      --external-voting-power-providers stringToString   External voting power providers mapping in format 'providerId=url' (e.g. '0x11223344556677889900=127.0.0.1:50051') (default [])
  -h, --help                                             help for network
```

### Options inherited from parent commands

```
      --log.level string   log level(info, debug, warn, error) (default "info")
      --log.mode string    log mode(pretty, text, json) (default "text")
```

### SEE ALSO

* [utils](utils.md)	 - Utils tool
* [utils network generate-genesis](utils_network_generate-genesis.md)	 - Generate genesis validator set header
* [utils network info](utils_network_info.md)	 - Print network information

