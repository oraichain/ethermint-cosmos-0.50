// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package keeper

import (
	errorsmod "cosmossdk.io/errors"
	"cosmossdk.io/store/prefix"
	storetypes "cosmossdk.io/store/types"
	"github.com/cosmos/cosmos-sdk/runtime"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/common"
	utils "github.com/evmos/ethermint/utils"
	"github.com/evmos/ethermint/x/erc20/types"
)

// CreateNewTokenPair creates a new token pair and stores it in the state.
func (k *Keeper) CreateNewTokenPair(ctx sdk.Context, denom string) (types.TokenPair, error) {
	pair, err := types.NewTokenPairSTRv2(denom)
	if err != nil {
		return types.TokenPair{}, err
	}
	k.SetToken(ctx, pair)
	return pair, nil
}

// SetToken stores a token pair, denom map and erc20 map.
func (k *Keeper) SetToken(ctx sdk.Context, pair types.TokenPair) {
	k.SetTokenPair(ctx, pair)
	k.SetDenomMap(ctx, pair.Denom, pair.GetID())
	k.SetERC20Map(ctx, pair.GetERC20Contract(), pair.GetID())
}

// GetTokenPairs gets all registered token tokenPairs.
func (k Keeper) GetTokenPairs(ctx sdk.Context) []types.TokenPair {
	tokenPairs := []types.TokenPair{}

	k.IterateTokenPairs(ctx, func(tokenPair types.TokenPair) (stop bool) {
		tokenPairs = append(tokenPairs, tokenPair)
		return false
	})

	return tokenPairs
}

// IterateTokenPairs iterates over all the stored token pairs.
func (k Keeper) IterateTokenPairs(ctx sdk.Context, cb func(tokenPair types.TokenPair) (stop bool)) {
	store := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	iterator := storetypes.KVStorePrefixIterator(store, types.KeyPrefixTokenPair)
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var tokenPair types.TokenPair
		k.cdc.MustUnmarshal(iterator.Value(), &tokenPair)

		if cb(tokenPair) {
			break
		}
	}
}

// GetTokenPairID returns the pair id for the specified token. Hex address or Denom can be used as token argument.
// If the token is not registered empty bytes are returned.
func (k Keeper) GetTokenPairID(ctx sdk.Context, token string) []byte {
	if common.IsHexAddress(token) {
		addr := common.HexToAddress(token)
		return k.GetERC20Map(ctx, addr)
	}
	return k.GetDenomMap(ctx, token)
}

// GetTokenPair gets a registered token pair from the identifier.
func (k Keeper) GetTokenPair(ctx sdk.Context, id []byte) (types.TokenPair, bool) {
	if id == nil {
		return types.TokenPair{}, false
	}
	store := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	prefixStore := prefix.NewStore(store, types.KeyPrefixTokenPair)
	var tokenPair types.TokenPair
	bz := prefixStore.Get(id)
	if len(bz) == 0 {
		return types.TokenPair{}, false
	}

	k.cdc.MustUnmarshal(bz, &tokenPair)
	return tokenPair, true
}

// SetTokenPair stores a token pair.
func (k Keeper) SetTokenPair(ctx sdk.Context, tokenPair types.TokenPair) {
	store := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	prefixStore := prefix.NewStore(store, types.KeyPrefixTokenPair)
	key := tokenPair.GetID()
	bz := k.cdc.MustMarshal(&tokenPair)
	prefixStore.Set(key, bz)
}

// DeleteTokenPair removes a token pair.
func (k Keeper) DeleteTokenPair(ctx sdk.Context, tokenPair types.TokenPair) {
	id := tokenPair.GetID()
	k.deleteTokenPair(ctx, id)
	k.deleteERC20Map(ctx, tokenPair.GetERC20Contract())
	k.deleteDenomMap(ctx, tokenPair.Denom)
}

// deleteTokenPair deletes the token pair for the given id.
func (k Keeper) deleteTokenPair(ctx sdk.Context, id []byte) {
	store := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	prefixStore := prefix.NewStore(store, types.KeyPrefixTokenPair)
	prefixStore.Delete(id)
}

// GetERC20Map returns the token pair id for the given address.
func (k Keeper) GetERC20Map(ctx sdk.Context, erc20 common.Address) []byte {
	store := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	prefixStore := prefix.NewStore(store, types.KeyPrefixTokenPairByERC20)
	return prefixStore.Get(erc20.Bytes())
}

// GetDenomMap returns the token pair id for the given denomination.
func (k Keeper) GetDenomMap(ctx sdk.Context, denom string) []byte {
	store := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	prefixStore := prefix.NewStore(store, types.KeyPrefixTokenPairByDenom)
	return prefixStore.Get([]byte(denom))
}

// SetERC20Map sets the token pair id for the given address.
func (k Keeper) SetERC20Map(ctx sdk.Context, erc20 common.Address, id []byte) {
	store := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	prefixStore := prefix.NewStore(store, types.KeyPrefixTokenPairByERC20)
	prefixStore.Set(erc20.Bytes(), id)
}

// deleteERC20Map deletes the token pair id for the given address.
func (k Keeper) deleteERC20Map(ctx sdk.Context, erc20 common.Address) {
	store := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	prefixStore := prefix.NewStore(store, types.KeyPrefixTokenPairByERC20)
	prefixStore.Delete(erc20.Bytes())
}

// SetDenomMap sets the token pair id for the denomination.
func (k Keeper) SetDenomMap(ctx sdk.Context, denom string, id []byte) {
	store := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	prefixStore := prefix.NewStore(store, types.KeyPrefixTokenPairByDenom)
	prefixStore.Set([]byte(denom), id)
}

// deleteDenomMap deletes the token pair id for the given denom.
func (k Keeper) deleteDenomMap(ctx sdk.Context, denom string) {
	store := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	prefixStore := prefix.NewStore(store, types.KeyPrefixTokenPairByDenom)
	prefixStore.Delete([]byte(denom))
}

// IsTokenPairRegistered - check if registered token tokenPair is registered.
func (k Keeper) IsTokenPairRegistered(ctx sdk.Context, id []byte) bool {
	store := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	prefixStore := prefix.NewStore(store, types.KeyPrefixTokenPair)
	return prefixStore.Has(id)
}

// IsERC20Registered check if registered ERC20 token is registered.
func (k Keeper) IsERC20Registered(ctx sdk.Context, erc20 common.Address) bool {
	store := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	prefixStore := prefix.NewStore(store, types.KeyPrefixTokenPairByERC20)
	return prefixStore.Has(erc20.Bytes())
}

// IsDenomRegistered check if registered coin denom is registered.
func (k Keeper) IsDenomRegistered(ctx sdk.Context, denom string) bool {
	store := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	prefixStore := prefix.NewStore(store, types.KeyPrefixTokenPairByDenom)
	return prefixStore.Has([]byte(denom))
}

// GetCoinAddress returns the corresponding ERC-20 contract address for the
// given denom.
// If the denom is not registered and its an IBC voucher, it returns the address
// from the hash of the ICS20's DenomTrace Path.
func (k Keeper) GetCoinAddress(ctx sdk.Context, denom string) (common.Address, error) {
	id := k.GetDenomMap(ctx, denom)
	if len(id) == 0 {
		// if the denom is not registered, check if it is an IBC voucher
		return utils.GetIBCDenomAddress(denom)
	}

	tokenPair, found := k.GetTokenPair(ctx, id)
	if !found {
		// safety check, should never happen
		return common.Address{}, errorsmod.Wrapf(
			types.ErrTokenPairNotFound, "coin '%s' not registered", denom,
		)
	}

	return tokenPair.GetERC20Contract(), nil
}

// GetTokenDenom returns the denom associated with the tokenAddress or an error
// if the TokenPair does not exist.
func (k Keeper) GetTokenDenom(ctx sdk.Context, tokenAddress common.Address) (string, error) {
	tokenPairID := k.GetERC20Map(ctx, tokenAddress)
	if len(tokenPairID) == 0 {
		return "", errorsmod.Wrapf(
			types.ErrTokenPairNotFound, "token '%s' not registered", tokenAddress,
		)
	}

	tokenPair, found := k.GetTokenPair(ctx, tokenPairID)
	if !found {
		// safety check, should never happen
		return "", errorsmod.Wrapf(
			types.ErrTokenPairNotFound, "token '%s' not registered", tokenAddress,
		)
	}

	return tokenPair.Denom, nil
}
