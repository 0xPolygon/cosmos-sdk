package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"
)

// Querier path constants
const (
	QueryBalance     = "balance"
	QueryAllBalances = "all_balances"
	QueryTotalSupply = "total_supply"
	QuerySupplyOf    = "supply_of"
)

// NewQueryBalanceRequest creates a new instance of QueryBalanceRequest.
func NewQueryBalanceRequest(addr sdk.AccAddress, denom string) *QueryBalanceRequest {
	return &QueryBalanceRequest{Address: addr.String(), Denom: denom}
}

// NewQueryAllBalancesRequest creates a new instance of QueryAllBalancesRequest.
func NewQueryAllBalancesRequest(addr sdk.AccAddress, req *query.PageRequest, resolveDenom bool) *QueryAllBalancesRequest {
	return &QueryAllBalancesRequest{Address: addr.String(), Pagination: req, ResolveDenom: resolveDenom}
}

// NewQuerySpendableBalancesRequest creates a new instance of a
// QuerySpendableBalancesRequest.
func NewQuerySpendableBalancesRequest(addr sdk.AccAddress, req *query.PageRequest) *QuerySpendableBalancesRequest {
	return &QuerySpendableBalancesRequest{Address: addr.String(), Pagination: req}
}

// NewQuerySpendableBalanceByDenomRequest creates a new instance of a
// QuerySpendableBalanceByDenomRequest.
func NewQuerySpendableBalanceByDenomRequest(addr sdk.AccAddress, denom string) *QuerySpendableBalanceByDenomRequest {
	return &QuerySpendableBalanceByDenomRequest{Address: addr.String(), Denom: denom}
}

// QueryTotalSupplyParams defines the params for the following queries:
// - 'custom/bank/totalSupply'
type QueryTotalSupplyParams struct {
	Page, Limit int
}

// NewQueryTotalSupplyParams creates a new instance to query the total supply
func NewQueryTotalSupplyParams(page, limit int) QueryTotalSupplyParams {
	return QueryTotalSupplyParams{page, limit}
}

// QuerySupplyOfParams defines the params for the following queries:
// - 'custom/bank/totalSupplyOf'
type QuerySupplyOfParams struct {
	Denom string
}

// NewQuerySupplyOfParams creates a new instance to query the total supply
// of a given denomination
func NewQuerySupplyOfParams(denom string) QuerySupplyOfParams {
	return QuerySupplyOfParams{denom}
}

// TODO HV2: from bank/querier.go in heimdall repo; might not be needed, double check
// NewQuerier returns a new sdk.Keeper instance.
// func NewQuerier(k Keeper) sdk.Querier {
// 	return func(ctx sdk.Context, path []string, req abci.RequestQuery) ([]byte, sdk.Error) {
// 		switch path[0] {
// 		case types.QueryBalance:
// 			return queryBalance(ctx, req, k)

// 		default:
// 			return nil, sdk.ErrUnknownRequest("unknown bank query endpoint")
// 		}
// 	}
// }

// TODO HV2: from bank/querier.go in heimdall repo; might not be needed, double check
// queryBalance fetch an account's balance for the supplied height.
// Height and account address are passed as first and second path components respectively.
// func queryBalance(ctx sdk.Context, req abci.RequestQuery, k Keeper) ([]byte, sdk.Error) {
// 	var params types.QueryBalanceParams

// 	if err := types.ModuleCdc.UnmarshalJSON(req.Data, &params); err != nil {
// 		return nil, sdk.ErrInternal(fmt.Sprintf("failed to parse params: %s", err))
// 	}

// 	bz, err := codec.MarshalJSONIndent(types.ModuleCdc, k.GetCoins(ctx, params.Address))
// 	if err != nil {
// 		return nil, sdk.ErrInternal(sdk.AppendMsgToErr("could not marshal result to JSON", err.Error()))
// 	}

// 	return bz, nil
// }
