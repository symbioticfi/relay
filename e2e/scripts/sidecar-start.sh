#!/bin/sh

DRIVER_ADDRESS=0x43C27243F96591892976FFf886511807B65a33d5

cat > /tmp/sidecar.yaml << EOFCONFIG
# Logging
log:
  level: "debug"
  mode: "json"

# API Server Configuration
api:
  listen: ":8080"
  http-gateway: true

# Metrics Configuration
metrics:
  pprof: true

# Driver Contract
driver:
  chain-id: 31337
  address: "$DRIVER_ADDRESS"

# P2P Configuration
p2p:
  listen: "/ip4/0.0.0.0/tcp/8880"
  bootnodes:
    - /dns4/relay-sidecar-1/tcp/8880/p2p/16Uiu2HAmFUiPYAJ7bE88Q8d7Kznrw5ifrje2e5QFyt7uFPk2G3iR
  dht-mode: "server"
  mdns: true

# EVM Configuration
evm:
  chains:
    - "http://anvil:8545"
    - "http://anvil-settlement:8546"
  max-calls: 30

# Retention config
retention:
  valset-epochs: 1000
  signature-epochs: 1000
  proof-epochs: 1000

sync:
  enabled: true
  period: 5s
  timeout: 1m
  epochs: 1000

pruner:
  enabled: true
  interval: 1m

tracing:
  enabled: false
  endpoint: "jaeger:4317"
  sample-rate: 1.0

EOFCONFIG

# Ensure environment variables are explicitly preserved
export MAX_VALIDATORS="${MAX_VALIDATORS:-}"

# Handle optional circuits directory parameter
if [ -n "$3" ] && [ -d "$3" ]; then
    echo "Using circuits directory: $3"
    echo "Starting relay_sidecar with MAX_VALIDATORS=$MAX_VALIDATORS"
    exec /app/relay_sidecar --config /tmp/sidecar.yaml --secret-keys "$1" --storage-dir "$2" --circuits-dir "$3"
else
    echo "No circuits directory provided or directory doesn't exist, running without circuits"
    echo "Starting relay_sidecar with MAX_VALIDATORS=$MAX_VALIDATORS"
    exec /app/relay_sidecar --config /tmp/sidecar.yaml --secret-keys "$1" --storage-dir "$2"
fi
