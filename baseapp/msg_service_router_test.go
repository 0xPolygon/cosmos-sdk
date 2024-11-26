package baseapp_test

import (
	"context"
	"math/big"
	"testing"

	abci "github.com/cometbft/cometbft/abci/types"
	"github.com/stretchr/testify/require"

	hApp "github.com/0xPolygon/heimdall-v2/app"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/cosmos/cosmos-sdk/testutil/testdata"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	authsigning "github.com/cosmos/cosmos-sdk/x/auth/signing"
	authtx "github.com/cosmos/cosmos-sdk/x/auth/tx"
	authTypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/bank/testutil"
)

func TestRegisterMsgService(t *testing.T) {
	app, _, _ := hApp.SetupApp(t, 1)
	hApp.RequestFinalizeBlock(t, app, app.LastBlockHeight()+1)

	require.Panics(t, func() {
		testdata.RegisterMsgServer(
			app.MsgServiceRouter(),
			testdata.MsgServerImpl{},
		)
	})

	// Register testdata Msg services, and rerun `RegisterMsgService`.
	testdata.RegisterInterfaces(app.InterfaceRegistry())

	require.NotPanics(t, func() {
		testdata.RegisterMsgServer(
			app.MsgServiceRouter(),
			testdata.MsgServerImpl{},
		)
	})
}

func TestRegisterMsgServiceTwice(t *testing.T) {
	// Setup baseapp.
	app, _, _ := hApp.SetupApp(t, 1)
	hApp.RequestFinalizeBlock(t, app, app.LastBlockHeight()+1)

	testdata.RegisterInterfaces(app.InterfaceRegistry())

	// First time registering service shouldn't panic.
	require.NotPanics(t, func() {
		testdata.RegisterMsgServer(
			app.MsgServiceRouter(),
			testdata.MsgServerImpl{},
		)
	})

	// Second time should panic.
	require.Panics(t, func() {
		testdata.RegisterMsgServer(
			app.MsgServiceRouter(),
			testdata.MsgServerImpl{},
		)
	})
}

func TestHybridHandlerByMsgName(t *testing.T) {
	// Setup baseapp and router.
	app, _, _ := hApp.SetupApp(t, 1)
	hApp.RequestFinalizeBlock(t, app, app.LastBlockHeight()+1)

	testdata.RegisterInterfaces(app.InterfaceRegistry())

	testdata.RegisterMsgServer(
		app.MsgServiceRouter(),
		testdata.MsgServerImpl{},
	)

	handler := app.MsgServiceRouter().HybridHandlerByMsgName("testpb.MsgCreateDog")

	require.NotNil(t, handler)
	ctx := app.NewContext(true)
	resp := new(testdata.MsgCreateDogResponse)
	err := handler(ctx, &testdata.MsgCreateDog{
		Dog:   &testdata.Dog{Name: "Spot"},
		Owner: "me",
	}, resp)
	require.NoError(t, err)
	require.Equal(t, resp.Name, "Spot")
}

func TestMsgService(t *testing.T) {
	priv, _, _ := testdata.KeyTestPubAddr()

	app, _, _ := hApp.SetupApp(t, 1)
	hApp.RequestFinalizeBlock(t, app, app.LastBlockHeight()+1)

	ctx := app.NewContext(false)

	addr := sdk.AccAddress(priv.PubKey().Address())
	acc := authTypes.NewBaseAccount(addr, priv.PubKey(), 1337, 0)
	require.NoError(t, testutil.FundAccount(ctx, app.BankKeeper, addr, sdk.NewCoins(sdk.NewInt64Coin("pol", 43*defaultFeeAmount))))

	app.AccountKeeper.SetAccount(ctx, acc)

	// patch in TxConfig instead of using an output from x/auth/tx
	txConfig := authtx.NewTxConfig(app.AppCodec(), authtx.DefaultSignModes)
	// set the TxDecoder in the BaseApp for minimal tx simulations
	app.SetTxDecoder(txConfig.TxDecoder())

	defaultSignMode, err := authsigning.APISignModeToInternal(txConfig.SignModeHandler().DefaultMode())
	require.NoError(t, err)

	testdata.RegisterInterfaces(app.InterfaceRegistry())
	testdata.RegisterMsgServer(
		app.MsgServiceRouter(),
		testdata.MsgServerImpl{},
	)

	hApp.RequestFinalizeBlock(t, app, app.LastBlockHeight()+1)

	msg := testdata.MsgCreateDog{
		Dog:   &testdata.Dog{Name: "Spot"},
		Owner: addr.String(),
	}

	txBuilder := txConfig.NewTxBuilder()
	txBuilder.SetFeeAmount(testdata.NewTestFeeAmount())
	txBuilder.SetGasLimit(testdata.NewTestGasLimit())
	err = txBuilder.SetMsgs(&msg)
	require.NoError(t, err)

	// First round: we gather all the signer infos. We use the "set empty
	// signature" hack to do that.
	sigV2 := signing.SignatureV2{
		PubKey: priv.PubKey(),
		Data: &signing.SingleSignatureData{
			SignMode:  defaultSignMode,
			Signature: nil,
		},
		Sequence: 0,
	}

	err = txBuilder.SetSignatures(sigV2)
	require.NoError(t, err)

	// Second round: all signer infos are set, so each signer can sign.
	signerData := authsigning.SignerData{
		ChainID:       "",
		AccountNumber: 1337,
		Sequence:      0,
		PubKey:        priv.PubKey(),
	}
	sigV2, err = tx.SignWithPrivKey(
		context.TODO(), defaultSignMode, signerData,
		txBuilder, priv, txConfig, 0)
	require.NoError(t, err)
	err = txBuilder.SetSignatures(sigV2)
	require.NoError(t, err)

	// Send the tx to the app
	txBytes, err := txConfig.TxEncoder()(txBuilder.GetTx())
	require.NoError(t, err)
	res := hApp.RequestFinalizeBlockWithTxs(t, app, app.LastBlockHeight()+1, txBytes)
	require.Equal(t, abci.CodeTypeOK, res.TxResults[1].Code, "res=%+v", res)
}

var defaultFeeAmount = big.NewInt(10).Exp(big.NewInt(10), big.NewInt(15), nil).Int64()
