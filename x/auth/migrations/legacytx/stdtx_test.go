package legacytx

import (
	"context"
	"fmt"
	"github.com/cometbft/cometbft/crypto/secp256k1"
	"testing"

	"github.com/cosmos/gogoproto/proto"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/anypb"

	basev1beta1 "cosmossdk.io/api/cosmos/base/v1beta1"
	txv1beta1 "cosmossdk.io/api/cosmos/tx/v1beta1"
	txsigning "cosmossdk.io/x/tx/signing"
	"cosmossdk.io/x/tx/signing/aminojson"

	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	cryptocodec "github.com/cosmos/cosmos-sdk/crypto/codec"
	"github.com/cosmos/cosmos-sdk/testutil/testdata"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
)

var (
	// TODO CHECK HEIMDALL-V2 imported comet secp256k1, check if tests pass
	priv = secp256k1.GenPrivKey()
	//priv = ed25519.GenPrivKey()
	addr = sdk.AccAddress(priv.PubKey().Address())
)

func TestStdSignBytes(t *testing.T) {
	type args struct {
		chainID       string
		accnum        uint64
		sequence      uint64
		timeoutHeight uint64
		fee           *txv1beta1.Fee
		msgs          []sdk.Msg
		memo          string
	}
	defaultFee := &txv1beta1.Fee{
		Amount:   []*basev1beta1.Coin{{Denom: "atom", Amount: "150"}},
		GasLimit: 100000,
	}
	msgStr := fmt.Sprintf(`{"type":"testpb/TestMsg","value":{"signers":["%s"]}}`, addr)
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			"with timeout height",
			args{"1234", 3, 6, 10, defaultFee, []sdk.Msg{testdata.NewTestMsg(addr)}, "memo"},
			fmt.Sprintf(`{"account_number":"3","chain_id":"1234","fee":{"amount":[{"amount":"150","denom":"atom"}],"gas":"100000"},"memo":"memo","msgs":[%s],"sequence":"6","timeout_height":"10"}`, msgStr),
		},
		{
			"no timeout height (omitempty)",
			args{"1234", 3, 6, 0, defaultFee, []sdk.Msg{testdata.NewTestMsg(addr)}, "memo"},
			fmt.Sprintf(`{"account_number":"3","chain_id":"1234","fee":{"amount":[{"amount":"150","denom":"atom"}],"gas":"100000"},"memo":"memo","msgs":[%s],"sequence":"6"}`, msgStr),
		},
		{
			"empty fee",
			args{"1234", 3, 6, 0, &txv1beta1.Fee{}, []sdk.Msg{testdata.NewTestMsg(addr)}, "memo"},
			fmt.Sprintf(`{"account_number":"3","chain_id":"1234","fee":{"amount":[],"gas":"0"},"memo":"memo","msgs":[%s],"sequence":"6"}`, msgStr),
		},
		{
			"no fee payer and fee granter (both omitempty)",
			args{"1234", 3, 6, 0, &txv1beta1.Fee{Amount: defaultFee.Amount, GasLimit: defaultFee.GasLimit}, []sdk.Msg{testdata.NewTestMsg(addr)}, "memo"},
			fmt.Sprintf(`{"account_number":"3","chain_id":"1234","fee":{"amount":[{"amount":"150","denom":"atom"}],"gas":"100000"},"memo":"memo","msgs":[%s],"sequence":"6"}`, msgStr),
		},
		{
			"with fee granter, no fee payer (omitempty)",
			args{"1234", 3, 6, 0, &txv1beta1.Fee{Amount: defaultFee.Amount, GasLimit: defaultFee.GasLimit, Granter: addr.String()}, []sdk.Msg{testdata.NewTestMsg(addr)}, "memo"},
			fmt.Sprintf(`{"account_number":"3","chain_id":"1234","fee":{"amount":[{"amount":"150","denom":"atom"}],"gas":"100000","granter":"%s"},"memo":"memo","msgs":[%s],"sequence":"6"}`, addr, msgStr),
		},
		{
			"with fee payer, no fee granter (omitempty)",
			args{"1234", 3, 6, 0, &txv1beta1.Fee{Amount: defaultFee.Amount, GasLimit: defaultFee.GasLimit, Payer: addr.String()}, []sdk.Msg{testdata.NewTestMsg(addr)}, "memo"},
			fmt.Sprintf(`{"account_number":"3","chain_id":"1234","fee":{"amount":[{"amount":"150","denom":"atom"}],"gas":"100000","payer":"%s"},"memo":"memo","msgs":[%s],"sequence":"6"}`, addr, msgStr),
		},
		{
			"with fee payer and fee granter",
			args{"1234", 3, 6, 0, &txv1beta1.Fee{Amount: defaultFee.Amount, GasLimit: defaultFee.GasLimit, Payer: addr.String(), Granter: addr.String()}, []sdk.Msg{testdata.NewTestMsg(addr)}, "memo"},
			fmt.Sprintf(`{"account_number":"3","chain_id":"1234","fee":{"amount":[{"amount":"150","denom":"atom"}],"gas":"100000","granter":"%s","payer":"%s"},"memo":"memo","msgs":[%s],"sequence":"6"}`, addr, addr, msgStr),
		},
	}
	handler := aminojson.NewSignModeHandler(aminojson.SignModeHandlerOptions{
		FileResolver: proto.HybridResolver,
	})
	for i, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			anyMsgs := make([]*anypb.Any, len(tc.args.msgs))
			for j, msg := range tc.args.msgs {
				legacyAny, err := codectypes.NewAnyWithValue(msg)
				require.NoError(t, err)
				anyMsgs[j] = &anypb.Any{
					TypeUrl: legacyAny.TypeUrl,
					Value:   legacyAny.Value,
				}
			}
			got, err := handler.GetSignBytes(
				context.TODO(),
				txsigning.SignerData{
					Address:       "foo",
					ChainID:       tc.args.chainID,
					AccountNumber: tc.args.accnum,
					Sequence:      tc.args.sequence,
				},
				txsigning.TxData{
					Body: &txv1beta1.TxBody{
						Memo:          tc.args.memo,
						Messages:      anyMsgs,
						TimeoutHeight: tc.args.timeoutHeight,
					},
					AuthInfo: &txv1beta1.AuthInfo{
						Fee: tc.args.fee,
					},
				},
			)
			require.NoError(t, err)
			require.Equal(t, tc.want, string(got), "Got unexpected result on test case i: %d", i)
		})
	}
}

func TestSignatureV2Conversions(t *testing.T) {
	_, pubKey, _ := testdata.KeyTestPubAddr()
	cdc := codec.NewLegacyAmino()
	sdk.RegisterLegacyAminoCodec(cdc)
	dummy := []byte("dummySig")
	sig := StdSignature{PubKey: pubKey, Signature: dummy}

	sigV2, err := StdSignatureToSignatureV2(cdc, sig)
	require.NoError(t, err)
	require.Equal(t, pubKey, sigV2.PubKey)
	require.Equal(t, &signing.SingleSignatureData{
		SignMode:  signing.SignMode_SIGN_MODE_LEGACY_AMINO_JSON,
		Signature: dummy,
	}, sigV2.Data)

	sigBz, err := SignatureDataToAminoSignature(cdc, sigV2.Data)
	require.NoError(t, err)
	require.Equal(t, dummy, sigBz)
}

func TestGetSignaturesV2(t *testing.T) {
	_, pubKey, _ := testdata.KeyTestPubAddr()
	dummy := []byte("dummySig")

	cdc := codec.NewLegacyAmino()
	sdk.RegisterLegacyAminoCodec(cdc)
	cryptocodec.RegisterCrypto(cdc)

	fee := NewStdFee(50000, sdk.Coins{sdk.NewInt64Coin("atom", 150)})
	sig := StdSignature{PubKey: pubKey, Signature: dummy}
	stdTx := NewStdTx([]sdk.Msg{testdata.NewTestMsg()}, fee, []StdSignature{sig}, "testsigs")

	sigs, err := stdTx.GetSignaturesV2()
	require.Nil(t, err)
	require.Equal(t, len(sigs), 1)

	require.Equal(t, cdc.MustMarshal(sigs[0].PubKey), cdc.MustMarshal(sig.GetPubKey()))
	require.Equal(t, sigs[0].Data, &signing.SingleSignatureData{
		SignMode:  signing.SignMode_SIGN_MODE_LEGACY_AMINO_JSON,
		Signature: sig.GetSignature(),
	})
}

// TODO CHECK HEIMDALL-V2 tests imported from heimdall. If needed, rewrite without deprecated methods/types
//func TestStdTx(t *testing.T) {
//	t.Parallel()
//
//	msg := testdata.NewTestMsg(addr)
//	sig := StdSignature{}
//
//	tx := NewStdTx(msg, sig, "")
//	require.Equal(t, msg, tx.GetMsgs()[0])
//	require.Equal(t, sig, tx.GetSignatures()[0])
//
//	feePayer := tx.GetSigners()[0]
//	require.Equal(t, addr, feePayer)
//}
//
//func TestTxValidateBasic(t *testing.T) {
//	t.Parallel()
//
//	ctx := sdk.NewContext(nil, abci.Header{ChainID: "mychainid"}, false, log.NewNopLogger())
//
//	// keys and addresses
//	priv1, _, addr1 := sdkAuth.KeyTestPubAddr()
//
//	// msg and signatures
//	msg1 := sdk.NewTestMsg(addr1)
//	tx := NewTestTx(ctx, msg1, priv1, uint64(0), uint64(0))
//
//	require.NotNil(t, msg1)
//
//	err := tx.ValidateBasic()
//	require.Nil(t, err)
//	require.NoError(t, err)
//	require.NotPanics(t, func() { msg1.GetSignBytes() })
//}
//
//func TestDefaultTxEncoder(t *testing.T) {
//	t.Parallel()
//
//	cdc := codec.New()
//	sdk.RegisterCodec(cdc)
//	RegisterCodec(cdc)
//	cdc.RegisterConcrete(sdk.TestMsg{}, "cosmos-sdk/Test", nil)
//	encoder := DefaultTxEncoder(cdc)
//
//	msg := sdk.NewTestMsg(addr)
//	tx := NewStdTx(msg, StdSignature{}, "")
//
//	cdcBytes, err := cdc.MarshalBinaryLengthPrefixed(tx)
//
//	require.NoError(t, err)
//	encoderBytes, err := encoder(tx)
//
//	require.NoError(t, err)
//	require.Equal(t, cdcBytes, encoderBytes)
//}
//
//func TestTxDecode(t *testing.T) {
//	t.Parallel()
//
//	tx, err := base64.StdEncoding.DecodeString("wWhvHPg6AQHY1wEBlP+zHe/ZNZTQii57ULFjrJulHewY2NcBAZT/sx3v2TWU0Ioue1CxY6ybpR3sGICEXTLzJQ==")
//	require.NoError(t, err)
//
//	expected := "c1686f1cf83a0101d8d7010194ffb31defd93594d08a2e7b50b163ac9ba51dec18d8d7010194ffb31defd93594d08a2e7b50b163ac9ba51dec1880845d32f325"
//	require.Equal(t, expected, hex.EncodeToString(tx), "Tx encoding should match")
//}
//
//func TestTxHash(t *testing.T) {
//	t.Parallel()
//
//	txStr := "AANQR/im+GCUHE8PBUoNahQVOC3A/YPGU1GIsiCAggP/oAUa5K2J62X6bWX065hIawNsvuv3z2qU4ObSU8l7Mgm0oMCuqfNQzHmirstq75vRV+hkFczlWh9VjSGNn8JQCo3YhF5C8VG4QZGyoPc937dVz4DrkdYdDRwnigW0qiIE+yMVS/Drcdt9FXol4Tzegb+1qIQbP+EXUnnFLFAuaeUF7A3Rs8WajjUBgA=="
//	txHashStr := "b4560c30b12ebae71977373bcca2b0b553ae510efc4b167b4ebe7925f6e98557"
//
//	txBz, err := base64.StdEncoding.DecodeString(txStr)
//	require.NoError(t, err)
//
//	var tx tmTypes.Tx = txBz
//
//	require.Equal(t, txHashStr, hex.EncodeToString(tx.Hash()))
//}
