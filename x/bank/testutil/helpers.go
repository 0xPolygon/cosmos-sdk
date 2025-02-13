package testutil

import (
	context "context"

	topupTypes "github.com/0xPolygon/heimdall-v2/x/topup/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
)

// FundAccount is a utility function that funds an account by minting and
// sending the coins to the address. This should be used for testing purposes
// only!
//
// TODO: Instead of using the mint module account, which has the
// permission of minting, create a "faucet" account. (@fdymylja)
func FundAccount(ctx context.Context, bankKeeper bankkeeper.Keeper, addr sdk.AccAddress, amounts sdk.Coins) error {
	if err := bankKeeper.MintCoins(ctx, topupTypes.ModuleName, amounts); err != nil {
		return err
	}

	return bankKeeper.SendCoinsFromModuleToAccount(ctx, topupTypes.ModuleName, addr, amounts)
}

// FundModuleAccount is a utility function that funds a module account by
// minting and sending the coins to the address. This should be used for testing
// purposes only!
//
// TODO: Instead of using the mint module account, which has the
// permission of minting, create a "faucet" account. (@fdymylja)
func FundModuleAccount(ctx context.Context, bankKeeper bankkeeper.Keeper, recipientMod string, amounts sdk.Coins) error {
	if err := bankKeeper.MintCoins(ctx, topupTypes.ModuleName, amounts); err != nil {
		return err
	}

	return bankKeeper.SendCoinsFromModuleToModule(ctx, topupTypes.ModuleName, recipientMod, amounts)
}
