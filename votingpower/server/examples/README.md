# Voting Power Server Example

This example starts a mock external voting power provider server.

## Run

```bash
cd votingpower/server/examples
go run main.go
```

It listens on `:9090` and implements:
- `votingpower.v1.VotingPowerProviderService/GetVotingPowersAt`
- `grpc.health.v1.Health/Check`
