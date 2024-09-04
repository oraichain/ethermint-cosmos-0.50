package hd

import (
	"fmt"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/ethereum/go-ethereum/common"

	amino "github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"

	cryptocodec "github.com/evmos/ethermint/crypto/codec"
	enccodec "github.com/evmos/ethermint/encoding/codec"
	ethermint "github.com/evmos/ethermint/types"
)

var TestCodec amino.Codec

func init() {
	cdc := amino.NewLegacyAmino()
	cryptocodec.RegisterCrypto(cdc)

	interfaceRegistry := types.NewInterfaceRegistry()
	TestCodec = amino.NewProtoCodec(interfaceRegistry)
	enccodec.RegisterInterfaces(interfaceRegistry)
}

const (
	mnemonic = "picnic rent average infant boat squirrel federal assault mercy purity very motor fossil wheel verify upset box fresh horse vivid copy predict square regret"

	// hdWalletFixEnv defines whether the standard (correct) bip39
	// derivation path was used, or if derivation was affected by
	// https://github.com/btcsuite/btcutil/issues/172
	hdWalletFixEnv = "GO_ETHEREUM_HDWALLET_FIX_ISSUE_179"
)

func TestKeyring(t *testing.T) {
	dir := t.TempDir()
	mockIn := strings.NewReader("")
	kr, err := keyring.New("ethermint", keyring.BackendTest, dir, mockIn, TestCodec, EthSecp256k1Option())
	require.NoError(t, err)

	// fail in retrieving key
	info, err := kr.Key("foo")
	require.Error(t, err)
	require.Nil(t, info)

	mockIn.Reset("password\npassword\n")
	info, mnemonic, err := kr.NewMnemonic("foo", keyring.English, ethermint.BIP44HDPath, keyring.DefaultBIP39Passphrase, EthSecp256k1)
	require.NoError(t, err)
	require.NotEmpty(t, mnemonic)
	require.Equal(t, "foo", info.Name)
	require.Equal(t, "local", info.GetType().String())
	pubKey, err := info.GetPubKey()
	require.NoError(t, err)
	require.Equal(t, string(EthSecp256k1Type), pubKey.Type())

	hdPath := ethermint.BIP44HDPath

	bz, err := EthSecp256k1.Derive()(mnemonic, keyring.DefaultBIP39Passphrase, hdPath)
	require.NoError(t, err)
	require.NotEmpty(t, bz)

	wrongBz, err := EthSecp256k1.Derive()(mnemonic, keyring.DefaultBIP39Passphrase, "/wrong/hdPath")
	require.Error(t, err)
	require.Empty(t, wrongBz)

	privkey := EthSecp256k1.Generate()(bz)
	addr := common.BytesToAddress(privkey.PubKey().Address().Bytes())

	os.Setenv(hdWalletFixEnv, "true")
	wallet, err := NewFromMnemonic(mnemonic)
	os.Setenv(hdWalletFixEnv, "")
	require.NoError(t, err)

	path := MustParseDerivationPath(hdPath)

	account, err := wallet.Derive(path, false)
	require.NoError(t, err)
	require.Equal(t, addr.String(), account.Address.String())
}

func TestDerivation(t *testing.T) {
	bz, err := EthSecp256k1.Derive()(mnemonic, keyring.DefaultBIP39Passphrase, ethermint.BIP44HDPath)
	require.NoError(t, err)
	require.NotEmpty(t, bz)

	badBz, err := EthSecp256k1.Derive()(mnemonic, keyring.DefaultBIP39Passphrase, "44'/60'/0'/0/0")
	require.NoError(t, err)
	require.NotEmpty(t, badBz)

	require.NotEqual(t, bz, badBz)

	privkey := EthSecp256k1.Generate()(bz)
	badPrivKey := EthSecp256k1.Generate()(badBz)

	require.False(t, privkey.Equals(badPrivKey))

	wallet, err := NewFromMnemonic(mnemonic)
	require.NoError(t, err)

	path := MustParseDerivationPath(ethermint.BIP44HDPath)
	account, err := wallet.Derive(path, false)
	require.NoError(t, err)

	badPath := MustParseDerivationPath("44'/60'/0'/0/0")
	badAccount, err := wallet.Derive(badPath, false)
	require.NoError(t, err)

	// Equality of Address BIP44
	require.Equal(t, account.Address.String(), "0xA588C66983a81e800Db4dF74564F09f91c026351")
	require.Equal(t, badAccount.Address.String(), "0xF8D6FDf2B8b488ea37e54903750dcd13F67E71cb")
	// Inequality of wrong derivation path address
	require.NotEqual(t, account.Address.String(), badAccount.Address.String())
	// Equality of Ethermint implementation
	require.Equal(t, common.BytesToAddress(privkey.PubKey().Address().Bytes()).String(), "0xA588C66983a81e800Db4dF74564F09f91c026351")
	require.Equal(t, common.BytesToAddress(badPrivKey.PubKey().Address().Bytes()).String(), "0xF8D6FDf2B8b488ea37e54903750dcd13F67E71cb")

	// Equality of Eth and Ethermint implementation
	require.Equal(t, common.BytesToAddress(privkey.PubKey().Address()).String(), account.Address.String())
	require.Equal(t, common.BytesToAddress(badPrivKey.PubKey().Address()).String(), badAccount.Address.String())

	// Inequality of wrong derivation path of Eth and Ethermint implementation
	require.NotEqual(t, common.BytesToAddress(privkey.PubKey().Address()).String(), badAccount.Address.String())
	require.NotEqual(t, common.BytesToAddress(badPrivKey.PubKey().Address()).String(), account.Address.Hex())
}

func TestArmorUnarmorPubKey(t *testing.T) {
	// Select the encryption and storage for your cryptostore

	tests := []struct {
		name              string
		uid               string
		backend           string
		userInput         io.Reader
		encryptPassphrase string
		importUID         string
		importPassphrase  string
	}{
		{
			name:              "export import",
			uid:               "testOne",
			backend:           keyring.BackendTest,
			userInput:         nil,
			encryptPassphrase: "apassphrase",
			importUID:         "importedKey",
			importPassphrase:  "apassphrase",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			kb, err := keyring.New("TestExport", tt.backend, t.TempDir(), tt.userInput, TestCodec, EthSecp256k1Option())
			require.NoError(t, err)
			k, _, err := kb.NewMnemonic(tt.uid, keyring.English, ethermint.BIP44HDPath, keyring.DefaultBIP39Passphrase, EthSecp256k1)
			require.NoError(t, err)
			require.NotNil(t, k)
			require.Equal(t, k.Name, tt.uid)

			record, err := kb.Key(tt.uid)
			require.NoError(t, err)
			require.Equal(t, record.Name, tt.uid)
			key, err := k.GetPubKey()
			require.NoError(t, err)
			fmt.Println("key: ", key)

			// will fail if codec not register ethsecp256k1 interface: enccodec.RegisterInterfaces(interfaceRegistry)
			armor, err := kb.ExportPrivKeyArmor(tt.uid, tt.encryptPassphrase)
			require.NoError(t, err)
			fmt.Println("armor: ", armor)
		})
	}
}
