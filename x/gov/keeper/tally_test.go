package keeper_test

import (
	// "context"
	"testing"

	"github.com/stretchr/testify/require"

	// "github.com/0xPolygon/heimdall-v2/helper"
	// chainmanagerKeeper "github.com/0xPolygon/heimdall-v2/x/chainmanager/keeper"
	stakeTypes "github.com/0xPolygon/heimdall-v2/x/stake/types"

	// "github.com/cosmos/cosmos-sdk/codec"
	// "github.com/cosmos/cosmos-sdk/codec/address"
	// "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	// bankKeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	// "github.com/cosmos/cosmos-sdk/x/gov/testutil"
)

// setupTestKeeper initializes the StakeKeeper and sets up CurrentValidators
// func setupTestKeeper(bankKeeper testutil.MockBankKeeper, ctx context.Context) (*stakeKeeper.Keeper, context.Context) {

// 	// Create a new StakeKeeper instance with required dependencies
// 	sk := stakeKeeper.NewKeeper(
// 		codec.NewProtoCodec(types.NewInterfaceRegistry()),
// 		nil, // Store key (mocked)
// 		bankKeeper,
// 		chainmanagerKeeper.Keeper{},
// 		address.HexCodec{},       // Address Codec
// 		&helper.ContractCaller{}, // Contract Caller (mocked)
// 	)

// 	// Initialize mock validators
// 	mockValidators := []*stakeType.Validator{
// 		{
// 			EndEpoch:         0,
// 			ValId:            1,
// 			StartEpoch:       0,
// 			Nonce:            0,
// 			VotingPower:      1000,
// 			PubKey:           []byte("Hello"),
// 			Signer:           "0xworld",
// 			LastUpdated:      "",
// 			Jailed:           false,
// 			ProposerPriority: 0,
// 		},
// 	}

// 	sk.AddValidator(ctx, mockValidators)

// 	return &sk, ctx
// }

func TestTally(t *testing.T) {
	govKeeper, _, _, sk, _, _, ctx := setupGovKeeper(t)

	// Create a minimal proposal
	tp := TestProposal
	accAddr, err := sdk.AccAddressFromHex("0xb316fa9fa91700d7084d377bfdc81eb9f232f5ff")
	proposal, err := govKeeper.SubmitProposal(ctx, tp, "", "title", "description", accAddr, false)
	mockValidators := []*stakeTypes.Validator{
		{
			EndEpoch:         0,
			ValId:            1,
			StartEpoch:       0,
			Nonce:            0,
			VotingPower:      1000,
			PubKey:           []byte("Hello"),
			Signer:           "0xworld",
			LastUpdated:      "",
			Jailed:           false,
			ProposerPriority: 0,
		},
	}

	sk.AddValidator(ctx, *mockValidators[0])

	// Call Tally function
	passes, burnDeposits, tallyResults, err := govKeeper.Tally(ctx, proposal)

	// Assertions
	require.NoError(t, err)
	require.NotNil(t, tallyResults)

	// Print results for debugging
	t.Logf("Passes: %v, BurnDeposits: %v, Tally Results: %+v", passes, burnDeposits, tallyResults)
}
