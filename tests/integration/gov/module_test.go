package gov_test

import (
	"testing"

	"gotest.tools/v3/assert"

	hApp "github.com/0xPolygon/heimdall-v2/app"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	_ "github.com/cosmos/cosmos-sdk/x/distribution"
	"github.com/cosmos/cosmos-sdk/x/gov/types"
	_ "github.com/cosmos/cosmos-sdk/x/mint"
)

func TestItCreatesModuleAccountOnInitBlock(t *testing.T) {
	setUpAppResult := hApp.SetupApp(t, 1)
	app := setUpAppResult.App
	hApp.RequestFinalizeBlock(t, app, app.LastBlockHeight()+1)
	ctx := app.BaseApp.NewContext(false)
	acc := app.AccountKeeper.GetAccount(ctx, authtypes.NewModuleAddress(types.ModuleName))
	assert.Assert(t, acc != nil)
}
