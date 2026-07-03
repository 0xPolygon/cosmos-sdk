package tx

import (
	"testing"

	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/tx"
	"github.com/stretchr/testify/require"
)

// decodeAuthInfo builds and decodes a tx with the given AuthInfo, leaving the
// body minimal, and returns it as an sdk.FeeTx. It mirrors what a decoded
// wire tx looks like when the fee field is present or omitted.
func decodeAuthInfo(t *testing.T, authInfo *tx.AuthInfo) sdk.FeeTx {
	t.Helper()
	cdc := codec.NewProtoCodec(codectypes.NewInterfaceRegistry())

	bodyBz, err := (&tx.TxBody{Memo: "foo"}).Marshal()
	require.NoError(t, err)
	authInfoBz, err := authInfo.Marshal()
	require.NoError(t, err)
	txBz, err := (&tx.TxRaw{BodyBytes: bodyBz, AuthInfoBytes: authInfoBz}).Marshal()
	require.NoError(t, err)

	decoded, err := DefaultTxDecoder(cdc)(txBz)
	require.NoError(t, err)
	feeTx, ok := decoded.(sdk.FeeTx)
	require.True(t, ok)
	return feeTx
}

// A tx whose AuthInfo omits the Fee field (proto field 2) decodes with
// AuthInfo.Fee == nil. The fee accessors — reached first by the ante chain's
// SetUpContextDecorator — must return zero values instead of dereferencing nil.
func TestNilFeeAccessorsDoNotPanic(t *testing.T) {
	feeTx := decodeAuthInfo(t, &tx.AuthInfo{})

	require.NotPanics(t, func() {
		require.Equal(t, uint64(0), feeTx.GetGas())
		require.Nil(t, feeTx.GetFee())
		require.Nil(t, feeTx.FeePayer())
		require.Nil(t, feeTx.FeeGranter())
	})
}

// An empty (but present) Fee — what newBuilder always sets — must keep working.
func TestEmptyFeeStillWorks(t *testing.T) {
	feeTx := decodeAuthInfo(t, &tx.AuthInfo{Fee: &tx.Fee{}})
	require.Equal(t, uint64(0), feeTx.GetGas())
	require.Nil(t, feeTx.GetFee())
}

// A populated Fee returns its values unchanged.
func TestPopulatedFeeUnchanged(t *testing.T) {
	feeTx := decodeAuthInfo(t, &tx.AuthInfo{Fee: &tx.Fee{GasLimit: 21000}})
	require.Equal(t, uint64(21000), feeTx.GetGas())
}
