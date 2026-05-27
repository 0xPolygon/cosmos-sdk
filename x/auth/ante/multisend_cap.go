package ante

import (
	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
)

// MsgMultiSendCapDecorator caps the aggregate bank-transfer output count per
// tx (MsgMultiSend contributes len(Outputs), MsgSend contributes 1). activeFn
// gates enforcement by block height; nil means always-on. Only inspects
// top-level tx messages — wrapper executors like authz.MsgExec dispatch inner
// msgs at handler time and bypass this cap; consumers enabling such wrappers
// must constrain or recursively check.
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
	totalOutputs := 0
	for _, msg := range tx.GetMsgs() {
		switch m := msg.(type) {
		case *banktypes.MsgMultiSend:
			totalOutputs += len(m.Outputs)
		case *banktypes.MsgSend:
			totalOutputs++
		default:
			continue
		}
		if totalOutputs > d.maxOutputs {
			return ctx, errorsmod.Wrapf(sdkerrors.ErrInvalidRequest,
				"tx has %d aggregate bank-transfer outputs, max allowed is %d", totalOutputs, d.maxOutputs)
		}
	}
	return next(ctx, tx, simulate)
}
