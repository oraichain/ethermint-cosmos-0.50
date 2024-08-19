package types

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPubkeyToEVMAddress(t *testing.T) {
	pubkey := "Ah4NweWyFaVG5xcOwY5I7Tm4mmfPgLtS+Qn3jvXLX0VP"
	actualEvmAddress, err := PubkeyToEVMAddress(pubkey)
	require.NoError(t, err)
	require.Equal(t, "0x39D8810d16Bc6E8888F78E7F01D8B9999CE03499", actualEvmAddress.Hex())
}
