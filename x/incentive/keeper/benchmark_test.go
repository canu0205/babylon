package keeper_test

import (
	"context"
	"crypto/sha256"
	"fmt"
	"sync"
	"testing"

	"cosmossdk.io/collections"
	corestore "cosmossdk.io/core/store"
	"cosmossdk.io/log"
	"cosmossdk.io/store"
	storetypes "cosmossdk.io/store/types"
	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"
	dbm "github.com/cosmos/cosmos-db"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/babylonlabs-io/babylon/v4/x/incentive/types"
)

// Test approaches
type RefundableApproach interface {
	IndexRefundableMsg(ctx context.Context, msgHash []byte)
	HasRefundableMsg(ctx context.Context, msgHash []byte) bool
	RemoveRefundableMsg(ctx context.Context, msgHash []byte)
	EndBlockCleanup(ctx context.Context)
}

// 1. In-memory map approach
type InMemoryMapApproach struct {
	refundableMsgSet map[string]struct{}
	mu               sync.RWMutex
}

func NewInMemoryMapApproach() *InMemoryMapApproach {
	return &InMemoryMapApproach{
		refundableMsgSet: make(map[string]struct{}),
	}
}

func (a *InMemoryMapApproach) IndexRefundableMsg(ctx context.Context, msgHash []byte) {
	a.mu.Lock()
	a.refundableMsgSet[string(msgHash)] = struct{}{}
	a.mu.Unlock()
}

func (a *InMemoryMapApproach) HasRefundableMsg(ctx context.Context, msgHash []byte) bool {
	a.mu.RLock()
	_, exists := a.refundableMsgSet[string(msgHash)]
	a.mu.RUnlock()
	return exists
}

func (a *InMemoryMapApproach) RemoveRefundableMsg(ctx context.Context, msgHash []byte) {
	a.mu.Lock()
	delete(a.refundableMsgSet, string(msgHash))
	a.mu.Unlock()
}

func (a *InMemoryMapApproach) EndBlockCleanup(ctx context.Context) {
	a.mu.Lock()
	a.refundableMsgSet = make(map[string]struct{})
	a.mu.Unlock()
}

// 2. Direct transient store approach
type TransientStoreApproach struct {
	tKey   *storetypes.TransientStoreKey
	prefix []byte
}

func NewTransientStoreApproach(tKey *storetypes.TransientStoreKey) *TransientStoreApproach {
	return &TransientStoreApproach{
		tKey:   tKey,
		prefix: types.RefundableMsgKeySetPrefix.Bytes(),
	}
}

func (a *TransientStoreApproach) IndexRefundableMsg(ctx context.Context, msgHash []byte) {
	store := sdk.UnwrapSDKContext(ctx).TransientStore(a.tKey)
	prefixedKey := append(a.prefix, msgHash...)
	store.Set(prefixedKey, []byte{1})
}

func (a *TransientStoreApproach) HasRefundableMsg(ctx context.Context, msgHash []byte) bool {
	store := sdk.UnwrapSDKContext(ctx).TransientStore(a.tKey)
	prefixedKey := append(a.prefix, msgHash...)
	return store.Has(prefixedKey)
}

func (a *TransientStoreApproach) RemoveRefundableMsg(ctx context.Context, msgHash []byte) {
	store := sdk.UnwrapSDKContext(ctx).TransientStore(a.tKey)
	prefixedKey := append(a.prefix, msgHash...)
	store.Delete(prefixedKey)
}

func (a *TransientStoreApproach) EndBlockCleanup(ctx context.Context) {
	// Transient store automatically cleans up
}

// 3. collections.KeySet with TransientStore approach
type CollectionsTransientApproach struct {
	refundableMsgKeySet collections.KeySet[[]byte]
}

func NewCollectionsTransientApproach(tKey *storetypes.TransientStoreKey) *CollectionsTransientApproach {
	transientStoreAccessor := func(ctx context.Context) corestore.KVStore {
		sdkCtx := sdk.UnwrapSDKContext(ctx)
		return &transientStoreWrapper{store: sdkCtx.TransientStore(tKey)}
	}

	sb := collections.NewSchemaBuilderFromAccessor(transientStoreAccessor)

	refundableMsgKeySet := collections.NewKeySet(
		sb,
		types.RefundableMsgKeySetPrefix,
		"refundable_msg_keyset",
		collections.BytesKey,
	)

	// Build schema
	_, err := sb.Build()
	if err != nil {
		panic(err) // Should not happen in benchmark
	}

	return &CollectionsTransientApproach{
		refundableMsgKeySet: refundableMsgKeySet,
	}
}

func (a *CollectionsTransientApproach) IndexRefundableMsg(ctx context.Context, msgHash []byte) {
	err := a.refundableMsgKeySet.Set(ctx, msgHash)
	if err != nil {
		panic(err) // Should not happen in benchmark
	}
}

func (a *CollectionsTransientApproach) HasRefundableMsg(ctx context.Context, msgHash []byte) bool {
	has, err := a.refundableMsgKeySet.Has(ctx, msgHash)
	if err != nil {
		panic(err) // Should not happen in benchmark
	}
	return has
}

func (a *CollectionsTransientApproach) RemoveRefundableMsg(ctx context.Context, msgHash []byte) {
	err := a.refundableMsgKeySet.Remove(ctx, msgHash)
	if err != nil {
		panic(err) // Should not happen in benchmark
	}
}

func (a *CollectionsTransientApproach) EndBlockCleanup(ctx context.Context) {
	// Transient store automatically cleans up
}

// transientStoreWrapper adapts SDK TransientStore to core.KVStore interface
type transientStoreWrapper struct {
	store storetypes.KVStore
}

func (w *transientStoreWrapper) Get(key []byte) ([]byte, error) {
	return w.store.Get(key), nil
}

func (w *transientStoreWrapper) Has(key []byte) (bool, error) {
	return w.store.Has(key), nil
}

func (w *transientStoreWrapper) Set(key, value []byte) error {
	w.store.Set(key, value)
	return nil
}

func (w *transientStoreWrapper) Delete(key []byte) error {
	w.store.Delete(key)
	return nil
}

func (w *transientStoreWrapper) Iterator(start, end []byte) (corestore.Iterator, error) {
	iter := w.store.Iterator(start, end)
	return &iteratorWrapper{iter: iter}, nil
}

func (w *transientStoreWrapper) ReverseIterator(start, end []byte) (corestore.Iterator, error) {
	iter := w.store.ReverseIterator(start, end)
	return &iteratorWrapper{iter: iter}, nil
}

// iteratorWrapper adapts SDK Iterator to core.Iterator interface
type iteratorWrapper struct {
	iter storetypes.Iterator
}

func (w *iteratorWrapper) Domain() ([]byte, []byte) {
	return w.iter.Domain()
}

func (w *iteratorWrapper) Valid() bool {
	return w.iter.Valid()
}

func (w *iteratorWrapper) Next() {
	w.iter.Next()
}

func (w *iteratorWrapper) Key() []byte {
	return w.iter.Key()
}

func (w *iteratorWrapper) Value() []byte {
	return w.iter.Value()
}

func (w *iteratorWrapper) Close() error {
	return w.iter.Close()
}

func (w *iteratorWrapper) Error() error {
	return nil
}

func generateDummyMsgHashes(count int) [][]byte {
	hashes := make([][]byte, count)
	for i := 0; i < count; i++ {
		hash := sha256.Sum256([]byte(fmt.Sprintf("dummy_msg_%d", i)))
		hashes[i] = hash[:]
	}
	return hashes
}

func setupSDKContext() (sdk.Context, *storetypes.TransientStoreKey) {
	db := dbm.NewMemDB()
	cms := store.NewCommitMultiStore(db, log.NewNopLogger(), nil)

	tKey := storetypes.NewTransientStoreKey("test_transient")
	cms.MountStoreWithDB(tKey, storetypes.StoreTypeTransient, nil)
	_ = cms.LoadLatestVersion()

	header := cmtproto.Header{}
	ctx := sdk.NewContext(cms, header, false, log.NewNopLogger())
	return ctx, tKey
}

func benchmarkRefundableApproach(b *testing.B, approachFunc func(*storetypes.TransientStoreKey) RefundableApproach, msgCount int) {
	ctx, tKey := setupSDKContext()
	approach := approachFunc(tKey)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		msgHashes := generateDummyMsgHashes(msgCount)

		for _, msgHash := range msgHashes {
			approach.IndexRefundableMsg(ctx, msgHash)
		}

		for _, msgHash := range msgHashes {
			if approach.HasRefundableMsg(ctx, msgHash) {
				approach.RemoveRefundableMsg(ctx, msgHash)
			}
		}

		approach.EndBlockCleanup(ctx)
	}
}

// Benchmarks for different message counts
func BenchmarkInMemoryMap_10msgs(b *testing.B) {
	benchmarkRefundableApproach(b, func(tKey *storetypes.TransientStoreKey) RefundableApproach {
		return NewInMemoryMapApproach()
	}, 10)
}

func BenchmarkInMemoryMap_100msgs(b *testing.B) {
	benchmarkRefundableApproach(b, func(tKey *storetypes.TransientStoreKey) RefundableApproach {
		return NewInMemoryMapApproach()
	}, 100)
}

func BenchmarkInMemoryMap_1000msgs(b *testing.B) {
	benchmarkRefundableApproach(b, func(tKey *storetypes.TransientStoreKey) RefundableApproach {
		return NewInMemoryMapApproach()
	}, 1000)
}

func BenchmarkTransientStore_10msgs(b *testing.B) {
	benchmarkRefundableApproach(b, func(tKey *storetypes.TransientStoreKey) RefundableApproach {
		return NewTransientStoreApproach(tKey)
	}, 10)
}

func BenchmarkTransientStore_100msgs(b *testing.B) {
	benchmarkRefundableApproach(b, func(tKey *storetypes.TransientStoreKey) RefundableApproach {
		return NewTransientStoreApproach(tKey)
	}, 100)
}

func BenchmarkTransientStore_1000msgs(b *testing.B) {
	benchmarkRefundableApproach(b, func(tKey *storetypes.TransientStoreKey) RefundableApproach {
		return NewTransientStoreApproach(tKey)
	}, 1000)
}

func BenchmarkCollectionsTransient_10msgs(b *testing.B) {
	benchmarkRefundableApproach(b, func(tKey *storetypes.TransientStoreKey) RefundableApproach {
		return NewCollectionsTransientApproach(tKey)
	}, 10)
}

func BenchmarkCollectionsTransient_100msgs(b *testing.B) {
	benchmarkRefundableApproach(b, func(tKey *storetypes.TransientStoreKey) RefundableApproach {
		return NewCollectionsTransientApproach(tKey)
	}, 100)
}

func BenchmarkCollectionsTransient_1000msgs(b *testing.B) {
	benchmarkRefundableApproach(b, func(tKey *storetypes.TransientStoreKey) RefundableApproach {
		return NewCollectionsTransientApproach(tKey)
	}, 1000)
}
