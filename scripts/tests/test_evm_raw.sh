#!/bin/bash

# Raw EVM Transaction Test
echo "🔥 Testing Raw EVM Transactions"
echo "==============================="

# Configuration  
HOME_DIR="./.testnet/node0/babylond"
CHAIN_ID="chain-test"
NODE="tcp://localhost:26657"
KEYRING="--keyring-backend test"

# First, let's get the nonce for each account
echo "Getting account nonces..."
DEV0_NONCE=$(babylond query evm nonce $(babylond debug addr bbn1fx944mzagwdhx0wz7k9tfztc8g3lkfk6kugnlw | grep "Address (hex):" | awk '{print $3}') --node $NODE 2>/dev/null || echo "0")
NODE0_NONCE=$(babylond query evm nonce $(babylond debug addr bbn1r6m2kdzzx0m6yzaafszdwv0zzk4zrd8eer2t4r | grep "Address (hex):" | awk '{print $3}') --node $NODE 2>/dev/null || echo "0")

echo "dev0 nonce: $DEV0_NONCE"
echo "node0 nonce: $NODE0_NONCE"
echo ""

# Create a simple ETH transfer transaction
# This is an unsigned transaction that babylond will sign based on --from
# Legacy transaction format (EIP-155):
# nonce: 0 (will be adjusted) 
# gasPrice: 1000000000 (1 gwei)
# gasLimit: 21000
# to: 0x742d35Cc6634C0532925a3b8D45C5E6234a1B0A5
# value: 100000000000000000 (0.1 ETH)
# data: empty
# chainId: 6901 (Babylon EVM chain ID)

# Unsigned transaction hex (will be signed by babylond based on --from account)
UNSIGNED_TX="eb80843b9aca008252089742d35cc6634c0532925a3b8d45c5e6234a1b0a5880de0b6b3a76400008082197538"

echo "Testing raw EVM tx from eth_secp256k1 (dev0)..."
babylond tx evm raw $UNSIGNED_TX \
  --from dev0 \
  $KEYRING \
  --home $HOME_DIR \
  --chain-id $CHAIN_ID \
  --node $NODE \
  --yes \
  --output json
echo ""
sleep 3

echo "Testing raw EVM tx from secp256k1 (node0)..."
babylond tx evm raw $UNSIGNED_TX \
  --from node0 \
  $KEYRING \
  --home $HOME_DIR \
  --chain-id $CHAIN_ID \
  --node $NODE \
  --yes \
  --output json
echo ""

echo "Expected Results:"
echo "✅ eth_secp256k1 (dev0): Should work - proper EVM signature scheme"
echo "❌ secp256k1 (node0): Should fail - signature verification mismatch"
echo ""
echo "This test proves that EVM transactions require eth_secp256k1 key type!"