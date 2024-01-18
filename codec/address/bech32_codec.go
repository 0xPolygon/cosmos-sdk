package address

import (
	"errors"
	"strings"

	"github.com/ethereum/go-ethereum/common"

	"cosmossdk.io/core/address"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type HexCodec struct {
	HexPrefix string
}

var _ address.Codec = &HexCodec{}

func NewHexCodec(prefix string) address.Codec {
	return HexCodec{prefix}
}

// StringToBytes encodes text to bytes
func (bc HexCodec) StringToBytes(text string) ([]byte, error) {
	if len(strings.TrimSpace(text)) == 0 {
		return []byte{}, errors.New("empty address string is not allowed")
	}

	bz := common.FromHex(text)

	if err := sdk.VerifyAddressFormat(bz); err != nil {
		return nil, err
	}

	return bz, nil
}

// BytesToString decodes bytes to text
func (bc HexCodec) BytesToString(bz []byte) (string, error) {
	if len(bz) == 0 {
		return "", nil
	}

	if err := sdk.VerifyAddressFormat(bz); err != nil {
		return "", err
	}

	text := common.Bytes2Hex(bz)

	return text, nil
}
