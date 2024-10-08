package v4

import (
	corestore "cosmossdk.io/core/store"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"

	v3types "github.com/evmos/ethermint/x/evm/migrations/v3/types"
	"github.com/evmos/ethermint/x/evm/types"
)

const evmDenom = "aorai"

// MigrateStore migrates the x/evm module state from the consensus version 3 to
// version 4. Specifically, it takes the parameters that are currently stored
// and managed by the Cosmos SDK params module and stores them directly into the x/evm module state.
func MigrateStore(
	ctx sdk.Context,
	storeService corestore.KVStoreService,
	legacySubspace types.Subspace,
	cdc codec.BinaryCodec,
) error {

	params := types.DefaultParams()

	chainCfgBz := cdc.MustMarshal(&params.ChainConfig)
	extraEIPsBz := cdc.MustMarshal(&v3types.ExtraEIPs{EIPs: params.ExtraEIPs})

	store := storeService.OpenKVStore(ctx)

	// our evm denom is aorai
	if err := store.Set(types.ParamStoreKeyEVMDenom, []byte(evmDenom)); err != nil {
		return err
	}

	if err := store.Set(types.ParamStoreKeyExtraEIPs, extraEIPsBz); err != nil {
		return err
	}

	if err := store.Set(types.ParamStoreKeyChainConfig, chainCfgBz); err != nil {
		return err
	}

	if params.AllowUnprotectedTxs {
		if err := store.Set(types.ParamStoreKeyAllowUnprotectedTxs, []byte{0x01}); err != nil {
			return err
		}
	}

	if params.EnableCall {
		if err := store.Set(types.ParamStoreKeyEnableCall, []byte{0x01}); err != nil {
			return err
		}
	}

	if params.EnableCreate {
		if err := store.Set(types.ParamStoreKeyEnableCreate, []byte{0x01}); err != nil {
			return err
		}
	}

	return nil
}
