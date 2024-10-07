package address

import (
	"errors"
	"strings"

	"github.com/ethereum/go-ethereum/common"

	"cosmossdk.io/core/address"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type HexCodec struct {
}

var _ address.Codec = &HexCodec{}

func NewHexCodec() address.Codec {
	return HexCodec{}
}

// StringToBytes encodes text to bytes
func (bc HexCodec) StringToBytes(hexAddr string) ([]byte, error) {
	if len(strings.TrimSpace(hexAddr)) == 0 {
		return []byte{}, errors.New("empty address string is not allowed")
	}

	hexAddr = strings.ToLower(hexAddr)

	if !has0xPrefix(hexAddr) {
		hexAddr = "0x" + hexAddr
	}

	bz := common.FromHex(hexAddr)

	if err := sdk.VerifyAddressFormat(bz); err != nil {
		return nil, err
	}

	return bz, nil
}

// BytesToString decodes bytes to text
func (bc HexCodec) BytesToString(bz []byte) (string, error) {
	if len(bz) == 0 || bz == nil {
		return "", nil
	}

	if err := sdk.VerifyAddressFormat(bz); err != nil {
		return "", err
	}

	hexAddr := common.Bytes2Hex(bz)

	hexAddr = strings.ToLower(hexAddr)

	if has0xPrefix(hexAddr) {
		return hexAddr, nil
	} else {
		return "0x" + hexAddr, nil
	}

}

// has0xPrefix validates str begins with '0x' or '0X'.
func has0xPrefix(str string) bool {
	return len(str) >= 2 && str[0] == '0' && (str[1] == 'x' || str[1] == 'X')
}
