package testutil

import (
	_ "embed"
	"encoding/json"

	evmtypes "github.com/evmos/ethermint/x/evm/types"
)

var (
	//go:embed ERC20Contract.json
	erc20JSON []byte

	// ERC20Contract is the compiled test erc20 contract
	ERC20Contract evmtypes.CompiledContract

	//go:embed SimpleStorageContract.json
	simpleStorageJSON []byte

	// SimpleStorageContract is the compiled test simple storage contract
	SimpleStorageContract evmtypes.CompiledContract

	//go:embed TestMessageCall.json
	testMessageCallJSON []byte

	// TestMessageCall is the compiled message call benchmark contract
	TestMessageCall evmtypes.CompiledContract
)

func init() {
	err := json.Unmarshal(erc20JSON, &ERC20Contract)
	if err != nil {
		panic(err)
	}

	if len(ERC20Contract.Bin) == 0 {
		panic("load contract failed")
	}

	err = json.Unmarshal(testMessageCallJSON, &TestMessageCall)
	if err != nil {
		panic(err)
	}

	if len(TestMessageCall.Bin) == 0 {
		panic("load contract failed")
	}

	err = json.Unmarshal(simpleStorageJSON, &SimpleStorageContract)
	if err != nil {
		panic(err)
	}

	if len(TestMessageCall.Bin) == 0 {
		panic("load contract failed")
	}
}
