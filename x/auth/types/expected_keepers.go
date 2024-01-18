package types

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// TODO HV2 check this file (it was deleted in heimdall). Is this needed? Does it clash with heimdall's gov/expected_keepers.go SupplyKeeper? In case, adapt it

// BankKeeper defines the contract needed for supply related APIs (noalias)
type BankKeeper interface {
	IsSendEnabledCoins(ctx context.Context, coins ...sdk.Coin) error
	SendCoins(ctx context.Context, from, to sdk.AccAddress, amt sdk.Coins) error
	SendCoinsFromAccountToModule(ctx context.Context, senderAddr sdk.AccAddress, recipientModule string, amt sdk.Coins) error
}
