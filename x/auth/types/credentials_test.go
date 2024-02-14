package types_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
)

func TestNewModuleCrendentials(t *testing.T) {
	// wrong derivation keys
	_, err := authtypes.NewModuleCredential("group", []byte{})
	require.Error(t, err, "derivation keys must be non empty")
	_, err = authtypes.NewModuleCredential("group", [][]byte{{0x0, 0x30}, {}}...)
	require.Error(t, err)

	credential, err := authtypes.NewModuleCredential("group")
	require.NoError(t, err, "must be able to create a Root Module credential (see ADR-33)")
	require.NoError(t, sdk.VerifyAddressFormat(credential.Address()))

	credential, err = authtypes.NewModuleCredential("group", [][]byte{{0x20}, {0x0}}...)
	require.NoError(t, err)
	require.Error(t, sdk.VerifyAddressFormat(credential.Address()))
	_, err = sdk.AccAddressFromHex(credential.Address().String())
	require.Error(t, err)

	c, err := authtypes.NewModuleCredential("group", [][]byte{{0x20}, {0x0}}...)
	require.NoError(t, err)
	require.True(t, credential.Equals(c))

	c, err = authtypes.NewModuleCredential("group", [][]byte{{0x20}, {0x1}}...)
	require.NoError(t, err)
	require.False(t, credential.Equals(c))

	c, err = authtypes.NewModuleCredential("group", []byte{0x20})
	require.NoError(t, err)
	require.False(t, credential.Equals(c))

	address := secp256k1.GenPrivKey().PubKey().Address()
	expected := sdk.MustAccAddressFromHex(address.String())
	c, err = authtypes.NewModuleCredential("group", address)
	require.NoError(t, err)
	require.Equal(t, strings.ToLower(expected.String()), "0x"+strings.ToLower(address.String()))
}

func TestNewBaseAccountWithPubKey(t *testing.T) {
	pubKey := secp256k1.GenPrivKey().PubKey()
	address := pubKey.Address()
	expected := sdk.MustAccAddressFromHex(address.String())

	_, err := authtypes.NewModuleCredential("group", address)
	require.NoError(t, err)
	account, err := authtypes.NewBaseAccountWithPubKey(pubKey)
	require.NoError(t, err)
	require.Equal(t, expected, account.GetAddress())
	require.Equal(t, pubKey, account.GetPubKey())
}

func TestNewBaseAccountWithPubKey_WrongCredentials(t *testing.T) {
	_, err := authtypes.NewBaseAccountWithPubKey(cryptotypes.PubKey(nil))
	require.Error(t, err)
}
