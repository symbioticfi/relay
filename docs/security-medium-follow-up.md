# Security notes

## Upgrading an old keystore

Old keystores used the empty password for individual entries, despite a password-protected store. Runtime can read these files (including read-only mounts), but this does **not** secure the original file at rest.

On a writable copy, run:

```sh
relay_utils keys migrate --path /secure/path/keystore.jks
```

The command prompts for the existing password and rewrites **all** private-key entries with that password using an atomic mode-0600 replacement. Replace the mounted secret with the migrated file. Treat old copies/backups as sensitive; rotate keys if an old file may have been exposed. Do not put passwords in command-line arguments.

## Protocol and deployment boundaries

- Version 1 keeps its existing voting-power hash leaves. Changing their padding silently changes historical header hashes and scheduler assignments. Updating this encoding requires a coordinated protocol/version upgrade. Numeric validation and fixed-width serialization remain in place.
- Signing APIs are trusted interfaces, not public anonymous services. Origin checks do not authenticate native callers. Use network isolation or an authenticated proxy when exposing the API.
- The driver-derived P2P swarm key is public network separation, not membership authentication. Do not use it as an authorization boundary.
- Self-hosted runner admission must be enforced outside PR-controlled workflow YAML.
- RPC URLs must be trusted and mapped to the intended chains by deployment configuration; a reported chain ID is not RPC authentication.
