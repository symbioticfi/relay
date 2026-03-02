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
  -c, --chains strings                               Chains rpc url, comma separated
      --driver.address string                        Driver contract address
      --driver.chainid uint                          Driver contract chain id
  -e, --epoch uint                                   Network epoch to fetch info
      --external-voting-power-provider stringArray   External voting power provider config in format 'id=<id>,url=<url>[,secure=<bool>][,ca-cert-file=<path>][,server-name=<name>][,timeout=<duration>][,headers=<k:v|k2:v2>]'
      --log.level string                             log level(info, debug, warn, error) (default "info")
      --log.mode string                              log mode(pretty, text, json) (default "text")
```

### SEE ALSO

* [utils network](utils_network.md)	 - Network tool

