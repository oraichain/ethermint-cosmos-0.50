package keeper

import (
	"context"
	"fmt"

	errorsmod "cosmossdk.io/errors"
	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	errortypes "github.com/cosmos/cosmos-sdk/types/errors"
	evmtypes "github.com/evmos/ethermint/x/evm/types"
)

// ConversionMultiplier is the conversion multiplier between evm and cosmos denom decimals (18 vs 6)
var ConversionMultiplier = sdkmath.NewInt(1_000_000_000_000)

var _ evmtypes.BankKeeper = EvmBankKeeper{}

// EvmBankKeeper is a BankKeeper wrapper for the x/evm module to allow the use
// of the 18 decimal aorai coin on the evm.
// x/evm consumes gas and send coins by minting and burning aorai coins in its module
// account and then sending the funds to the target account.
// This keeper uses both the ukava coin and a separate aorai balance to manage the
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
func (k EvmBankKeeper) GetBalance(ctx context.Context, addr sdk.AccAddress, denom string) sdk.Coin {
	if denom != k.EvmDenom {
		panic(fmt.Errorf("only evm denom %s is supported by EvmBankKeeper", k.EvmDenom))
	}

	spendableCoins := k.bk.SpendableCoins(ctx, addr)
	cosmosAmount := spendableCoins.AmountOf(k.CosmosDenom)
	evmAmount := spendableCoins.AmountOf(denom)
	total := cosmosAmount.Mul(ConversionMultiplier).Add(evmAmount)
	return sdk.NewCoin(k.EvmDenom, total)
}

// BurnCoins burns aorai coins by burning the equivalent orai coins and any remaining aorai coins.
// It will panic if the module account does not exist or is unauthorized.
func (ebk EvmBankKeeper) BurnCoins(ctx context.Context, moduleName string, amt sdk.Coins) error {
	orai, _, err := SplitAoraiCoins(amt, ebk.EvmDenom, ebk.CosmosDenom)
	if err != nil {
		return err
	}

	if orai.IsPositive() {
		if err := ebk.bk.BurnCoins(ctx, moduleName, sdk.NewCoins(orai)); err != nil {
			return err
		}
	}

	return nil
}

// IsSendEnabledCoins implements types.BankKeeper.
func (e EvmBankKeeper) IsSendEnabledCoins(ctx context.Context, coins ...sdk.Coin) error {
	return e.bk.IsSendEnabledCoins(ctx, coins...)
}

// MintCoins mints aorai coins by minting the equivalent ukava coins and any remaining aorai coins.
// It will panic if the module account does not exist or is unauthorized.
func (ebk EvmBankKeeper) MintCoins(ctx context.Context, moduleName string, amt sdk.Coins) error {
	orai, _, err := SplitAoraiCoins(amt, ebk.EvmDenom, ebk.CosmosDenom)
	if err != nil {
		return err
	}

	if orai.IsPositive() {
		if err := ebk.bk.MintCoins(ctx, moduleName, sdk.NewCoins(orai)); err != nil {
			return err
		}
	}

	return nil
}

// SendCoins implements types.BankKeeper.
func (e EvmBankKeeper) SendCoins(ctx context.Context, from sdk.AccAddress, to sdk.AccAddress, amt sdk.Coins) error {
	return e.bk.SendCoins(ctx, from, to, amt)
}

// SendCoinsFromAccountToModule transfers aorai coins from an AccAddress to a ModuleAccount.
// It will panic if the module account does not exist.
func (ebk EvmBankKeeper) SendCoinsFromAccountToModule(ctx context.Context, senderAddr sdk.AccAddress, recipientModule string, amt sdk.Coins) error {
	orai, _, err := SplitAoraiCoins(amt, ebk.EvmDenom, ebk.CosmosDenom)
	if err != nil {
		return err
	}

	if orai.IsPositive() {
		if err := ebk.bk.SendCoinsFromAccountToModule(ctx, senderAddr, recipientModule, sdk.NewCoins(orai)); err != nil {
			return err
		}
	}

	return nil
}

// SendCoinsFromModuleToAccount transfers aorai coins from a ModuleAccount to an AccAddress.
// It will panic if the module account does not exist. An error is returned if the recipient
// address is black-listed or if sending the tokens fails.
func (ebk EvmBankKeeper) SendCoinsFromModuleToAccount(ctx context.Context, senderModule string, recipientAddr sdk.AccAddress, amt sdk.Coins) error {
	orai, _, err := SplitAoraiCoins(amt, ebk.EvmDenom, ebk.CosmosDenom)
	if err != nil {
		return err
	}

	if orai.Amount.IsPositive() {
		if err := ebk.bk.SendCoinsFromModuleToAccount(ctx, senderModule, recipientAddr, sdk.NewCoins(orai)); err != nil {
			return err
		}
	}

	return nil
}

// SpendableCoins implements types.BankKeeper.
func (e EvmBankKeeper) SpendableCoins(ctx context.Context, addr sdk.AccAddress) sdk.Coins {
	return e.bk.SpendableCoins(ctx, addr)
}

// SplitAoraiCoins splits aorai coins to the equivalent orai coins and any remaining aorai balance.
// An error will be returned if the coins are not valid or if the coins are not the aorai denom.
func SplitAoraiCoins(coins sdk.Coins, evmDenom string, cosmosDenom string) (sdk.Coin, sdkmath.Int, error) {
	aorai := sdkmath.ZeroInt()
	orai := sdk.NewCoin(cosmosDenom, sdkmath.ZeroInt())

	if len(coins) == 0 {
		return orai, aorai, nil
	}

	if err := ValidateEvmCoins(coins, evmDenom); err != nil {
		return orai, aorai, err
	}

	// note: we should always have len(coins) == 1 here since coins cannot have dup denoms after we validate.
	coin := coins[0]
	remainingBalance := coin.Amount.Mod(ConversionMultiplier)
	if remainingBalance.IsPositive() {
		aorai = remainingBalance
	}

	oraiAmount := coin.Amount.Quo(ConversionMultiplier)
	if oraiAmount.IsPositive() {
		orai = sdk.NewCoin(cosmosDenom, oraiAmount)
	}

	return orai, aorai, nil
}

// ValidateEvmCoins validates the coins from evm is valid and is the EvmDenom (aorai).
func ValidateEvmCoins(coins sdk.Coins, evmDenom string) error {
	if len(coins) == 0 {
		return nil
	}

	// validate that coins are non-negative, sorted, and no dup denoms
	if err := coins.Validate(); err != nil {
		return errorsmod.Wrap(errortypes.ErrInvalidCoins, coins.String())
	}

	// validate that coin denom is aorai
	if len(coins) != 1 || coins[0].Denom != evmDenom {
		return errorsmod.Wrapf(errortypes.ErrInvalidCoins, "invalid evm coin denom, only %s is supported", evmDenom)
	}

	return nil
}
