# `utils operator` Command Reference

## utils operator

Operator tool

### Options

```
  -c, --chains strings                  Chains rpc url, comma separated
      --driver.address string           Driver contract address
      --driver.chainid uint             Driver contract chain id
  -h, --help                            help for operator
      --voting-provider-chain-id uint   Voting power provider chain id
```

### Options inherited from parent commands

```
      --log.level string   log level(info, debug, warn, error) (default "info")
      --log.mode string    log mode(pretty, text, json) (default "text")
```

### SEE ALSO

* [utils](utils.md)	 - Utils tool
* [utils operator info](utils_operator_info.md)	 - Print operator information
* [utils operator invalidate-old-signatures](utils_operator_invalidate-old-signatures.md)	 - Invalidate old signatures for operator
* [utils operator register-key](utils_operator_register-key.md)	 - Register operator key in key registry
* [utils operator register-operator](utils_operator_register-operator.md)	 - Register operator on-chain via VotingPowerProvider
* [utils operator register-operator-with-signature](utils_operator_register-operator-with-signature.md)	 - Generate EIP-712 signature for operator registration
* [utils operator unregister-operator](utils_operator_unregister-operator.md)	 - Unregister operator on-chain via VotingPowerProvider
* [utils operator unregister-operator-with-signature](utils_operator_unregister-operator-with-signature.md)	 - Generate EIP-712 signature for operator unregistration

