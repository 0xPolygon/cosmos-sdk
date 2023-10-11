package types

import (
	"cosmossdk.io/collections"
	"github.com/cosmos/cosmos-sdk/types"
)

const (
	// module name
	ModuleName = "auth"

	// StoreKey is string representation of the store key for auth
	StoreKey = "acc"

	// FeeCollectorName the root string for the fee collector account address
	FeeCollectorName = "fee_collector"

	// TODO CHECK HEIMDALL-V2 check usage of FeeToken in heimdall and implement eventual changes
	// FeeToken fee token name
	FeeToken = "matic"
)

var (
	// ParamsKey is the prefix for params key
	ParamsKey = collections.NewPrefix(0)

	// AddressStoreKeyPrefix prefix for account-by-address store
	AddressStoreKeyPrefix = collections.NewPrefix(1)

	// GlobalAccountNumberKey identifies the prefix where the monotonically increasing
	// account number is stored.
	GlobalAccountNumberKey = collections.NewPrefix(2)

	// AccountNumberStoreKeyPrefix prefix for account-by-id store
	AccountNumberStoreKeyPrefix = collections.NewPrefix("accountNumber")

	// TODO CHECK HEIMDALL-V2 changed byte to collections. Is it ok?
	// ProposerKeyPrefix prefix for proposer
	// ProposerKeyPrefix = []byte("proposer")
	ProposerKeyPrefix = collections.NewPrefix("proposer")
)

// TODO CHECK HEIMDALL-V2 check those 2 functions and import HeimdallAddress
// TODO CHECK HEIMDALL-V2 AddressStoreKey is moved (and edited) to x/auth/keeper/migrations.go

// AddressStoreKey turn an address to key used to get it from the account store
func AddressStoreKey(addr types.HeimdallAddress) []byte {
	return append(AddressStoreKeyPrefix, addr.Bytes()...)
}

// ProposerKey returns proposer key
func ProposerKey() []byte {
	return ProposerKeyPrefix
}
