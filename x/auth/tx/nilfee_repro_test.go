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
// body minimal. It mirrors what a peer-gossiped tx looks like on the wire.
func decodeAuthInfo(t *testing.T, authInfo *tx.AuthInfo) sdk.Tx {
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
	return decoded
}

// A tx whose AuthInfo omits the Fee field (proto field 2) decodes with
// AuthInfo.Fee == nil. The first ante decorator (SetUpContextDecorator) calls
// GetGas() — this must not panic.
func TestNilFeeAccessorsDoNotPanic(t *testing.T) {
	feeTx, ok := decodeAuthInfo(t, &tx.AuthInfo{}).(sdk.FeeTx)
	require.True(t, ok)

	require.NotPanics(t, func() {
		require.Equal(t, uint64(0), feeTx.GetGas())
		require.Nil(t, feeTx.GetFee())
	})
}

// An empty (but present) Fee — what newBuilder always sets — must keep working.
func TestEmptyFeeStillWorks(t *testing.T) {
	feeTx := decodeAuthInfo(t, &tx.AuthInfo{Fee: &tx.Fee{}}).(sdk.FeeTx)
	require.Equal(t, uint64(0), feeTx.GetGas())
	require.Nil(t, feeTx.GetFee())
}

// A populated Fee returns its values unchanged.
func TestPopulatedFeeUnchanged(t *testing.T) {
	feeTx := decodeAuthInfo(t, &tx.AuthInfo{Fee: &tx.Fee{GasLimit: 21000}}).(sdk.FeeTx)
	require.Equal(t, uint64(21000), feeTx.GetGas())
}
