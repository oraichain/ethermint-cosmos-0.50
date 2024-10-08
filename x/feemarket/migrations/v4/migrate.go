package v4

import (
	corestore "cosmossdk.io/core/store"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/evmos/ethermint/x/feemarket/types"
)

// MigrateStore migrates the x/evm module state from the consensus version 3 to
// version 4. Specifically, it takes the parameters that are currently stored
// and managed by the Cosmos SDK params module and stores them directly into the x/evm module state.
func MigrateStore(
	ctx sdk.Context,
	storeService corestore.KVStoreService,
	cdc codec.BinaryCodec,
) error {
	store := storeService.OpenKVStore(ctx)
	params := types.DefaultParams()

	bz, err := cdc.Marshal(&params)
	if err != nil {
		return err
	}

	err = store.Set(types.ParamsKey, bz)
	if err != nil {
		return err
	}

	return nil
}
