require("@nomicfoundation/hardhat-toolbox");

/** @type import('hardhat/config').HardhatUserConfig */
module.exports = {
  solidity: "0.8.19",
  networks: {
    babylon: {
      url: "http://localhost:8545", // Babylon EVM JSON-RPC endpoint
      chainId: 6901,
      accounts: [
        // Actual private keys from testnet keyring (without 0x prefix)
        "8a36c69d940a92fcea94b36d0f2928c7a0ee19a90073eda769693298dfa9603b", // dev0 (eth_secp256k1)
        "33abf2380893085b45dbf560129dfc234346c2ed26f516cd224b2faf98b6981d"  // node0 (secp256k1) - leading 0 removed
      ]
    }
  }
};