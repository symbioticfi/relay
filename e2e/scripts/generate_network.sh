#!/bin/bash

# Symbiotic Network Infrastructure Generator
# This script generates a Docker Compose setup for blockchain infrastructure
# (anvil chains, deployer, genesis-generator) with configurable parameters
#
# Environment Variables:
#   OPERATORS        - Number of operators (default: 4, max: 999)
#   COMMITERS        - Number of commiters (default: 1)
#   AGGREGATORS      - Number of aggregators (default: 1)
#   VERIFICATION_TYPE - Verification type: 0=BLS-BN254-ZK, 1=BLS-BN254-SIMPLE (default: 1)
#   EPOCH_TIME       - Time for new epochs in relay network (default: 30)
#   BLOCK_TIME       - Block time in seconds for anvil interval mining (default: 1)
#   FINALITY_BLOCKS  - Number of blocks for finality (default: 1)
#
# Example usage:
#   OPERATORS=6 COMMITERS=2 AGGREGATORS=1 VERIFICATION_TYPE=0 EPOCH_TIME=32 BLOCK_TIME=2 ./generate_network.sh

set -e

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# Default values
DEFAULT_OPERATORS=4
DEFAULT_COMMITERS=1
DEFAULT_AGGREGATORS=1
DEFAULT_VERIFICATION_TYPE=1  # BLS-BN254-SIMPLE
DEFAULT_EPOCH_TIME=30
DEFAULT_BLOCK_TIME=1
DEFAULT_FINALITY_BLOCKS=2
MAX_OPERATORS=999
DEFAULT_COMMITTER_SLOT_DURATION=10


print_status() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

print_header() {
    echo -e "${BLUE}================================${NC}"
    echo -e "${BLUE}$1${NC}"
    echo -e "${BLUE}================================${NC}"
}


validate_number() {
    local num=$1
    local name=$2
    if ! [[ "$num" =~ ^[0-9]+$ ]] || [ "$num" -lt 1 ]; then
        print_error "$name must be a positive integer"
        exit 1
    fi
}

validate_verification_type() {
    local type=$1
    local name=$2
    if ! [[ "$type" =~ ^[0-1]$ ]]; then
        print_error "$name must be 0 (BLS-BN254-ZK) or 1 (BLS-BN254-SIMPLE)"
        exit 1
    fi
}

get_config_from_env() {
    echo
    print_header "Symbiotic Network Configuration"
    echo
    
    # Read from environment variables with defaults
    operators=${OPERATORS:-$DEFAULT_OPERATORS}
    commiters=${COMMITERS:-$DEFAULT_COMMITERS}
    aggregators=${AGGREGATORS:-$DEFAULT_AGGREGATORS}
    verification_type=${VERIFICATION_TYPE:-$DEFAULT_VERIFICATION_TYPE}
    epoch_size=${EPOCH_TIME:-$DEFAULT_EPOCH_TIME}
    block_time=${BLOCK_TIME:-$DEFAULT_BLOCK_TIME}
    finality_blocks=${FINALITY_BLOCKS:-$DEFAULT_FINALITY_BLOCKS}
    committer_slot_duration=${COMMITTER_SLOT_DURATION:-$DEFAULT_COMMITTER_SLOT_DURATION}
    
    # Validate inputs
    validate_number "$operators" "Number of operators (OPERATORS env var)"
    validate_number "$commiters" "Number of commiters (COMMITERS env var)"
    validate_number "$aggregators" "Number of aggregators (AGGREGATORS env var)"
    validate_verification_type "$verification_type" "Verification type (VERIFICATION_TYPE env var)"
    validate_number "$epoch_size" "Epoch size (EPOCH_TIME env var)"
    validate_number "$block_time" "Block time (BLOCK_TIME env var)"
    validate_number "$finality_blocks" "Finality blocks (FINALITY_BLOCKS env var)"
    validate_number "$committer_slot_duration" "Committer slot duration (COMMITTER_SLOT_DURATION env var)"

    # Validate that commiters + aggregators <= operators
    total_special_roles=$((commiters + aggregators))
    if [ "$total_special_roles" -gt "$operators" ]; then
        print_error "Total commiters ($commiters) + aggregators ($aggregators) cannot exceed total operators ($operators)"
        exit 1
    fi

    if [ "$operators" -gt $MAX_OPERATORS ]; then
        print_error "Maximum $MAX_OPERATORS operators supported. Requested: $operators"
        exit 1
    fi
    
    # Convert verification type to description
    local verification_desc
    case $verification_type in
        0) verification_desc="BLS-BN254-ZK" ;;
        1) verification_desc="BLS-BN254-SIMPLE" ;;
    esac
    
    print_status "Configuration (from environment variables):"
    print_status "  Operators: $operators (OPERATORS=${OPERATORS:-default})"
    print_status "  Committers: $commiters (COMMITERS=${COMMITERS:-default})"
    print_status "  Aggregators: $aggregators (AGGREGATORS=${AGGREGATORS:-default})"
    print_status "  Regular signers: $((operators - total_special_roles))"
    print_status "  Verification type: $verification_type ($verification_desc) (VERIFICATION_TYPE=${VERIFICATION_TYPE:-default})"
    print_status "  Epoch size: $epoch_size slots (EPOCH_TIME=${EPOCH_TIME:-default})"
    print_status "  Block time: $block_time seconds (BLOCK_TIME=${BLOCK_TIME:-default})"
    print_status "  Finality blocks: $finality_blocks (FINALITY_BLOCKS=${FINALITY_BLOCKS:-default})"
    print_status "  Committer slot duration: $committer_slot_duration seconds (COMMITTER_SLOT_DURATION=${COMMITTER_SLOT_DURATION:-default})"
}

# Function to generate Docker Compose file
generate_docker_compose() {
    local operators=$1
    local commiters=$2
    local aggregators=$3
    local verification_type=$4
    local epoch_size=$5
    local block_time=$6
    local finality_blocks=$7
    local committer_slot_duration=$8
    
    local network_dir="temp-network"

    if [ -d "$network_dir" ]; then
        print_status "Cleaning up existing $network_dir directory..."
        rm -rf "$network_dir"
    fi

    mkdir -p "$network_dir/deploy-data"
    # Ensure deploy-data directory is writable for Docker containers
    chmod 777 "$network_dir/deploy-data"

    # Create cache and broadcast directories with proper permissions
    print_status "Creating out, cache and broadcast directories..."
    mkdir -p "$network_dir/out" "$network_dir/cache" "$network_dir/broadcast"
    chmod 777 "$network_dir/out" "$network_dir/cache" "$network_dir/broadcast"

    local deploy_config_src="contracts/script/my-relay-deploy.toml"
    local deploy_config_dst="$network_dir/my-relay-deploy.toml"
    if [ ! -f "$deploy_config_src" ]; then
        print_error "Deployment config not found at $deploy_config_src"
        exit 1
    fi
    print_status "Copying deployment config to $deploy_config_dst"
    cp "$deploy_config_src" "$deploy_config_dst"

    for i in $(seq 1 $operators); do
        local storage_dir="$network_dir/data-$(printf "%02d" $i)"
        mkdir -p "$storage_dir"
        # Make sure the directory is writable
        chmod 777 "$storage_dir"
    done

    local anvil_port=8545
    local anvil_settlement_port=8546
    local relay_start_port=8081
    local sum_start_port=9091
    
    # Calculate timestamp as current unix timestamp + 5 seconds
    local timestamp=$(($(date +%s) + 5))
    
    cat > "$network_dir/docker-compose.yml" << EOF
services:
  # Main Anvil local Ethereum network (Chain ID: 31337)
  anvil:
    image: ghcr.io/foundry-rs/foundry:v1.4.3
    container_name: symbiotic-anvil
    entrypoint: ["anvil"]
    command: "--port 8545 --chain-id 31337 --timestamp $timestamp --auto-impersonate --slots-in-an-epoch $finality_blocks --accounts 10 --balance 10000 --gas-limit 30000000"
    environment:
      - ANVIL_IP_ADDR=0.0.0.0
    ports:
      - "8545:8545"
    networks:
      - symbiotic-network
    healthcheck:
      test: ["CMD", "cast", "client", "--rpc-url", "http://localhost:8545"]
      interval: 2s
      timeout: 1s
      retries: 10

  # Settlement Anvil local Ethereum network (Chain ID: 31338)
  anvil-settlement:
    image: ghcr.io/foundry-rs/foundry:v1.4.3
    container_name: symbiotic-anvil-settlement
    entrypoint: ["anvil"]
    command: "--port 8546 --chain-id 31338 --timestamp $timestamp --auto-impersonate --slots-in-an-epoch $finality_blocks --accounts 10 --balance 10000 --gas-limit 30000000"
    environment:
      - ANVIL_IP_ADDR=0.0.0.0
    ports:
      - "8546:8546"
    networks:
      - symbiotic-network
    healthcheck:
      test: ["CMD", "cast", "client", "--rpc-url", "http://localhost:8546"]
      interval: 2s
      timeout: 1s
      retries: 10

  # Contract deployment service for main chain
  deployer:
    build:
      context: ..
      dockerfile: scripts/deployer.Dockerfile
    image: symbiotic-deployer
    container_name: symbiotic-deployer
    user: "1000:1000"
    volumes:
      - ../contracts/:/app
      - ../scripts:/app/deploy-scripts
      - ../temp-network:/app/temp-network
      - ./cache:/app/cache
      - ./broadcast:/app/broadcast
      - ./out:/app/out
      - ./deploy-data:/deploy-data
      - ./my-relay-deploy.toml:/my-relay-deploy.toml
    working_dir: /app
    command: ./deploy-scripts/deploy.sh
    depends_on:
      anvil:
        condition: service_healthy
      anvil-settlement:
        condition: service_healthy
    networks:
      - symbiotic-network
    environment:
      - OPERATOR_COUNT=$operators
      - VERIFICATION_TYPE=$verification_type
      - BLOCK_TIME=$block_time
      - EPOCH_TIME=$epoch_size
      - FOUNDRY_CACHE_PATH=/tmp/.foundry-cache
      - NUM_AGGREGATORS=$aggregators
      - NUM_COMMITTERS=$commiters
      - COMMITTER_SLOT_DURATION=$committer_slot_duration

  # Genesis generation service
  genesis-generator:
    image: relay_sidecar:dev
    container_name: symbiotic-genesis-generator
    volumes:
      - ../:/workspace
      - ./deploy-data:/deploy-data
    working_dir: /workspace
    command: ./scripts/genesis-generator.sh
    depends_on:
      deployer:
        condition: service_completed_successfully
    networks:
      - symbiotic-network

EOF

    local committer_count=0
    local aggregator_count=0
    local signer_count=0
    
    # Calculate symb private key properly
    # ECDSA secp256k1 private keys must be 32 bytes (64 hex chars) and within range [1, n-1]
    # where n = 0xFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFEBAAEDCE6AF48A03BBFD25E8CD0364141
    BASE_PRIVATE_KEY=1000000000000000000

    for i in $(seq 1 $operators); do
        local port=$((relay_start_port + i - 1))
        local storage_dir="data-$(printf "%02d" $i)"
        local key_index=$((i - 1))
        
        SYMB_PRIVATE_KEY_DECIMAL=$(($BASE_PRIVATE_KEY + $key_index))
        SYMB_SECONDARY_PRIVATE_KEY_DECIMAL=$(($BASE_PRIVATE_KEY + $key_index + 10000))
        SYMB_PRIVATE_KEY_HEX=$(printf "%064x" $SYMB_PRIVATE_KEY_DECIMAL)
        SYMB_SECONDARY_PRIVATE_KEY_HEX=$(printf "%064x" $SYMB_SECONDARY_PRIVATE_KEY_DECIMAL)

        # Validate ECDSA secp256k1 private key range (must be between 1 and n-1)
        # Maximum valid key: 0xFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFEBAAEDCE6AF48A03BBFD25E8CD0364140
        if [ $SYMB_PRIVATE_KEY_DECIMAL -eq 0 ]; then
            echo "ERROR: Generated private key is zero (invalid for ECDSA)"
            exit 1
        fi
        
        # Set circuits directory parameter based on verification type
        if [ "$verification_type" = "0" ]; then
            circuits_param="/app/circuits"
        else
            circuits_param=""
        fi

        cat >> "$network_dir/docker-compose.yml" << EOF

  # Relay sidecar $i
  relay-sidecar-$i:
    image: relay_sidecar:dev
    container_name: symbiotic-relay-$i
    command:
      - sh
      - -c
      - "chmod 777 /app/$storage_dir /deploy-data 2>/dev/null || true && /workspace/scripts/sidecar-start.sh symb/0/15/0x$SYMB_PRIVATE_KEY_HEX,symb/0/11/0x$SYMB_SECONDARY_PRIVATE_KEY_HEX,symb/1/0/0x$SYMB_PRIVATE_KEY_HEX,evm/1/31337/0x$SYMB_PRIVATE_KEY_HEX,evm/1/31338/0x$SYMB_PRIVATE_KEY_HEX,p2p/1/1/$SYMB_PRIVATE_KEY_HEX /app/$storage_dir $circuits_param"
    ports:
      - "$port:8080"
    volumes:
      - ../:/workspace
      - ./$storage_dir:/app/$storage_dir
      - ./deploy-data:/deploy-data
EOF

        # Add circuits volume only if verification type is 0
        if [ "$verification_type" = "0" ]; then
            cat >> "$network_dir/docker-compose.yml" << EOF
      - ./circuits:/app/circuits
EOF
        fi

        cat >> "$network_dir/docker-compose.yml" << EOF
    depends_on:
      genesis-generator:
        condition: service_completed_successfully
    networks:
      - symbiotic-network
    restart: unless-stopped
    environment:
      - MAX_VALIDATORS=10,100
    healthcheck:
      test: ["CMD", "wget", "--quiet", "--tries=1", "--spider", "http://localhost:8080/healthz"]
      interval: 30s
      timeout: 10s
      retries: 3
      start_period: 40s

EOF
    done
    
    cat >> "$network_dir/docker-compose.yml" << EOF

networks:
  symbiotic-network:
    driver: bridge

EOF
}


# Main execution
main() {
    print_header "Symbiotic Network Generator"

    # Check if required tools are available
    if ! command -v docker &> /dev/null; then
        print_error "Docker is not installed or not in PATH"
        exit 1
    fi

    if ! command -v docker-compose &> /dev/null && ! docker compose version &> /dev/null; then
        print_error "Docker Compose is not installed or not in PATH"
        exit 1
    fi
    
    get_config_from_env
    

    print_status "Generating Docker Compose configuration..."
    print_status "Creating $operators new operator accounts..."
    generate_docker_compose "$operators" "$commiters" "$aggregators" "$verification_type" "$epoch_size" "$block_time" "$finality_blocks" "$committer_slot_duration"
}

main "$@" 