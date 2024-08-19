package types

import (
	"encoding/base64"

	errorsmod "cosmossdk.io/errors"
	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

// second half of go-ethereum/core/types/transaction_signing.go:recoverPlain
func PubkeyToEVMAddress(pub string) (*common.Address, error) {
	pubKeyBytes, err := base64.StdEncoding.DecodeString(pub)
	if err != nil {
		return nil, err
	}
	// Decompress the public key
	pubKey, err := btcec.ParsePubKey(pubKeyBytes)
	if err != nil {
		return nil, err
	}

	// Convert to uncompressed format
	uncompressedPubKeyBytes := pubKey.SerializeUncompressed()
	// Parse the public key
	ethPubkey, err := crypto.UnmarshalPubkey(uncompressedPubKeyBytes)
	if err != nil {
		return nil, err
	}
	evmAddress := crypto.PubkeyToAddress(*ethPubkey)
	return &evmAddress, nil
}

func pubkeyBytesToPubkey(pub string) (*secp256k1.PubKey, error) {
	pubKeyBytes, err := base64.StdEncoding.DecodeString(pub)
	if err != nil {
		return nil, err
	}
	pubKey, err := btcec.ParsePubKey(pubKeyBytes)
	if err != nil {
		return nil, err
	}
	return &secp256k1.PubKey{Key: pubKey.SerializeCompressed()}, nil
}

func pubkeyToAddress(pub secp256k1.PubKey) sdk.AccAddress {
	return sdk.AccAddress(pub.Address())
}

func PubkeyToAddress(pub string) (sdk.AccAddress, error) {
	pubkey, err := pubkeyBytesToPubkey(pub)
	if err != nil {
		return nil, errorsmod.Wrap(sdkerrors.ErrInvalidPubKey, "The pubkey is invalid")
	}
	cosmosAddress := pubkeyToAddress(*pubkey)
	return cosmosAddress, nil
}
