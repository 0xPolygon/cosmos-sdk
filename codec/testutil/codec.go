package testutil

import (
	"github.com/cosmos/gogoproto/proto"

	"cosmossdk.io/x/tx/signing"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/address"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
)

// CodecOptions are options for creating a test codec.
type CodecOptions struct {
	AccAddressPrefix string
	ValAddressPrefix string
}

// NewInterfaceRegistry returns a new InterfaceRegistry with the given options.
func (o CodecOptions) NewInterfaceRegistry() codectypes.InterfaceRegistry {
	accAddressPrefix := o.AccAddressPrefix
	valAddressPrefix := o.ValAddressPrefix

	ir, err := codectypes.NewInterfaceRegistryWithOptions(codectypes.InterfaceRegistryOptions{
		ProtoFiles: proto.HybridResolver,
		SigningOptions: signing.Options{
			AddressCodec:          address.NewHexCodec(accAddressPrefix),
			ValidatorAddressCodec: address.NewHexCodec(valAddressPrefix),
		},
	})
	if err != nil {
		panic(err)
	}
	return ir
}

// NewCodec returns a new codec with the given options.
func (o CodecOptions) NewCodec() *codec.ProtoCodec {
	return codec.NewProtoCodec(o.NewInterfaceRegistry())
}
