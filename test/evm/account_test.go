package evm

import (
	"crypto/rand"
	"testing"

	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	"github.com/cosmos/evm/crypto/ethsecp256k1"
)

func TestAccount(t *testing.T) {
	// Generate 32 bytes of entropy for consistent seed
	seed := make([]byte, 32)
	_, err := rand.Read(seed)
	if err != nil {
		t.Fatal(err)
	}

	// Generate cosmos secp256k1 key from seed
	cosmosPriv := secp256k1.GenPrivKeyFromSecret(seed)

	// Use the same cosmosPriv as well since there's no Gen function from seed to priv in ethsecp256k1
	ethPriv := &ethsecp256k1.PrivKey{Key: cosmosPriv.Bytes()}

	cosmosPub := cosmosPriv.PubKey()
	ethPub := ethPriv.PubKey()

	cosmosAddr := cosmosPub.Address()
	ethAddr := ethPub.Address()

	t.Logf("Seed: %x", seed)
	t.Logf("Cosmos pubkey: %s", cosmosPub.String())
	t.Logf("Ethereum pubkey: %s", ethPub.String())
	t.Logf("Cosmos secp256k1 address: %s", cosmosAddr.String())
	t.Logf("Ethereum ethsecp256k1 address: %s", ethAddr.String())

	// Verify they are different (they should be since they use different key types)
	if cosmosAddr.String() == ethAddr.String() {
		t.Error("Addresses should be different when using different key types")
	} else {
		t.Log("✓ Addresses are different as expected")
	}
}
