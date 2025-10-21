#!/bin/sh

apk add --no-cache jq

echo 'Waiting for deployment completion...'
until [ -f /deploy-data/deployment-complete.marker ]; do sleep 2; done

echo 'Waiting for relay contracts file...'
until [ -f /deploy-data/relay_contracts.json ]; do sleep 2; done

DRIVER_ADDRESS=$(jq -r '.driver.addr' /deploy-data/relay_contracts.json)
echo "Driver address from relay_contracts.json: $DRIVER_ADDRESS"

MAX_RETRIES=50
RETRY_DELAY=2
attempt=1

while [ $attempt -le $MAX_RETRIES ]; do
    echo "Attempt $attempt of $MAX_RETRIES: Generating network genesis..."
    
    if /app/relay_utils network \
            --chains http://anvil:8545,http://anvil-settlement:8546 \
            --driver.address "$DRIVER_ADDRESS" \
            --driver.chainid 31337 \
          generate-genesis \
            --commit \
            --secret-keys 31337:0xac0974bec39a17e36ba4a6b4d238ff944bacb478cbed5efcae784d7bf4f2ff80,31338:0xac0974bec39a17e36ba4a6b4d238ff944bacb478cbed5efcae784d7bf4f2ff80; then
        echo 'Genesis generation completed successfully!'
        
        # Create genesis completion marker
        echo "$(date): Genesis generation completed successfully" > /deploy-data/genesis-complete.marker
        echo "Genesis completion marker created"

        echo "Waiting few seconds before exiting..."
        sleep 5
        
        exit 0
    else
        echo "Genesis generation failed on attempt $attempt"
        if [ $attempt -lt $MAX_RETRIES ]; then
            echo "Waiting $RETRY_DELAY seconds before retry..."
            sleep $RETRY_DELAY
        else
            echo "All $MAX_RETRIES attempts failed. Exiting with error."
            exit 1
        fi
        attempt=$((attempt + 1))
    fi
done 