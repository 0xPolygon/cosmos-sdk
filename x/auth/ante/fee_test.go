package ante_test

import (
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"

	"cosmossdk.io/math"

	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/cosmos/cosmos-sdk/testutil/testdata"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	"github.com/cosmos/cosmos-sdk/x/auth/ante"
)

func TestDeductFeeDecorator_ZeroGas(t *testing.T) {
	s := SetupTestSuite(t, false)
	s.txBuilder = s.clientCtx.TxConfig.NewTxBuilder()

	mfd := ante.NewDeductFeeDecorator(s.accountKeeper, s.bankKeeper, s.feeGrantKeeper, nil)
	antehandler := sdk.ChainAnteDecorators(mfd)

	// keys and addresses
	accs := s.CreateTestAccounts(1)

	// msg and signatures
	msg := testdata.NewTestMsg(accs[0].acc.GetAddress())
	require.NoError(t, s.txBuilder.SetMsgs(msg))

	// set zero gas
	s.txBuilder.SetGasLimit(0)

	privs, accNums, accSeqs := []cryptotypes.PrivKey{accs[0].priv}, []uint64{0}, []uint64{0}
	tx, err := s.CreateTestTx(s.ctx, privs, accNums, accSeqs, s.ctx.ChainID(), signing.SignMode_SIGN_MODE_DIRECT)
	require.NoError(t, err)

	// Set IsCheckTx to true
	s.ctx = s.ctx.WithIsCheckTx(true)

	_, err = antehandler(s.ctx, tx, false)
	require.Error(t, err)

	// zero gas is accepted in simulation mode
	_, err = antehandler(s.ctx, tx, true)
	require.NoError(t, err)
}

func TestEnsureMempoolFees(t *testing.T) {
	s := SetupTestSuite(t, true) // setup
	s.txBuilder = s.clientCtx.TxConfig.NewTxBuilder()

	mfd := ante.NewDeductFeeDecorator(s.accountKeeper, s.bankKeeper, s.feeGrantKeeper, nil)
	antehandler := sdk.ChainAnteDecorators(mfd)

	// keys and addresses
	accs := s.CreateTestAccounts(1)

	// msg and signatures
	msg := testdata.NewTestMsg(accs[0].acc.GetAddress())
	feeAmount := testdata.NewTestFeeAmount()
	gasLimit := uint64(15)
	require.NoError(t, s.txBuilder.SetMsgs(msg))
	s.txBuilder.SetFeeAmount(feeAmount)
	s.txBuilder.SetGasLimit(gasLimit)

	// called once more than vanilla cosmos-sdk as decorator will not throw an error for heimdall (CheckTx is disabled)
	s.bankKeeper.EXPECT().SendCoinsFromAccountToModule(gomock.Any(), accs[0].acc.GetAddress(), authtypes.FeeCollectorName, feeAmount).Return(nil).Times(4)

	privs, accNums, accSeqs := []cryptotypes.PrivKey{accs[0].priv}, []uint64{0}, []uint64{0}
	tx, err := s.CreateTestTx(s.ctx, privs, accNums, accSeqs, s.ctx.ChainID(), signing.SignMode_SIGN_MODE_DIRECT)
	require.NoError(t, err)

	// Set high gas price so standard test fee fails
	polPrice := sdk.NewDecCoinFromDec("pol", math.LegacyNewDec(20))
	highGasPrice := []sdk.DecCoin{polPrice}
	s.ctx = s.ctx.WithMinGasPrices(highGasPrice)

	// Set IsCheckTx to true
	s.ctx = s.ctx.WithIsCheckTx(true)

	// antehandler errors with insufficient fees
	_, err = antehandler(s.ctx, tx, false)
	require.Nil(t, err, "Decorator will not throw an error for heimdall as CheckTx is disabled")

	// antehandler should not error since we do not check minGasPrice in simulation mode
	cacheCtx, _ := s.ctx.CacheContext()
	_, err = antehandler(cacheCtx, tx, true)
	require.Nil(t, err, "Decorator should not have errored in simulation mode")

	// Set IsCheckTx to false
	s.ctx = s.ctx.WithIsCheckTx(false)

	// antehandler should not error since we do not check minGasPrice in DeliverTx
	_, err = antehandler(s.ctx, tx, false)
	require.Nil(t, err, "MempoolFeeDecorator returned error in DeliverTx")

	// Set IsCheckTx back to true for testing sufficient mempool fee
	s.ctx = s.ctx.WithIsCheckTx(true)

	polPrice = sdk.NewDecCoinFromDec("pol", math.LegacyNewDec(0).Quo(math.LegacyNewDec(1000000000000000)))
	lowGasPrice := []sdk.DecCoin{polPrice}
	s.ctx = s.ctx.WithMinGasPrices(lowGasPrice)

	newCtx, err := antehandler(s.ctx, tx, false)
	require.Nil(t, err, "Decorator should not have errored on fee higher than local gasPrice")
	// Priority is the smallest gas price amount in any denom. Since we have only 1 gas price
	// of 1000000000pol, the priority here is 1000000000.
	require.Equal(t, int64(1000000000), newCtx.Priority())
}

func TestDeductFees(t *testing.T) {
	s := SetupTestSuite(t, false)
	s.txBuilder = s.clientCtx.TxConfig.NewTxBuilder()

	// keys and addresses
	accs := s.CreateTestAccounts(1)

	// msg and signatures
	msg := testdata.NewTestMsg(accs[0].acc.GetAddress())
	feeAmount := testdata.NewTestFeeAmount()
	gasLimit := testdata.NewTestGasLimit()
	require.NoError(t, s.txBuilder.SetMsgs(msg))
	s.txBuilder.SetFeeAmount(feeAmount)
	s.txBuilder.SetGasLimit(gasLimit)

	privs, accNums, accSeqs := []cryptotypes.PrivKey{accs[0].priv}, []uint64{0}, []uint64{0}
	tx, err := s.CreateTestTx(s.ctx, privs, accNums, accSeqs, s.ctx.ChainID(), signing.SignMode_SIGN_MODE_DIRECT)
	require.NoError(t, err)

	dfd := ante.NewDeductFeeDecorator(s.accountKeeper, s.bankKeeper, nil, nil)
	antehandler := sdk.ChainAnteDecorators(dfd)
	s.bankKeeper.EXPECT().SendCoinsFromAccountToModule(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(sdkerrors.ErrInsufficientFunds)

	_, err = antehandler(s.ctx, tx, false)

	require.NotNil(t, err, "Tx did not error when fee payer had insufficient funds")

	s.bankKeeper.EXPECT().SendCoinsFromAccountToModule(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
	_, err = antehandler(s.ctx, tx, false)

	require.Nil(t, err, "Tx errored after account has been set with sufficient funds")
}
