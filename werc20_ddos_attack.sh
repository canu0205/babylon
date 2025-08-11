#!/bin/bash

# DDoS attack vector test
# install babylond
make install

# init babylon evm testnet
rm -rf ./.testnet
babylond testnet \
  --v                     1 \
  --output-dir            ./.testnet \
  --starting-ip-address   192.168.10.2 \
  --keyring-backend       test \
  --chain-id              chain-test \
  --additional-sender-account true

# start babylon evm testnet
babylond start --home ./.testnet/node0/babylond --log_level info --bls-password-file ./.testnet/node0/babylond/config/bls_password.txt > dev/null 2>&1 & BABYLOND_PID=$!

sleep 10

echo "dev0 account BABY balance:"
babylond q evm balance-bank 0x498B5AeC5D439b733dC2F58AB489783A23FB26dA ubbn --keyring-backend test --home ./.testnet/node0/babylond/

echo "dev0 account WBABY balance:"
babylond q evm balance-erc20 0x498B5AeC5D439b733dC2F58AB489783A23FB26dA 0xD4949664cD82660AaE99bEdc034a0deA8A0bd517 --keyring-backend test --home ./.testnet/node0/babylond/
WBABY_BEFORE=$(babylond q evm balance-erc20 0x498B5AeC5D439b733dC2F58AB489783A23FB26dA 0xD4949664cD82660AaE99bEdc034a0deA8A0bd517 --keyring-backend test --home ./.testnet/node0/babylond/ | grep "amount:" | awk '{print $2}')

echo "transfer WBABY from dev0 to node0 with 9.5WBABY and 1BABY tx fee"
TX_HASH=$(cast send 0xD4949664cD82660AaE99bEdc034a0deA8A0bd517 \
  "transfer(address,uint256)" 0x4cfAB673bE7704Fff569e0ccA8464eB6bBc3e01f 9500000 \
  --private-key 8A36C69D940A92FCEA94B36D0F2928C7A0EE19A90073EDA769693298DFA9603B \
  --rpc-url http://localhost:8545 \
  --gas-price 20000000000 \
  --gas-limit 50000000 \
  --async)

sleep 5

cast receipt $TX_HASH

STATUS=$(cast receipt $TX_HASH | grep "status" | awk '{print $2}')

if [[ "$STATUS" == "1" ]]; then
    echo "✅ Transaction SUCCEEDED"
elif [[ "$STATUS" == "0" ]]; then
    echo "❌ Transaction FAILED - DDoS attack confirmed!"
else
    echo "⚠️  Unknown status: $STATUS"
fi

echo "check dev0 account BABY WBABY balance:"
babylond q evm balance-erc20 0x498B5AeC5D439b733dC2F58AB489783A23FB26dA 0xD4949664cD82660AaE99bEdc034a0deA8A0bd517 --keyring-backend test --home ./dev/babylon/.testnet/node0/babylond/
WBABY_AFTER=$(babylond q evm balance-erc20 0x498B5AeC5D439b733dC2F58AB489783A23FB26dA 0xD4949664cD82660AaE99bEdc034a0deA8A0bd517 --keyring-backend test --home ./.testnet/node0/babylond/ | grep "amount:" | awk '{print $2}')

ATTACK_COST=$((WBABY_BEFORE - WBABY_AFTER))
echo "DDoS attack cost: $ATTACK_COST (ubbn)"

kill $BABYLOND_PID
echo "babylond stopped"