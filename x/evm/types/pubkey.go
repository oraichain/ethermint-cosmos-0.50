package types

import (
	"encoding/base64"

	errorsmod "cosmossdk.io/errors"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/ethereum/go-ethereum/common"
	"github.com/evmos/ethermint/crypto/ethsecp256k1"
)

// second half of go-ethereum/core/types/transaction_signing.go:recoverPlain
func PubkeyToEVMAddress(pub string) (*common.Address, error) {
	pubKeyBytes, err := base64.StdEncoding.DecodeString(pub)
	if err != nil {
		return nil, err
	}

	return PubkeyBytesToEVMAddress(pubKeyBytes)
}

func PubkeyBytesToEVMAddress(pubKeyBytes []byte) (*common.Address, error) {
	// Decompress the public key
	pubKey := ethsecp256k1.PubKey{Key: pubKeyBytes}
	address := common.Address(pubKey.Address().Bytes())
	return &address, nil
}

func PubkeyToCosmosAddress(pub string) (sdk.AccAddress, error) {
	pubKeyBytes, err := base64.StdEncoding.DecodeString(pub)
	if err != nil {
		return nil, err
	}
	return PubkeyBytesToCosmosAddress(pubKeyBytes)
}

func PubkeyBytesToCosmosAddress(pubKeyBytes []byte) (sdk.AccAddress, error) {
	pubkey := secp256k1.PubKey{Key: pubKeyBytes}
	if len(pubkey.Key) != secp256k1.PubKeySize {
		return nil, errorsmod.Wrap(sdkerrors.ErrInvalidPubKey, "length of pubkey is incorrect")
	}
	cosmosAddress := sdk.AccAddress(pubkey.Address().Bytes())
	return cosmosAddress, nil
}
