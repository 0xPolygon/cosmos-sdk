package types_test

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"github.com/cosmos/cosmos-sdk/codec/address"
	mathrand "math/rand"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"sigs.k8s.io/yaml"

	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/cosmos/cosmos-sdk/types"
)

type addressTestSuite struct {
	suite.Suite
}

func TestAddressTestSuite(t *testing.T) {
	suite.Run(t, new(addressTestSuite))
}

func (s *addressTestSuite) SetupSuite() {
	s.T().Parallel()
}

var invalidStrs = []string{
	"hello, world!",
	"0xAA",
	"AAA",
}

func (s *addressTestSuite) testMarshal(original, res interface{}, marshal func() ([]byte, error), unmarshal func([]byte) error) {
	bz, err := marshal()
	s.Require().Nil(err)
	s.Require().Nil(unmarshal(bz))
	s.Require().Equal(original, res)
}

func (s *addressTestSuite) TestEmptyAddresses() {
	s.T().Parallel()
	s.Require().Equal((types.AccAddress{}).String(), "")
	s.Require().Equal((types.ValAddress{}).String(), "")
	s.Require().Equal((types.ConsAddress{}).String(), "")

	accAddr, err := types.AccAddressFromHex("")
	s.Require().True(accAddr.Empty())
	s.Require().Error(err)

	valAddr, err := types.ValAddressFromHex("")
	s.Require().True(valAddr.Empty())
	s.Require().Error(err)

	consAddr, err := types.ConsAddressFromHex("")
	s.Require().True(consAddr.Empty())
	s.Require().Error(err)
}

func (s *addressTestSuite) TestYAMLMarshalers() {
	addr := secp256k1.GenPrivKey().PubKey().Address()

	acc := types.AccAddress(addr)
	val := types.ValAddress(addr)
	cons := types.ConsAddress(addr)

	got, _ := yaml.Marshal(&acc)
	s.Require().Equal(acc.String()+"\n", string(got))

	got, _ = yaml.Marshal(&val)
	s.Require().Equal(val.String()+"\n", string(got))

	got, _ = yaml.Marshal(&cons)
	s.Require().Equal(cons.String()+"\n", string(got))
}

func (s *addressTestSuite) TestRandHexAccAddrConsistency() {
	pubBz := make([]byte, ed25519.PubKeySize)
	pub := &ed25519.PubKey{Key: pubBz}

	for i := 0; i < 1000; i++ {
		rand.Read(pub.Key)

		acc := types.AccAddress(pub.Address())
		res := types.AccAddress{}

		s.testMarshal(&acc, &res, acc.MarshalJSON, (&res).UnmarshalJSON)
		s.testMarshal(&acc, &res, acc.Marshal, (&res).Unmarshal)

		str := acc.String()
		res, err := types.AccAddressFromHex(str)
		s.Require().Nil(err)
		s.Require().Equal(acc, res)

		str = hex.EncodeToString(acc)
		res, err = types.AccAddressFromHex(str)
		s.Require().Nil(err)
		s.Require().Equal(acc, res)
	}

	for _, str := range invalidStrs {
		_, err := types.AccAddressFromHex(str)
		s.Require().NotNil(err)

		_, err = types.AccAddressFromHex(str)
		s.Require().NotNil(err)

		err = (*types.AccAddress)(nil).UnmarshalJSON([]byte("\"" + str + "\""))
		s.Require().NotNil(err)
	}

	_, err := types.AccAddressFromHex("")
	s.Require().Contains(err.Error(), types.ErrEmptyHexAddress)
}

// This test is inherited from cosmos-sdk upstream
// It used to test that the account address cache ignores the bech32 prefix setting, retrieving addresses from the cache.
// See https://github.com/cosmos/cosmos-sdk/issues/15317.
// In heimdall, the test just checks that two addresses generated from the same key are equal.
// Kept the test to reduce the diff with upstream.
func (s *addressTestSuite) TestAddrCache() {
	// Use a random key
	pubBz := make([]byte, ed25519.PubKeySize)
	pub := &ed25519.PubKey{Key: pubBz}
	rand.Read(pub.Key)

	acc := types.AccAddress(pub.Address())
	osmoAddrHex := acc.String()

	addrCosmos := types.AccAddress(pub.Address())
	cosmosAddrHex := addrCosmos.String()

	// The default behavior will retrieve the address from cache
	s.Require().Equal(osmoAddrHex, cosmosAddrHex)
}

// Test that the bech32 prefix is respected when the address cache is disabled.
// This causes AccAddress.String() to print out the expected prefixes if the config is changed between bech32 lookups.
// See https://github.com/cosmos/cosmos-sdk/issues/15317.
func (s *addressTestSuite) TestAddrCacheDisabled() {
	types.SetAddrCacheEnabled(false)

	// Use a random key
	pubBz := make([]byte, ed25519.PubKeySize)
	pub := &ed25519.PubKey{Key: pubBz}
	rand.Read(pub.Key)

	acc := types.AccAddress(pub.Address())
	osmoAddrHex := acc.String()

	addrCosmos := types.AccAddress(pub.Address())
	cosmosAddrHex := addrCosmos.String()

	// retrieve the address from the cache
	s.Require().Equal(osmoAddrHex, cosmosAddrHex)
}

func (s *addressTestSuite) TestValAddr() {
	pubBz := make([]byte, ed25519.PubKeySize)
	pub := &ed25519.PubKey{Key: pubBz}

	for i := 0; i < 20; i++ {
		rand.Read(pub.Key)

		acc := types.ValAddress(pub.Address())
		res := types.ValAddress{}

		s.testMarshal(&acc, &res, acc.MarshalJSON, (&res).UnmarshalJSON)
		s.testMarshal(&acc, &res, acc.Marshal, (&res).Unmarshal)

		str := acc.String()
		res, err := types.ValAddressFromHex(str)
		s.Require().Nil(err)
		s.Require().Equal(acc, res)

		str = hex.EncodeToString(acc)
		res, err = types.ValAddressFromHex(str)
		s.Require().Nil(err)
		s.Require().Equal(acc, res)

	}

	for _, str := range invalidStrs {
		_, err := types.ValAddressFromHex(str)
		s.Require().NotNil(err)

		_, err = types.ValAddressFromHex(str)
		s.Require().NotNil(err)

		err = (*types.ValAddress)(nil).UnmarshalJSON([]byte("\"" + str + "\""))
		s.Require().NotNil(err)
	}

	// test empty string
	_, err := types.ValAddressFromHex("")
	s.Require().ErrorContains(err, "empty address string is not allowed")
}

func (s *addressTestSuite) TestConsAddress() {
	pubBz := make([]byte, ed25519.PubKeySize)
	pub := &ed25519.PubKey{Key: pubBz}

	for i := 0; i < 20; i++ {
		rand.Read(pub.Key[:])

		acc := types.ConsAddress(pub.Address())
		res := types.ConsAddress{}

		s.testMarshal(&acc, &res, acc.MarshalJSON, (&res).UnmarshalJSON)
		s.testMarshal(&acc, &res, acc.Marshal, (&res).Unmarshal)

		str := acc.String()
		res, err := types.ConsAddressFromHex(str)
		s.Require().Nil(err)
		s.Require().Equal(acc, res)

		str = hex.EncodeToString(acc)
		res, err = types.ConsAddressFromHex(str)
		s.Require().Nil(err)
		s.Require().Equal(acc, res)
	}

	for _, str := range invalidStrs {
		_, err := types.ConsAddressFromHex(str)
		s.Require().NotNil(err)

		_, err = types.ConsAddressFromHex(str)
		s.Require().NotNil(err)

		err = (*types.ConsAddress)(nil).UnmarshalJSON([]byte("\"" + str + "\""))
		s.Require().NotNil(err)
	}

	// test empty string
	_, err := types.ConsAddressFromHex("")
	s.Require().ErrorContains(err, "empty address string is not allowed")
}

const letterBytes = "abcdefghijklmnopqrstuvwxyz"

func RandString(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[mathrand.Intn(len(letterBytes))]
	}
	return string(b)
}

// HV2: removed as irrelevant to Heimdall
/*
func (s *addressTestSuite) TestConfiguredPrefix() {
	pubBz := make([]byte, ed25519.PubKeySize)
	pub := &ed25519.PubKey{Key: pubBz}
	for length := 1; length < 10; length++ {
		for times := 1; times < 20; times++ {
			rand.Read(pub.Key[:])
			// Test if randomly generated prefix of a given length works
			prefix := RandString(length)

			// Assuming that GetConfig is not sealed.
			config := types.GetConfig()
			config.SetBech32PrefixForAccount(
				prefix+types.PrefixAccount,
				prefix+types.PrefixPublic)

			acc := types.AccAddress(pub.Address())
			s.Require().True(strings.HasPrefix(
				acc.String(),
				prefix+types.PrefixAccount), acc.String())

			bech32Pub := legacybech32.MustMarshalPubKey(legacybech32.AccPK, pub) //nolint:staticcheck // SA1019: legacybech32 is deprecated: use the bech32 package instead.
			s.Require().True(strings.HasPrefix(
				bech32Pub,
				prefix+types.PrefixPublic))

			config.SetBech32PrefixForValidator(
				prefix+types.PrefixValidator+types.PrefixAddress,
				prefix+types.PrefixValidator+types.PrefixPublic)

			val := types.ValAddress(pub.Address())
			s.Require().True(strings.HasPrefix(
				val.String(),
				prefix+types.PrefixValidator+types.PrefixAddress))

			bech32ValPub := legacybech32.MustMarshalPubKey(legacybech32.ValPK, pub) //nolint:staticcheck // SA1019: legacybech32 is deprecated: use the bech32 package instead.
			s.Require().True(strings.HasPrefix(
				bech32ValPub,
				prefix+types.PrefixValidator+types.PrefixPublic))

			config.SetBech32PrefixForConsensusNode(
				prefix+types.PrefixConsensus+types.PrefixAddress,
				prefix+types.PrefixConsensus+types.PrefixPublic)

			cons := types.ConsAddress(pub.Address())
			s.Require().True(strings.HasPrefix(
				cons.String(),
				prefix+types.PrefixConsensus+types.PrefixAddress))

			bech32ConsPub := legacybech32.MustMarshalPubKey(legacybech32.ConsPK, pub) //nolint:staticcheck // SA1019: legacybech32 is deprecated: use the bech32 package instead.
			s.Require().True(strings.HasPrefix(
				bech32ConsPub,
				prefix+types.PrefixConsensus+types.PrefixPublic))
		}
	}
}
*/

func (s *addressTestSuite) TestAddressInterface() {
	pubBz := make([]byte, ed25519.PubKeySize)
	pub := &ed25519.PubKey{Key: pubBz}
	rand.Read(pub.Key)

	addrs := []types.Address{
		types.ConsAddress(pub.Address()),
		types.ValAddress(pub.Address()),
		types.AccAddress(pub.Address()),
	}

	for _, addr := range addrs {
		switch addr := addr.(type) {
		case types.AccAddress:
			_, err := types.AccAddressFromHex(addr.String())
			s.Require().Nil(err)
		case types.ValAddress:
			_, err := types.ValAddressFromHex(addr.String())
			s.Require().Nil(err)
		case types.ConsAddress:
			_, err := types.ConsAddressFromHex(addr.String())
			s.Require().Nil(err)
		default:
			s.T().Fail()
		}
	}
}

func (s *addressTestSuite) TestVerifyAddressFormat() {
	addr1, err := address.NewHexCodec().StringToBytes("0x000000000000000000000000000000000000dead")
	require.NoError(s.T(), err)

	err = types.VerifyAddressFormat(addr1)
	s.Require().Nil(err)

	err = types.VerifyAddressFormat([]byte{})
	s.Require().NotNil(err)

	addr2, err := address.NewHexCodec().StringToBytes("zx000000000000000000000000000000000000dead")
	require.NotNil(s.T(), err)
	err = types.VerifyAddressFormat(addr2)
	s.Require().NotNil(err)
}

func (s *addressTestSuite) TestCustomAddressVerifier() {
	addr, err := address.NewHexCodec().StringToBytes("0x000000000000000000000000000000000000dead")
	require.NoError(s.T(), err)
	accBech := types.AccAddress(addr).String()
	valBech := types.ValAddress(addr).String()
	consBech := types.ConsAddress(addr).String()

	err = types.VerifyAddressFormat(addr)
	s.Require().Nil(err)
	_, err = types.AccAddressFromHex(accBech)
	s.Require().Nil(err)
	_, err = types.ValAddressFromHex(valBech)
	s.Require().Nil(err)
	_, err = types.ConsAddressFromHex(consBech)
	s.Require().Nil(err)

	/* HV2: not needed in Heimdall
	// Set a custom address verifier only accepts 20 byte addresses
	types.GetConfig().SetAddressVerifier(func(bz []byte) error {
		n := len(bz)
		if n == 20 {
			return nil
		}
		return fmt.Errorf("incorrect address length %d", n)
	})

	// Verifiy that the custom logic rejects this 10 byte address
	err = types.VerifyAddressFormat(addr)
	s.Require().NotNil(err)
	_, err = types.AccAddressFromHex(accBech)
	s.Require().NotNil(err)
	_, err = types.ValAddressFromHex(valBech)
	s.Require().NotNil(err)
	_, err = types.ConsAddressFromHex(consBech)
	s.Require().NotNil(err)

	// Reinitialize the global config to default address verifier (nil)
	types.GetConfig().SetAddressVerifier(nil)
	*/
}

func (s *addressTestSuite) TestBech32ifyAddressBytes() {
	s.T().Skip("skipping test for HV2 (bech32 not relevant)")
	addr10byte := []byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}
	addr20byte := []byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19}
	type args struct {
		prefix string
		bs     []byte
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{"empty address", args{"prefixa", []byte{}}, "", false},
		{"empty prefix", args{"", addr20byte}, "", true},
		{"10-byte address", args{"prefixa", addr10byte}, "prefixa1qqqsyqcyq5rqwzqf3953cc", false},
		{"10-byte address", args{"prefixb", addr10byte}, "prefixb1qqqsyqcyq5rqwzqf20xxpc", false},
		{"20-byte address", args{"prefixa", addr20byte}, "prefixa1qqqsyqcyq5rqwzqfpg9scrgwpugpzysn7hzdtn", false},
		{"20-byte address", args{"prefixb", addr20byte}, "prefixb1qqqsyqcyq5rqwzqfpg9scrgwpugpzysnrujsuw", false},
	}
	for _, tt := range tests {
		tt := tt
		s.T().Run(tt.name, func(t *testing.T) {
			got, err := types.HexifyAddressBytes(tt.args.bs)
			if (err != nil) != tt.wantErr {
				t.Errorf("Bech32ifyBytes() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			require.Equal(t, tt.want, got)
		})
	}
}

func (s *addressTestSuite) TestMustHexifyAddressBytes() {
	ac := address.NewHexCodec()
	addrBytes, err := ac.StringToBytes("0x000000000000000000000000000000000000dead")
	require.NoError(s.T(), err)
	type args struct {
		prefix string
		bs     []byte
	}
	tests := []struct {
		name      string
		args      args
		want      string
		wantPanic bool
	}{
		{"empty address", args{"prefixa", []byte{}}, "", false},
		{"empty prefix", args{"", addrBytes}, "0x000000000000000000000000000000000000dead", false},
		{"10-byte address", args{"prefixa", addrBytes}, "0x000000000000000000000000000000000000dead", false},
		{"10-byte address", args{"prefixb", addrBytes}, "0x000000000000000000000000000000000000dead", false},
		{"20-byte address", args{"prefixa", addrBytes}, "0x000000000000000000000000000000000000dead", false},
		{"20-byte address", args{"prefixb", addrBytes}, "0x000000000000000000000000000000000000dead", false},
	}
	for _, tt := range tests {
		tt := tt
		s.T().Run(tt.name, func(t *testing.T) {
			if tt.wantPanic {
				require.Panics(t, func() { types.MustHexifyAddressBytes(tt.args.bs) })
				return
			}
			require.Equal(t, tt.want, types.MustHexifyAddressBytes(tt.args.bs))
		})
	}
}

func (s *addressTestSuite) TestAddressTypesEquals() {
	addr1 := secp256k1.GenPrivKey().PubKey().Address()
	accAddr1 := types.AccAddress(addr1)
	consAddr1 := types.ConsAddress(addr1)
	valAddr1 := types.ValAddress(addr1)

	addr2 := secp256k1.GenPrivKey().PubKey().Address()
	accAddr2 := types.AccAddress(addr2)
	consAddr2 := types.ConsAddress(addr2)
	valAddr2 := types.ValAddress(addr2)

	// equality
	s.Require().True(accAddr1.Equals(accAddr1))   //nolint:gocritic // checking if these are the same
	s.Require().True(consAddr1.Equals(consAddr1)) //nolint:gocritic // checking if these are the same
	s.Require().True(valAddr1.Equals(valAddr1))   //nolint:gocritic // checking if these are the same

	// emptiness
	s.Require().True(types.AccAddress{}.Equals(types.AccAddress{})) //nolint:gocritic // checking if these are the same
	s.Require().True(types.AccAddress{}.Equals(types.AccAddress(nil)))
	s.Require().True(types.AccAddress(nil).Equals(types.AccAddress{}))
	s.Require().True(types.AccAddress(nil).Equals(types.AccAddress(nil))) //nolint:gocritic // checking if these are the same

	s.Require().True(types.ConsAddress{}.Equals(types.ConsAddress{})) //nolint:gocritic // checking if these are the same
	s.Require().True(types.ConsAddress{}.Equals(types.ConsAddress(nil)))
	s.Require().True(types.ConsAddress(nil).Equals(types.ConsAddress{}))
	s.Require().True(types.ConsAddress(nil).Equals(types.ConsAddress(nil))) //nolint:gocritic // checking if these are the same

	s.Require().True(types.ValAddress{}.Equals(types.ValAddress{})) //nolint:gocritic // checking if these are the same
	s.Require().True(types.ValAddress{}.Equals(types.ValAddress(nil)))
	s.Require().True(types.ValAddress(nil).Equals(types.ValAddress{}))
	s.Require().True(types.ValAddress(nil).Equals(types.ValAddress(nil))) //nolint:gocritic // checking if these are the same

	s.Require().False(accAddr1.Equals(accAddr2))
	s.Require().Equal(accAddr1.Equals(accAddr2), accAddr2.Equals(accAddr1))
	s.Require().False(consAddr1.Equals(consAddr2))
	s.Require().Equal(consAddr1.Equals(consAddr2), consAddr2.Equals(consAddr1))
	s.Require().False(valAddr1.Equals(valAddr2))
	s.Require().Equal(valAddr1.Equals(valAddr2), valAddr2.Equals(valAddr1))
}

func (s *addressTestSuite) TestNilAddressTypesEmpty() {
	s.Require().True(types.AccAddress(nil).Empty())
	s.Require().True(types.ConsAddress(nil).Empty())
	s.Require().True(types.ValAddress(nil).Empty())
}

func (s *addressTestSuite) TestGetConsAddress() {
	pk := secp256k1.GenPrivKey().PubKey()
	s.Require().NotEqual(types.GetConsAddress(pk), pk.Address())
	s.Require().True(bytes.Equal(types.GetConsAddress(pk).Bytes(), pk.Address().Bytes()))
	s.Require().Panics(func() { types.GetConsAddress(cryptotypes.PubKey(nil)) })
}

func (s *addressTestSuite) TestGetFromHex() {
	_, err := types.GetFromHex("")
	s.Require().Error(err)
	s.Require().Equal("empty address string is not allowed", err.Error())
	_, err = types.GetFromHex("0x000000000000000000000000000000000000dead")
	s.Require().NoError(err)
}
