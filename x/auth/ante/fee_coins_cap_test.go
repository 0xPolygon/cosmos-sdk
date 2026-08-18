package ante_test

import (
	"fmt"
	"testing"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/auth/ante"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

// feeCapTestTx is a minimal sdk.FeeTx exposing only what FeeCoinsCapDecorator reads.
// Keep separate from capTestTx: it must NOT implement sdk.FeeTx, since
// TestFeeCoinsCapDecoratorWrongTxType relies on that type-assertion failing.
type feeCapTestTx struct {
	fee sdk.Coins
}

func (t *feeCapTestTx) GetMsgs() []sdk.Msg                  { return nil }
func (t *feeCapTestTx) GetMsgsV2() ([]proto.Message, error) { return nil, nil }
func (t *feeCapTestTx) GetGas() uint64                      { return 0 }
func (t *feeCapTestTx) GetFee() sdk.Coins                   { return t.fee }
func (t *feeCapTestTx) FeePayer() []byte                    { return nil }
func (t *feeCapTestTx) FeeGranter() []byte                  { return nil }

// manyDistinctValidCoins builds n coins that each individually pass
// Coins.Validate() (positive amount, valid denom, sorted, no duplicates),
// unlike manyInvalidCoins - exercises the count cap independently of Validate.
func manyDistinctValidCoins(n int) sdk.Coins {
	coins := make(sdk.Coins, n)
	for i := range coins {
		coins[i] = sdk.Coin{Denom: fmt.Sprintf("aaa%06d", i), Amount: math.OneInt()}
	}
	return coins
}

func manyInvalidCoins(n int) sdk.Coins {
	coins := make(sdk.Coins, n)
	for i := range coins {
		coins[i] = sdk.Coin{Denom: "", Amount: math.ZeroInt()}
	}
	return coins
}

func TestFeeCoinsCapDecorator(t *testing.T) {
	const cap = 1

	alwaysActive := func(int64) bool { return true }
	alwaysInactive := func(int64) bool { return false }
	activateAt100 := func(h int64) bool { return h >= 100 }

	tests := []struct {
		name       string
		decorator  ante.FeeCoinsCapDecorator
		height     int64
		fee        sdk.Coins
		recheckTx  bool
		wantReject bool
		wantErrIs  error
	}{
		{
			name:      "empty fee accepts",
			decorator: ante.NewFeeCoinsCapDecorator(cap, alwaysActive),
			fee:       sdk.Coins{},
		},
		{
			name:      "single valid coin accepts",
			decorator: ante.NewFeeCoinsCapDecorator(cap, alwaysActive),
			fee:       sdk.NewCoins(sdk.NewInt64Coin("pol", 1)),
		},
		{
			name:       "two coins rejects (over cap)",
			decorator:  ante.NewFeeCoinsCapDecorator(cap, alwaysActive),
			fee:        sdk.NewCoins(sdk.NewInt64Coin("aaa", 1), sdk.NewInt64Coin("bbb", 1)),
			wantReject: true,
			wantErrIs:  sdkerrors.ErrInsufficientFee,
		},
		{
			name:       "many malformed coins (zero amount, empty denom) rejects",
			decorator:  ante.NewFeeCoinsCapDecorator(cap, alwaysActive),
			fee:        manyInvalidCoins(100000),
			wantReject: true,
			wantErrIs:  sdkerrors.ErrInsufficientFee,
		},
		{
			name:       "many individually-valid distinct-denom coins still rejects",
			decorator:  ante.NewFeeCoinsCapDecorator(cap, alwaysActive),
			fee:        manyDistinctValidCoins(10000),
			wantReject: true,
			wantErrIs:  sdkerrors.ErrInsufficientFee,
		},
		{
			name:       "single coin with zero amount rejects (caught by Validate, not count)",
			decorator:  ante.NewFeeCoinsCapDecorator(cap, alwaysActive),
			fee:        sdk.Coins{sdk.Coin{Denom: "pol", Amount: math.ZeroInt()}},
			wantReject: true,
			wantErrIs:  sdkerrors.ErrInsufficientFee,
		},
		{
			name:       "single coin with empty denom rejects",
			decorator:  ante.NewFeeCoinsCapDecorator(cap, alwaysActive),
			fee:        sdk.Coins{sdk.Coin{Denom: "", Amount: math.OneInt()}},
			wantReject: true,
			wantErrIs:  sdkerrors.ErrInsufficientFee,
		},
		{
			name:      "inactive (activeFn=false) skips enforcement even over cap",
			decorator: ante.NewFeeCoinsCapDecorator(cap, alwaysInactive),
			fee:       manyInvalidCoins(1000),
		},
		{
			name:       "activeFn nil acts as always-on, rejects over cap",
			decorator:  ante.NewFeeCoinsCapDecorator(cap, nil),
			fee:        manyInvalidCoins(1000),
			wantReject: true,
		},
		{
			name:      "below activation height: over cap accepts",
			decorator: ante.NewFeeCoinsCapDecorator(cap, activateAt100),
			height:    99,
			fee:       manyInvalidCoins(1000),
		},
		{
			name:       "at activation height: over cap rejects",
			decorator:  ante.NewFeeCoinsCapDecorator(cap, activateAt100),
			height:     100,
			fee:        manyInvalidCoins(1000),
			wantReject: true,
		},
		{
			name:      "ReCheckTx skips enforcement even over cap and past activation",
			decorator: ante.NewFeeCoinsCapDecorator(cap, activateAt100),
			height:    100,
			fee:       manyInvalidCoins(1000),
			recheckTx: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := sdk.Context{}.WithBlockHeight(tt.height).WithIsReCheckTx(tt.recheckTx)
			tx := &feeCapTestTx{fee: tt.fee}
			_, err := tt.decorator.AnteHandle(ctx, tx, false, terminalHandler)
			if tt.wantReject {
				require.Error(t, err)
				if tt.wantErrIs != nil {
					require.ErrorIs(t, err, tt.wantErrIs)
				}
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestFeeCoinsCapDecoratorWrongTxType(t *testing.T) {
	decorator := ante.NewFeeCoinsCapDecorator(1, nil)
	ctx := sdk.Context{}
	_, err := decorator.AnteHandle(ctx, &capTestTx{}, false, terminalHandler)
	require.Error(t, err)
	require.ErrorIs(t, err, sdkerrors.ErrTxDecode)
}
