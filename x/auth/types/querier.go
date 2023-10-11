package types

import "github.com/cosmos/cosmos-sdk/types"

// query endpoints supported by the auth Querier
const (
	QueryAccount = "account"
	QueryParams  = "params"
)

// TODO CHECK HEIMDALL-V2 these two methods have been removed > replace implementation when called
// QueryAccountParams defines the params for querying accounts.
type QueryAccountParams struct {
	Address types.HeimdallAddress
}

// NewQueryAccountParams creates a new instance of QueryAccountParams.
func NewQueryAccountParams(addr types.HeimdallAddress) QueryAccountParams {
	return QueryAccountParams{Address: addr}
}
