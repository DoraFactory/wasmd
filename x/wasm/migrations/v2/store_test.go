package v2_test

import (
	"testing"

	"github.com/cometbft/cometbft/libs/rand"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	storetypes "github.com/cosmos/cosmos-sdk/store/v2/types"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/runtime"
	"github.com/cosmos/cosmos-sdk/testutil"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/address"
	moduletestutil "github.com/cosmos/cosmos-sdk/types/module/testutil"

	"github.com/CosmWasm/wasmd/x/wasm"
	"github.com/CosmWasm/wasmd/x/wasm/exported"
	v2 "github.com/CosmWasm/wasmd/x/wasm/migrations/v2"
	"github.com/CosmWasm/wasmd/x/wasm/types"
)

// parameterFixture supplies the migration's read-only Subspace interface without
// restoring the x/params runtime removed in SDK 0.55. Parameters are encoded in
// a separate KV store; this checks migration reading/writing, not x/params itself.
type parameterFixture struct {
	key   *storetypes.KVStoreKey
	amino *codec.LegacyAmino
}

func (f parameterFixture) GetParamSet(ctx sdk.Context, ps exported.ParamSet) {
	for _, pair := range ps.ParamSetPairs() {
		f.amino.MustUnmarshalJSON(ctx.KVStore(f.key).Get(pair.Key), pair.Value)
	}
}

func TestMigrate(t *testing.T) {
	cfg := moduletestutil.MakeTestEncodingConfig(wasm.AppModuleBasic{})
	cdc := cfg.Codec
	var (
		wasmStoreKey   = storetypes.NewKVStoreKey(types.StoreKey)
		paramsStoreKey = storetypes.NewKVStoreKey("legacy-params-test")
		myAddress      = sdk.AccAddress(rand.Bytes(address.Len))
	)
	specs := map[string]struct {
		src v2.Params
	}{
		"one address": {
			src: v2.Params{
				CodeUploadAccess: v2.AccessConfig{
					Permission: v2.AccessTypeOnlyAddress,
					Address:    myAddress.String(),
				},
				InstantiateDefaultPermission: v2.AccessTypeNobody,
			},
		},
		"multiple addresses": {
			src: v2.Params{
				CodeUploadAccess: v2.AccessConfig{
					Permission: v2.AccessTypeAnyOfAddresses,
					Addresses:  []string{myAddress.String(), sdk.AccAddress(rand.Bytes(address.Len)).String()},
				},
				InstantiateDefaultPermission: v2.AccessTypeEverybody,
			},
		},
		"everybody": {
			src: v2.Params{
				CodeUploadAccess: v2.AccessConfig{
					Permission: v2.AccessTypeEverybody,
				},
				InstantiateDefaultPermission: v2.AccessTypeEverybody,
			},
		},
		"nobody": {
			src: v2.Params{
				CodeUploadAccess: v2.AccessConfig{
					Permission: v2.AccessTypeNobody,
				},
				InstantiateDefaultPermission: v2.AccessTypeNobody,
			},
		},
	}
	for name, spec := range specs {
		t.Run(name, func(t *testing.T) {
			ctx := testutil.DefaultContextWithKeys(
				map[string]*storetypes.KVStoreKey{
					"legacy-params-test": paramsStoreKey,
					types.StoreKey:       wasmStoreKey,
				},
				nil,
				nil,
			)

			// register legacy parameters
			params := spec.src
			subspace := parameterFixture{key: paramsStoreKey, amino: cfg.Amino}
			for _, pair := range params.ParamSetPairs() {
				ctx.KVStore(paramsStoreKey).Set(pair.Key, cfg.Amino.MustMarshalJSON(pair.Value))
			}

			// when
			require.NoError(t, v2.MigrateStore(ctx, runtime.NewKVStoreService(wasmStoreKey), subspace, cdc))

			var res v2.Params
			bz := ctx.KVStore(wasmStoreKey).Get(types.ParamsKey)
			require.NoError(t, cdc.Unmarshal(bz, &res))
			assert.Equal(t, params, res)
		})
	}
}
