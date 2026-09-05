# Security findings follow-up

This follow-up addresses the reproduced defects from the second review:

- P2P signature sync bounds request IDs and database lookups, including misses, and checks cancellation.
- Aggregation catch-up budgets every inspected record and retains a cursor between cycles.
- HTTP mutation endpoints reject foreign origins, cross-site requests and non-JSON content types. HTTP request bodies are bounded.
- Signature-request pagination has a 4 MiB encoded-value budget in both storage backends, before value copying/decoding. Oversized individual stored records return an error.
- Keystore reads do not mutate disk or shared state. Mutations are serialized and atomically replace the file; failed writes leave the live store unchanged.
- Commit catch-up cannot raise its cursor based on incomplete settlement RPC results. The status cursor advances only through consecutive locally checked epochs.
- Derivation checks already committed headers before returning a validator set. Startup status checks precede signing workers.
- Invalid or overflowing voting powers return errors instead of panicking.

## Upgrading an old keystore

Old keystores used the empty password for individual entries, despite a password-protected store. Runtime can read these files (including read-only mounts), but this does **not** secure the original file at rest.

On a writable copy, run:

```sh
relay_utils keys migrate --path /secure/path/keystore.jks
```

The command prompts for the existing password and rewrites **all** private-key entries with that password using an atomic mode-0600 replacement. Replace the mounted secret with the migrated file. Treat old copies/backups as sensitive; rotate keys if an old file may have been exposed. Do not put passwords in command-line arguments.

## Protocol and deployment boundaries

- Version 1 keeps its existing voting-power hash leaves. Changing their padding silently changes historical header hashes and scheduler assignments. Finding 32 requires a coordinated protocol/version upgrade; this branch intentionally does not activate that breaking change. Numeric validation and fixed-width serialization remain in place.
- Signing APIs are trusted interfaces, not public anonymous services. Origin checks do not authenticate native callers. Findings 1/34 require network isolation or an authenticated proxy in deployments exposing the API.
- The driver-derived P2P swarm key is public network separation, not membership authentication (14). Do not use it as an authorization boundary.
- Self-hosted runner admission must be enforced outside PR-controlled workflow YAML (16).
- RPC URLs must be trusted and mapped to the intended chains by deployment configuration (26); a reported chain ID is not RPC authentication.

These deployment/protocol conditions are not claimed as resolved by local unit or E2E tests.

## Validation

- Base: remote main `51b1b538d250186da3e406af5bee484fc37db54f`, rechecked before publication.
- Go 1.26.1: build, all 36 non-E2E test packages, and race tests for the changed use cases and both repositories passed. Lint: zero issues.
- Full Bbolt/simple-verifier Docker E2E: 19 top-level scenarios passed in 934.938 seconds (4 operators, 1 committer, 1 aggregator, 60-second epochs, 2 finality blocks).
- Final-image restart on existing Bbolt data: API connectivity, validator-set retrieval, epoch progression and non-header signing passed (4 scenarios, 42.446 seconds), after the no-settlement cursor regression was added.
- Both storage backends have byte-budget regression tests. The full Badger, multi-aggregator and ZK E2E profiles were not rerun in this follow-up.
- Docker Hub metadata requests for the repository Dockerfile timed out. E2E used the same source built against cached Go 1.26.5 / Alpine 3.23 images; the repository Dockerfile was not changed. This is not a successful validation of its Go 1.27 / Alpine latest build.
- An initial E2E attempt used inconsistent epoch-duration environment values and was discarded. The successful run used matching environment values after recreating the network.
