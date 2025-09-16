package ante_test

import (
	"github.com/babylonlabs-io/babylon/v4/app/ante"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"testing"

	icacontrollertypes "github.com/cosmos/ibc-go/v10/modules/apps/27-interchain-accounts/controller/types"
	ibctransfertypes "github.com/cosmos/ibc-go/v10/modules/apps/transfer/types"
	clienttypes "github.com/cosmos/ibc-go/v10/modules/core/02-client/types"

	btccheckpointtypes "github.com/babylonlabs-io/babylon/v4/x/btccheckpoint/types"
	btclightclient "github.com/babylonlabs-io/babylon/v4/x/btclightclient/types"
	bstypes "github.com/babylonlabs-io/babylon/v4/x/btcstaking/types"
	ftypes "github.com/babylonlabs-io/babylon/v4/x/finality/types"
)

// Benchmark IBCMsgSizeDecorator with different transaction types
func BenchmarkIBCMsgSizeDecorator_IBCTransfer_Small(b *testing.B) {
	tx := createIBCTransferTx(3)
	ctx := sdk.Context{}.WithIsCheckTx(true)
	decorator := ante.NewIBCMsgSizeDecorator()
	next := noOpNext()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := decorator.AnteHandle(ctx, tx, false, next)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkIBCMsgSizeDecorator_IBCTransfer_Large(b *testing.B) {
	tx := createIBCTransferTx(50)
	ctx := sdk.Context{}.WithIsCheckTx(true)
	decorator := ante.NewIBCMsgSizeDecorator()
	next := noOpNext()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := decorator.AnteHandle(ctx, tx, false, next)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkIBCMsgSizeDecorator_Mixed_Small(b *testing.B) {
	tx := createMixedTx(3)
	ctx := sdk.Context{}.WithIsCheckTx(true)
	decorator := ante.NewIBCMsgSizeDecorator()
	next := noOpNext()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := decorator.AnteHandle(ctx, tx, false, next)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkIBCMsgSizeDecorator_Mixed_Large(b *testing.B) {
	tx := createMixedTx(50)
	ctx := sdk.Context{}.WithIsCheckTx(true)
	decorator := ante.NewIBCMsgSizeDecorator()
	next := noOpNext()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := decorator.AnteHandle(ctx, tx, false, next)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkIBCMsgSizeDecorator_NoRelevantMsgs_Small(b *testing.B) {
	tx := createBTCHeadersTx(3)
	ctx := sdk.Context{}.WithIsCheckTx(true)
	decorator := ante.NewIBCMsgSizeDecorator()
	next := noOpNext()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := decorator.AnteHandle(ctx, tx, false, next)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkIBCMsgSizeDecorator_NoRelevantMsgs_Large(b *testing.B) {
	tx := createBTCHeadersTx(50)
	ctx := sdk.Context{}.WithIsCheckTx(true)
	decorator := ante.NewIBCMsgSizeDecorator()
	next := noOpNext()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := decorator.AnteHandle(ctx, tx, false, next)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkGetMsgsIteration_Small(b *testing.B) {
	tx := createMixedTx(3)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, msg := range tx.GetMsgs() {
			_ = msg
		}
	}
}

func BenchmarkGetMsgsIteration_Large(b *testing.B) {
	tx := createMixedTx(50)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, msg := range tx.GetMsgs() {
			_ = msg
		}
	}
}

// Helper function to create no-op next handler
func noOpNext() sdk.AnteHandler {
	return func(ctx sdk.Context, tx sdk.Tx, simulate bool) (sdk.Context, error) {
		return ctx, nil
	}
}

// Test transaction creators with realistic message content
func createIBCTransferTx(msgCount int) mockTx {
	msgs := make([]sdk.Msg, msgCount)
	for i := 0; i < msgCount; i++ {
		msgs[i] = &ibctransfertypes.MsgTransfer{
			SourcePort:    "transfer",
			SourceChannel: "channel-0",
			Sender:        "cosmos1sender",
			Receiver:      "cosmos1receiver",
			Memo:          "test memo",
		}
	}
	return mockTx{msgs: msgs}
}

func createBTCHeadersTx(msgCount int) mockTx {
	msgs := make([]sdk.Msg, msgCount)
	for i := 0; i < msgCount; i++ {
		msgs[i] = &btclightclient.MsgInsertHeaders{
			Signer: "cosmos1signer",
		}
	}
	return mockTx{msgs: msgs}
}

func createMixedTx(msgCount int) mockTx {
	msgs := make([]sdk.Msg, msgCount)
	for i := 0; i < msgCount; i++ {
		switch i % 7 {
		case 0:
			msgs[i] = &ibctransfertypes.MsgTransfer{Sender: "cosmos1sender", Receiver: "cosmos1receiver"}
		case 1:
			msgs[i] = &icacontrollertypes.MsgSendTx{Owner: "cosmos1owner"}
		case 2:
			msgs[i] = &clienttypes.MsgUpdateClient{ClientId: "07-tendermint-0", Signer: "cosmos1signer"}
		case 3:
			msgs[i] = &btclightclient.MsgInsertHeaders{Signer: "cosmos1signer"}
		case 4:
			msgs[i] = &btccheckpointtypes.MsgInsertBTCSpvProof{Submitter: "cosmos1submitter"}
		case 5:
			msgs[i] = &bstypes.MsgAddCovenantSigs{Signer: "cosmos1signer"}
		case 6:
			msgs[i] = &ftypes.MsgAddFinalitySig{Signer: "cosmos1signer"}
		}
	}
	return mockTx{msgs: msgs}
}
