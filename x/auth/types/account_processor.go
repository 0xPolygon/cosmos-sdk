package types

import (
	"github.com/maticnetwork/heimdall/auth/exported"
)

// TODO CHECK HEIMDALL-V2 this is imported from heimdall > merge/move (e.g. exported is now in types/account.go)

// AccountProcessor is an interface to process account as per module
type AccountProcessor func(*GenesisAccount, *BaseAccount) exported.Account
