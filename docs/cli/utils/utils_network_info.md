# `utils network info` Command Reference

## utils network info

Print network information

```
utils network info [flags]
```

### Options

```
  -a, --addresses         Print addresses
  -h, --help              help for info
  -s, --settlement        Print settlement info
  -v, --validators        Print compact validators info
  -V, --validators-full   Print full validators info
```

### Options inherited from parent commands

```
  -c, --chains strings                                   Chains rpc url, comma separated
      --config string                                    Path to config file with external-voting-power-providers settings
      --driver.address string                            Driver contract address
      --driver.chainid uint                              Driver contract chain id
  -e, --epoch uint                                       Network epoch to fetch info
      --external-voting-power-providers stringToString   External voting power providers mapping in format 'providerId=url' (e.g. '0x11223344556677889900=127.0.0.1:50051') (default [])
      --log.level string                                 log level(info, debug, warn, error) (default "info")
      --log.mode string                                  log mode(pretty, text, json) (default "text")
```

### SEE ALSO

* [utils network](utils_network.md)	 - Network tool

