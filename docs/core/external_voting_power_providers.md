# External Voting Power Providers

## Description

External voting power providers allow relay to derive operator voting power from off-chain gRPC services instead of EVM `VotingPowerProvider` contracts.

Routing is selected by on-chain voting power provider `chainId`:

- `4_000_000_000 .. 4_100_000_000` (inclusive): external gRPC provider
- all other chain IDs: EVM provider contract

This routing is used during validator set derivation (see [Validator Set Derivation](./valset_derivation.md)).

## Process Overview

1. Relay reads voting power providers from on-chain network config.
2. For each provider:
   - external range `chainId` -> query external gRPC service
   - non-external `chainId` -> query EVM provider
3. Provider queries run in parallel with shared limit `10`.
4. Relay aggregates provider outputs into validator voting power.

Derivation is fail-closed:

- any provider fetch error fails derivation for that epoch
- external provider referenced on-chain but missing in local config fails derivation

## External Provider Identity

External provider lookup uses:

1. on-chain `CrossChainAddress.ChainId` (must be in reserved external range)
2. provider ID encoded in `CrossChainAddress.Address`

Provider ID is the first `10` bytes of the provider address (`20` hex chars).  
Local relay config maps provider ID to gRPC endpoint and transport/auth settings.

## gRPC Contract

External services must implement:

- `votingpower.v1.VotingPowerProviderService/GetVotingPowersAt`

See:

- `votingpower/proto/v1/votingpower.proto`
- `docs/votingpower/v1/doc.md`

Input:

- `timestamp` (`uint64`)

Output rows:

- `operator` (hex EVM address)
- `voting_power` (decimal string, non-negative integer)

Relay behavior for response parsing:

- empty list is valid
- duplicate operators in one response are merged by sum
- output is sorted by operator address for determinism
- invalid operator or invalid/negative voting power fails request

## Relay Configuration

Configure external mappings in relay config:

- `external-voting-power-providers`

Fields:

- `id` (required): provider ID (`10` bytes hex, `0x` optional)
- `url` (required): gRPC target
- `secure` (optional, default `false`): TLS enabled
- `ca-cert-file` (optional): CA PEM file
- `server-name` (optional): TLS server name override
- `headers` (optional): outbound gRPC metadata
- `timeout` (optional, default `5s`): dial/request timeout

Example:

```yaml
external-voting-power-providers:
  - id: "0x11223344556677889900"
    url: "dns:///beacon-vp:50051"
    secure: false
    # ca-cert-file: "/path/to/ca.pem"
    # server-name: "beacon-vp.internal"
    # timeout: 5s
    # headers:
    #   authorization: "Bearer <token>"
```

See:

- `example.config.yaml`

## Add / Remove Flow

To add an external provider:

1. Run external gRPC service implementing `GetVotingPowersAt`.
2. Add local relay mapping in `external-voting-power-providers`.
3. Register provider on-chain in `ValSetDriver.addVotingPowerProvider((chainId, addr))`:
   - `chainId` in `4_000_000_000..4_100_000_000`
   - `addr` with provider ID in first `10` bytes

To remove provider:

- call `ValSetDriver.removeVotingPowerProvider((chainId, addr))`

Configuration changes apply by epoch, not mid-epoch.

## Operational Notes

- duplicate provider IDs in local config are rejected at relay startup
- relay establishes gRPC connections on startup and fails startup if provider connection is not ready
- external request failures are not retried by relay
- transport is insecure by default (`secure: false`)

## CLI Usage

Utility commands can load external provider mappings from:

- `--config` (`external-voting-power-providers` section)
- `--external-voting-power-providers` flag (`providerId=url`)

See:

- `docs/cli/utils/utils_network_info.md`
- `docs/cli/utils/utils_operator_info.md`
