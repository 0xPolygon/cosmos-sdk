package keys

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	cryptocodec "github.com/cosmos/cosmos-sdk/crypto/codec"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	kmultisig "github.com/cosmos/cosmos-sdk/crypto/keys/multisig"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	"github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func generatePubKeys(n int) []types.PubKey {
	pks := make([]types.PubKey, n)
	for i := 0; i < n; i++ {
		pks[i] = secp256k1.GenPrivKey().PubKey()
	}
	return pks
}

func TestBech32KeysOutput(t *testing.T) {
	sk := secp256k1.PrivKey{Key: []byte{154, 49, 3, 117, 55, 232, 249, 20, 205, 216, 102, 7, 136, 72, 177, 2, 131, 202, 234, 81, 31, 208, 46, 244, 179, 192, 167, 163, 142, 117, 246, 13}}
	tmpKey := sk.PubKey()
	multisigPk := kmultisig.NewLegacyAminoPubKey(1, []types.PubKey{tmpKey})

	k, err := keyring.NewMultiRecord("multisig", multisigPk)
	require.NotNil(t, k)
	require.NoError(t, err)
	pubKey, err := k.GetPubKey()
	require.NoError(t, err)
	accAddr := sdk.AccAddress(pubKey.Address())
	expectedOutput, err := NewKeyOutput(k.Name, k.GetType(), accAddr, multisigPk)
	require.NoError(t, err)

	out, err := MkAccKeyOutput(k)
	require.NoError(t, err)
	require.Equal(t, expectedOutput, out)
	require.Equal(t, "{Name:multisig Type:multi Address:0xe7e1add937d6cd091d6a0db2ea44eef5c1613c72 PubKey:{\"@type\":\"/cosmos.crypto.multisig.LegacyAminoPubKey\",\"threshold\":1,\"public_keys\":[{\"@type\":\"/cosmos.crypto.secp256k1.PubKey\",\"key\":\"BOrroA7jvfPd1AadmmOvWM2rJSwipXfRf8yD6pLbA2DJRBvMr2v1o7/Tt5MFOm9RcJhxfM8axlq8SseN3Dl07xg=\"}]} Mnemonic:}", fmt.Sprintf("%+v", out))
}

// TestBech32KeysOutputNestedMsig tests that the output of a nested multisig key is correct
func TestBech32KeysOutputNestedMsig(t *testing.T) {
	sk := secp256k1.PrivKey{Key: []byte{154, 49, 3, 117, 55, 232, 249, 20, 205, 216, 102, 7, 136, 72, 177, 2, 131, 202, 234, 81, 31, 208, 46, 244, 179, 192, 167, 163, 142, 117, 246, 13}}
	tmpKey := sk.PubKey()
	nestedMultiSig := kmultisig.NewLegacyAminoPubKey(1, []types.PubKey{tmpKey})
	multisigPk := kmultisig.NewLegacyAminoPubKey(2, []types.PubKey{tmpKey, nestedMultiSig})
	k, err := keyring.NewMultiRecord("multisig", multisigPk)
	require.NotNil(t, k)
	require.NoError(t, err)

	pubKey, err := k.GetPubKey()
	require.NoError(t, err)

	accAddr := sdk.AccAddress(pubKey.Address())
	expectedOutput, err := NewKeyOutput(k.Name, k.GetType(), accAddr, multisigPk)
	require.NoError(t, err)

	out, err := MkAccKeyOutput(k)
	require.NoError(t, err)

	require.Equal(t, expectedOutput, out)
	require.Equal(t, "{Name:multisig Type:multi Address:0x7d1c3ef842dcc796df25d46a0716005d30de5a9a PubKey:{\"@type\":\"/cosmos.crypto.multisig.LegacyAminoPubKey\",\"threshold\":2,\"public_keys\":[{\"@type\":\"/cosmos.crypto.secp256k1.PubKey\",\"key\":\"BOrroA7jvfPd1AadmmOvWM2rJSwipXfRf8yD6pLbA2DJRBvMr2v1o7/Tt5MFOm9RcJhxfM8axlq8SseN3Dl07xg=\"},{\"@type\":\"/cosmos.crypto.multisig.LegacyAminoPubKey\",\"threshold\":1,\"public_keys\":[{\"@type\":\"/cosmos.crypto.secp256k1.PubKey\",\"key\":\"BOrroA7jvfPd1AadmmOvWM2rJSwipXfRf8yD6pLbA2DJRBvMr2v1o7/Tt5MFOm9RcJhxfM8axlq8SseN3Dl07xg=\"}]}]} Mnemonic:}", fmt.Sprintf("%+v", out))
}

func TestProtoMarshalJSON(t *testing.T) {
	require := require.New(t)
	pubkeys := generatePubKeys(3)
	msig := kmultisig.NewLegacyAminoPubKey(2, pubkeys)

	registry := codectypes.NewInterfaceRegistry()
	cryptocodec.RegisterInterfaces(registry)
	cdc := codec.NewProtoCodec(registry)

	bz, err := cdc.MarshalInterfaceJSON(msig)
	require.NoError(err)

	var pk2 types.PubKey
	err = cdc.UnmarshalInterfaceJSON(bz, &pk2)
	require.NoError(err)
	require.True(pk2.Equals(msig))

	// Test that we can correctly unmarshal key from output
	k, err := keyring.NewMultiRecord("my multisig", msig)
	require.NoError(err)
	ko, err := MkAccKeyOutput(k)
	require.NoError(err)
	require.Equal(ko.Address, sdk.AccAddress(pk2.Address()).String())
	require.Equal(ko.PubKey, string(bz))
}
