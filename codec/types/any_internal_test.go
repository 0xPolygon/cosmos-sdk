package types

import (
	"testing"

	"github.com/cosmos/gogoproto/proto"
	"github.com/stretchr/testify/require"
)

type Dog struct {
	Name string `protobuf:"bytes,1,opt,name=size,proto3" json:"size,omitempty"`
}

func (d Dog) Greet() string { return d.Name }

// We implement a minimal proto.Message interface
func (d *Dog) Reset()                  { d.Name = "" }
func (d *Dog) String() string          { return d.Name }
func (d *Dog) ProtoMessage()           {}
func (d *Dog) XXX_MessageName() string { return "tests/dog" } //nolint:revive // XXX_ prefix is required

type Animal interface {
	Greet() string
}

var (
	_ Animal        = (*Dog)(nil)
	_ proto.Message = (*Dog)(nil)
)

func TestAnyPackUnpack(t *testing.T) {
	registry := NewInterfaceRegistry()
	registry.RegisterInterface("Animal", (*Animal)(nil))
	registry.RegisterImplementations(
		(*Animal)(nil),
		&Dog{},
	)

	spot := &Dog{Name: "Spot"}
	var animal Animal

	// with cache
	a, err := NewAnyWithValue(spot)
	require.NoError(t, err)
	require.Equal(t, spot, a.GetCachedValue())
	err = registry.UnpackAny(a, &animal)
	require.NoError(t, err)
	require.Equal(t, spot, animal)

	// without cache
	a.cachedValue = nil
	err = registry.UnpackAny(a, &animal)
	require.NoError(t, err)
	require.Equal(t, spot, animal)
}

func TestString(t *testing.T) {
	req := require.New(t)
	spot := &Dog{Name: "Spot"}
	a, err := NewAnyWithValue(spot)
	req.NoError(err)

	req.Equal("&Any{TypeUrl:/tests/dog,Value:[10 4 83 112 111 116],XXX_unrecognized:[]}", a.String())
	req.Equal(`&Any{TypeUrl: "/tests/dog",
  Value: []byte{0xa, 0x4, 0x53, 0x70, 0x6f, 0x74}
}`, a.GoString())
}
