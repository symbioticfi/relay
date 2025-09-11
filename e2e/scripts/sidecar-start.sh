#!/bin/sh

apk add --no-cache jq

echo "Waiting for relay_contracts.json file..."
until [ -f /deploy-data/relay_contracts.json ]; do sleep 2; done

DRIVER_ADDRESS=$(jq -r '.driver.addr' /deploy-data/relay_contracts.json)
echo "Driver address from relay_contracts.json: $DRIVER_ADDRESS"

cat > /tmp/sidecar.yaml << EOFCONFIG
driver:
  chain-id: 31337
  address: "$DRIVER_ADDRESS"
log-level: "debug"
log-mode: "json"
signer: true
chains:
  - "http://anvil:8545"
  - "http://anvil-settlement:8546"
bootnodes:
  - /dns4/relay-sidecar-1/tcp/8880/p2p/16Uiu2HAmFUiPYAJ7bE88Q8d7Kznrw5ifrje2e5QFyt7uFPk2G3iR
http-listen: ":8080"
p2p-listen: "/ip4/0.0.0.0/tcp/8880"
enable-mdns: true
EOFCONFIG

exec /app/relay_sidecar --config /tmp/sidecar.yaml --secret-keys "$1" --storage-dir "$2" $3