package keeper_test

import (
	borTypes "github.com/0xPolygon/heimdall-v2/x/bor/types"
	checkpointTypes "github.com/0xPolygon/heimdall-v2/x/checkpoint/types"
	milestoneTypes "github.com/0xPolygon/heimdall-v2/x/milestone/types"
	"math/big"
	"strings"
	"time"

	authv1beta1 "cosmossdk.io/api/cosmos/auth/v1beta1"
	bankv1beta1 "cosmossdk.io/api/cosmos/bank/v1beta1"
	consensusv1 "cosmossdk.io/api/cosmos/consensus/v1"
	govv1 "cosmossdk.io/api/cosmos/gov/v1"
	sdkmath "cosmossdk.io/math"
	chainmanagertypes "github.com/0xPolygon/heimdall-v2/x/chainmanager/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	consensustypes "github.com/cosmos/cosmos-sdk/x/consensus/types"

	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
	"github.com/cosmos/cosmos-sdk/testutil/testdata"
	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	v1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	"github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
)

const (
	abc = "abc"
	o1  = "-0.1"
)

var longAddressError = "address max length is 255"

func (suite *KeeperTestSuite) TestSubmitProposalReq() {
	suite.reset()
	govAcct := suite.govKeeper.GetGovernanceAccount(suite.ctx).GetAddress()
	addrs := suite.addrs
	proposer := addrs[0]

	coins := sdk.NewCoins(sdk.NewCoin("matic", sdkmath.NewInt(100000000000000000)))
	initialDeposit := coins
	params, _ := suite.govKeeper.Params.Get(suite.ctx)
	minDeposit := params.MinDeposit
	tp := TestProposal
	tpWithBadSigner, _ := v1.NewLegacyContent(v1beta1.NewTextProposal("Title", "description"), addrs[0].String())

	cases := map[string]struct {
		preRun    func() (*v1.MsgSubmitProposal, error)
		expErr    bool
		expErrMsg string
	}{
		"invalid addr": {
			preRun: func() (*v1.MsgSubmitProposal, error) {
				return v1.NewMsgSubmitProposal(
					tp,
					initialDeposit,
					"",
					strings.Repeat("1", 300),
					"Proposal",
					"description of proposal",
					false,
				)
			},
			expErr:    true,
			expErrMsg: "invalid proposer address",
		},
		"empty msgs and metadata": {
			preRun: func() (*v1.MsgSubmitProposal, error) {
				return v1.NewMsgSubmitProposal(
					nil,
					initialDeposit,
					proposer.String(),
					"",
					"Proposal",
					"description of proposal",
					false,
				)
			},
			expErr:    true,
			expErrMsg: "no messages proposed",
		},
		"empty title": {
			preRun: func() (*v1.MsgSubmitProposal, error) {
				return v1.NewMsgSubmitProposal(
					tp,
					initialDeposit,
					proposer.String(),
					"",
					"",
					"description of proposal",
					false,
				)
			},
			expErr:    true,
			expErrMsg: "proposal title cannot be empty",
		},
		"empty description": {
			preRun: func() (*v1.MsgSubmitProposal, error) {
				return v1.NewMsgSubmitProposal(
					tp,
					initialDeposit,
					proposer.String(),
					"",
					"Proposal",
					"",
					false,
				)
			},
			expErr:    true,
			expErrMsg: "proposal summary cannot be empty",
		},
		"title != metadata.title": {
			preRun: func() (*v1.MsgSubmitProposal, error) {
				return v1.NewMsgSubmitProposal(
					tp,
					initialDeposit,
					proposer.String(),
					"{\"title\":\"Proposal\", \"description\":\"description of proposal\"}",
					"Proposal2",
					"description of proposal",
					false,
				)
			},
			expErr:    true,
			expErrMsg: "metadata title 'Proposal' must equal proposal title 'Proposal2",
		},
		"summary != metadata.summary": {
			preRun: func() (*v1.MsgSubmitProposal, error) {
				return v1.NewMsgSubmitProposal(
					tp,
					initialDeposit,
					proposer.String(),
					"{\"title\":\"Proposal\", \"description\":\"description of proposal\"}",
					"Proposal",
					"description",
					false,
				)
			},
			expErr:    true,
			expErrMsg: "metadata summary '' must equal proposal summary 'description'",
		},
		"metadata too long": {
			preRun: func() (*v1.MsgSubmitProposal, error) {
				return v1.NewMsgSubmitProposal(
					tp,
					initialDeposit,
					proposer.String(),
					strings.Repeat("1", 300),
					"Proposal",
					"description of proposal",
					false,
				)
			},
			expErr:    true,
			expErrMsg: "metadata too long",
		},
		"summary too long": {
			preRun: func() (*v1.MsgSubmitProposal, error) {
				return v1.NewMsgSubmitProposal(
					tp,
					initialDeposit,
					proposer.String(),
					"",
					"Proposal",
					strings.Repeat("1", 300*40),
					false,
				)
			},
			expErr:    true,
			expErrMsg: "summary too long",
		},
		// HV2: testdata.TestMsg is not a valid proposal message type
		"invalid proposal message type": {
			preRun: func() (*v1.MsgSubmitProposal, error) {
				return v1.NewMsgSubmitProposal(
					[]sdk.Msg{testdata.NewTestMsg(govAcct, addrs[0])},
					initialDeposit,
					proposer.String(),
					"",
					"Proposal",
					"description of proposal",
					false,
				)
			},
			expErr:    true,
			expErrMsg: " invalid proposal message type",
		},
		"signer isn't gov account": {
			preRun: func() (*v1.MsgSubmitProposal, error) {
				return v1.NewMsgSubmitProposal(
					[]sdk.Msg{tpWithBadSigner},
					initialDeposit,
					proposer.String(),
					"",
					"Proposal",
					"description of proposal",
					false,
				)
			},
			expErr:    true,
			expErrMsg: "expected gov account as only signer for proposal message",
		},
		"type not supported": {
			preRun: func() (*v1.MsgSubmitProposal, error) {
				return v1.NewMsgSubmitProposal(
					[]sdk.Msg{testdata.NewTestMsg(govAcct)},
					initialDeposit,
					proposer.String(),
					"",
					"Proposal",
					"description of proposal",
					false,
				)
			},
			expErr:    true,
			expErrMsg: "invalid proposal message type",
		},
		// HV2: banktypes.NewMsgSend is not a valid proposal message type
		"msg send not supported": {
			preRun: func() (*v1.MsgSubmitProposal, error) {
				return v1.NewMsgSubmitProposal(
					[]sdk.Msg{banktypes.NewMsgSend(govAcct, addr, sdk.NewCoins(sdk.NewCoin("matic", sdkmath.NewInt(1000))))},
					initialDeposit,
					proposer.String(),
					"",
					"Proposal",
					"description of proposal",
					false,
				)
			},
			expErr:    true,
			expErrMsg: "invalid proposal message type",
		},
		"invalid deposited coin": {
			preRun: func() (*v1.MsgSubmitProposal, error) {
				return v1.NewMsgSubmitProposal(
					tp,
					[]sdk.Coin{sdk.NewCoin("invalid", sdkmath.NewInt(100))},
					proposer.String(),
					"",
					"Proposal",
					"description of proposal",
					false,
				)
			},
			expErr:    true,
			expErrMsg: "deposited 100invalid, but gov accepts only the following denom(s): [matic]: invalid deposit denom",
		},
		"invalid deposited coin (multiple)": {
			preRun: func() (*v1.MsgSubmitProposal, error) {
				return v1.NewMsgSubmitProposal(
					tp,
					initialDeposit.Add(sdk.NewCoin("invalid", sdkmath.NewInt(100))),
					proposer.String(),
					"",
					"Proposal",
					"description of proposal",
					false,
				)
			},
			expErr:    true,
			expErrMsg: "deposited 100invalid,100000000000000000matic, but gov accepts only the following denom(s): [matic]: invalid deposit denom",
		},
		"all good": {
			preRun: func() (*v1.MsgSubmitProposal, error) {
				return v1.NewMsgSubmitProposal(
					tp,
					initialDeposit,
					proposer.String(),
					"",
					"Proposal",
					"description of proposal",
					false,
				)
			},
			expErr: false,
		},
		"all good with min deposit": {
			preRun: func() (*v1.MsgSubmitProposal, error) {
				return v1.NewMsgSubmitProposal(
					tp,
					minDeposit,
					proposer.String(),
					"",
					"Proposal",
					"description of proposal",
					false,
				)
			},
			expErr: false,
		},
		"all good with v1.MsgUpdateParams": {
			preRun: func() (*v1.MsgSubmitProposal, error) {
				return v1.NewMsgSubmitProposal(
					[]sdk.Msg{&v1.MsgUpdateParams{Authority: govAcct.String()}},
					initialDeposit,
					proposer.String(),
					"",
					"Proposal",
					"description of proposal",
					false,
				)
			},
			expErr: false,
		},
		"all good with govv1.MsgUpdateParams": {
			preRun: func() (*v1.MsgSubmitProposal, error) {
				return v1.NewMsgSubmitProposal(
					[]sdk.Msg{&govv1.MsgUpdateParams{Authority: govAcct.String()}},
					initialDeposit,
					proposer.String(),
					"",
					"Proposal",
					"description of proposal",
					false,
				)
			},
			expErr: false,
		},
		"all good with banktypes.MsgUpdateParams": {
			preRun: func() (*v1.MsgSubmitProposal, error) {
				return v1.NewMsgSubmitProposal(
					[]sdk.Msg{&banktypes.MsgUpdateParams{Authority: govAcct.String()}},
					initialDeposit,
					proposer.String(),
					"",
					"Proposal",
					"description of proposal",
					false,
				)
			},
			expErr: false,
		},
		"all good with bankv1beta1.MsgUpdateParams": {
			preRun: func() (*v1.MsgSubmitProposal, error) {
				return v1.NewMsgSubmitProposal(
					[]sdk.Msg{&bankv1beta1.MsgUpdateParams{Authority: govAcct.String()}},
					initialDeposit,
					proposer.String(),
					"",
					"Proposal",
					"description of proposal",
					false,
				)
			},
			expErr: false,
		},
		"all good with authtypes.MsgUpdateParams": {
			preRun: func() (*v1.MsgSubmitProposal, error) {
				return v1.NewMsgSubmitProposal(
					[]sdk.Msg{&authtypes.MsgUpdateParams{Authority: govAcct.String()}},
					initialDeposit,
					proposer.String(),
					"",
					"Proposal",
					"description of proposal",
					false,
				)
			},
			expErr: false,
		},
		"all good with authv1beta1.MsgUpdateParams": {
			preRun: func() (*v1.MsgSubmitProposal, error) {
				return v1.NewMsgSubmitProposal(
					[]sdk.Msg{&authv1beta1.MsgUpdateParams{Authority: govAcct.String()}},
					initialDeposit,
					proposer.String(),
					"",
					"Proposal",
					"description of proposal",
					false,
				)
			},
			expErr: false,
		},
		"all good with consensustypes.MsgUpdateParams": {
			preRun: func() (*v1.MsgSubmitProposal, error) {
				return v1.NewMsgSubmitProposal(
					[]sdk.Msg{&consensustypes.MsgUpdateParams{Authority: govAcct.String()}},
					initialDeposit,
					proposer.String(),
					"",
					"Proposal",
					"description of proposal",
					false,
				)
			},
			expErr: false,
		},
		"all good with with consensusv1.MsgUpdateParams": {
			preRun: func() (*v1.MsgSubmitProposal, error) {
				return v1.NewMsgSubmitProposal(
					[]sdk.Msg{&consensusv1.MsgUpdateParams{Authority: govAcct.String()}},
					initialDeposit,
					proposer.String(),
					"",
					"Proposal",
					"description of proposal",
					false,
				)
			},
			expErr: false,
		},
		"all good with borTypes.MsgUpdateParams": {
			preRun: func() (*v1.MsgSubmitProposal, error) {
				return v1.NewMsgSubmitProposal(
					[]sdk.Msg{&borTypes.MsgUpdateParams{Authority: govAcct.String()}},
					initialDeposit,
					proposer.String(),
					"",
					"Proposal",
					"description of proposal",
					false,
				)
			},
			expErr: false,
		},
		"all good with checkpointTypes.MsgUpdateParams": {
			preRun: func() (*v1.MsgSubmitProposal, error) {
				return v1.NewMsgSubmitProposal(
					[]sdk.Msg{&checkpointTypes.MsgUpdateParams{Authority: govAcct.String()}},
					initialDeposit,
					proposer.String(),
					"",
					"Proposal",
					"description of proposal",
					false,
				)
			},
			expErr: false,
		},
		"all good with milestoneTypes.MsgUpdateParams": {
			preRun: func() (*v1.MsgSubmitProposal, error) {
				return v1.NewMsgSubmitProposal(
					[]sdk.Msg{&milestoneTypes.MsgUpdateParams{Authority: govAcct.String()}},
					initialDeposit,
					proposer.String(),
					"",
					"Proposal",
					"description of proposal",
					false,
				)
			},
			expErr: false,
		},
		"all good with chainmanagertypes.MsgUpdateParams": {
			preRun: func() (*v1.MsgSubmitProposal, error) {
				return v1.NewMsgSubmitProposal(
					[]sdk.Msg{&chainmanagertypes.MsgUpdateParams{Authority: govAcct.String()}},
					initialDeposit,
					proposer.String(),
					"",
					"Proposal",
					"description of proposal",
					false,
				)
			},
			expErr: false,
		},
	}

	for name, tc := range cases {
		suite.Run(name, func() {
			msg, err := tc.preRun()
			suite.Require().NoError(err)
			res, err := suite.msgSrvr.SubmitProposal(suite.ctx, msg)
			if tc.expErr {
				suite.Require().Error(err)
				suite.Require().Contains(err.Error(), tc.expErrMsg)
			} else {
				suite.Require().NoError(err)
				suite.Require().NotNil(res.ProposalId)
			}
		})
	}
}

func (suite *KeeperTestSuite) TestCancelProposalReq() {
	addrs := suite.addrs
	proposer := addrs[0]

	coins := sdk.NewCoins(sdk.NewCoin("matic", sdkmath.NewInt(100000000000000000)))

	tp := TestProposal
	msg, err := v1.NewMsgSubmitProposal(
		tp,
		coins,
		proposer.String(),
		"", "title", "summary",
		false,
	)
	suite.Require().NoError(err)

	res, err := suite.msgSrvr.SubmitProposal(suite.ctx, msg)
	suite.Require().NoError(err)
	suite.Require().NotNil(res.ProposalId)
	proposalID := res.ProposalId

	cases := map[string]struct {
		preRun     func() uint64
		expErr     bool
		expErrMsg  string
		proposalID uint64
		depositor  sdk.AccAddress
	}{
		"wrong proposal id": {
			preRun: func() uint64 {
				return 0
			},
			depositor: proposer,
			expErr:    true,
			expErrMsg: "not found",
		},
		"valid proposal but invalid proposer": {
			preRun: func() uint64 {
				return proposalID
			},
			depositor: addrs[1],
			expErr:    true,
			expErrMsg: "invalid proposer",
		},
		"empty proposer": {
			preRun: func() uint64 {
				return proposalID
			},
			depositor: sdk.AccAddress{},
			expErr:    true,
			expErrMsg: "invalid proposer address: empty address string is not allowed",
		},
		"all good": {
			preRun: func() uint64 {
				msg, err := v1.NewMsgSubmitProposal(
					tp,
					coins,
					proposer.String(),
					"",
					"Proposal",
					"description of proposal",
					false,
				)
				suite.Require().NoError(err)

				res, err := suite.msgSrvr.SubmitProposal(suite.ctx, msg)
				suite.Require().NoError(err)
				suite.Require().NotNil(res.ProposalId)
				return res.ProposalId
			},
			depositor: proposer,
			expErr:    false,
		},
	}

	for name, tc := range cases {
		suite.Run(name, func() {
			proposalID := tc.preRun()
			cancelProposalReq := v1.NewMsgCancelProposal(proposalID, tc.depositor.String())
			_, err := suite.msgSrvr.CancelProposal(suite.ctx, cancelProposalReq)
			if tc.expErr {
				suite.Require().Error(err)
				suite.Require().Contains(err.Error(), tc.expErrMsg)
			} else {
				suite.Require().NoError(err)
			}
		})
	}
}

func (suite *KeeperTestSuite) TestVoteReq() {
	suite.reset()
	addrs := suite.addrs
	proposer := addrs[0]

	coins := sdk.NewCoins(sdk.NewCoin("matic", sdkmath.NewInt(100000000000000000)))
	params, _ := suite.govKeeper.Params.Get(suite.ctx)
	minDeposit := params.MinDeposit
	tp := TestProposal

	msg, err := v1.NewMsgSubmitProposal(
		tp,
		minDeposit,
		proposer.String(),
		"",
		"Proposal",
		"description of proposal",
		false,
	)
	suite.Require().NoError(err)

	res, err := suite.msgSrvr.SubmitProposal(suite.ctx, msg)
	suite.Require().NoError(err)
	suite.Require().NotNil(res.ProposalId)
	proposalID := res.ProposalId

	cases := map[string]struct {
		preRun    func() uint64
		expErr    bool
		expErrMsg string
		option    v1.VoteOption
		metadata  string
		voter     sdk.AccAddress
	}{
		"empty voter": {
			preRun: func() uint64 {
				return proposalID
			},
			option:    v1.VoteOption_VOTE_OPTION_YES,
			voter:     sdk.AccAddress{},
			metadata:  "",
			expErr:    true,
			expErrMsg: "invalid voter address",
		},
		"wrong vote option": {
			preRun: func() uint64 {
				return proposalID
			},
			option:    v1.VoteOption(0x13),
			voter:     proposer,
			metadata:  "",
			expErr:    true,
			expErrMsg: "invalid vote option",
		},
		"vote on inactive proposal": {
			preRun: func() uint64 {
				msg, err := v1.NewMsgSubmitProposal(
					tp,
					coins,
					proposer.String(),
					"",
					"Proposal",
					"description of proposal",
					false,
				)
				suite.Require().NoError(err)

				res, err := suite.msgSrvr.SubmitProposal(suite.ctx, msg)
				suite.Require().NoError(err)
				suite.Require().NotNil(res.ProposalId)
				return res.ProposalId
			},
			option:    v1.VoteOption_VOTE_OPTION_YES,
			voter:     proposer,
			metadata:  "",
			expErr:    true,
			expErrMsg: "inactive proposal",
		},
		"metadata too long": {
			preRun: func() uint64 {
				return proposalID
			},
			option:    v1.VoteOption_VOTE_OPTION_YES,
			voter:     proposer,
			metadata:  strings.Repeat("a", 300),
			expErr:    true,
			expErrMsg: "metadata too long",
		},
		"voter error": {
			preRun: func() uint64 {
				return proposalID
			},
			option:    v1.VoteOption_VOTE_OPTION_YES,
			voter:     sdk.AccAddress(strings.Repeat("a", 300)),
			metadata:  "",
			expErr:    true,
			expErrMsg: "invalid address",
		},
		"all good": {
			preRun: func() uint64 {
				msg, err := v1.NewMsgSubmitProposal(
					tp,
					minDeposit,
					proposer.String(),
					"",
					"Proposal",
					"description of proposal",
					false,
				)
				suite.Require().NoError(err)

				res, err := suite.msgSrvr.SubmitProposal(suite.ctx, msg)
				suite.Require().NoError(err)
				suite.Require().NotNil(res.ProposalId)
				return res.ProposalId
			},
			option:   v1.VoteOption_VOTE_OPTION_YES,
			voter:    proposer,
			metadata: "",
			expErr:   false,
		},
	}

	for name, tc := range cases {
		suite.Run(name, func() {
			pID := tc.preRun()
			voteReq := v1.NewMsgVote(tc.voter, pID, tc.option, tc.metadata)
			_, err := suite.msgSrvr.Vote(suite.ctx, voteReq)
			if tc.expErr {
				suite.Require().Error(err)
				suite.Require().Contains(err.Error(), tc.expErrMsg)
			} else {
				suite.Require().NoError(err)
			}
		})
	}
}

func (suite *KeeperTestSuite) TestVoteWeightedReq() {
	suite.reset()

	accAmt := sdkmath.NewIntFromBigInt(new(big.Int).Mul(big.NewInt(10), new(big.Int).Exp(big.NewInt(10), big.NewInt(18), nil)))
	proposer := simtestutil.AddTestAddrsIncremental(suite.bankKeeper, suite.stakingKeeper, suite.ctx, 1, accAmt.Mul(sdkmath.NewInt(5)))[0]

	coins := sdk.NewCoins(sdk.NewCoin("matic", sdkmath.NewInt(100000000000000000)))
	params, _ := suite.govKeeper.Params.Get(suite.ctx)
	minDeposit := params.MinDeposit
	tp := TestProposal

	msg, err := v1.NewMsgSubmitProposal(
		tp,
		minDeposit,
		proposer.String(),
		"",
		"Proposal",
		"description of proposal",
		false,
	)
	suite.Require().NoError(err)

	res, err := suite.msgSrvr.SubmitProposal(suite.ctx, msg)
	suite.Require().NoError(err)
	suite.Require().NotNil(res.ProposalId)
	proposalID := res.ProposalId

	cases := map[string]struct {
		preRun    func() uint64
		vote      *v1.MsgVote
		expErr    bool
		expErrMsg string
		option    v1.WeightedVoteOptions
		metadata  string
		voter     sdk.AccAddress
	}{
		"empty voter": {
			preRun: func() uint64 {
				return proposalID
			},
			option: v1.WeightedVoteOptions{
				v1.NewWeightedVoteOption(v1.OptionYes, sdkmath.LegacyNewDec(1)),
			},
			voter:     sdk.AccAddress{},
			metadata:  "",
			expErr:    true,
			expErrMsg: "weighted vote options not supported",
		},
		"weights sum > 1": {
			preRun: func() uint64 {
				return proposalID
			},
			option: v1.WeightedVoteOptions{
				v1.NewWeightedVoteOption(v1.OptionYes, sdkmath.LegacyNewDec(1)),
				v1.NewWeightedVoteOption(v1.OptionAbstain, sdkmath.LegacyNewDec(1)),
			},
			voter:     proposer,
			metadata:  "",
			expErr:    true,
			expErrMsg: "weighted vote options not supported",
		},
		"duplicate vote options": {
			preRun: func() uint64 {
				return proposalID
			},
			option: v1.WeightedVoteOptions{
				v1.NewWeightedVoteOption(v1.OptionYes, sdkmath.LegacyNewDecWithPrec(5, 1)),
				v1.NewWeightedVoteOption(v1.OptionYes, sdkmath.LegacyNewDecWithPrec(5, 1)),
			},
			voter:     proposer,
			metadata:  "",
			expErr:    true,
			expErrMsg: "weighted vote options not supported",
		},
		"zero weight": {
			preRun: func() uint64 {
				return proposalID
			},
			option: v1.WeightedVoteOptions{
				v1.NewWeightedVoteOption(v1.OptionYes, sdkmath.LegacyNewDec(0)),
			},
			voter:     proposer,
			metadata:  "",
			expErr:    true,
			expErrMsg: "weighted vote options not supported",
		},
		"negative weight": {
			preRun: func() uint64 {
				return proposalID
			},
			option: v1.WeightedVoteOptions{
				v1.NewWeightedVoteOption(v1.OptionYes, sdkmath.LegacyNewDec(-1)),
			},
			voter:     proposer,
			metadata:  "",
			expErr:    true,
			expErrMsg: "weighted vote options not supported",
		},
		"empty options": {
			preRun: func() uint64 {
				return proposalID
			},
			option:    v1.WeightedVoteOptions{},
			voter:     proposer,
			metadata:  "",
			expErr:    true,
			expErrMsg: "weighted vote options not supported",
		},
		"invalid vote option": {
			preRun: func() uint64 {
				return proposalID
			},
			option:    v1.NewNonSplitVoteOption(v1.VoteOption(0x13)),
			voter:     proposer,
			metadata:  "",
			expErr:    true,
			expErrMsg: "weighted vote options not supported",
		},
		"weight sum < 1": {
			preRun: func() uint64 {
				return proposalID
			},
			option: v1.WeightedVoteOptions{ // weight sum <1
				v1.NewWeightedVoteOption(v1.OptionYes, sdkmath.LegacyNewDecWithPrec(5, 1)),
			},
			voter:     proposer,
			metadata:  "",
			expErr:    true,
			expErrMsg: "weighted vote options not supported",
		},
		"vote on inactive proposal": {
			preRun: func() uint64 {
				msg, err := v1.NewMsgSubmitProposal(
					tp,
					coins,
					proposer.String(),
					"",
					"Proposal",
					"description of proposal",
					false,
				)
				suite.Require().NoError(err)

				res, err := suite.msgSrvr.SubmitProposal(suite.ctx, msg)
				suite.Require().NoError(err)
				suite.Require().NotNil(res.ProposalId)
				return res.ProposalId
			},
			option:    v1.NewNonSplitVoteOption(v1.VoteOption_VOTE_OPTION_YES),
			voter:     proposer,
			metadata:  "",
			expErr:    true,
			expErrMsg: "weighted vote options not supported",
		},
		"metadata too long": {
			preRun: func() uint64 {
				return proposalID
			},
			option:    v1.NewNonSplitVoteOption(v1.VoteOption_VOTE_OPTION_YES),
			voter:     proposer,
			metadata:  strings.Repeat("a", 300),
			expErr:    true,
			expErrMsg: "weighted vote options not supported",
		},
		"voter error": {
			preRun: func() uint64 {
				return proposalID
			},
			option:    v1.NewNonSplitVoteOption(v1.VoteOption_VOTE_OPTION_YES),
			voter:     sdk.AccAddress(strings.Repeat("a", 300)),
			metadata:  "",
			expErr:    true,
			expErrMsg: "weighted vote options not supported",
		},
		"all good": {
			preRun: func() uint64 {
				msg, err := v1.NewMsgSubmitProposal(
					tp,
					minDeposit,
					proposer.String(),
					"",
					"Proposal",
					"description of proposal",
					false,
				)
				suite.Require().NoError(err)

				res, err := suite.msgSrvr.SubmitProposal(suite.ctx, msg)
				suite.Require().NoError(err)
				suite.Require().NotNil(res.ProposalId)
				return res.ProposalId
			},
			option:    v1.NewNonSplitVoteOption(v1.VoteOption_VOTE_OPTION_YES),
			voter:     proposer,
			metadata:  "",
			expErr:    true,
			expErrMsg: "weighted vote options not supported",
		},
		"all good with split votes": {
			preRun: func() uint64 {
				msg, err := v1.NewMsgSubmitProposal(
					tp,
					minDeposit,
					proposer.String(),
					"",
					"Proposal",
					"description of proposal",
					false,
				)
				suite.Require().NoError(err)

				res, err := suite.msgSrvr.SubmitProposal(suite.ctx, msg)
				suite.Require().NoError(err)
				suite.Require().NotNil(res.ProposalId)
				return res.ProposalId
			},
			option: v1.WeightedVoteOptions{
				v1.NewWeightedVoteOption(v1.OptionYes, sdkmath.LegacyNewDecWithPrec(5, 1)),
				v1.NewWeightedVoteOption(v1.OptionAbstain, sdkmath.LegacyNewDecWithPrec(5, 1)),
			},
			voter:     proposer,
			metadata:  "",
			expErr:    true,
			expErrMsg: "weighted vote options not supported",
		},
	}

	for name, tc := range cases {
		suite.Run(name, func() {
			pID := tc.preRun()
			voteReq := v1.NewMsgVoteWeighted(tc.voter, pID, tc.option, tc.metadata)
			_, err := suite.msgSrvr.VoteWeighted(suite.ctx, voteReq)
			if tc.expErr {
				suite.Require().Error(err)
				suite.Require().Contains(err.Error(), tc.expErrMsg)
			} else {
				suite.Require().NoError(err)
			}
		})
	}
}

func (suite *KeeperTestSuite) TestDepositReq() {
	addrs := suite.addrs
	proposer := addrs[0]

	coins := sdk.NewCoins(sdk.NewCoin("matic", sdkmath.NewInt(100000000000000000)))
	params, _ := suite.govKeeper.Params.Get(suite.ctx)
	minDeposit := sdk.Coins(params.MinDeposit)
	tp := TestProposal

	msg, err := v1.NewMsgSubmitProposal(
		tp,
		coins,
		proposer.String(),
		"",
		"Proposal",
		"description of proposal",
		false,
	)
	suite.Require().NoError(err)

	res, err := suite.msgSrvr.SubmitProposal(suite.ctx, msg)
	suite.Require().NoError(err)
	suite.Require().NotNil(res.ProposalId)
	pID := res.ProposalId

	cases := map[string]struct {
		preRun     func() uint64
		expErr     bool
		proposalID uint64
		depositor  sdk.AccAddress
		deposit    sdk.Coins
		expErrMsg  string
	}{
		"wrong proposal id": {
			preRun: func() uint64 {
				return 0
			},
			depositor: proposer,
			deposit:   coins,
			expErr:    true,
			expErrMsg: "not found",
		},
		"empty depositor": {
			preRun: func() uint64 {
				return pID
			},
			depositor: sdk.AccAddress{},
			deposit:   minDeposit,
			expErr:    true,
			expErrMsg: "invalid depositor address",
		},
		"invalid deposited coin ": {
			preRun: func() uint64 {
				return pID
			},
			depositor: proposer,
			deposit:   []sdk.Coin{sdk.NewCoin("ibc/badcoin", sdkmath.NewInt(1000))},
			expErr:    true,
			expErrMsg: "deposited 1000ibc/badcoin, but gov accepts only the following denom(s): [matic]",
		},
		"invalid deposited coin (multiple)": {
			preRun: func() uint64 {
				return pID
			},
			depositor: proposer,
			deposit:   minDeposit.Add(sdk.NewCoin("ibc/badcoin", sdkmath.NewInt(1000))),
			expErr:    true,
			expErrMsg: "deposited 1000ibc/badcoin,10000000000000000000matic, but gov accepts only the following denom(s): [matic]: invalid deposit denom",
		},
		"all good": {
			preRun: func() uint64 {
				return pID
			},
			depositor: proposer,
			deposit:   minDeposit,
			expErr:    false,
		},
	}

	for name, tc := range cases {
		suite.Run(name, func() {
			proposalID := tc.preRun()
			depositReq := v1.NewMsgDeposit(tc.depositor, proposalID, tc.deposit)
			_, err := suite.msgSrvr.Deposit(suite.ctx, depositReq)
			if tc.expErr {
				suite.Require().Error(err)
				suite.Require().Contains(err.Error(), tc.expErrMsg)
			} else {
				suite.Require().NoError(err)
			}
		})
	}
}

// legacy msg server tests
func (suite *KeeperTestSuite) TestLegacyMsgSubmitProposal() {
	accAmt := sdkmath.NewIntFromBigInt(new(big.Int).Mul(big.NewInt(10), new(big.Int).Exp(big.NewInt(10), big.NewInt(18), nil)))
	proposer := simtestutil.AddTestAddrsIncremental(suite.bankKeeper, suite.stakingKeeper, suite.ctx, 1, accAmt.Mul(sdkmath.NewInt(5)))[0]
	coins := sdk.NewCoins(sdk.NewCoin("matic", sdkmath.NewInt(100000000000000000)))
	initialDeposit := coins
	params, _ := suite.govKeeper.Params.Get(suite.ctx)
	minDeposit := params.MinDeposit

	cases := map[string]struct {
		preRun    func() (*v1beta1.MsgSubmitProposal, error)
		expErr    bool
		expErrMsg string
	}{
		"empty title": {
			preRun: func() (*v1beta1.MsgSubmitProposal, error) {
				content := v1beta1.NewTextProposal("", "I am test")
				return v1beta1.NewMsgSubmitProposal(
					content,
					initialDeposit,
					proposer,
				)
			},
			expErr:    true,
			expErrMsg: "proposal title cannot be blank",
		},
		"empty description": {
			preRun: func() (*v1beta1.MsgSubmitProposal, error) {
				content := v1beta1.NewTextProposal("test", "")
				return v1beta1.NewMsgSubmitProposal(
					content,
					initialDeposit,
					proposer,
				)
			},
			expErr:    true,
			expErrMsg: "proposal description cannot be blank",
		},
		"empty proposer": {
			preRun: func() (*v1beta1.MsgSubmitProposal, error) {
				content := v1beta1.NewTextProposal("test", "I am test")
				return v1beta1.NewMsgSubmitProposal(
					content,
					initialDeposit,
					sdk.AccAddress{},
				)
			},
			expErr:    true,
			expErrMsg: "invalid proposer address: empty address string is not allowed",
		},
		"title text length > max limit allowed": {
			preRun: func() (*v1beta1.MsgSubmitProposal, error) {
				content := v1beta1.NewTextProposal(strings.Repeat("#", v1beta1.MaxTitleLength*2), "I am test")
				return v1beta1.NewMsgSubmitProposal(
					content,
					initialDeposit,
					proposer,
				)
			},
			expErr:    true,
			expErrMsg: "proposal title is longer than max length of 140: invalid proposal content",
		},
		"description text length > max limit allowed": {
			preRun: func() (*v1beta1.MsgSubmitProposal, error) {
				content := v1beta1.NewTextProposal("test", strings.Repeat("#", v1beta1.MaxDescriptionLength*2))
				return v1beta1.NewMsgSubmitProposal(
					content,
					initialDeposit,
					proposer,
				)
			},
			expErr:    true,
			expErrMsg: "proposal description is longer than max length of 10000: invalid proposal content",
		},
		"all good": {
			preRun: func() (*v1beta1.MsgSubmitProposal, error) {
				return v1beta1.NewMsgSubmitProposal(
					v1beta1.NewTextProposal("test", "I am test"),
					initialDeposit,
					proposer,
				)
			},
			expErr: false,
		},
		"all good with min deposit": {
			preRun: func() (*v1beta1.MsgSubmitProposal, error) {
				return v1beta1.NewMsgSubmitProposal(
					v1beta1.NewTextProposal("test", "I am test"),
					minDeposit,
					proposer,
				)
			},
			expErr: false,
		},
	}

	for name, c := range cases {
		suite.Run(name, func() {
			msg, err := c.preRun()
			suite.Require().NoError(err)
			res, err := suite.legacyMsgSrvr.SubmitProposal(suite.ctx, msg)
			if c.expErr {
				suite.Require().Error(err)
				suite.Require().Contains(err.Error(), c.expErrMsg)
			} else {
				suite.Require().NoError(err)
				suite.Require().NotNil(res.ProposalId)
			}
		})
	}
}

func (suite *KeeperTestSuite) TestLegacyMsgVote() {
	addrs := suite.addrs
	proposer := addrs[0]

	coins := sdk.NewCoins(sdk.NewCoin("matic", sdkmath.NewInt(100000000000000000)))
	params, _ := suite.govKeeper.Params.Get(suite.ctx)
	minDeposit := params.MinDeposit
	tp := TestProposal

	msg, err := v1.NewMsgSubmitProposal(
		tp,
		minDeposit,
		proposer.String(),
		"",
		"Proposal",
		"description of proposal",
		false,
	)
	suite.Require().NoError(err)

	res, err := suite.msgSrvr.SubmitProposal(suite.ctx, msg)
	suite.Require().NoError(err)
	suite.Require().NotNil(res.ProposalId)
	proposalID := res.ProposalId

	cases := map[string]struct {
		preRun    func() uint64
		expErr    bool
		expErrMsg string
		option    v1beta1.VoteOption
		metadata  string
		voter     sdk.AccAddress
	}{
		"empty voter": {
			preRun: func() uint64 {
				return proposalID
			},
			option:    v1beta1.OptionYes,
			voter:     sdk.AccAddress{},
			metadata:  "",
			expErr:    true,
			expErrMsg: "invalid address",
		},
		"wrong vote option": {
			preRun: func() uint64 {
				return proposalID
			},
			option:    v1beta1.VoteOption(0x13),
			voter:     proposer,
			metadata:  "",
			expErr:    true,
			expErrMsg: "invalid vote option",
		},
		"vote on inactive proposal": {
			preRun: func() uint64 {
				msg, err := v1.NewMsgSubmitProposal(
					tp,
					coins,
					proposer.String(),
					"",
					"Proposal",
					"description of proposal",
					false,
				)
				suite.Require().NoError(err)

				res, err := suite.msgSrvr.SubmitProposal(suite.ctx, msg)
				suite.Require().NoError(err)
				suite.Require().NotNil(res.ProposalId)
				return res.ProposalId
			},
			option:    v1beta1.OptionYes,
			voter:     proposer,
			metadata:  "",
			expErr:    true,
			expErrMsg: "inactive proposal",
		},
		"voter error": {
			preRun: func() uint64 {
				return proposalID
			},
			option:    v1beta1.OptionYes,
			voter:     sdk.AccAddress(strings.Repeat("a", 300)),
			metadata:  "",
			expErr:    true,
			expErrMsg: "invalid address",
		},
		"all good": {
			preRun: func() uint64 {
				msg, err := v1.NewMsgSubmitProposal(
					tp,
					minDeposit,
					proposer.String(),
					"",
					"Proposal",
					"description of proposal",
					false,
				)
				suite.Require().NoError(err)

				res, err := suite.msgSrvr.SubmitProposal(suite.ctx, msg)
				suite.Require().NoError(err)
				suite.Require().NotNil(res.ProposalId)
				return res.ProposalId
			},
			option:   v1beta1.OptionYes,
			voter:    proposer,
			metadata: "",
			expErr:   false,
		},
	}

	for name, tc := range cases {
		suite.Run(name, func() {
			pID := tc.preRun()
			voteReq := v1beta1.NewMsgVote(tc.voter, pID, tc.option)
			_, err := suite.legacyMsgSrvr.Vote(suite.ctx, voteReq)
			if tc.expErr {
				suite.Require().Error(err)
				suite.Require().Contains(err.Error(), tc.expErrMsg)
			} else {
				suite.Require().NoError(err)
			}
		})
	}
}

func (suite *KeeperTestSuite) TestLegacyVoteWeighted() {
	suite.reset()
	addrs := suite.addrs
	proposer := addrs[0]

	coins := sdk.NewCoins(sdk.NewCoin("matic", sdkmath.NewInt(100000000000000000)))
	params, _ := suite.govKeeper.Params.Get(suite.ctx)
	minDeposit := params.MinDeposit
	tp := TestProposal

	msg, err := v1.NewMsgSubmitProposal(
		tp,
		minDeposit,
		proposer.String(),
		"",
		"Proposal",
		"description of proposal",
		false,
	)
	suite.Require().NoError(err)

	res, err := suite.msgSrvr.SubmitProposal(suite.ctx, msg)
	suite.Require().NoError(err)
	suite.Require().NotNil(res.ProposalId)
	proposalID := res.ProposalId

	cases := map[string]struct {
		preRun    func() uint64
		vote      *v1beta1.MsgVote
		expErr    bool
		expErrMsg string
		option    v1beta1.WeightedVoteOptions
		metadata  string
		voter     sdk.AccAddress
	}{
		"empty voter": {
			preRun: func() uint64 {
				return proposalID
			},
			option: v1beta1.WeightedVoteOptions{
				v1beta1.WeightedVoteOption{
					Option: v1beta1.OptionYes,
					Weight: sdkmath.LegacyNewDec(1),
				},
			},
			voter:     sdk.AccAddress{},
			metadata:  "",
			expErr:    true,
			expErrMsg: "weighted vote options not supported",
		},
		"weights sum > 1": {
			preRun: func() uint64 {
				return proposalID
			},
			option: v1beta1.WeightedVoteOptions{
				v1beta1.WeightedVoteOption{
					Option: v1beta1.OptionYes,
					Weight: sdkmath.LegacyNewDec(1),
				},
				v1beta1.WeightedVoteOption{
					Option: v1beta1.OptionAbstain,
					Weight: sdkmath.LegacyNewDec(1),
				},
			},
			voter:     proposer,
			metadata:  "",
			expErr:    true,
			expErrMsg: "weighted vote options not supported",
		},
		"duplicate vote options": {
			preRun: func() uint64 {
				return proposalID
			},
			option: v1beta1.WeightedVoteOptions{
				v1beta1.WeightedVoteOption{
					Option: v1beta1.OptionYes,
					Weight: sdkmath.LegacyNewDecWithPrec(5, 1),
				},
				v1beta1.WeightedVoteOption{
					Option: v1beta1.OptionYes,
					Weight: sdkmath.LegacyNewDecWithPrec(5, 1),
				},
			},
			voter:     proposer,
			metadata:  "",
			expErr:    true,
			expErrMsg: "weighted vote options not supported",
		},
		"zero weight": {
			preRun: func() uint64 {
				return proposalID
			},
			option: v1beta1.WeightedVoteOptions{
				v1beta1.WeightedVoteOption{
					Option: v1beta1.OptionYes,
					Weight: sdkmath.LegacyNewDec(0),
				},
			},
			voter:     proposer,
			metadata:  "",
			expErr:    true,
			expErrMsg: "weighted vote options not supported",
		},
		"negative weight": {
			preRun: func() uint64 {
				return proposalID
			},
			option: v1beta1.WeightedVoteOptions{
				v1beta1.WeightedVoteOption{
					Option: v1beta1.OptionYes,
					Weight: sdkmath.LegacyNewDec(-1),
				},
			},
			voter:     proposer,
			metadata:  "",
			expErr:    true,
			expErrMsg: "weighted vote options not supported",
		},
		"empty options": {
			preRun: func() uint64 {
				return proposalID
			},
			option:    v1beta1.WeightedVoteOptions{},
			voter:     proposer,
			metadata:  "",
			expErr:    true,
			expErrMsg: "weighted vote options not supported",
		},
		"invalid vote option": {
			preRun: func() uint64 {
				return proposalID
			},
			option: v1beta1.WeightedVoteOptions{
				v1beta1.WeightedVoteOption{
					Option: v1beta1.VoteOption(0x13),
					Weight: sdkmath.LegacyNewDecWithPrec(5, 1),
				},
			},
			voter:     proposer,
			metadata:  "",
			expErr:    true,
			expErrMsg: "weighted vote options not supported",
		},
		"weight sum < 1": {
			preRun: func() uint64 {
				return proposalID
			},
			option: v1beta1.WeightedVoteOptions{
				v1beta1.WeightedVoteOption{
					Option: v1beta1.OptionYes,
					Weight: sdkmath.LegacyNewDecWithPrec(5, 1),
				},
			},
			voter:     proposer,
			metadata:  "",
			expErr:    true,
			expErrMsg: "weighted vote options not supported",
		},
		"vote on inactive proposal": {
			preRun: func() uint64 {
				msg, err := v1.NewMsgSubmitProposal(
					tp,
					coins,
					proposer.String(),
					"",
					"Proposal",
					"description of proposal",
					false,
				)
				suite.Require().NoError(err)

				res, err := suite.msgSrvr.SubmitProposal(suite.ctx, msg)
				suite.Require().NoError(err)
				suite.Require().NotNil(res.ProposalId)
				return res.ProposalId
			},
			option: v1beta1.WeightedVoteOptions{
				v1beta1.WeightedVoteOption{
					Option: v1beta1.OptionYes,
					Weight: sdkmath.LegacyNewDec(1),
				},
			},
			voter:     proposer,
			metadata:  "",
			expErr:    true,
			expErrMsg: "weighted vote options not supported",
		},
		"voter error": {
			preRun: func() uint64 {
				return proposalID
			},
			option: v1beta1.WeightedVoteOptions{
				v1beta1.WeightedVoteOption{
					Option: v1beta1.OptionYes,
					Weight: sdkmath.LegacyNewDec(1),
				},
			},
			voter:     sdk.AccAddress(strings.Repeat("a", 300)),
			metadata:  "",
			expErr:    true,
			expErrMsg: "weighted vote options not supported",
		},
		"all good": {
			preRun: func() uint64 {
				msg, err := v1.NewMsgSubmitProposal(
					tp,
					minDeposit,
					proposer.String(),
					"",
					"Proposal",
					"description of proposal",
					false,
				)
				suite.Require().NoError(err)

				res, err := suite.msgSrvr.SubmitProposal(suite.ctx, msg)
				suite.Require().NoError(err)
				suite.Require().NotNil(res.ProposalId)
				return res.ProposalId
			},
			option: v1beta1.WeightedVoteOptions{
				v1beta1.WeightedVoteOption{
					Option: v1beta1.OptionYes,
					Weight: sdkmath.LegacyNewDec(1),
				},
			},
			voter:     proposer,
			metadata:  "",
			expErr:    true,
			expErrMsg: "weighted vote options not supported",
		},
	}

	for name, tc := range cases {
		suite.Run(name, func() {
			pID := tc.preRun()
			voteReq := v1beta1.NewMsgVoteWeighted(tc.voter, pID, tc.option)
			_, err := suite.legacyMsgSrvr.VoteWeighted(suite.ctx, voteReq)
			if tc.expErr {
				suite.Require().Error(err)
				suite.Require().Contains(err.Error(), tc.expErrMsg)
			} else {
				suite.Require().NoError(err)
			}
		})
	}
}

func (suite *KeeperTestSuite) TestLegacyMsgDeposit() {
	addrs := suite.addrs
	proposer := addrs[0]

	coins := sdk.NewCoins(sdk.NewCoin("matic", sdkmath.NewInt(100000000000000000)))
	params, _ := suite.govKeeper.Params.Get(suite.ctx)
	minDeposit := params.MinDeposit
	tp := TestProposal

	msg, err := v1.NewMsgSubmitProposal(
		tp,
		coins,
		proposer.String(),
		"",
		"Proposal",
		"description of proposal",
		false,
	)
	suite.Require().NoError(err)

	res, err := suite.msgSrvr.SubmitProposal(suite.ctx, msg)
	suite.Require().NoError(err)
	suite.Require().NotNil(res.ProposalId)
	pID := res.ProposalId

	cases := map[string]struct {
		preRun     func() uint64
		expErr     bool
		expErrMsg  string
		proposalID uint64
		depositor  sdk.AccAddress
		deposit    sdk.Coins
	}{
		"wrong proposal id": {
			preRun: func() uint64 {
				return 0
			},
			depositor: proposer,
			deposit:   coins,
			expErr:    true,
			expErrMsg: "not found",
		},
		"empty depositer": {
			preRun: func() uint64 {
				return pID
			},
			depositor: sdk.AccAddress{},
			deposit:   coins,
			expErr:    true,
			expErrMsg: "invalid depositor address: empty address string is not allowed",
		},
		"all good": {
			preRun: func() uint64 {
				return pID
			},
			depositor: proposer,
			deposit:   minDeposit,
			expErr:    false,
		},
	}

	for name, tc := range cases {
		suite.Run(name, func() {
			proposalID := tc.preRun()
			depositReq := v1beta1.NewMsgDeposit(tc.depositor, proposalID, tc.deposit)
			_, err := suite.legacyMsgSrvr.Deposit(suite.ctx, depositReq)
			if tc.expErr {
				suite.Require().Error(err)
				suite.Require().Contains(err.Error(), tc.expErrMsg)
			} else {
				suite.Require().NoError(err)
			}
		})
	}
}

func (suite *KeeperTestSuite) TestMsgUpdateParams() {
	authority := suite.govKeeper.GetAuthority()
	params := v1.DefaultParams()
	testCases := []struct {
		name      string
		input     func() *v1.MsgUpdateParams
		expErr    bool
		expErrMsg string
	}{
		{
			name: "valid",
			input: func() *v1.MsgUpdateParams {
				return &v1.MsgUpdateParams{
					Authority: authority,
					Params:    params,
				}
			},
			expErr: false,
		},
		{
			name: "invalid authority",
			input: func() *v1.MsgUpdateParams {
				return &v1.MsgUpdateParams{
					Authority: "authority",
					Params:    params,
				}
			},
			expErr:    true,
			expErrMsg: "invalid authority",
		},
		{
			name: "invalid min deposit",
			input: func() *v1.MsgUpdateParams {
				params1 := params
				params1.MinDeposit = nil

				return &v1.MsgUpdateParams{
					Authority: authority,
					Params:    params1,
				}
			},
			expErr:    true,
			expErrMsg: "invalid minimum deposit",
		},
		{
			name: "negative deposit",
			input: func() *v1.MsgUpdateParams {
				params1 := params
				params1.MinDeposit = sdk.Coins{{
					Denom:  sdk.DefaultBondDenom,
					Amount: sdkmath.NewInt(-100),
				}}

				return &v1.MsgUpdateParams{
					Authority: authority,
					Params:    params1,
				}
			},
			expErr:    true,
			expErrMsg: "invalid minimum deposit",
		},
		{
			name: "invalid max deposit period",
			input: func() *v1.MsgUpdateParams {
				params1 := params
				params1.MaxDepositPeriod = nil

				return &v1.MsgUpdateParams{
					Authority: authority,
					Params:    params1,
				}
			},
			expErr:    true,
			expErrMsg: "maximum deposit period must not be nil",
		},
		{
			name: "zero max deposit period",
			input: func() *v1.MsgUpdateParams {
				params1 := params
				duration := time.Duration(0)
				params1.MaxDepositPeriod = &duration

				return &v1.MsgUpdateParams{
					Authority: authority,
					Params:    params1,
				}
			},
			expErr:    true,
			expErrMsg: "maximum deposit period must be positive",
		},
		{
			name: "invalid quorum",
			input: func() *v1.MsgUpdateParams {
				params1 := params
				params1.Quorum = abc

				return &v1.MsgUpdateParams{
					Authority: authority,
					Params:    params1,
				}
			},
			expErr:    true,
			expErrMsg: "invalid quorum string",
		},
		{
			name: "negative quorum",
			input: func() *v1.MsgUpdateParams {
				params1 := params
				params1.Quorum = o1

				return &v1.MsgUpdateParams{
					Authority: authority,
					Params:    params1,
				}
			},
			expErr:    true,
			expErrMsg: "quorom cannot be negative",
		},
		{
			name: "quorum > 1",
			input: func() *v1.MsgUpdateParams {
				params1 := params
				params1.Quorum = "2"

				return &v1.MsgUpdateParams{
					Authority: authority,
					Params:    params1,
				}
			},
			expErr:    true,
			expErrMsg: "quorom too large",
		},
		{
			name: "invalid threshold",
			input: func() *v1.MsgUpdateParams {
				params1 := params
				params1.Threshold = abc

				return &v1.MsgUpdateParams{
					Authority: authority,
					Params:    params1,
				}
			},
			expErr:    true,
			expErrMsg: "invalid threshold string",
		},
		{
			name: "negative threshold",
			input: func() *v1.MsgUpdateParams {
				params1 := params
				params1.Threshold = o1

				return &v1.MsgUpdateParams{
					Authority: authority,
					Params:    params1,
				}
			},
			expErr:    true,
			expErrMsg: "vote threshold must be positive",
		},
		{
			name: "threshold > 1",
			input: func() *v1.MsgUpdateParams {
				params1 := params
				params1.Threshold = "2"

				return &v1.MsgUpdateParams{
					Authority: authority,
					Params:    params1,
				}
			},
			expErr:    true,
			expErrMsg: "vote threshold too large",
		},
		{
			name: "invalid veto threshold",
			input: func() *v1.MsgUpdateParams {
				params1 := params
				params1.VetoThreshold = abc

				return &v1.MsgUpdateParams{
					Authority: authority,
					Params:    params1,
				}
			},
			expErr:    true,
			expErrMsg: "invalid vetoThreshold string",
		},
		{
			name: "negative veto threshold",
			input: func() *v1.MsgUpdateParams {
				params1 := params
				params1.VetoThreshold = o1

				return &v1.MsgUpdateParams{
					Authority: authority,
					Params:    params1,
				}
			},
			expErr:    true,
			expErrMsg: "veto threshold must be positive",
		},
		{
			name: "veto threshold > 1",
			input: func() *v1.MsgUpdateParams {
				params1 := params
				params1.VetoThreshold = "2"

				return &v1.MsgUpdateParams{
					Authority: authority,
					Params:    params1,
				}
			},
			expErr:    true,
			expErrMsg: "veto threshold too large",
		},
		{
			name: "invalid voting period",
			input: func() *v1.MsgUpdateParams {
				params1 := params
				params1.VotingPeriod = nil

				return &v1.MsgUpdateParams{
					Authority: authority,
					Params:    params1,
				}
			},
			expErr:    true,
			expErrMsg: "voting period must not be nil",
		},
		{
			name: "zero voting period",
			input: func() *v1.MsgUpdateParams {
				params1 := params
				duration := time.Duration(0)
				params1.VotingPeriod = &duration

				return &v1.MsgUpdateParams{
					Authority: authority,
					Params:    params1,
				}
			},
			expErr:    true,
			expErrMsg: "voting period must be positive",
		},
	}

	for _, tc := range testCases {
		tc := tc
		suite.Run(tc.name, func() {
			msg := tc.input()
			exec := func(updateParams *v1.MsgUpdateParams) error {
				if _, err := suite.msgSrvr.UpdateParams(suite.ctx, updateParams); err != nil {
					return err
				}
				return nil
			}

			err := exec(msg)
			if tc.expErr {
				suite.Require().Error(err)
				suite.Require().Contains(err.Error(), tc.expErrMsg)
			} else {
				suite.Require().NoError(err)
			}
		})
	}
}

func (suite *KeeperTestSuite) TestSubmitProposal_InitialDeposit() {
	const meetsDepositValue = baseDepositTestAmount * baseDepositTestPercent / 100
	baseDepositRatioDec := sdkmath.LegacyNewDec(baseDepositTestPercent).Quo(sdkmath.LegacyNewDec(100))

	testcases := map[string]struct {
		minDeposit             sdk.Coins
		minInitialDepositRatio sdkmath.LegacyDec
		initialDeposit         sdk.Coins
		accountBalance         sdk.Coins

		expectError bool
	}{
		"meets initial deposit, enough balance - success": {
			minDeposit:             sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdkmath.NewInt(baseDepositTestAmount))),
			minInitialDepositRatio: baseDepositRatioDec,
			initialDeposit:         sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdkmath.NewInt(meetsDepositValue))),
			accountBalance:         sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdkmath.NewInt(meetsDepositValue))),
		},
		"does not meet initial deposit, enough balance - error": {
			minDeposit:             sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdkmath.NewInt(baseDepositTestAmount))),
			minInitialDepositRatio: baseDepositRatioDec,
			initialDeposit:         sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdkmath.NewInt(meetsDepositValue-1))),
			accountBalance:         sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdkmath.NewInt(meetsDepositValue))),

			expectError: true,
		},
		"meets initial deposit, not enough balance - error": {
			minDeposit:             sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdkmath.NewInt(baseDepositTestAmount))),
			minInitialDepositRatio: baseDepositRatioDec,
			initialDeposit:         sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdkmath.NewInt(meetsDepositValue))),
			accountBalance:         sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdkmath.NewInt(meetsDepositValue-1))),

			expectError: true,
		},
		"does not meet initial deposit and not enough balance - error": {
			minDeposit:             sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdkmath.NewInt(baseDepositTestAmount))),
			minInitialDepositRatio: baseDepositRatioDec,
			initialDeposit:         sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdkmath.NewInt(meetsDepositValue-1))),
			accountBalance:         sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdkmath.NewInt(meetsDepositValue-1))),

			expectError: true,
		},
	}

	for name, tc := range testcases {
		suite.Run(name, func() {
			// Setup
			govKeeper, ctx := suite.govKeeper, suite.ctx
			address := simtestutil.AddTestAddrs(suite.bankKeeper, suite.stakingKeeper, ctx, 1, tc.accountBalance[0].Amount)[0]

			params := v1.DefaultParams()
			params.MinDeposit = tc.minDeposit
			params.MinInitialDepositRatio = tc.minInitialDepositRatio.String()
			govKeeper.Params.Set(ctx, params)

			msg, err := v1.NewMsgSubmitProposal(TestProposal, tc.initialDeposit, address.String(), "test", "Proposal", "description of proposal", false)
			suite.Require().NoError(err)

			// System under test
			_, err = suite.msgSrvr.SubmitProposal(ctx, msg)

			// Assertions
			if tc.expectError {
				suite.Require().Error(err)
				return
			}
			suite.Require().NoError(err)
		})
	}
}
