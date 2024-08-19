package keeper_test

import (
	_ "embed"
	"math/big"
	"os"
	"testing"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	sdkmath "cosmossdk.io/math"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"cosmossdk.io/simapp"
	tmjson "github.com/cometbft/cometbft/libs/json"
	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/evmos/ethermint/contracts"
	"github.com/evmos/ethermint/testutil"
	"github.com/evmos/ethermint/testutil/tx"
	feemarkettypes "github.com/evmos/ethermint/x/feemarket/types"

	"github.com/evmos/ethermint/app"
	"github.com/evmos/ethermint/crypto/ethsecp256k1"
	"github.com/evmos/ethermint/tests"
	ethermint "github.com/evmos/ethermint/types"
	"github.com/evmos/ethermint/x/erc20/types"
	"github.com/evmos/ethermint/x/evm/statedb"
	evmtypes "github.com/evmos/ethermint/x/evm/types"

	"github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/params"

	abci "github.com/cometbft/cometbft/abci/types"
	"github.com/cometbft/cometbft/crypto/tmhash"
	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	tmversion "github.com/cometbft/cometbft/proto/tendermint/version"
	"github.com/cometbft/cometbft/version"
)

type KeeperTestSuite struct {
	suite.Suite

	ctx            sdk.Context
	app            *app.EthermintApp
	queryClient    types.QueryClient
	queryClientEvm evmtypes.QueryClient
	address        common.Address
	consAddress    sdk.ConsAddress

	// for generate test tx
	clientCtx client.Context
	ethSigner ethtypes.Signer

	appCodec codec.Codec
	signer   keyring.Signer

	enableFeemarket  bool
	mintFeeCollector bool
	denom            string
}

var s *KeeperTestSuite

func TestKeeperTestSuite(t *testing.T) {
	if os.Getenv("benchmark") != "" {
		t.Skip("Skipping Gingko Test")
	}
	s = new(KeeperTestSuite)
	s.enableFeemarket = false
	suite.Run(t, s)

	// Run Ginkgo integration tests
	RegisterFailHandler(Fail)
	RunSpecs(t, "Keeper Suite")
}

func (suite *KeeperTestSuite) SetupTest() {
	checkTx := false
	suite.app = app.Setup(checkTx, nil)
	suite.SetupApp(checkTx)
}

func (suite *KeeperTestSuite) SetupTestWithT(t require.TestingT) {
	checkTx := false
	suite.app = app.Setup(checkTx, nil)
	suite.SetupAppWithT(checkTx, t)
}

func (suite *KeeperTestSuite) SetupApp(checkTx bool) {
	suite.SetupAppWithT(checkTx, suite.T())
}

// SetupApp setup test environment, it uses`require.TestingT` to support both `testing.T` and `testing.B`.
func (suite *KeeperTestSuite) SetupAppWithT(checkTx bool, t require.TestingT) {
	// account key, use a constant account to keep unit test deterministic.
	ecdsaPriv, err := crypto.HexToECDSA("b71c71a67e1177ad4e901695e1b4b9ee17ae16c6668d313eac2f96dbcda3f291")
	require.NoError(t, err)
	priv := &ethsecp256k1.PrivKey{
		Key: crypto.FromECDSA(ecdsaPriv),
	}
	suite.address = common.BytesToAddress(priv.PubKey().Address().Bytes())
	suite.signer = tests.NewSigner(priv)

	// consensus key
	priv, err = ethsecp256k1.GenerateKey()
	require.NoError(t, err)
	suite.consAddress = sdk.ConsAddress(priv.PubKey().Address())

	suite.app = app.Setup(checkTx, func(app *app.EthermintApp, genesis simapp.GenesisState) simapp.GenesisState {
		feemarketGenesis := feemarkettypes.DefaultGenesisState()
		if suite.enableFeemarket {
			feemarketGenesis.Params.EnableHeight = 1
			feemarketGenesis.Params.NoBaseFee = false
		} else {
			feemarketGenesis.Params.NoBaseFee = true
		}
		genesis[feemarkettypes.ModuleName] = app.AppCodec().MustMarshalJSON(feemarketGenesis)

		return genesis
	})

	if suite.mintFeeCollector {
		// mint some coin to fee collector
		coins := sdk.NewCoins(sdk.NewCoin(evmtypes.DefaultEVMDenom, sdkmath.NewInt(int64(params.TxGas)-1)))
		genesisState := app.NewTestGenesisState(suite.app.AppCodec())
		balances := []banktypes.Balance{
			{
				Address: suite.app.AccountKeeper.GetModuleAddress(authtypes.FeeCollectorName).String(),
				Coins:   coins,
			},
		}
		var bankGenesis banktypes.GenesisState
		suite.app.AppCodec().MustUnmarshalJSON(genesisState[banktypes.ModuleName], &bankGenesis)
		// Update balances and total supply
		bankGenesis.Balances = append(bankGenesis.Balances, balances...)
		bankGenesis.Supply = bankGenesis.Supply.Add(coins...)
		genesisState[banktypes.ModuleName] = suite.app.AppCodec().MustMarshalJSON(&bankGenesis)

		// we marshal the genesisState of all module to a byte array
		stateBytes, err := tmjson.MarshalIndent(genesisState, "", " ")
		require.NoError(t, err)

		// Initialize the chain
		suite.app.InitChain(
			&abci.RequestInitChain{
				ChainId:         "ethermint_9000-1",
				Validators:      []abci.ValidatorUpdate{},
				ConsensusParams: app.DefaultConsensusParams,
				AppStateBytes:   stateBytes,
			},
		)
	}

	suite.ctx = suite.app.BaseApp.NewContextLegacy(checkTx, tmproto.Header{
		Height:          1,
		ChainID:         "ethermint_9000-1",
		Time:            time.Now().UTC(),
		ProposerAddress: suite.consAddress.Bytes(),
		Version: tmversion.Consensus{
			Block: version.BlockProtocol,
		},
		LastBlockId: tmproto.BlockID{
			Hash: tmhash.Sum([]byte("block_id")),
			PartSetHeader: tmproto.PartSetHeader{
				Total: 11,
				Hash:  tmhash.Sum([]byte("partset_header")),
			},
		},
		AppHash:            tmhash.Sum([]byte("app")),
		DataHash:           tmhash.Sum([]byte("data")),
		EvidenceHash:       tmhash.Sum([]byte("evidence")),
		ValidatorsHash:     tmhash.Sum([]byte("validators")),
		NextValidatorsHash: tmhash.Sum([]byte("next_validators")),
		ConsensusHash:      tmhash.Sum([]byte("consensus")),
		LastResultsHash:    tmhash.Sum([]byte("last_result")),
	})

	queryHelper := baseapp.NewQueryServerTestHelper(suite.ctx, suite.app.InterfaceRegistry())
	types.RegisterQueryServer(queryHelper, suite.app.Erc20Keeper)
	suite.queryClient = types.NewQueryClient(queryHelper)
	evmtypes.RegisterQueryServer(queryHelper, suite.app.EvmKeeper)
	suite.queryClientEvm = evmtypes.NewQueryClient(queryHelper)

	accNum := suite.app.AccountKeeper.NextAccountNumber(suite.ctx)
	acc := &ethermint.EthAccount{
		BaseAccount: authtypes.NewBaseAccount(sdk.AccAddress(suite.address.Bytes()), nil, accNum, 0),
		CodeHash:    common.BytesToHash(crypto.Keccak256(nil)).String(),
	}

	suite.app.AccountKeeper.SetAccount(suite.ctx, acc)

	valAddr := sdk.ValAddress(suite.address.Bytes())
	validator, err := stakingtypes.NewValidator(valAddr.String(), priv.PubKey(), stakingtypes.Description{})
	require.NoError(t, err)
	err = suite.app.StakingKeeper.SetValidatorByConsAddr(suite.ctx, validator)
	require.NoError(t, err)
	err = suite.app.StakingKeeper.SetValidatorByConsAddr(suite.ctx, validator)
	require.NoError(t, err)
	suite.app.StakingKeeper.SetValidator(suite.ctx, validator)

	suite.clientCtx = client.Context{}.WithTxConfig(suite.app.TxConfig())
	suite.ethSigner = ethtypes.LatestSignerForChainID(suite.app.EvmKeeper.ChainID())
	suite.appCodec = suite.app.AppCodec()
	suite.denom = evmtypes.DefaultEVMDenom
}

func (suite *KeeperTestSuite) Commit() {
	header := suite.ctx.BlockHeader()
	_, err := suite.app.FinalizeBlock(&abci.RequestFinalizeBlock{
		Height: header.Height,
	})
	suite.Require().NoError(err)

	_, err = suite.app.Commit()
	suite.Require().NoError(err)

	header.Height += 1
	suite.app.FinalizeBlock(&abci.RequestFinalizeBlock{
		Height: header.Height,
	})

	// update ctx
	suite.ctx = suite.app.BaseApp.NewContextLegacy(false, header)

	queryHelper := baseapp.NewQueryServerTestHelper(suite.ctx, suite.app.InterfaceRegistry())
	types.RegisterQueryServer(queryHelper, suite.app.Erc20Keeper)
	suite.queryClient = types.NewQueryClient(queryHelper)
	evmtypes.RegisterQueryServer(queryHelper, suite.app.EvmKeeper)
	suite.queryClientEvm = evmtypes.NewQueryClient(queryHelper)
}

func (suite *KeeperTestSuite) StateDB() *statedb.StateDB {
	return statedb.New(suite.ctx, suite.app.EvmKeeper, statedb.NewEmptyTxConfig(common.BytesToHash(suite.ctx.HeaderHash())))
}

func (suite *KeeperTestSuite) DeployContract(name, symbol string, decimals uint8) (common.Address, error) {
	from, priv := tests.NewAddrKey()

	chainID := suite.app.EvmKeeper.ChainID()

	nonce := suite.app.EvmKeeper.GetNonce(suite.ctx, from)

	ctorArgs, err := contracts.ERC20MinterBurnerDecimalsContract.ABI.Pack("", name, symbol, decimals)
	if err != nil {
		return common.Address{}, err
	}

	data := append(evmtypes.ERC20Contract.Bin, ctorArgs...) //nolint:gocritic
	gas, err := tx.GasLimit(suite.ctx, from, data, suite.queryClientEvm)
	if err != nil {
		return common.Address{}, err
	}

	msgEthereumTx := evmtypes.NewTx(
		chainID,
		nonce,
		nil,
		nil,
		gas,
		big.NewInt(200_000),
		suite.app.FeeMarketKeeper.GetBaseFee(suite.ctx),
		big.NewInt(1),
		data,
		&ethtypes.AccessList{},
	)
	msgEthereumTx.From = from.String()

	_, err = testutil.DeliverEthTx(suite.app, priv, msgEthereumTx)
	if err != nil {
		return common.Address{}, err
	}

	return crypto.CreateAddress(from, nonce), nil
}
