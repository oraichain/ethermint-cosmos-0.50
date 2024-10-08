package v4_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/evmos/ethermint/x/evm/types"

	storetypes "cosmossdk.io/store/types"
	"github.com/cosmos/cosmos-sdk/runtime"
	"github.com/cosmos/cosmos-sdk/testutil"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/evmos/ethermint/encoding"
	v3 "github.com/evmos/ethermint/x/evm/migrations/v3"
	v3types "github.com/evmos/ethermint/x/evm/migrations/v3/types"
)

type mockSubspace struct {
	ps types.Params
}

func newMockSubspace(ps types.Params) mockSubspace {
	return mockSubspace{ps: ps}
}

func (ms mockSubspace) GetParamSetIfExists(ctx sdk.Context, ps types.LegacyParams) {
	*ps.(*types.Params) = ms.ps
}

func TestMigrate(t *testing.T) {
	encCfg := encoding.MakeTestEncodingConfig()
	cdc := encCfg.Codec

	storeKey := storetypes.NewKVStoreKey(types.ModuleName)
	tKey := storetypes.NewTransientStoreKey(types.TransientKey)
	ctx := testutil.DefaultContext(storeKey, tKey)
	storeService := runtime.NewKVStoreService(storeKey)

	legacySubspace := newMockSubspace(types.DefaultParams())
	require.NoError(t, v3.MigrateStore(ctx, storeService, legacySubspace, cdc))

	kvStore := storeService.OpenKVStore(ctx)

	// Get all the new parameters from the kvStore
	var evmDenom string
	bz, err := kvStore.Get(types.ParamStoreKeyEVMDenom)
	require.NoError(t, err)
	evmDenom = string(bz)

	allowUnprotectedTx, err := kvStore.Has(types.ParamStoreKeyAllowUnprotectedTxs)
	require.NoError(t, err)
	enableCreate, err := kvStore.Has(types.ParamStoreKeyEnableCreate)
	require.NoError(t, err)
	enableCall, err := kvStore.Has(types.ParamStoreKeyEnableCall)
	require.NoError(t, err)

	var chainCfg v3types.V4ChainConfig
	bz, err = kvStore.Get(types.ParamStoreKeyChainConfig)
	require.NoError(t, err)
	cdc.MustUnmarshal(bz, &chainCfg)

	var extraEIPs v3types.ExtraEIPs
	bz, err = kvStore.Get(types.ParamStoreKeyExtraEIPs)
	require.NoError(t, err)
	cdc.MustUnmarshal(bz, &extraEIPs)
	require.Equal(t, []int64(nil), extraEIPs.EIPs)

	params := v3types.V4Params{
		EvmDenom:            evmDenom,
		AllowUnprotectedTxs: allowUnprotectedTx,
		EnableCreate:        enableCreate,
		EnableCall:          enableCall,
		V4ChainConfig:       chainCfg,
		ExtraEIPs:           extraEIPs,
	}

	require.Equal(t, legacySubspace.ps.EnableCall, params.EnableCall)
	require.Equal(t, legacySubspace.ps.EnableCreate, params.EnableCreate)
	require.Equal(t, legacySubspace.ps.AllowUnprotectedTxs, params.AllowUnprotectedTxs)
	require.Equal(t, legacySubspace.ps.ExtraEIPs, params.ExtraEIPs.EIPs)
	require.EqualValues(t, legacySubspace.ps.ChainConfig, params.V4ChainConfig)
}
