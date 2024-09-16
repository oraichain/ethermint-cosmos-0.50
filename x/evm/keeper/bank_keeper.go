package keeper

import (
	"context"
	"fmt"

	sdkmath "cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	evmtypes "github.com/evmos/ethermint/x/evm/types"
)

// ConversionMultiplier is the conversion multiplier between evm and cosmos denom decimals (18 vs 6)
var ConversionMultiplier = sdkmath.NewInt(1_000_000_000_000)

var _ evmtypes.BankKeeper = EvmBankKeeper{}

// EvmBankKeeper is a BankKeeper wrapper for the x/evm module to allow the use
// of the 18 decimal orai coin on the evm.
// x/evm consumes gas and send coins by minting and burning akava coins in its module
// account and then sending the funds to the target account.
// This keeper uses both the ukava coin and a separate akava balance to manage the
// extra percision needed by the evm.
type EvmBankKeeper struct {
	bk          evmtypes.BankKeeper
	ak          evmtypes.AccountKeeper
	EvmDenom    string
	CosmosDenom string
}

func NewEvmBankKeeperWithDenoms(bk evmtypes.BankKeeper, ak evmtypes.AccountKeeper, evmDenom, cosmosDenom string) EvmBankKeeper {
	return EvmBankKeeper{
		bk:          bk,
		ak:          ak,
		EvmDenom:    evmDenom,
		CosmosDenom: cosmosDenom,
	}
}

// GetBalance returns the total **spendable** balance of aorai for a given account by address.
func (k EvmBankKeeper) GetBalance(context context.Context, addr sdk.AccAddress, denom string) sdk.Coin {
	ctx := sdk.UnwrapSDKContext(context)
	if denom != k.EvmDenom {
		panic(fmt.Errorf("only evm denom %s is supported by EvmBankKeeper", k.EvmDenom))
	}

	spendableCoins := k.bk.SpendableCoins(ctx, addr)
	cosmosAmount := spendableCoins.AmountOf(k.CosmosDenom)
	evmAmount := spendableCoins.AmountOf(denom)
	total := cosmosAmount.Mul(ConversionMultiplier).Add(evmAmount)
	return sdk.NewCoin(k.EvmDenom, total)
}

// BurnCoins implements types.BankKeeper.
func (e EvmBankKeeper) BurnCoins(ctx context.Context, moduleName string, amt types.Coins) error {
	return e.bk.BurnCoins(ctx, moduleName, amt)
}

// IsSendEnabledCoins implements types.BankKeeper.
func (e EvmBankKeeper) IsSendEnabledCoins(ctx context.Context, coins ...types.Coin) error {
	return e.bk.IsSendEnabledCoins(ctx, coins...)
}

// MintCoins implements types.BankKeeper.
func (e EvmBankKeeper) MintCoins(ctx context.Context, moduleName string, amt types.Coins) error {
	return e.bk.MintCoins(ctx, moduleName, amt)
}

// SendCoins implements types.BankKeeper.
func (e EvmBankKeeper) SendCoins(ctx context.Context, from types.AccAddress, to types.AccAddress, amt types.Coins) error {
	return e.bk.SendCoins(ctx, from, to, amt)
}

// SendCoinsFromAccountToModule implements types.BankKeeper.
func (e EvmBankKeeper) SendCoinsFromAccountToModule(ctx context.Context, senderAddr types.AccAddress, recipientModule string, amt types.Coins) error {
	return e.bk.SendCoinsFromAccountToModule(ctx, senderAddr, recipientModule, amt)
}

// SendCoinsFromModuleToAccount implements types.BankKeeper.
func (e EvmBankKeeper) SendCoinsFromModuleToAccount(ctx context.Context, senderModule string, recipientAddr types.AccAddress, amt types.Coins) error {
	return e.bk.SendCoinsFromModuleToAccount(ctx, senderModule, recipientAddr, amt)
}

// SpendableCoins implements types.BankKeeper.
func (e EvmBankKeeper) SpendableCoins(ctx context.Context, addr types.AccAddress) types.Coins {
	return e.bk.SpendableCoins(ctx, addr)
}
