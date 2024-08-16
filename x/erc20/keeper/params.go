// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package keeper

import (
	"slices"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/evmos/ethermint/x/erc20/types"
)

var isTrue = []byte("0x01")

const addressLength = 42

// GetParams returns the total set of erc20 parameters.
func (k Keeper) GetParams(ctx sdk.Context) (params types.Params) {
	enableErc20 := k.IsERC20Enabled(ctx)
	dynamicPrecompiles := k.getDynamicPrecompiles(ctx)
	nativePrecompiles := k.getNativePrecompiles(ctx)
	return types.NewParams(enableErc20, nativePrecompiles, dynamicPrecompiles)
}

// SetParams sets the erc20 parameters to the param space.
func (k Keeper) SetParams(ctx sdk.Context, params types.Params) error {
	// and keep params equal between different executions
	slices.Sort(params.DynamicPrecompiles)
	slices.Sort(params.NativePrecompiles)

	if err := params.Validate(); err != nil {
		return err
	}

	k.setERC20Enabled(ctx, params.EnableErc20)
	k.setDynamicPrecompiles(ctx, params.DynamicPrecompiles)
	k.setNativePrecompiles(ctx, params.NativePrecompiles)
	return nil
}

// IsERC20Enabled returns true if the module logic is enabled
func (k Keeper) IsERC20Enabled(ctx sdk.Context) bool {
	store := k.storeService.OpenKVStore(ctx)
	has, err := store.Has(types.ParamStoreKeyEnableErc20)
	if err != nil {
		return false
	}
	return has
}

// setERC20Enabled sets the EnableERC20 param in the store
func (k Keeper) setERC20Enabled(ctx sdk.Context, enable bool) {
	store := k.storeService.OpenKVStore(ctx)
	if enable {
		store.Set(types.ParamStoreKeyEnableErc20, isTrue)
		return
	}
	store.Delete(types.ParamStoreKeyEnableErc20)
}

// setDynamicPrecompiles sets the DynamicPrecompiles param in the store
func (k Keeper) setDynamicPrecompiles(ctx sdk.Context, dynamicPrecompiles []string) {
	store := k.storeService.OpenKVStore(ctx)
	bz := make([]byte, 0, addressLength*len(dynamicPrecompiles))
	for _, str := range dynamicPrecompiles {
		bz = append(bz, []byte(str)...)
	}
	store.Set(types.ParamStoreKeyDynamicPrecompiles, bz)
}

// getDynamicPrecompiles returns the DynamicPrecompiles param from the store
func (k Keeper) getDynamicPrecompiles(ctx sdk.Context) (dynamicPrecompiles []string) {
	store := k.storeService.OpenKVStore(ctx)
	bz, err := store.Get(types.ParamStoreKeyDynamicPrecompiles)
	if err != nil {
		return nil
	}

	for i := 0; i < len(bz); i += addressLength {
		dynamicPrecompiles = append(dynamicPrecompiles, string(bz[i:i+addressLength]))
	}
	return dynamicPrecompiles
}

// setNativePrecompiles sets the NativePrecompiles param in the store
func (k Keeper) setNativePrecompiles(ctx sdk.Context, nativePrecompiles []string) {
	store := k.storeService.OpenKVStore(ctx)
	bz := make([]byte, 0, addressLength*len(nativePrecompiles))
	for _, str := range nativePrecompiles {
		bz = append(bz, []byte(str)...)
	}
	store.Set(types.ParamStoreKeyNativePrecompiles, bz)
}

// getNativePrecompiles returns the NativePrecompiles param from the store
func (k Keeper) getNativePrecompiles(ctx sdk.Context) (nativePrecompiles []string) {
	store := k.storeService.OpenKVStore(ctx)
	bz, err := store.Get(types.ParamStoreKeyNativePrecompiles)
	if err != nil {
		return nil
	}
	for i := 0; i < len(bz); i += addressLength {
		nativePrecompiles = append(nativePrecompiles, string(bz[i:i+addressLength]))
	}
	return nativePrecompiles
}
