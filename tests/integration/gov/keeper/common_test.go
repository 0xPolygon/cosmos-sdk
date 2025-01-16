package keeper_test

import (
	"testing"

	"gotest.tools/v3/assert"

	"cosmossdk.io/math"

	stakeTypes "github.com/0xPolygon/heimdall-v2/x/stake/types"
	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/gov/types"
	v1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	"github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
)

var TestProposal = getTestProposal()

func getTestProposal() []sdk.Msg {
	legacyProposalMsg, err := v1.NewLegacyContent(v1beta1.NewTextProposal("Title", "description"), authtypes.NewModuleAddress(types.ModuleName).String())
	if err != nil {
		panic(err)
	}
	testProposal := v1beta1.NewTextProposal("Proposal", "testing proposal")
	legacyProposalMsg2, err := v1.NewLegacyContent(testProposal, authtypes.NewModuleAddress(types.ModuleName).String())
	if err != nil {
		panic(err)
	}

	return []sdk.Msg{
		legacyProposalMsg,
		legacyProposalMsg2,
	}
}

func createValidators(t *testing.T, f *fixture, powers []int64) ([]sdk.AccAddress, []sdk.ValAddress) {
	amount := math.NewInt(30000000)
	addrs := simtestutil.AddTestAddrsIncremental(f.bankKeeper, f.ctx, 5, amount)
	valAddrs := simtestutil.ConvertAddrsToValAddrs(addrs)
	pks := simtestutil.CreateTestPubKeys(5)

	if len(pks) < len(valAddrs) {
		t.Fatal("PubKeys length is less than ValAddrs length")
	}

	for i := 0; i < len(valAddrs); i++ {
		val, err := stakeTypes.NewValidator(uint64(i+1), 0, 0, 0, sdk.TokensToConsensusPower(amount, sdk.DefaultPowerReduction), pks[0], valAddrs[0].String())
		assert.NilError(t, err)
		err = f.stakingKeeper.AddValidator(f.ctx, *val)
		assert.NilError(t, err)
	}

	f.stakingKeeper.EndBlocker(f.ctx)

	return addrs, valAddrs
}
