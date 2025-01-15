package gov_test

import (
	"testing"
	"time"

	"math/big"
	"math/rand"

	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/require"

	"cosmossdk.io/collections"
	"cosmossdk.io/math"

	sideTxs "github.com/0xPolygon/heimdall-v2/sidetxs"
	stakeKeeper "github.com/0xPolygon/heimdall-v2/x/stake/keeper"
	stakeTypes "github.com/0xPolygon/heimdall-v2/x/stake/types"
	topupTypes "github.com/0xPolygon/heimdall-v2/x/topup/types"
	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/simulation"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	bankKeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	"github.com/cosmos/cosmos-sdk/x/gov"
	"github.com/cosmos/cosmos-sdk/x/gov/keeper"
	"github.com/cosmos/cosmos-sdk/x/gov/types"
	v1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	"github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
)

func TestUnregisteredProposal_InactiveProposalFails(t *testing.T) {
	suite := createTestSuite(t)
	app := suite.App
	ctx := suite.App.BaseApp.NewContext(false)
	addrs := simtestutil.AddTestAddrs(app.BankKeeper, ctx, 10, valTokens)

	// manually set proposal in store
	startTime, endTime := time.Now().Add(-4*time.Hour), ctx.BlockHeader().Time
	proposal, err := v1.NewProposal([]sdk.Msg{
		&v1.Proposal{}, // invalid proposal message
	}, 1, startTime, startTime, "", "Unsupported proposal", "Unsupported proposal", addrs[0], false)
	require.NoError(t, err)

	err = app.GovKeeper.SetProposal(ctx, proposal)
	require.NoError(t, err)

	// manually set proposal in inactive proposal queue
	err = app.GovKeeper.InactiveProposalsQueue.Set(ctx, collections.Join(endTime, proposal.Id), proposal.Id)
	require.NoError(t, err)

	checkInactiveProposalsQueue(t, ctx, &app.GovKeeper)

	err = gov.EndBlocker(ctx, &app.GovKeeper)
	require.NoError(t, err)

	_, err = app.GovKeeper.Proposals.Get(ctx, proposal.Id)
	require.Error(t, err, collections.ErrNotFound)
}

func TestUnregisteredProposal_ActiveProposalFails(t *testing.T) {
	suite := createTestSuite(t)
	app := suite.App
	ctx := suite.App.BaseApp.NewContext(false)
	addrs := simtestutil.AddTestAddrs(app.BankKeeper, ctx, 10, valTokens)

	// manually set proposal in store
	startTime, endTime := time.Now().Add(-4*time.Hour), ctx.BlockHeader().Time
	proposal, err := v1.NewProposal([]sdk.Msg{
		&v1.Proposal{}, // invalid proposal message
	}, 1, startTime, startTime, "", "Unsupported proposal", "Unsupported proposal", addrs[0], false)
	require.NoError(t, err)
	proposal.Status = v1.StatusVotingPeriod
	proposal.VotingEndTime = &endTime

	err = app.GovKeeper.SetProposal(ctx, proposal)
	require.NoError(t, err)

	// manually set proposal in active proposal queue
	err = app.GovKeeper.ActiveProposalsQueue.Set(ctx, collections.Join(endTime, proposal.Id), proposal.Id)
	require.NoError(t, err)

	checkActiveProposalsQueue(t, ctx, &app.GovKeeper)

	err = gov.EndBlocker(ctx, &app.GovKeeper)
	require.NoError(t, err)

	p, err := app.GovKeeper.Proposals.Get(ctx, proposal.Id)
	require.NoError(t, err)
	require.Equal(t, v1.StatusFailed, p.Status)
}

func TestTickExpiredDepositPeriod(t *testing.T) {
	suite := createTestSuite(t)
	app := suite.App
	ctx := app.BaseApp.NewContext(false)
	addrs := simtestutil.AddTestAddrs(app.BankKeeper, ctx, 10, valTokens)

	govMsgSvr := keeper.NewMsgServerImpl(&app.GovKeeper)

	checkInactiveProposalsQueue(t, ctx, &app.GovKeeper)

	newProposalMsg, err := v1.NewMsgSubmitProposal(
		[]sdk.Msg{mkTestLegacyContent(t)},
		sdk.Coins{sdk.NewInt64Coin(sdk.DefaultBondDenom, 100000000000000000)},
		addrs[0].String(),
		"",
		"Proposal",
		"description of proposal",
		false,
	)
	require.NoError(t, err)

	res, err := govMsgSvr.SubmitProposal(ctx, newProposalMsg)
	require.NoError(t, err)
	require.NotNil(t, res)

	checkInactiveProposalsQueue(t, ctx, &app.GovKeeper)

	newHeader := ctx.BlockHeader()
	newHeader.Time = ctx.BlockHeader().Time.Add(time.Duration(1) * time.Second)
	ctx = ctx.WithBlockHeader(newHeader)

	checkInactiveProposalsQueue(t, ctx, &app.GovKeeper)

	params, _ := app.GovKeeper.Params.Get(ctx)
	newHeader = ctx.BlockHeader()
	newHeader.Time = ctx.BlockHeader().Time.Add(*params.MaxDepositPeriod)
	ctx = ctx.WithBlockHeader(newHeader)

	checkInactiveProposalsQueue(t, ctx, &app.GovKeeper)

	err = gov.EndBlocker(ctx, &app.GovKeeper)
	require.NoError(t, err)

	checkInactiveProposalsQueue(t, ctx, &app.GovKeeper)
}

func TestTickMultipleExpiredDepositPeriod(t *testing.T) {
	suite := createTestSuite(t)
	app := suite.App
	ctx := app.BaseApp.NewContext(false)
	addrs := simtestutil.AddTestAddrs(app.BankKeeper, ctx, 10, valTokens)

	govMsgSvr := keeper.NewMsgServerImpl(&app.GovKeeper)

	checkInactiveProposalsQueue(t, ctx, &app.GovKeeper)

	newProposalMsg, err := v1.NewMsgSubmitProposal(
		[]sdk.Msg{mkTestLegacyContent(t)},
		sdk.Coins{sdk.NewInt64Coin(sdk.DefaultBondDenom, 100000000000000000)},
		addrs[0].String(),
		"",
		"Proposal",
		"description of proposal",
		false,
	)
	require.NoError(t, err)

	res, err := govMsgSvr.SubmitProposal(ctx, newProposalMsg)
	require.NoError(t, err)
	require.NotNil(t, res)

	checkInactiveProposalsQueue(t, ctx, &app.GovKeeper)

	newHeader := ctx.BlockHeader()
	newHeader.Time = ctx.BlockHeader().Time.Add(time.Duration(2) * time.Second)
	ctx = ctx.WithBlockHeader(newHeader)

	checkInactiveProposalsQueue(t, ctx, &app.GovKeeper)

	newProposalMsg2, err := v1.NewMsgSubmitProposal(
		[]sdk.Msg{mkTestLegacyContent(t)},
		sdk.Coins{sdk.NewInt64Coin(sdk.DefaultBondDenom, 100000000000000000)},
		addrs[0].String(),
		"",
		"Proposal",
		"description of proposal",
		false,
	)
	require.NoError(t, err)

	res, err = govMsgSvr.SubmitProposal(ctx, newProposalMsg2)
	require.NoError(t, err)
	require.NotNil(t, res)

	newHeader = ctx.BlockHeader()
	params, _ := app.GovKeeper.Params.Get(ctx)
	newHeader.Time = ctx.BlockHeader().Time.Add(*params.MaxDepositPeriod).Add(time.Duration(-1) * time.Second)
	ctx = ctx.WithBlockHeader(newHeader)

	checkInactiveProposalsQueue(t, ctx, &app.GovKeeper)
	require.NoError(t, gov.EndBlocker(ctx, &app.GovKeeper))
	checkInactiveProposalsQueue(t, ctx, &app.GovKeeper)

	newHeader = ctx.BlockHeader()
	newHeader.Time = ctx.BlockHeader().Time.Add(time.Duration(5) * time.Second)
	ctx = ctx.WithBlockHeader(newHeader)

	checkInactiveProposalsQueue(t, ctx, &app.GovKeeper)
	require.NoError(t, gov.EndBlocker(ctx, &app.GovKeeper))
	checkInactiveProposalsQueue(t, ctx, &app.GovKeeper)
}

func TestTickPassedDepositPeriod(t *testing.T) {
	suite := createTestSuite(t)
	app := suite.App
	ctx := app.BaseApp.NewContext(false)
	addrs := simtestutil.AddTestAddrs(app.BankKeeper, ctx, 10, valTokens)

	govMsgSvr := keeper.NewMsgServerImpl(&app.GovKeeper)

	newProposalMsg, err := v1.NewMsgSubmitProposal(
		[]sdk.Msg{mkTestLegacyContent(t)},
		sdk.Coins{sdk.NewInt64Coin(sdk.DefaultBondDenom, 100000000000000000)},
		addrs[0].String(),
		"",
		"Proposal",
		"description of proposal",
		false,
	)
	require.NoError(t, err)

	res, err := govMsgSvr.SubmitProposal(ctx, newProposalMsg)
	require.NoError(t, err)
	require.NotNil(t, res)

	proposalID := res.ProposalId

	checkInactiveProposalsQueue(t, ctx, &app.GovKeeper)

	newHeader := ctx.BlockHeader()
	newHeader.Time = ctx.BlockHeader().Time.Add(time.Duration(1) * time.Second)
	ctx = ctx.WithBlockHeader(newHeader)

	checkInactiveProposalsQueue(t, ctx, &app.GovKeeper)

	newDepositMsg := v1.NewMsgDeposit(addrs[1], proposalID, sdk.Coins{sdk.NewInt64Coin(sdk.DefaultBondDenom, 100000000000000000)})

	res1, err := govMsgSvr.Deposit(ctx, newDepositMsg)
	require.NoError(t, err)
	require.NotNil(t, res1)

	checkActiveProposalsQueue(t, ctx, &app.GovKeeper)
}

func TestTickPassedVotingPeriod(t *testing.T) {
	testcases := []struct {
		name      string
		expedited bool
	}{
		{
			name: "regular - deleted",
		},
		{
			name:      "expedited - converted to regular",
			expedited: true,
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			suite := createTestSuite(t)
			app := suite.App
			ctx := app.BaseApp.NewContext(false)
			depositMultiplier := getDepositMultiplier(tc.expedited)
			addrs := simtestutil.AddTestAddrs(app.BankKeeper, ctx, 10, valTokens.Mul(math.NewInt(depositMultiplier)))

			SortAddresses(addrs)

			govMsgSvr := keeper.NewMsgServerImpl(&app.GovKeeper)

			checkInactiveProposalsQueue(t, ctx, &app.GovKeeper)
			checkActiveProposalsQueue(t, ctx, &app.GovKeeper)

			proposalCoins := sdk.Coins{sdk.NewCoin(sdk.DefaultBondDenom, sdk.TokensFromConsensusPower(5*depositMultiplier, sdk.DefaultPowerReduction))}
			newProposalMsg, err := v1.NewMsgSubmitProposal([]sdk.Msg{mkTestLegacyContent(t)}, proposalCoins, addrs[0].String(), "", "Proposal", "description of proposal", tc.expedited)
			require.NoError(t, err)

			res, err := govMsgSvr.SubmitProposal(ctx, newProposalMsg)
			require.NoError(t, err)
			require.NotNil(t, res)

			proposalID := res.ProposalId

			newHeader := ctx.BlockHeader()
			newHeader.Time = ctx.BlockHeader().Time.Add(time.Duration(1) * time.Second)
			ctx = ctx.WithBlockHeader(newHeader)

			newDepositMsg := v1.NewMsgDeposit(addrs[1], proposalID, proposalCoins)

			res1, err := govMsgSvr.Deposit(ctx, newDepositMsg)
			require.NoError(t, err)
			require.NotNil(t, res1)

			params, _ := app.GovKeeper.Params.Get(ctx)
			votingPeriod := params.VotingPeriod
			if tc.expedited {
				votingPeriod = params.ExpeditedVotingPeriod
			}

			newHeader = ctx.BlockHeader()
			newHeader.Time = ctx.BlockHeader().Time.Add(*params.MaxDepositPeriod).Add(*votingPeriod)
			ctx = ctx.WithBlockHeader(newHeader)

			checkInactiveProposalsQueue(t, ctx, &app.GovKeeper)
			checkActiveProposalsQueue(t, ctx, &app.GovKeeper)

			proposal, err := app.GovKeeper.Proposals.Get(ctx, res.ProposalId)
			require.NoError(t, err)
			require.Equal(t, v1.StatusVotingPeriod, proposal.Status)

			err = gov.EndBlocker(ctx, &app.GovKeeper)
			require.NoError(t, err)

			if !tc.expedited {
				checkActiveProposalsQueue(t, ctx, &app.GovKeeper)
				return
			}

			// If expedited, it should be converted to a regular proposal instead.
			checkActiveProposalsQueue(t, ctx, &app.GovKeeper)

			proposal, err = app.GovKeeper.Proposals.Get(ctx, res.ProposalId)
			require.Nil(t, err)
			require.Equal(t, v1.StatusVotingPeriod, proposal.Status)
			require.False(t, proposal.Expedited)
			require.Equal(t, proposal.VotingStartTime.Add(*params.VotingPeriod), *proposal.VotingEndTime)
		})
	}
}

func TestProposalPassedEndblocker(t *testing.T) {
	testcases := []struct {
		name      string
		expedited bool
	}{
		{
			name: "regular",
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			suite := createTestSuite(t)
			app := suite.App
			ctx := app.BaseApp.NewContext(false)
			depositMultiplier := getDepositMultiplier(tc.expedited)
			addrs := simtestutil.AddTestAddrs(app.BankKeeper, ctx, 10, valTokens.Mul(math.NewInt(depositMultiplier)))

			valTokenAmount := sdk.TokensFromConsensusPower(1000000000000, sdk.DefaultPowerReduction).Mul(math.NewInt(depositMultiplier))
			fundAccounts(t, ctx, app.BankKeeper, addrs, valTokenAmount)

			govMsgSvr := keeper.NewMsgServerImpl(&app.GovKeeper)
			stakeMsgSvr := stakeKeeper.NewMsgServerImpl(&app.StakeKeeper)
			stakeSideMsgSvr := stakeKeeper.NewSideMsgServerImpl(&app.StakeKeeper)

			valAddr := sdk.ValAddress(addrs[0])
			proposer := addrs[0]

			createValidators(t, stakeMsgSvr, stakeSideMsgSvr, ctx, []sdk.ValAddress{valAddr}, []math.Int{valTokenAmount})
			app.StakeKeeper.EndBlocker(ctx)

			macc := app.GovKeeper.GetGovernanceAccount(ctx)
			require.NotNil(t, macc)
			initialModuleAccCoins := app.BankKeeper.GetAllBalances(ctx, macc.GetAddress())

			proposal, err := app.GovKeeper.SubmitProposal(ctx, []sdk.Msg{mkTestLegacyContent(t)}, "", "title", "summary", proposer, tc.expedited)
			require.NoError(t, err)

			proposalCoins := sdk.Coins{sdk.NewCoin(sdk.DefaultBondDenom, sdk.TokensFromConsensusPower(10*depositMultiplier, sdk.DefaultPowerReduction))}
			newDepositMsg := v1.NewMsgDeposit(addrs[0], proposal.Id, proposalCoins)

			res, err := govMsgSvr.Deposit(ctx, newDepositMsg)
			require.NoError(t, err)
			require.NotNil(t, res)

			macc = app.GovKeeper.GetGovernanceAccount(ctx)
			require.NotNil(t, macc)
			moduleAccCoins := app.BankKeeper.GetAllBalances(ctx, macc.GetAddress())

			deposits := initialModuleAccCoins.Add(proposal.TotalDeposit...).Add(proposalCoins...)
			require.True(t, moduleAccCoins.Equal(deposits))

			err = app.GovKeeper.AddVote(ctx, proposal.Id, addrs[0], v1.NewNonSplitVoteOption(v1.OptionYes), "")
			require.NoError(t, err)

			newHeader := ctx.BlockHeader()
			params, _ := app.GovKeeper.Params.Get(ctx)
			newHeader.Time = ctx.BlockHeader().Time.Add(*params.MaxDepositPeriod).Add(*params.VotingPeriod)
			ctx = ctx.WithBlockHeader(newHeader)

			gov.EndBlocker(ctx, &app.GovKeeper)

			macc = app.GovKeeper.GetGovernanceAccount(ctx)
			require.NotNil(t, macc)

			// This happens because of truncation in division in DistributeAndDeleteDeposits
			// 1 pol is left at the end
			require.True(t, app.BankKeeper.GetAllBalances(ctx, macc.GetAddress()).Equal(sdk.NewCoins(sdk.NewCoin("pol", math.NewInt(1)))))
		})
	}
}

func TestEndBlockerProposalHandlerFailed(t *testing.T) {
	suite := createTestSuite(t)
	app := suite.App
	ctx := app.BaseApp.NewContext(false)
	depositMultiplier := getDepositMultiplier(true)
	addrs := simtestutil.AddTestAddrs(app.BankKeeper, ctx, 1, valTokens)

	valTokenAmount := sdk.TokensFromConsensusPower(1000000000000, sdk.DefaultPowerReduction).Mul(math.NewInt(depositMultiplier))
	fundAccounts(t, ctx, app.BankKeeper, addrs, valTokenAmount)

	stakeMsgSvr := stakeKeeper.NewMsgServerImpl(&app.StakeKeeper)
	stakeSideMsgSvr := stakeKeeper.NewSideMsgServerImpl(&app.StakeKeeper)

	valAddr := sdk.ValAddress(addrs[0])
	proposer := addrs[0]

	createValidators(t, stakeMsgSvr, stakeSideMsgSvr, ctx, []sdk.ValAddress{valAddr}, []math.Int{valTokenAmount})
	app.StakeKeeper.EndBlocker(ctx)

	legacyProposalMsg1, err := v1.NewLegacyContent(v1beta1.NewTextProposal("Title1", "description1"), authtypes.NewModuleAddress(types.ModuleName).String())
	require.NoError(t, err)
	msg := []sdk.Msg{legacyProposalMsg1}
	proposal, err := app.GovKeeper.SubmitProposal(ctx, msg, "", "title", "summary", proposer, false)
	require.NoError(t, err)

	proposalCoins := sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdk.TokensFromConsensusPower(10, sdk.DefaultPowerReduction)))
	newDepositMsg := v1.NewMsgDeposit(addrs[0], proposal.Id, proposalCoins)

	govMsgSvr := keeper.NewMsgServerImpl(&app.GovKeeper)
	res, err := govMsgSvr.Deposit(ctx, newDepositMsg)
	require.NoError(t, err)
	require.NotNil(t, res)

	err = app.GovKeeper.AddVote(ctx, proposal.Id, addrs[0], v1.NewNonSplitVoteOption(v1.OptionYes), "")
	require.NoError(t, err)

	params, _ := app.GovKeeper.Params.Get(ctx)
	newHeader := ctx.BlockHeader()
	newHeader.Time = ctx.BlockHeader().Time.Add(*params.MaxDepositPeriod).Add(*params.VotingPeriod)
	ctx = ctx.WithBlockHeader(newHeader)

	// validate that the proposal fails/has been rejected
	gov.EndBlocker(ctx, &app.GovKeeper)

	// check proposal events
	events := ctx.EventManager().Events()
	attr, eventOk := events.GetAttributes(types.AttributeKeyProposalLog)
	require.True(t, eventOk)
	// HV2: rejected because of missing implementation of `IterateCurrentValidatorsAndApplyFn`.
	// That function will be implemented in heimdall's staking module.
	require.Contains(t, attr[0].Value, "rejected")

	proposal, err = app.GovKeeper.Proposals.Get(ctx, proposal.Id)
	require.Nil(t, err)
	require.Equal(t, v1.StatusRejected, proposal.Status)
}

func TestExpeditedProposal_PassAndConversionToRegular(t *testing.T) {
	t.Skip("In HV2, expedited proposals are not supported.")
	testcases := []struct {
		name string
		// indicates whether the expedited proposal passes.
		expeditedPasses bool
		// indicates whether the converted regular proposal is expected to eventually pass
		regularEventuallyPassing bool
	}{
		{
			name:            "expedited passes and not converted to regular",
			expeditedPasses: true,
		},
		{
			name:                     "expedited fails, converted to regular - regular eventually passes",
			expeditedPasses:          false,
			regularEventuallyPassing: true,
		},
		{
			name:                     "expedited fails, converted to regular - regular eventually fails",
			expeditedPasses:          false,
			regularEventuallyPassing: false,
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			suite := createTestSuite(t)
			app := suite.App
			ctx := app.BaseApp.NewContext(false)
			depositMultiplier := getDepositMultiplier(true)
			addrs := []sdk.AccAddress{}
			for _, pubkey := range pubkeys {
				accAddr := sdk.AccAddress(pubkey.Address())
				addrs = append(addrs, accAddr)
			}

			valTokenAmount := sdk.TokensFromConsensusPower(1000000000000, sdk.DefaultPowerReduction).Mul(math.NewInt(depositMultiplier))
			fundAccounts(t, ctx, app.BankKeeper, addrs, valTokenAmount)

			params, err := app.GovKeeper.Params.Get(ctx)
			require.NoError(t, err)

			govMsgSvr := keeper.NewMsgServerImpl(&app.GovKeeper)
			stakeMsgSvr := stakeKeeper.NewMsgServerImpl(&app.StakeKeeper)
			stakeSideMsgSvr := stakeKeeper.NewSideMsgServerImpl(&app.StakeKeeper)

			valAddr := sdk.ValAddress(addrs[0])
			proposer := addrs[0]

			// Create a validator so that able to vote on proposal.
			createValidators(t, stakeMsgSvr, stakeSideMsgSvr, ctx, []sdk.ValAddress{valAddr}, []math.Int{valTokenAmount})
			app.StakeKeeper.EndBlocker(ctx)

			checkInactiveProposalsQueue(t, ctx, &app.GovKeeper)
			checkActiveProposalsQueue(t, ctx, &app.GovKeeper)

			macc := app.GovKeeper.GetGovernanceAccount(ctx)
			require.NotNil(t, macc)
			initialModuleAccCoins := app.BankKeeper.GetAllBalances(ctx, macc.GetAddress())

			submitterInitialBalance := app.BankKeeper.GetAllBalances(ctx, addrs[0])
			depositorInitialBalance := app.BankKeeper.GetAllBalances(ctx, addrs[1])

			proposalCoins := sdk.Coins{sdk.NewCoin(sdk.DefaultBondDenom, sdk.TokensFromConsensusPower(5*depositMultiplier, sdk.DefaultPowerReduction))}
			newProposalMsg, err := v1.NewMsgSubmitProposal([]sdk.Msg{}, proposalCoins, proposer.String(), "metadata", "title", "summary", true)
			require.NoError(t, err)

			res, err := govMsgSvr.SubmitProposal(ctx, newProposalMsg)
			require.NoError(t, err)
			require.NotNil(t, res)

			proposalID := res.ProposalId

			newHeader := ctx.BlockHeader()
			newHeader.Time = ctx.BlockHeader().Time.Add(time.Duration(1) * time.Second)
			ctx = ctx.WithBlockHeader(newHeader)

			newDepositMsg := v1.NewMsgDeposit(addrs[1], proposalID, proposalCoins)

			res1, err := govMsgSvr.Deposit(ctx, newDepositMsg)
			require.NoError(t, err)
			require.NotNil(t, res1)

			newHeader = ctx.BlockHeader()
			newHeader.Time = ctx.BlockHeader().Time.Add(*params.MaxDepositPeriod).Add(*params.ExpeditedVotingPeriod)
			ctx = ctx.WithBlockHeader(newHeader)

			checkInactiveProposalsQueue(t, ctx, &app.GovKeeper)
			checkActiveProposalsQueue(t, ctx, &app.GovKeeper)

			proposal, err := app.GovKeeper.Proposals.Get(ctx, res.ProposalId)
			require.Nil(t, err)
			require.Equal(t, v1.StatusVotingPeriod, proposal.Status)

			if tc.expeditedPasses {
				// Validator votes YES, letting the expedited proposal pass.
				err = app.GovKeeper.AddVote(ctx, proposal.Id, addrs[0], v1.NewNonSplitVoteOption(v1.OptionYes), "metadata")
				require.NoError(t, err)
			}

			// Here the expedited proposal is converted to regular after expiry.
			gov.EndBlocker(ctx, &app.GovKeeper)

			if tc.expeditedPasses {
				checkActiveProposalsQueue(t, ctx, &app.GovKeeper)

				proposal, err = app.GovKeeper.Proposals.Get(ctx, res.ProposalId)
				require.Nil(t, err)

				require.Equal(t, v1.StatusPassed, proposal.Status)

				submitterEventualBalance := app.BankKeeper.GetAllBalances(ctx, addrs[0])
				depositorEventualBalance := app.BankKeeper.GetAllBalances(ctx, addrs[1])

				eventualModuleAccCoins := app.BankKeeper.GetAllBalances(ctx, macc.GetAddress())

				// Module account has refunded the deposit
				require.Equal(t, initialModuleAccCoins, eventualModuleAccCoins)

				require.Equal(t, submitterInitialBalance, submitterEventualBalance)
				require.Equal(t, depositorInitialBalance, depositorEventualBalance)
				return
			}

			// Expedited proposal should be converted to a regular proposal instead.
			checkActiveProposalsQueue(t, ctx, &app.GovKeeper)
			proposal, err = app.GovKeeper.Proposals.Get(ctx, res.ProposalId)
			require.Nil(t, err)
			require.Equal(t, v1.StatusVotingPeriod, proposal.Status)
			require.False(t, proposal.Expedited)
			require.Equal(t, proposal.VotingStartTime.Add(*params.VotingPeriod), *proposal.VotingEndTime)

			// We also want to make sure that the deposit is not refunded yet and is still present in the module account
			macc = app.GovKeeper.GetGovernanceAccount(ctx)
			require.NotNil(t, macc)
			intermediateModuleAccCoins := app.BankKeeper.GetAllBalances(ctx, macc.GetAddress())
			require.NotEqual(t, initialModuleAccCoins, intermediateModuleAccCoins)

			// Submit proposal deposit + 1 extra top up deposit
			expectedIntermediateMofuleAccCoings := initialModuleAccCoins.Add(proposalCoins...).Add(proposalCoins...)
			require.Equal(t, expectedIntermediateMofuleAccCoings, intermediateModuleAccCoins)

			// block header time at the voting period
			newHeader.Time = ctx.BlockHeader().Time.Add(*params.MaxDepositPeriod).Add(*params.VotingPeriod)
			ctx = ctx.WithBlockHeader(newHeader)

			checkInactiveProposalsQueue(t, ctx, &app.GovKeeper)
			checkActiveProposalsQueue(t, ctx, &app.GovKeeper)

			if tc.regularEventuallyPassing {
				// Validator votes YES, letting the converted regular proposal pass.
				err = app.GovKeeper.AddVote(ctx, proposal.Id, addrs[0], v1.NewNonSplitVoteOption(v1.OptionYes), "metadata")
				require.NoError(t, err)
			}

			// Here we validate the converted regular proposal
			gov.EndBlocker(ctx, &app.GovKeeper)

			macc = app.GovKeeper.GetGovernanceAccount(ctx)
			require.NotNil(t, macc)
			eventualModuleAccCoins := app.BankKeeper.GetAllBalances(ctx, macc.GetAddress())

			submitterEventualBalance := app.BankKeeper.GetAllBalances(ctx, addrs[0])
			depositorEventualBalance := app.BankKeeper.GetAllBalances(ctx, addrs[1])

			checkActiveProposalsQueue(t, ctx, &app.GovKeeper)

			proposal, err = app.GovKeeper.Proposals.Get(ctx, res.ProposalId)
			require.Nil(t, err)

			if tc.regularEventuallyPassing {
				// Module account has refunded the deposit
				require.Equal(t, initialModuleAccCoins, eventualModuleAccCoins)
				require.Equal(t, submitterInitialBalance, submitterEventualBalance)
				require.Equal(t, depositorInitialBalance, depositorEventualBalance)

				require.Equal(t, v1.StatusPassed, proposal.Status)
				return
			}

			// Not enough votes - module account has returned the deposit
			require.Equal(t, initialModuleAccCoins, eventualModuleAccCoins)
			require.Equal(t, submitterInitialBalance, submitterEventualBalance)
			require.Equal(t, depositorInitialBalance, depositorEventualBalance)

			require.Equal(t, v1.StatusRejected, proposal.Status)
		})
	}
}

func createValidators(t *testing.T, stakeMsgSvr stakeTypes.MsgServer, sideMsgSvr sideTxs.SideMsgServer, ctx sdk.Context, addrs []sdk.ValAddress, powerAmt []math.Int) {
	require.True(t, len(addrs) <= len(pubkeys), "Not enough pubkeys specified at top of file.")

	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	txHash := common.Hash{}.Bytes()
	index := simulation.RandIntBetween(r, 0, 100)
	logIndex := uint64(index)
	blockNumber := big.NewInt(10)
	nonce := big.NewInt(3)

	for i := 0; i < len(addrs); i++ {
		valCreateMsg, err := stakeTypes.NewMsgValidatorJoin(
			addrs[i].String(),
			uint64(i+2),
			uint64(1),
			powerAmt[i],
			pubkeys[i],
			txHash,
			logIndex,
			blockNumber.Uint64(),
			nonce.Uint64(),
		)
		require.NoError(t, err)
		res, err := stakeMsgSvr.ValidatorJoin(ctx, valCreateMsg)
		require.NoError(t, err)
		require.NotNil(t, res)

		joinValHandler := sideMsgSvr.PostTxHandler(sdk.MsgTypeURL(valCreateMsg))
		require.NotNil(t, joinValHandler)

		joinValHandler(ctx, valCreateMsg, sideTxs.Vote_VOTE_YES)
	}
}

// With expedited proposal's minimum deposit set higher than the default deposit, we must
// initialize and deposit an amount depositMultiplier times larger
// than the regular min deposit amount.
func getDepositMultiplier(expedited bool) int64 {
	if expedited {
		return v1.DefaultMinExpeditedDepositTokensRatio
	}

	return 1
}

func fundAccounts(t *testing.T, ctx sdk.Context, bankKeeper bankKeeper.Keeper, addrs []sdk.AccAddress, accAmt math.Int) {
	coins := sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, accAmt))

	for _, addr := range addrs {
		err := bankKeeper.MintCoins(ctx, topupTypes.ModuleName, coins)
		require.NoError(t, err)
		err = bankKeeper.SendCoinsFromModuleToAccount(ctx, topupTypes.ModuleName, addr, coins)
		require.NoError(t, err)
	}
}

func checkActiveProposalsQueue(t *testing.T, ctx sdk.Context, k *keeper.Keeper) {
	err := k.ActiveProposalsQueue.Walk(ctx, collections.NewPrefixUntilPairRange[time.Time, uint64](ctx.BlockTime()), func(key collections.Pair[time.Time, uint64], value uint64) (stop bool, err error) {
		return false, err
	})

	require.NoError(t, err)
}

func checkInactiveProposalsQueue(t *testing.T, ctx sdk.Context, k *keeper.Keeper) {
	err := k.InactiveProposalsQueue.Walk(ctx, collections.NewPrefixUntilPairRange[time.Time, uint64](ctx.BlockTime()), func(key collections.Pair[time.Time, uint64], value uint64) (stop bool, err error) {
		return false, err
	})

	require.NoError(t, err)
}
