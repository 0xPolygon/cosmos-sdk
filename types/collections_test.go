package types

import (
	"testing"
	"time"

	"cosmossdk.io/collections/colltest"
	"github.com/cosmos/cosmos-sdk/codec/address"
	"github.com/stretchr/testify/require"
)

func TestCollectionsCorrectness(t *testing.T) {
	ac := address.HexCodec{}
	addr, err := ac.StringToBytes("0x000000000000000000000000000000000000dead")
	require.NoError(t, err)

	t.Run("AccAddress", func(t *testing.T) {
		colltest.TestKeyCodec(t, AccAddressKey, addr)
	})

	t.Run("ValAddress", func(t *testing.T) {
		colltest.TestKeyCodec(t, ValAddressKey, addr)
	})

	t.Run("ConsAddress", func(t *testing.T) {
		colltest.TestKeyCodec(t, ConsAddressKey, addr)
	})

	t.Run("AddressIndexingKey", func(t *testing.T) {
		colltest.TestKeyCodec(t, LengthPrefixedAddressKey(AccAddressKey), addr)
	})

	t.Run("Time", func(t *testing.T) {
		colltest.TestKeyCodec(t, TimeKey, time.Time{})
	})
}
