package app

import (
	"os"
	"testing"

	"cosmossdk.io/log"
	"github.com/cometbft/cometbft/abci/types"
	dbm "github.com/cosmos/cosmos-db"
	"github.com/cosmos/cosmos-sdk/baseapp"
	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
	"github.com/evmos/ethermint/encoding"
	"github.com/stretchr/testify/require"
)

func TestEthermintAppExport(t *testing.T) {
	db := dbm.NewMemDB()
	app := SetupWithDB(false, nil, db)
	// finalize block so we have CheckTx state set
	_, err := app.FinalizeBlock(&types.RequestFinalizeBlock{
		Height: app.LastBlockHeight() + 1,
	})
	require.NoError(t, err, "ExportAppStateAndValidators FinalizeBlock failed")
	app.Commit()

	// Making a new app object with the db, so that initchain hasn't been called
	app2 := NewEthermintApp(log.NewLogger(os.Stdout), db, nil, true, map[int64]bool{}, DefaultNodeHome, 0, encoding.MakeConfig(ModuleBasics), simtestutil.NewAppOptionsWithFlagHome(DefaultNodeHome),
		baseapp.SetChainID(ChainID),
	)
	_, err = app2.ExportAppStateAndValidators(false, []string{}, []string{})
	require.NoError(t, err, "ExportAppStateAndValidators should not have an error")
}
