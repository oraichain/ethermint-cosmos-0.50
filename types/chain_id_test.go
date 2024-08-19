package types

import (
	"math/big"
	"strings"
	"testing"

	"fmt"

	"github.com/stretchr/testify/require"
)

func TestParseChainID(t *testing.T) {
	fmt.Println(hashChainIdToInt("Oraichain"))
	testCases := []struct {
		name     string
		chainID  string
		expError bool
		expInt   *big.Int
	}{
		{
			"valid chain-id, single digit", "ethermint_1-1", false, big.NewInt(1),
		},
		{
			"valid chain-id, multiple digits", "aragonchain_256-1", false, big.NewInt(256),
		},
		{
			"valid chain-id, single digit", "Oraichain", false, hashChainIdToInt("Oraichain"),
		},
		{
			"invalid chain-id, double dash", "aragonchain-1-1", false, hashChainIdToInt("aragonchain-1-1"),
		},
		{
			"invalid chain-id, double underscore", "aragonchain_1_1", false, hashChainIdToInt("aragonchain_1_1"),
		},
		{
			"invalid chain-id, dash only", "-", false, hashChainIdToInt("-"),
		},
		{
			"invalid chain-id, undefined identifier and EIP155", "-1", false, hashChainIdToInt(("-1")),
		},
		{
			"invalid chain-id, undefined identifier", "_1-1", false, hashChainIdToInt("_1-1"),
		},
		{
			"invalid chain-id, uppercases", "ETHERMINT_1-1", false, hashChainIdToInt("ETHERMINT_1-1"),
		},
		{
			"invalid chain-id, mixed cases", "Ethermint_1-1", false, hashChainIdToInt("Ethermint_1-1"),
		},
		{
			"invalid chain-id, special chars", "$&*#!_1-1", false, hashChainIdToInt("$&*#!_1-1"),
		},
		{
			"invalid eip155 chain-id, cannot start with 0", "ethermint_001-1", false, hashChainIdToInt("ethermint_001-1"),
		},
		{
			"invalid eip155 chain-id, cannot invalid base", "ethermint_0x212-1", false, hashChainIdToInt("ethermint_0x212-1"),
		},
		{
			"invalid eip155 chain-id, non-integer", "ethermint_ethermint_9000-1", false, hashChainIdToInt("ethermint_ethermint_9000-1"),
		},
		{
			"invalid epoch, undefined", "ethermint_-", false, hashChainIdToInt("ethermint_-"),
		},
		{
			"blank chain ID", " ", true, nil,
		},
		{
			"empty chain ID", "", true, nil,
		},
		{
			"empty content for chain id, eip155 and epoch numbers", "_-", false, hashChainIdToInt("_-"),
		},
		{
			"long chain-id", "ethermint_" + strings.Repeat("1", 40) + "-1", true, nil,
		},
	}

	for _, tc := range testCases {
		chainIDEpoch, err := ParseChainID(tc.chainID)
		if tc.expError {
			require.Error(t, err, tc.name)
			require.Nil(t, chainIDEpoch)

			require.False(t, IsValidChainID(tc.chainID), tc.name)
		} else {
			require.NoError(t, err, tc.name)
			require.Equal(t, tc.expInt, chainIDEpoch, tc.name)
			require.True(t, IsValidChainID(tc.chainID))
		}
	}
}
