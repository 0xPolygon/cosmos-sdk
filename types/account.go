package types

import (
	"github.com/cosmos/gogoproto/proto"

	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
)

// AccountI is an interface used to store coins at a given address within state.
// It presumes a notion of sequence numbers for replay protection,
// a notion of account numbers for replay protection for previously pruned accounts,
// and a pubkey for authentication purposes.
//
// Many complex conditions can be used in the concrete struct which implements AccountI.

// TODO HV2 converted AccAddress into types.HeimdallAddress (see heimdall's auth/exported/exported.go)
//
//	Is this enough? Check al other interfaces implementing AccountI
//	Or do we want to separate AccountI from a different interface which is heimdall specific? Like AccountHeimdallI (less impact on upstream)
type AccountI interface {
	proto.Message

	GetAddress() HeimdallAddress
	SetAddress(HeimdallAddress) error // errors if already set.

	GetPubKey() cryptotypes.PubKey // can return nil.
	SetPubKey(cryptotypes.PubKey) error

	GetAccountNumber() uint64
	SetAccountNumber(uint64) error

	GetSequence() uint64
	SetSequence(uint64) error

	// Ensure that account implements stringer
	String() string
}

// ModuleAccountI defines an account interface for modules that hold tokens in
// an escrow.
type ModuleAccountI interface {
	AccountI

	GetName() string
	GetPermissions() []string
	HasPermission(string) bool
}
