package testutil

import (
	"fmt"
	"time"

	errorsmod "cosmossdk.io/errors"
	abci "github.com/cometbft/cometbft/abci/types"
	tmtypes "github.com/cometbft/cometbft/types"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	errortypes "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/evmos/ethermint/app"
	"github.com/evmos/ethermint/testutil/tx"
)

// Commit commits a block at a given time. Reminder: At the end of each
// CometBFT Consensus round the following methods are run
//  1. FinalizeBlock
//  2. Commit
func Commit(ctx sdk.Context, app *app.EthermintApp, t time.Duration, vs *tmtypes.ValidatorSet) (sdk.Context, error) {
	header := ctx.BlockHeader()

	if vs != nil {
		res, err := app.FinalizeBlock(&abci.RequestFinalizeBlock{
			Height: header.Height,
		})
		if err != nil {
			return ctx, err
		}

		nextVals, err := applyValSetChanges(vs, res.ValidatorUpdates)
		if err != nil {
			return ctx, err
		}
		header.ValidatorsHash = vs.Hash()
		header.NextValidatorsHash = nextVals.Hash()
	} else {
		_, err := app.FinalizeBlock(&abci.RequestFinalizeBlock{
			Height: header.Height,
		})
		if err != nil {
			return ctx, err
		}
	}

	_, err := app.Commit()
	if err != nil {
		return ctx, err
	}

	header.Height++
	header.Time = header.Time.Add(t)
	header.AppHash = app.LastCommitID().Hash

	return ctx.WithBlockHeader(header), nil
}

// DeliverEthTx generates and broadcasts a Cosmos Tx populated with MsgEthereumTx messages.
// If a private key is provided, it will attempt to sign all messages with the given private key,
// otherwise, it will assume the messages have already been signed.
func DeliverEthTx(
	appEvmos *app.EthermintApp,
	priv cryptotypes.PrivKey,
	msgs ...sdk.Msg,
) (abci.ExecTxResult, error) {

	txConfig := appEvmos.TxConfig()

	tx, err := tx.PrepareEthTx(txConfig, appEvmos, priv, msgs...)
	if err != nil {
		return abci.ExecTxResult{}, err
	}
	res, err := BroadcastTxBytes(appEvmos, txConfig.TxEncoder(), tx)
	if err != nil {
		return res, err
	}

	codec := appEvmos.AppCodec()
	if _, err := CheckEthTxResponse(res, codec); err != nil {
		return res, err
	}
	return res, nil
}

// BroadcastTxBytes encodes a transaction and calls DeliverTx on the app.
func BroadcastTxBytes(app *app.EthermintApp, txEncoder sdk.TxEncoder, tx sdk.Tx) (abci.ExecTxResult, error) {
	// bz are bytes to be broadcasted over the network
	bz, err := txEncoder(tx)
	if err != nil {
		return abci.ExecTxResult{}, err
	}

	req := abci.RequestFinalizeBlock{Txs: [][]byte{bz}}

	res, err := app.BaseApp.FinalizeBlock(&req)
	if err != nil {
		return abci.ExecTxResult{}, err
	}
	if len(res.TxResults) != 1 {
		return abci.ExecTxResult{}, fmt.Errorf("unexpected transaction results. Expected 1, got: %d", len(res.TxResults))
	}
	txRes := res.TxResults[0]
	if txRes.Code != 0 {
		return abci.ExecTxResult{}, errorsmod.Wrapf(errortypes.ErrInvalidRequest, txRes.Log)
	}

	return *txRes, nil
}

// applyValSetChanges takes in tmtypes.ValidatorSet and []abci.ValidatorUpdate and will return a new tmtypes.ValidatorSet which has the
// provided validator updates applied to the provided validator set.
func applyValSetChanges(valSet *tmtypes.ValidatorSet, valUpdates []abci.ValidatorUpdate) (*tmtypes.ValidatorSet, error) {
	updates, err := tmtypes.PB2TM.ValidatorUpdates(valUpdates)
	if err != nil {
		return nil, err
	}

	// must copy since validator set will mutate with UpdateWithChangeSet
	newVals := valSet.Copy()
	err = newVals.UpdateWithChangeSet(updates)
	if err != nil {
		return nil, err
	}

	return newVals, nil
}
