package auth_test

import (
	"testing"

	hApp "github.com/0xPolygon/heimdall-v2/app"
	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/x/auth/types"
)

func TestItCreatesModuleAccountOnInitBlock(t *testing.T) {
	setUpAppResult := hApp.SetupApp(t, 1)
	app := setUpAppResult.App
	hApp.RequestFinalizeBlock(t, app, app.LastBlockHeight()+1)
	ctx := app.BaseApp.NewContext(false)
	acc := app.AccountKeeper.GetAccount(ctx, types.NewModuleAddress(types.FeeCollectorName))
	require.NotNil(t, acc)
}
