package keeper_test

import (
	"testing"

	hApp "github.com/0xPolygon/heimdall-v2/app"
	stakeKeeper "github.com/0xPolygon/heimdall-v2/x/stake/keeper"
	"github.com/cosmos/cosmos-sdk/baseapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	"github.com/cosmos/cosmos-sdk/x/gov/keeper"
	v1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	"github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
)

type fixture struct {
	ctx sdk.Context

	queryClient       v1.QueryClient
	legacyQueryClient v1beta1.QueryClient

	bankKeeper    bankkeeper.Keeper
	stakingKeeper *stakeKeeper.Keeper
	govKeeper     *keeper.Keeper
}

func initFixture(t *testing.T) *fixture {

	setUpAppResult := hApp.SetupApp(t, 1)
	app := setUpAppResult.App
	hApp.RequestFinalizeBlock(t, app, app.LastBlockHeight()+1)

	ctx := app.NewContext(false)

	queryHelper := baseapp.NewQueryServerTestHelper(ctx, app.InterfaceRegistry())
	v1.RegisterQueryServer(queryHelper, keeper.NewQueryServer(&app.GovKeeper))
	legacyQueryHelper := baseapp.NewQueryServerTestHelper(ctx, app.InterfaceRegistry())
	v1beta1.RegisterQueryServer(legacyQueryHelper, keeper.NewLegacyQueryServer(&app.GovKeeper))
	queryClient := v1.NewQueryClient(queryHelper)
	legacyQueryClient := v1beta1.NewQueryClient(legacyQueryHelper)

	return &fixture{
		ctx:               ctx,
		queryClient:       queryClient,
		legacyQueryClient: legacyQueryClient,
		bankKeeper:        app.BankKeeper,
		stakingKeeper:     &app.StakeKeeper,
		govKeeper:         &app.GovKeeper,
	}
}
