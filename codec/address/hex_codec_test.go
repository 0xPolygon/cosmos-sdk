package address

import (
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/require"
)

func TestHexCodec_StringToBytes_ValidHex(t *testing.T) {
	codec := NewHexCodec()
	text := "0xa316fa9fa91700d7084d377bfdc81eb9f232f5ff"
	expected := common.FromHex(text)

	result, err := codec.StringToBytes(text)
	require.NoError(t, err)
	require.Equal(t, expected, result)
}

func TestHexCodec_StringToBytes_EmptyString(t *testing.T) {
	codec := NewHexCodec()
	text := ""

	_, err := codec.StringToBytes(text)
	require.Error(t, err)
	require.Equal(t, "empty address string is not allowed", err.Error())
}

func TestHexCodec_StringToBytes_InvalidHex(t *testing.T) {
	codec := NewHexCodec()
	text := "invalid_hex"

	_, err := codec.StringToBytes(text)
	require.Error(t, err)
}

func TestHexCodec_BytesToString_ValidBytes(t *testing.T) {
	codec := NewHexCodec()
	bz := common.FromHex("0xa316fa9fa91700d7084d377bfdc81eb9f232f5ff")
	expected := "0x" + common.Bytes2Hex(bz)

	result, err := codec.BytesToString(bz)
	require.NoError(t, err)
	require.Equal(t, expected, result)
}

func TestHexCodec_BytesToString_EmptyBytes(t *testing.T) {
	codec := NewHexCodec()
	var bz []byte

	result, err := codec.BytesToString(bz)
	require.NoError(t, err)
	require.Equal(t, "", result)
}

func TestHexCodec_BytesToString_InvalidBytes(t *testing.T) {
	codec := NewHexCodec()
	bz := []byte{0x01, 0x02, 0x03}

	_, err := codec.BytesToString(bz)
	require.Error(t, err)
}

func TestHexCodec_StringToBytes_BytesToString_Symmetry(t *testing.T) {
	codec := NewHexCodec()
	originalText := "0xa316fa9fa91700d7084d377bfdc81eb9f232f5ff"

	bytes, err := codec.StringToBytes(originalText)
	require.NoError(t, err)

	resultText, err := codec.BytesToString(bytes)
	require.NoError(t, err)
	require.Equal(t, originalText, resultText)
}
