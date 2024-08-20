package keeper_test

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/runtime"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	"github.com/evmos/ethermint/contracts"
	"github.com/evmos/ethermint/tests"
	evmtypes "github.com/evmos/ethermint/x/evm/types"
	"github.com/stretchr/testify/mock"

	"github.com/evmos/ethermint/x/erc20/keeper"
	"github.com/evmos/ethermint/x/erc20/types"
	erc20mocks "github.com/evmos/ethermint/x/erc20/types/mocks"
)

func (suite *KeeperTestSuite) TestBalanceOf() {
	var mockEVMKeeper *erc20mocks.EVMKeeper
	contract := tests.GenerateAddress()
	testCases := []struct {
		name       string
		malleate   func()
		expBalance int64
		res        bool
	}{
		{
			"Failed to call Evm",
			func() {
				mockEVMKeeper.On("CallEVM", mock.Anything, mock.Anything, mock.Anything, mock.Anything,
					mock.Anything, mock.Anything, mock.Anything).Return(nil, fmt.Errorf("forced ApplyMessage error"))
			},
			int64(0),
			false,
		},
		{
			"Incorrect res",
			func() {
				mockEVMKeeper.On("CallEVM", mock.Anything, mock.Anything, mock.Anything, mock.Anything,
					mock.Anything, mock.Anything, mock.Anything).Return(&evmtypes.MsgEthereumTxResponse{Ret: []uint8{0, 0}}, nil).Once()
			},
			int64(0),
			false,
		},
		{
			"Correct Execution",
			func() {
				balance := make([]uint8, 32)
				balance[31] = uint8(10)
				mockEVMKeeper.On("CallEVM", mock.Anything, mock.Anything, mock.Anything, mock.Anything,
					mock.Anything, mock.Anything, mock.Anything).Return(&evmtypes.MsgEthereumTxResponse{Ret: balance}, nil).Once()
			},
			int64(10),
			true,
		},
	}
	for _, tc := range testCases {
		suite.SetupTest() // reset
		mockEVMKeeper = &erc20mocks.EVMKeeper{}
		suite.app.Erc20Keeper = keeper.NewKeeper(
			runtime.NewKVStoreService(suite.app.GetKey("erc20")), suite.app.AppCodec(),
			authtypes.NewModuleAddress(govtypes.ModuleName),
			suite.app.AccountKeeper, suite.app.BankKeeper,
			mockEVMKeeper, suite.app.StakingKeeper,
			suite.app.AuthzKeeper,
		)

		tc.malleate()

		abi := contracts.ERC20MinterBurnerDecimalsContract.ABI
		balance := suite.app.Erc20Keeper.BalanceOf(suite.ctx, abi, contract, tests.GenerateAddress())
		if tc.res {
			suite.Require().Equal(balance.Int64(), tc.expBalance)
		} else {
			suite.Require().Nil(balance)
		}
	}
}

func (suite *KeeperTestSuite) TestQueryERC20ForceFail() {
	var mockEVMKeeper *erc20mocks.EVMKeeper
	contract := tests.GenerateAddress()
	testCases := []struct {
		name     string
		malleate func()
		res      bool
	}{
		{
			"Failed to call Evm",
			func() {
				mockEVMKeeper.On("CallEVM", mock.Anything, mock.Anything, mock.Anything, mock.Anything,
					mock.Anything, mock.Anything).Return(nil, fmt.Errorf("forced ApplyMessage error"))
			},
			false,
		},
		{
			"Incorrect res",
			func() {
				mockEVMKeeper.On("CallEVM", mock.Anything, mock.Anything, mock.Anything, mock.Anything,
					mock.Anything, mock.Anything).Return(&evmtypes.MsgEthereumTxResponse{Ret: []uint8{0, 0}}, nil).Once()
			},
			false,
		},
		{
			"Correct res for name - incorrect for symbol",
			func() {
				ret := []uint8{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 32, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 10, 67, 111, 105, 110, 32, 84, 111, 107, 101, 110, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}
				mockEVMKeeper.On("ApplyMessage", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(&evmtypes.MsgEthereumTxResponse{Ret: ret}, nil).Once()
				mockEVMKeeper.On("CallEVM", mock.Anything, mock.Anything, mock.Anything, mock.Anything,
					mock.Anything, mock.Anything).Return(&evmtypes.MsgEthereumTxResponse{VmError: "Error"}, nil).Once()
			},
			false,
		},
		{
			"incorrect symbol res",
			func() {
				ret := []uint8{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 32, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 10, 67, 111, 105, 110, 32, 84, 111, 107, 101, 110, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}
				mockEVMKeeper.On("CallEVM", mock.Anything, mock.Anything, mock.Anything, mock.Anything,
					mock.Anything, mock.Anything).Return(&evmtypes.MsgEthereumTxResponse{Ret: ret}, nil).Once()
				mockEVMKeeper.On("CallEVM", mock.Anything, mock.Anything, mock.Anything, mock.Anything,
					mock.Anything, mock.Anything).Return(&evmtypes.MsgEthereumTxResponse{Ret: []uint8{0, 0}}, nil).Once()
			},
			false,
		},
		{
			"Correct res for name - incorrect for symbol",
			func() {
				ret := []uint8{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 32, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 10, 67, 111, 105, 110, 32, 84, 111, 107, 101, 110, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}
				retSymbol := []uint8{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 32, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 4, 67, 84, 75, 78, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}
				mockEVMKeeper.On("CallEVM", mock.Anything, mock.Anything, mock.Anything, mock.Anything,
					mock.Anything, mock.Anything).Return(&evmtypes.MsgEthereumTxResponse{Ret: ret}, nil).Once()
				mockEVMKeeper.On("CallEVM", mock.Anything, mock.Anything, mock.Anything, mock.Anything,
					mock.Anything, mock.Anything).Return(&evmtypes.MsgEthereumTxResponse{Ret: retSymbol}, nil).Once()
				mockEVMKeeper.On("CallEVM", mock.Anything, mock.Anything, mock.Anything, mock.Anything,
					mock.Anything, mock.Anything).Return(&evmtypes.MsgEthereumTxResponse{VmError: "Error"}, nil).Once()
			},
			false,
		},
		{
			"incorrect symbol res",
			func() {
				ret := []uint8{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 32, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 10, 67, 111, 105, 110, 32, 84, 111, 107, 101, 110, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}
				retSymbol := []uint8{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 32, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 4, 67, 84, 75, 78, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}
				mockEVMKeeper.On("CallEVM", mock.Anything, mock.Anything, mock.Anything, mock.Anything,
					mock.Anything, mock.Anything).Return(&evmtypes.MsgEthereumTxResponse{Ret: ret}, nil).Once()
				mockEVMKeeper.On("CallEVM", mock.Anything, mock.Anything, mock.Anything, mock.Anything,
					mock.Anything, mock.Anything).Return(&evmtypes.MsgEthereumTxResponse{Ret: retSymbol}, nil).Once()
				mockEVMKeeper.On("CallEVM", mock.Anything, mock.Anything, mock.Anything, mock.Anything,
					mock.Anything, mock.Anything).Return(&evmtypes.MsgEthereumTxResponse{Ret: []uint8{0, 0}}, nil).Once()
			},
			false,
		},
	}
	for _, tc := range testCases {
		suite.SetupTest() // reset

		// TODO: what's the reason we are using mockEVMKeeper here? Instead of just passing the suite.app.EvmKeeper?
		mockEVMKeeper = &erc20mocks.EVMKeeper{}
		suite.app.Erc20Keeper = keeper.NewKeeper(
			runtime.NewKVStoreService(suite.app.GetKey("erc20")), suite.app.AppCodec(),
			authtypes.NewModuleAddress(govtypes.ModuleName), suite.app.AccountKeeper,
			suite.app.BankKeeper, mockEVMKeeper, suite.app.StakingKeeper,
			suite.app.AuthzKeeper,
		)

		tc.malleate()

		res, err := suite.app.Erc20Keeper.QueryERC20(suite.ctx, contract)
		if tc.res {
			suite.Require().NoError(err)
			suite.Require().Equal(
				types.ERC20Data{Name: "coin", Symbol: "token", Decimals: uint8(18)},
				res,
			)
		} else {
			suite.Require().Error(err)
		}
	}
}
