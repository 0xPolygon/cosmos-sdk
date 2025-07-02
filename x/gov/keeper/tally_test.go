package keeper_test

import (
	"testing"

	staketypes "github.com/0xPolygon/heimdall-v2/x/stake/types"
	"github.com/cosmos/cosmos-sdk/types"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
)

func TestTally(t *testing.T) {
	govKeeper, _, _, sk, _, _, ctx := setupGovKeeper(t)
	sk.EXPECT().AddValidator(gomock.Any(), gomock.Any()).AnyTimes()
	sk.EXPECT().IterateCurrentValidatorsAndApplyFn(gomock.Any(), gomock.Any()).Return(nil)

	// Create a minimal proposal
	tp := TestProposal
	accAddr, err := types.AccAddressFromHex("0xb316fa9fa91700d7084d377bfdc81eb9f232f5ff")
	proposal, err := govKeeper.SubmitProposal(ctx, tp, "", "title", "description", accAddr, false)
	mockValidators := []*staketypes.Validator{
		{
			EndEpoch:         0,
			ValId:            1,
			StartEpoch:       0,
			Nonce:            0,
			VotingPower:      1000,
			PubKey:           []byte("Hello"),
			Signer:           accAddr.String(),
			LastUpdated:      "",
			Jailed:           false,
			ProposerPriority: 0,
		},
	}

	err = sk.AddValidator(ctx, *mockValidators[0])
	require.NoError(t, err)

	// Call tally function
	passes, burnDeposits, tallyResults, err := govKeeper.Tally(ctx, proposal)

	// Assertions
	require.NoError(t, err)
	require.NotNil(t, tallyResults)

	// Print results for debugging
	t.Logf("Passes: %v, BurnDeposits: %v, Tally Results: %+v", passes, burnDeposits, tallyResults)
}
