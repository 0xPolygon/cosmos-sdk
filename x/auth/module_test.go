package auth_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"cosmossdk.io/depinject"
	"cosmossdk.io/log"

	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
	"github.com/cosmos/cosmos-sdk/x/auth/keeper"
	"github.com/cosmos/cosmos-sdk/x/auth/testutil"
	"github.com/cosmos/cosmos-sdk/x/auth/types"
)

func TestItCreatesModuleAccountOnInitBlock(t *testing.T) {
	// TODO HV2: To fix this test, we need to implement https://polygon.atlassian.net/browse/POS-2540
	t.Skip("skipping test as it simApp staking module instead of heimdall-v2 custom stake module")

	var accountKeeper keeper.AccountKeeper
	app, err := simtestutil.SetupAtGenesis(
		depinject.Configs(
			testutil.AppConfig,
			depinject.Supply(log.NewNopLogger()),
		),
		&accountKeeper)
	require.NoError(t, err)

	ctx := app.BaseApp.NewContext(false)
	acc := accountKeeper.GetAccount(ctx, types.NewModuleAddress(types.FeeCollectorName))
	require.NotNil(t, acc)
}
