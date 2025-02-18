//go:build ledger
// +build ledger

package ledger

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/codec/legacy"
	"github.com/cosmos/cosmos-sdk/crypto/hd"
	"github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/cosmos/cosmos-sdk/testutil/testdata"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func TestErrorHandling(t *testing.T) {
	// first, try to generate a key, must return an error
	// (no panic)
	path := *hd.NewParams(44, 555, 0, false, 0)
	_, err := NewPrivKeySecp256k1Unsafe(path)
	require.Error(t, err)
}

func TestPublicKeyUnsafe(t *testing.T) {
	path := *hd.NewFundraiserParams(0, sdk.CoinType, 0)
	priv, err := NewPrivKeySecp256k1Unsafe(path)
	require.NoError(t, err)
	checkDefaultPubKey(t, priv)
}

func checkDefaultPubKey(t *testing.T, priv types.LedgerPrivKey) {
	require.NotNil(t, priv)
	expectedPkStr := "PubKeySecp256k1{034FEF9CD7C4C63588D3B03FEB5281B9D232CBA34D6F3D71AEE59211FFBFE1FE870000000000000000000000000000000000000000000000000000000000000000}"
	require.Equal(t, "6ca3b68f41034fef9cd7c4c63588d3b03feb5281b9d232cba34d6f3d71aee59211ffbfe1fe870000000000000000000000000000000000000000000000000000000000000000",
		fmt.Sprintf("%x", cdc.Amino.MustMarshalBinaryBare(priv.PubKey())),
		"Is your device using test mnemonic: %s ?", testdata.TestMnemonic)
	require.Equal(t, expectedPkStr, priv.PubKey().String())
	addr := sdk.AccAddress(priv.PubKey().Address()).String()
	require.Equal(t, "0xeb57861cadb108ccd1797fa05fae372fef639f01",
		addr, "Is your device using test mnemonic: %s ?", testdata.TestMnemonic)
}

func TestPublicKeyUnsafeHDPath(t *testing.T) {
	expectedAnswers := []string{
		"PubKeySecp256k1{034FEF9CD7C4C63588D3B03FEB5281B9D232CBA34D6F3D71AEE59211FFBFE1FE870000000000000000000000000000000000000000000000000000000000000000}",
		"PubKeySecp256k1{0260D0487A3DFCE9228EEE2D0D83A40F6131F551526C8E52066FE7FE1E4A5096660000000000000000000000000000000000000000000000000000000000000000}",
		"PubKeySecp256k1{03A2670393D02B162D0ED06A08041E80D86BE36C0564335254DF7462447EB69AB30000000000000000000000000000000000000000000000000000000000000000}",
		"PubKeySecp256k1{033222FC61795077791665544A90740E8EAD638A391A3B8F9261F4A226B396C0420000000000000000000000000000000000000000000000000000000000000000}",
		"PubKeySecp256k1{03F577473348D7B01E7AF2F245E36B98D181BC935EC8B552CDE5932B646DC7BE040000000000000000000000000000000000000000000000000000000000000000}",
		"PubKeySecp256k1{0222B1A5486BE0A2D5F3C5866BE46E05D1BDE8CDA5EA1C4C77A9BC48D2FA2753BC0000000000000000000000000000000000000000000000000000000000000000}",
		"PubKeySecp256k1{0377A1C826D3A03CA4EE94FC4DEA6BCCB2BAC5F2AC0419A128C29F8E88F1FF295A0000000000000000000000000000000000000000000000000000000000000000}",
		"PubKeySecp256k1{031B75C84453935AB76F8C8D0B6566C3FCC101CC5C59D7000BFC9101961E9308D90000000000000000000000000000000000000000000000000000000000000000}",
		"PubKeySecp256k1{038905A42433B1D677CC8AFD36861430B9A8529171B0616F733659F131C3F802210000000000000000000000000000000000000000000000000000000000000000}",
		"PubKeySecp256k1{038BE7F348902D8C20BC88D32294F4F3B819284548122229DECD1ADF1A7EB0848B0000000000000000000000000000000000000000000000000000000000000000}",
	}

	const numIters = 10

	privKeys := make([]types.LedgerPrivKey, numIters)

	// Check with device
	for i := uint32(0); i < 10; i++ {
		path := *hd.NewFundraiserParams(0, sdk.CoinType, i)
		t.Logf("Checking keys at %v\n", path)

		priv, err := NewPrivKeySecp256k1Unsafe(path)
		require.NoError(t, err)
		require.NotNil(t, priv)

		// Check other methods
		tmp := priv.(PrivKeyLedgerSecp256k1)
		require.NoError(t, tmp.ValidateKey())
		(&tmp).AssertIsPrivKeyInner()

		// in this test we are chekcking if the generated keys are correct.
		require.Equal(t, expectedAnswers[i], priv.PubKey().String(),
			"Is your device using test mnemonic: %s ?", testdata.TestMnemonic)

		// Store and restore
		serializedPk := priv.Bytes()
		require.NotNil(t, serializedPk)
		require.True(t, len(serializedPk) >= 50)

		privKeys[i] = priv
	}

	// Now check equality
	for i := 0; i < 10; i++ {
		for j := 0; j < 10; j++ {
			require.Equal(t, i == j, privKeys[i].Equals(privKeys[j]))
			require.Equal(t, i == j, privKeys[j].Equals(privKeys[i]))
		}
	}
}

func TestPublicKeySafe(t *testing.T) {
	path := *hd.NewFundraiserParams(0, sdk.CoinType, 0)
	priv, addr, err := NewPrivKeySecp256k1(path, "cosmos")

	require.NoError(t, err)
	require.NotNil(t, priv)
	require.Nil(t, ShowAddress(path, priv.PubKey(), ""))
	checkDefaultPubKey(t, priv)

	addr2 := sdk.AccAddress(priv.PubKey().Address()).String()
	require.Equal(t, addr, addr2)
}

func TestPublicKeyHDPath(t *testing.T) {
	expectedPubKeys := []string{
		"PubKeySecp256k1{034FEF9CD7C4C63588D3B03FEB5281B9D232CBA34D6F3D71AEE59211FFBFE1FE870000000000000000000000000000000000000000000000000000000000000000}",
		"PubKeySecp256k1{0260D0487A3DFCE9228EEE2D0D83A40F6131F551526C8E52066FE7FE1E4A5096660000000000000000000000000000000000000000000000000000000000000000}",
		"PubKeySecp256k1{03A2670393D02B162D0ED06A08041E80D86BE36C0564335254DF7462447EB69AB30000000000000000000000000000000000000000000000000000000000000000}",
		"PubKeySecp256k1{033222FC61795077791665544A90740E8EAD638A391A3B8F9261F4A226B396C0420000000000000000000000000000000000000000000000000000000000000000}",
		"PubKeySecp256k1{03F577473348D7B01E7AF2F245E36B98D181BC935EC8B552CDE5932B646DC7BE040000000000000000000000000000000000000000000000000000000000000000}",
		"PubKeySecp256k1{0222B1A5486BE0A2D5F3C5866BE46E05D1BDE8CDA5EA1C4C77A9BC48D2FA2753BC0000000000000000000000000000000000000000000000000000000000000000}",
		"PubKeySecp256k1{0377A1C826D3A03CA4EE94FC4DEA6BCCB2BAC5F2AC0419A128C29F8E88F1FF295A0000000000000000000000000000000000000000000000000000000000000000}",
		"PubKeySecp256k1{031B75C84453935AB76F8C8D0B6566C3FCC101CC5C59D7000BFC9101961E9308D90000000000000000000000000000000000000000000000000000000000000000}",
		"PubKeySecp256k1{038905A42433B1D677CC8AFD36861430B9A8529171B0616F733659F131C3F802210000000000000000000000000000000000000000000000000000000000000000}",
		"PubKeySecp256k1{038BE7F348902D8C20BC88D32294F4F3B819284548122229DECD1ADF1A7EB0848B0000000000000000000000000000000000000000000000000000000000000000}",
	}

	expectedAddrs := []string{
		"0xeb57861cadb108ccd1797fa05fae372fef639f01",
		"0xe377bd431aa149aa250eb737b858a0e00591f6c0",
		"0x7265d1ac62b2391ec548ad65ea91a39343f4d777",
		"0xbbe799da39dfcb240bd58a0772d3f6af70bf8acf",
		"0x1e0cd52b02c89fb4edcbedd1c88e89f4ee83a3f6",
		"0xf36c94b4227ce6c52e9692f3eae0519eb50d909b",
		"0x337b4a0ffd69b597cbb38f1d9369c8d7d21d8c07",
		"0x07bafa9ad0db633575f04ac0ab7c3ff4c7465630",
		"0x332694a86fe9cdc2ea712e3ec94683f9bdd49328",
		"0x8c9894c41fe25ff62edd64cf5aa748fedf7172e5",
	}

	const numIters = 10

	privKeys := make([]types.LedgerPrivKey, numIters)

	// Check with device
	for i := 0; i < len(expectedAddrs); i++ {
		path := *hd.NewFundraiserParams(0, sdk.CoinType, uint32(i))
		t.Logf("Checking keys at %s\n", path.String())

		priv, addr, err := NewPrivKeySecp256k1(path, "cosmos")
		require.NoError(t, err)
		require.NotNil(t, addr)
		require.NotNil(t, priv)

		addr2 := sdk.AccAddress(priv.PubKey().Address()).String()
		require.Equal(t, addr2, addr)
		require.Equal(t,
			expectedAddrs[i], addr,
			"Is your device using test mnemonic: %s ?", testdata.TestMnemonic)

		// Check other methods
		tmp := priv.(PrivKeyLedgerSecp256k1)
		require.NoError(t, tmp.ValidateKey())
		(&tmp).AssertIsPrivKeyInner()

		// in this test we are chekcking if the generated keys are correct and stored in a right path.
		require.Equal(t,
			expectedPubKeys[i], priv.PubKey().String(),
			"Is your device using test mnemonic: %s ?", testdata.TestMnemonic)

		// Store and restore
		serializedPk := priv.Bytes()
		require.NotNil(t, serializedPk)
		require.True(t, len(serializedPk) >= 50)

		privKeys[i] = priv
	}

	// Now check equality
	for i := 0; i < 10; i++ {
		for j := 0; j < 10; j++ {
			require.Equal(t, i == j, privKeys[i].Equals(privKeys[j]))
			require.Equal(t, i == j, privKeys[j].Equals(privKeys[i]))
		}
	}
}

func getFakeTx(accountNumber uint32) []byte {
	tmp := fmt.Sprintf(
		`{"account_number":"%d","chain_id":"1234","fee":{"amount":[{"amount":"150","denom":"atom"}],"gas":"5000"},"memo":"memo","msgs":[[""]],"sequence":"6"}`,
		accountNumber)

	return []byte(tmp)
}

func TestSignaturesHD(t *testing.T) {
	t.Skip("TODO HV2: fix and enable this test")
	for account := uint32(0); account < 100; account += 30 {
		msg := getFakeTx(account)

		path := *hd.NewFundraiserParams(account, sdk.CoinType, account/5)
		t.Logf("Checking signature at %v    ---   PLEASE REVIEW AND ACCEPT IN THE DEVICE\n", path)

		priv, err := NewPrivKeySecp256k1Unsafe(path)
		require.NoError(t, err)

		pub := priv.PubKey()
		sig, err := priv.Sign(msg)
		require.NoError(t, err)

		valid := pub.VerifySignature(msg, sig)
		require.True(t, valid, "Is your device using test mnemonic: %s ?", testdata.TestMnemonic)
	}
}

func TestRealDeviceSecp256k1(t *testing.T) {
	t.Skip("TODO HV2: fix and enable this test")
	msg := getFakeTx(50)
	path := *hd.NewFundraiserParams(0, sdk.CoinType, 0)
	priv, err := NewPrivKeySecp256k1Unsafe(path)
	require.NoError(t, err)

	pub := priv.PubKey()
	sig, err := priv.Sign(msg)
	require.NoError(t, err)

	valid := pub.VerifySignature(msg, sig)
	require.True(t, valid)

	// now, let's serialize the public key and make sure it still works
	bs := cdc.Amino.MustMarshalBinaryBare(priv.PubKey())
	pub2, err := legacy.PubKeyFromBytes(bs)
	require.Nil(t, err, "%+v", err)

	// make sure we get the same pubkey when we load from disk
	require.Equal(t, pub, pub2)

	// signing with the loaded key should match the original pubkey
	sig, err = priv.Sign(msg)
	require.NoError(t, err)
	valid = pub.VerifySignature(msg, sig)
	require.True(t, valid)

	// make sure pubkeys serialize properly as well
	bs = legacy.Cdc.MustMarshal(pub)
	bpub, err := legacy.PubKeyFromBytes(bs)
	require.NoError(t, err)
	require.Equal(t, pub, bpub)
}
