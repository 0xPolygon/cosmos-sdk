package keeper_test

import (
	"context"
	"testing"
	"time"

	"github.com/golang/mock/gomock"

	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/codec/address"
	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/gov"
	"github.com/cosmos/cosmos-sdk/x/gov/keeper"
	"github.com/cosmos/cosmos-sdk/x/gov/types"
	v1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"

	stakeTypes "github.com/0xPolygon/heimdall-v2/x/stake/types"
)

var _ types.GovHooks = &MockGovHooksReceiver{}

// GovHooks event hooks for governance proposal object (noalias)
type MockGovHooksReceiver struct {
	AfterProposalSubmissionValid        bool
	AfterProposalDepositValid           bool
	AfterProposalVoteValid              bool
	AfterProposalFailedMinDepositValid  bool
	AfterProposalVotingPeriodEndedValid bool
}

func (h *MockGovHooksReceiver) AfterProposalSubmission(ctx context.Context, proposalID uint64) error {
	h.AfterProposalSubmissionValid = true
	return nil
}

func (h *MockGovHooksReceiver) AfterProposalDeposit(ctx context.Context, proposalID uint64, depositorAddr sdk.AccAddress) error {
	h.AfterProposalDepositValid = true
	return nil
}

func (h *MockGovHooksReceiver) AfterProposalVote(ctx context.Context, proposalID uint64, voterAddr sdk.AccAddress) error {
	h.AfterProposalVoteValid = true
	return nil
}

func (h *MockGovHooksReceiver) AfterProposalFailedMinDeposit(ctx context.Context, proposalID uint64) error {
	h.AfterProposalFailedMinDepositValid = true
	return nil
}

func (h *MockGovHooksReceiver) AfterProposalVotingPeriodEnded(ctx context.Context, proposalID uint64) error {
	h.AfterProposalVotingPeriodEndedValid = true
	return nil
}

func TestHooks(t *testing.T) {
	minDeposit := v1.DefaultParams().MinDeposit
	govKeeper, authKeeper, bankKeeper, stakingKeeper, _, _, ctx := setupGovKeeper(t)
	addrs := simtestutil.AddTestAddrs(bankKeeper, ctx, 1, minDeposit[0].Amount)

	authKeeper.EXPECT().AddressCodec().Return(address.NewHexCodec()).AnyTimes()
	stakingKeeper.EXPECT().ValidatorAddressCodec().Return(address.NewHexCodec()).AnyTimes()

	mockValidators := []stakeTypes.Validator{
		{Signer: "0xb316fa9fa91700d7084d377bfdc81eb9f232f5ff"},
		{Signer: "0xb316fa9fa91700d7084d377bfdc81eb9f232f5fd"},
	}

	stakingKeeper.EXPECT().IterateCurrentValidatorsAndApplyFn(gomock.Any(), gomock.Any()).DoAndReturn(
		func(ctx context.Context, fn func(stakeTypes.Validator) bool) error {
			for _, validator := range mockValidators {
				if stop := fn(validator); stop {
					break
				}
			}
			return nil
		},
	).AnyTimes()

	govHooksReceiver := MockGovHooksReceiver{}

	keeper.UnsafeSetHooks(
		govKeeper, types.NewMultiGovHooks(&govHooksReceiver),
	)

	require.False(t, govHooksReceiver.AfterProposalSubmissionValid)
	require.False(t, govHooksReceiver.AfterProposalDepositValid)
	require.False(t, govHooksReceiver.AfterProposalVoteValid)
	require.False(t, govHooksReceiver.AfterProposalFailedMinDepositValid)
	require.False(t, govHooksReceiver.AfterProposalVotingPeriodEndedValid)

	tp := TestProposal
	accAddr, err := sdk.AccAddressFromHex("0xb316fa9fa91700d7084d377bfdc81eb9f232f5ff")
	require.NoError(t, err)
	_, err = govKeeper.SubmitProposal(ctx, tp, "", "test", "summary", accAddr, false)
	require.NoError(t, err)
	require.True(t, govHooksReceiver.AfterProposalSubmissionValid)

	params, _ := govKeeper.Params.Get(ctx)
	newHeader := ctx.BlockHeader()
	newHeader.Time = ctx.BlockHeader().Time.Add(*params.MaxDepositPeriod).Add(time.Duration(1) * time.Second)
	ctx = ctx.WithBlockHeader(newHeader)
	gov.EndBlocker(ctx, govKeeper)

	require.True(t, govHooksReceiver.AfterProposalFailedMinDepositValid)

	accAddr, err = sdk.AccAddressFromHex("0xb316fa9fa91700d7084d377bfdc81eb9f232f5ff")
	require.NoError(t, err)
	p2, err := govKeeper.SubmitProposal(ctx, tp, "", "test", "summary", accAddr, false)
	require.NoError(t, err)

	activated, err := govKeeper.AddDeposit(ctx, p2.Id, addrs[0], minDeposit)
	require.True(t, activated)
	require.NoError(t, err)
	require.True(t, govHooksReceiver.AfterProposalDepositValid)

	err = govKeeper.AddVote(ctx, p2.Id, addrs[0], v1.NewNonSplitVoteOption(v1.OptionYes), "")
	require.NoError(t, err)
	require.True(t, govHooksReceiver.AfterProposalVoteValid)

	newHeader = ctx.BlockHeader()
	newHeader.Time = ctx.BlockHeader().Time.Add(*params.VotingPeriod).Add(time.Duration(1) * time.Second)
	ctx = ctx.WithBlockHeader(newHeader)
	gov.EndBlocker(ctx, govKeeper)
	require.True(t, govHooksReceiver.AfterProposalVotingPeriodEndedValid)
}
