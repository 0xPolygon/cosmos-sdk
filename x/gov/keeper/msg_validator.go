package keeper

import (
	"fmt"
	"github.com/cosmos/cosmos-sdk/x/gov/types"

	authv1beta1 "cosmossdk.io/api/cosmos/auth/v1beta1"
	bankv1beta1 "cosmossdk.io/api/cosmos/bank/v1beta1"
	consensusv1 "cosmossdk.io/api/cosmos/consensus/v1"
	govv1 "cosmossdk.io/api/cosmos/gov/v1"
	stakingv1beta1 "cosmossdk.io/api/cosmos/staking/v1beta1"
	errorsmod "cosmossdk.io/errors"

	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	consensustypes "github.com/cosmos/cosmos-sdk/x/consensus/types"
	v1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

// ValidateGovMsgType validates the type of the message submitted within a proposal.
// It only accepts `MsgExecLegacyContent` and `MsgUpdateParams` for the heimdall-v2 enabled modules.
func ValidateGovMsgType(msg sdk.Msg) error {
	switch msg.(type) {
	// TODO HV2: list to be eventually extended
	case *v1.MsgExecLegacyContent,
		*v1.MsgUpdateParams, *govv1.MsgUpdateParams,
		*authtypes.MsgUpdateParams, *authv1beta1.MsgUpdateParams,
		*banktypes.MsgUpdateParams, *bankv1beta1.MsgUpdateParams,
		*consensustypes.MsgUpdateParams, *consensusv1.MsgUpdateParams,
		*stakingtypes.MsgUpdateParams, *stakingv1beta1.MsgUpdateParams:
		return nil
	default:
		return errorsmod.Wrap(types.ErrInvalidProposalMsgType, fmt.Sprintf("type not supported: %T", msg))
	}
}

// ValidateGovMsgContentType validates the type of the msg content submitted within a proposal.
// It only accepts `TextProposal` and `ParamChange` for the heimdall-v2 enabled modules.
func ValidateGovMsgContentType(msg *v1.MsgExecLegacyContent) error {
	switch msg.Content.TypeUrl {
	// TODO HV2: list to be eventually extended
	case "/cosmos.gov.v1beta1.TextProposal",
		"/cosmos.params.v1beta1.ParameterChangeProposal", "/cosmos.params.v1beta1.ParamChange":
		return nil
	default:
		return errorsmod.Wrap(types.ErrInvalidProposalContentType, fmt.Sprintf("type not supported: %T", msg.Content))
	}
}
