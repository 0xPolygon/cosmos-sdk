package types

import (
	"testing"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/address"
	"github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
)

func TestMsgUnjailGetSignBytes(t *testing.T) {
	ac := address.NewHexCodec()
	addr, err := ac.StringToBytes("0x000000000000000000000000000000000000dead")
	require.NoError(t, err)
	msg := NewMsgUnjail(sdk.ValAddress(addr).String())
	pc := codec.NewProtoCodec(types.NewInterfaceRegistry())
	bytes, err := pc.MarshalAminoJSON(msg)
	require.NoError(t, err)
	require.Equal(
		t,
		`{"type":"cosmos-sdk/MsgUnjail","value":{"address":"0x000000000000000000000000000000000000dead"}}`,
		string(bytes),
	)
}
