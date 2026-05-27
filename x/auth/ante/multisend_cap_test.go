package ante_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/auth/ante"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

// capTestTx is a minimal sdk.Tx exposing GetMsgs / GetMsgsV2.
type capTestTx struct {
	msgs []sdk.Msg
}

func (t *capTestTx) GetMsgs() []sdk.Msg { return t.msgs }

func (t *capTestTx) GetMsgsV2() ([]proto.Message, error) {
	out := make([]proto.Message, 0, len(t.msgs))
	for _, m := range t.msgs {
		if pm, ok := m.(proto.Message); ok {
			out = append(out, pm)
		}
	}
	return out, nil
}

func newMultiSend(outputs int) *banktypes.MsgMultiSend {
	outs := make([]banktypes.Output, outputs)
	for i := range outs {
		outs[i] = banktypes.Output{Address: "cosmos1xyz", Coins: sdk.NewCoins(sdk.NewInt64Coin("stake", 1))}
	}
	return &banktypes.MsgMultiSend{
		Inputs:  []banktypes.Input{{Address: "cosmos1abc", Coins: sdk.NewCoins(sdk.NewInt64Coin("stake", int64(outputs)))}},
		Outputs: outs,
	}
}

func terminalHandler(ctx sdk.Context, _ sdk.Tx, _ bool) (sdk.Context, error) {
	return ctx, nil
}

func TestMsgMultiSendCapDecorator(t *testing.T) {
	const cap = 16

	alwaysActive := func(int64) bool { return true }
	alwaysInactive := func(int64) bool { return false }
	// Activates exactly at block height H = 100.
	activateAt100 := func(h int64) bool { return h >= 100 }

	tests := []struct {
		name       string
		decorator  ante.MsgMultiSendCapDecorator
		height     int64
		msgs       []sdk.Msg
		wantReject bool
	}{
		{
			name:       "active + over cap rejects",
			decorator:  ante.NewMsgMultiSendCapDecorator(cap, alwaysActive),
			msgs:       []sdk.Msg{newMultiSend(cap + 1)},
			wantReject: true,
		},
		{
			name:      "active + at cap accepts",
			decorator: ante.NewMsgMultiSendCapDecorator(cap, alwaysActive),
			msgs:      []sdk.Msg{newMultiSend(cap)},
		},
		{
			name:      "active + under cap accepts",
			decorator: ante.NewMsgMultiSendCapDecorator(cap, alwaysActive),
			msgs:      []sdk.Msg{newMultiSend(1)},
		},
		{
			name:      "inactive (activeFn=false) skips enforcement even over cap",
			decorator: ante.NewMsgMultiSendCapDecorator(cap, alwaysInactive),
			msgs:      []sdk.Msg{newMultiSend(cap + 100)},
		},
		{
			name:       "activeFn nil acts as always-on, rejects over cap",
			decorator:  ante.NewMsgMultiSendCapDecorator(cap, nil),
			msgs:       []sdk.Msg{newMultiSend(cap + 1)},
			wantReject: true,
		},
		{
			name:      "single MsgSend (1 output) passes",
			decorator: ante.NewMsgMultiSendCapDecorator(cap, alwaysActive),
			msgs:      []sdk.Msg{&banktypes.MsgSend{FromAddress: "cosmos1a", ToAddress: "cosmos1b", Amount: sdk.NewCoins(sdk.NewInt64Coin("stake", 1))}},
		},
		{
			name:      "1 MsgSend + MsgMultiSend at cap-1 (aggregate at cap) accepts",
			decorator: ante.NewMsgMultiSendCapDecorator(cap, alwaysActive),
			msgs: []sdk.Msg{
				&banktypes.MsgSend{FromAddress: "cosmos1a", ToAddress: "cosmos1b", Amount: sdk.NewCoins(sdk.NewInt64Coin("stake", 1))},
				newMultiSend(cap - 1),
			},
		},
		{
			name:      "1 MsgSend + MsgMultiSend at cap (aggregate over cap) rejects",
			decorator: ante.NewMsgMultiSendCapDecorator(cap, alwaysActive),
			msgs: []sdk.Msg{
				&banktypes.MsgSend{FromAddress: "cosmos1a", ToAddress: "cosmos1b", Amount: sdk.NewCoins(sdk.NewInt64Coin("stake", 1))},
				newMultiSend(cap),
			},
			wantReject: true,
		},
		{
			name:      "1 MsgSend + MsgMultiSend over cap rejects",
			decorator: ante.NewMsgMultiSendCapDecorator(cap, alwaysActive),
			msgs: []sdk.Msg{
				&banktypes.MsgSend{FromAddress: "cosmos1a", ToAddress: "cosmos1b", Amount: sdk.NewCoins(sdk.NewInt64Coin("stake", 1))},
				newMultiSend(cap + 1),
			},
			wantReject: true,
		},
		{
			name:      "below activation height: over cap accepts",
			decorator: ante.NewMsgMultiSendCapDecorator(cap, activateAt100),
			height:    99,
			msgs:      []sdk.Msg{newMultiSend(cap + 1)},
		},
		{
			name:       "at activation height: over cap rejects",
			decorator:  ante.NewMsgMultiSendCapDecorator(cap, activateAt100),
			height:     100,
			msgs:       []sdk.Msg{newMultiSend(cap + 1)},
			wantReject: true,
		},
		{
			name:       "above activation height: over cap rejects",
			decorator:  ante.NewMsgMultiSendCapDecorator(cap, activateAt100),
			height:     101,
			msgs:       []sdk.Msg{newMultiSend(cap + 1)},
			wantReject: true,
		},
		{
			name:      "multi-message: each under cap, aggregate under cap accepts",
			decorator: ante.NewMsgMultiSendCapDecorator(cap, alwaysActive),
			msgs: []sdk.Msg{
				newMultiSend(cap / 2),
				newMultiSend(cap / 2),
			},
		},
		{
			name:      "multi-message: each under cap, aggregate at cap accepts",
			decorator: ante.NewMsgMultiSendCapDecorator(cap, alwaysActive),
			msgs: []sdk.Msg{
				newMultiSend(cap / 2),
				newMultiSend(cap - cap/2),
			},
		},
		{
			name:      "multi-message: each under cap but aggregate over cap rejects",
			decorator: ante.NewMsgMultiSendCapDecorator(cap, alwaysActive),
			msgs: []sdk.Msg{
				newMultiSend(cap),
				newMultiSend(1),
			},
			wantReject: true,
		},
		{
			name:      "multi-message bypass attempt: 10 messages of half-cap each rejects",
			decorator: ante.NewMsgMultiSendCapDecorator(cap, alwaysActive),
			msgs: []sdk.Msg{
				newMultiSend(cap / 2),
				newMultiSend(cap / 2),
				newMultiSend(cap / 2),
				newMultiSend(cap / 2),
				newMultiSend(cap / 2),
				newMultiSend(cap / 2),
				newMultiSend(cap / 2),
				newMultiSend(cap / 2),
				newMultiSend(cap / 2),
				newMultiSend(cap / 2),
			},
			wantReject: true,
		},
		{
			name:      "MsgSend count under cap accepts",
			decorator: ante.NewMsgMultiSendCapDecorator(cap, alwaysActive),
			msgs: func() []sdk.Msg {
				out := make([]sdk.Msg, 0, cap)
				for i := 0; i < cap; i++ {
					out = append(out, &banktypes.MsgSend{FromAddress: "cosmos1a", ToAddress: "cosmos1b", Amount: sdk.NewCoins(sdk.NewInt64Coin("stake", 1))})
				}
				return out
			}(),
		},
		{
			name:      "MsgSend count over cap rejects (bypass attempt: 17 MsgSend in one tx)",
			decorator: ante.NewMsgMultiSendCapDecorator(cap, alwaysActive),
			msgs: func() []sdk.Msg {
				out := make([]sdk.Msg, 0, cap+1)
				for i := 0; i < cap+1; i++ {
					out = append(out, &banktypes.MsgSend{FromAddress: "cosmos1a", ToAddress: "cosmos1b", Amount: sdk.NewCoins(sdk.NewInt64Coin("stake", 1))})
				}
				return out
			}(),
			wantReject: true,
		},
		{
			name:      "mixed MsgSend + MsgMultiSend over aggregate cap rejects",
			decorator: ante.NewMsgMultiSendCapDecorator(cap, alwaysActive),
			msgs: []sdk.Msg{
				&banktypes.MsgSend{FromAddress: "cosmos1a", ToAddress: "cosmos1b", Amount: sdk.NewCoins(sdk.NewInt64Coin("stake", 1))},
				newMultiSend(cap),
			},
			wantReject: true,
		},
		{
			name:      "mixed MsgSend + MsgMultiSend at exact aggregate cap accepts",
			decorator: ante.NewMsgMultiSendCapDecorator(cap, alwaysActive),
			msgs: []sdk.Msg{
				&banktypes.MsgSend{FromAddress: "cosmos1a", ToAddress: "cosmos1b", Amount: sdk.NewCoins(sdk.NewInt64Coin("stake", 1))},
				newMultiSend(cap - 1),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := sdk.Context{}.WithBlockHeight(tt.height)
			tx := &capTestTx{msgs: tt.msgs}
			_, err := tt.decorator.AnteHandle(ctx, tx, false, terminalHandler)
			if tt.wantReject {
				require.Error(t, err)
				require.ErrorIs(t, err, sdkerrors.ErrInvalidRequest)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
