package ante

import (
	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// FeeCoinsCapDecorator bounds the number of coins a tx may declare in its
// fee. activeFn gates enforcement by block height; nil means always-on.
type FeeCoinsCapDecorator struct {
	maxFeeCoins int
	activeFn    func(int64) bool
}

func NewFeeCoinsCapDecorator(maxFeeCoins int, activeFn func(int64) bool) FeeCoinsCapDecorator {
	return FeeCoinsCapDecorator{maxFeeCoins: maxFeeCoins, activeFn: activeFn}
}

func (d FeeCoinsCapDecorator) AnteHandle(ctx sdk.Context, tx sdk.Tx, simulate bool, next sdk.AnteHandler) (sdk.Context, error) {
	// ReCheckTx never decides block inclusion, so skipping here is safe -
	// the real check runs again when the tx is actually included.
	if ctx.IsReCheckTx() {
		return next(ctx, tx, simulate)
	}

	if d.activeFn != nil && !d.activeFn(ctx.BlockHeight()) {
		return next(ctx, tx, simulate)
	}

	feeTx, ok := tx.(sdk.FeeTx)
	if !ok {
		return ctx, errorsmod.Wrap(sdkerrors.ErrTxDecode, "invalid tx type")
	}

	fee := feeTx.GetFee()
	if len(fee) > d.maxFeeCoins {
		return ctx, errorsmod.Wrapf(sdkerrors.ErrInsufficientFee,
			"tx declares %d fee coins, max allowed is %d", len(fee), d.maxFeeCoins)
	}

	// Bound the count before validating - Validate is an O(n) scan.
	if err := fee.Validate(); err != nil {
		return ctx, errorsmod.Wrapf(sdkerrors.ErrInsufficientFee, "invalid fee: %s", err)
	}

	return next(ctx, tx, simulate)
}
