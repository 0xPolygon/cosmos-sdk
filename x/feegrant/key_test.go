package feegrant_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"cosmossdk.io/x/feegrant"

	codecaddress "github.com/cosmos/cosmos-sdk/codec/address"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func TestMarshalAndUnmarshalFeegrantKey(t *testing.T) {
	addressCodec := codecaddress.NewHexCodec()
	grantee, err := addressCodec.StringToBytes("0x100000000000000000000000000000000000dead")
	require.NoError(t, err)
	granter, err := addressCodec.StringToBytes("0x200000000000000000000000000000000000dead")
	require.NoError(t, err)

	key := feegrant.FeeAllowanceKey(granter, grantee)
	require.Len(t, key, len(grantee)+len(granter)+3)
	require.Equal(t, feegrant.FeeAllowancePrefixByGrantee(grantee), key[:len(grantee)+2])

	g1, g2 := feegrant.ParseAddressesFromFeeAllowanceKey(key)
	require.Equal(t, granter, g1)
	require.Equal(t, grantee, g2)
}

func TestMarshalAndUnmarshalFeegrantKeyQueueKey(t *testing.T) {
	addressCodec := codecaddress.NewHexCodec()
	grantee, err := addressCodec.StringToBytes("0x100000000000000000000000000000000000dead")
	require.NoError(t, err)
	granter, err := addressCodec.StringToBytes("0x200000000000000000000000000000000000dead")
	require.NoError(t, err)

	exp := time.Now()
	expBytes := sdk.FormatTimeBytes(exp)

	key := feegrant.FeeAllowancePrefixQueue(&exp, feegrant.FeeAllowanceKey(granter, grantee)[1:])
	require.Len(t, key, len(grantee)+len(granter)+3+len(expBytes))

	granter1, grantee1 := feegrant.ParseAddressesFromFeeAllowanceQueueKey(key)
	require.Equal(t, granter, granter1)
	require.Equal(t, grantee, grantee1)
}
