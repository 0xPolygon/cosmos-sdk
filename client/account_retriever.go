package client

import (
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/cosmos/cosmos-sdk/types"
)

// Account defines a read-only version of the auth module's AccountI.
type Account interface {
	GetAddress() types.AccAddress
	GetPubKey() cryptotypes.PubKey // can return nil.
	GetAccountNumber() uint64
	GetSequence() uint64
}

// AccountRetriever defines the interfaces required by transactions to
// ensure an account exists and to be able to query for account fields necessary
// for signing.
type AccountRetriever interface {
	GetAccount(clientCtx Context, addr types.AccAddress) (Account, error)
	GetAccountWithHeight(clientCtx Context, addr types.AccAddress) (Account, int64, error)
	EnsureExists(clientCtx Context, addr types.AccAddress) error
	GetAccountNumberSequence(clientCtx Context, addr types.AccAddress) (accNum, accSeq uint64, err error)
}

var _ AccountRetriever = (*MockAccountRetriever)(nil)

// MockAccountRetriever defines a no-op basic AccountRetriever that can be used
// in mocked contexts. Tests or context that need more sophisticated testing
// state should implement their own mock AccountRetriever.
type MockAccountRetriever struct {
	ReturnAccNum, ReturnAccSeq uint64
}

func (mar MockAccountRetriever) GetAccount(_ Context, _ types.AccAddress) (Account, error) {
	return nil, nil
}

func (mar MockAccountRetriever) GetAccountWithHeight(_ Context, _ types.AccAddress) (Account, int64, error) {
	return nil, 0, nil
}

func (mar MockAccountRetriever) EnsureExists(_ Context, _ types.AccAddress) error {
	return nil
}

func (mar MockAccountRetriever) GetAccountNumberSequence(_ Context, _ types.AccAddress) (uint64, uint64, error) {
	return mar.ReturnAccNum, mar.ReturnAccSeq, nil
}
