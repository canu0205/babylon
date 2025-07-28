#!/bin/bash

# Simple Dual Key Type Test
echo "🔑 Testing Dual Key Types"
echo "========================="

# Configuration
HOME_DIR="./.testnet/node0/babylond"
CHAIN_ID="chain-test"
NODE="tcp://localhost:26657"
KEYRING="--keyring-backend test"
FEES="5000ubbn"

# Account addresses
DEV0_ADDR="bbn1fx944mzagwdhx0wz7k9tfztc8g3lkfk6kugnlw"  # eth_secp256k1
NODE0_ADDR="bbn1r6m2kdzzx0m6yzaafszdwv0zzk4zrd8eer2t4r" # secp256k1
TEST_ETHSECP="bbn13qe8m93tct2e6vnr6rnw4yrqfm9s9q5luwf3vk" # eth_secp256k1
TEST_SECP="bbn19gx33w9gh2xcpuqjxksv8urdekapkwae8sldna" # secp256k1

echo "Testing eth_secp256k1 → secp256k1..."
babylond tx bank send dev0 $NODE0_ADDR 50000ubbn $KEYRING --home $HOME_DIR --chain-id $CHAIN_ID --fees $FEES --yes --output json
echo ""
sleep 3

echo "Testing secp256k1 → eth_secp256k1..." 
babylond tx bank send node0 $DEV0_ADDR 50000ubbn $KEYRING --home $HOME_DIR --chain-id $CHAIN_ID --fees $FEES --yes --output json
echo ""
sleep 3

echo "Testing eth_secp256k1 → eth_secp256k1..."
babylond tx bank send dev0 $TEST_ETHSECP 50000ubbn $KEYRING --home $HOME_DIR --chain-id $CHAIN_ID --fees $FEES --yes --output json
echo ""
sleep 3

echo "Testing secp256k1 → secp256k1..."
babylond tx bank send node0 $TEST_SECP 50000ubbn $KEYRING --home $HOME_DIR --chain-id $CHAIN_ID --fees $FEES --yes --output json
echo ""


