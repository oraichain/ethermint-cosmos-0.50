package types

import (
	"testing"

	"github.com/ethereum/go-ethereum/params"

	"github.com/stretchr/testify/require"
)

const (
	validEthAddress   = "0xc0ffee254729296a45a3885639AC7E10F9d54979"
	invalidEthAddress = "0xc0ffee254729296a45a3885639AC7E10F9d5497"
)

func TestParamsValidate(t *testing.T) {
	extraEips := []int64{2929, 1884, 1344}
	testCases := []struct {
		name      string
		getParams func() Params
		expError  bool
	}{
		{
			name:      "default",
			getParams: DefaultParams,
			expError:  false,
		},
		{
			name: "valid",
			getParams: func() Params {
				return NewParams("ara", false, true, true, DefaultChainConfig(), extraEips, []EIP712AllowedMsg{})
			},
			expError: false,
		},
		{
			name: "empty",
			getParams: func() Params {
				return Params{}
			},
			expError: true,
		},
		{
			name: "invalid evm denom",
			getParams: func() Params {
				return Params{
					EvmDenom: "@!#!@$!@5^32",
				}
			},
			expError: true,
		},
		{
			name: "invalid eip",
			getParams: func() Params {
				return Params{
					EvmDenom:  "stake",
					ExtraEIPs: []int64{1},
				}
			},
			expError: true,
		},
		{
			name: "valid enabled precompiles",
			getParams: func() Params {
				params := DefaultParams()
				params.EnabledPrecompiles = []string{validEthAddress}
				return params
			},
			expError: false,
		},
		{
			name: "invalid enabled precompiles",
			getParams: func() Params {
				params := DefaultParams()
				params.EnabledPrecompiles = []string{invalidEthAddress}
				return params
			},
			expError: true,
		},
	}

	for _, tc := range testCases {
		err := tc.getParams().Validate()

		if tc.expError {
			require.Error(t, err, tc.name)
		} else {
			require.NoError(t, err, tc.name)
		}
	}
}

func TestParamsEIPs(t *testing.T) {
	extraEips := []int64{2929, 1884, 1344}
	params := NewParams("ara", false, true, true, DefaultChainConfig(), extraEips, []EIP712AllowedMsg{})
	actual := params.EIPs()

	require.Equal(t, []int([]int{2929, 1884, 1344}), actual)
}

func TestParamsValidatePriv(t *testing.T) {
	require.Error(t, validateEVMDenom(false))
	require.NoError(t, validateEVMDenom("inj"))
	require.Error(t, validateBool(""))
	require.NoError(t, validateBool(true))
	require.Error(t, validateEIPs(""))
	require.NoError(t, validateEIPs([]int64{1884}))

	require.Error(t, validateEnabledPrecompiles([]string{""}))
	require.Error(t, validateEnabledPrecompiles([]string{invalidEthAddress}))
	require.Error(t, validateEnabledPrecompiles([]string{validEthAddress, invalidEthAddress}))
	require.NoError(t, validateEnabledPrecompiles(nil))
	require.NoError(t, validateEnabledPrecompiles([]string{}))
	require.NoError(t, validateEnabledPrecompiles([]string{validEthAddress}))
}

func TestValidateChainConfig(t *testing.T) {
	testCases := []struct {
		name     string
		i        interface{}
		expError bool
	}{
		{
			"invalid chain config type",
			"string",
			true,
		},
		{
			"valid chain config type",
			DefaultChainConfig(),
			false,
		},
	}
	for _, tc := range testCases {
		err := validateChainConfig(tc.i)

		if tc.expError {
			require.Error(t, err, tc.name)
		} else {
			require.NoError(t, err, tc.name)
		}
	}
}

func TestIsLondon(t *testing.T) {
	testCases := []struct {
		name   string
		height int64
		result bool
	}{
		{
			"Before london block",
			5,
			false,
		},
		{
			"After london block",
			12_965_001,
			true,
		},
		{
			"london block",
			12_965_000,
			true,
		},
	}

	for _, tc := range testCases {
		ethConfig := params.MainnetChainConfig
		require.Equal(t, IsLondon(ethConfig, tc.height), tc.result)
	}
}
