package ante

import (
	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
)

// MsgMultiSendCapDecorator rejects MsgMultiSend with more than maxOutputs.
// activeFn, when non-nil, gates enforcement by block height; nil means always-on.
type MsgMultiSendCapDecorator struct {
	maxOutputs int
	activeFn   func(int64) bool
}

func NewMsgMultiSendCapDecorator(maxOutputs int, activeFn func(int64) bool) MsgMultiSendCapDecorator {
	return MsgMultiSendCapDecorator{maxOutputs: maxOutputs, activeFn: activeFn}
}

func (d MsgMultiSendCapDecorator) AnteHandle(ctx sdk.Context, tx sdk.Tx, simulate bool, next sdk.AnteHandler) (sdk.Context, error) {
	if d.activeFn != nil && !d.activeFn(ctx.BlockHeight()) {
		return next(ctx, tx, simulate)
	}
	for _, msg := range tx.GetMsgs() {
		ms, ok := msg.(*banktypes.MsgMultiSend)
		if !ok {
			continue
		}
		if got := len(ms.Outputs); got > d.maxOutputs {
			return ctx, errorsmod.Wrapf(sdkerrors.ErrInvalidRequest,
				"MsgMultiSend has %d outputs, max allowed is %d", got, d.maxOutputs)
		}
	}
	return next(ctx, tx, simulate)
}
