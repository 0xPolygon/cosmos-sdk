package bank_test

import (
	"math/big"
	"testing"

	abci "github.com/cometbft/cometbft/abci/types"
	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	sdkmath "cosmossdk.io/math"
	"cosmossdk.io/simapp"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
	sdk "github.com/cosmos/cosmos-sdk/types"
	moduletestutil "github.com/cosmos/cosmos-sdk/types/module/testutil"
	_ "github.com/cosmos/cosmos-sdk/x/auth"
	_ "github.com/cosmos/cosmos-sdk/x/auth/tx/config"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	"github.com/cosmos/cosmos-sdk/x/bank/testutil"
	"github.com/cosmos/cosmos-sdk/x/bank/types"
	_ "github.com/cosmos/cosmos-sdk/x/consensus"
	_ "github.com/cosmos/cosmos-sdk/x/distribution"
	distrkeeper "github.com/cosmos/cosmos-sdk/x/distribution/keeper"
	_ "github.com/cosmos/cosmos-sdk/x/gov"
	govv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	_ "github.com/cosmos/cosmos-sdk/x/params"
	_ "github.com/cosmos/cosmos-sdk/x/staking"
)

type (
	expectedBalance struct {
		addr  sdk.AccAddress
		coins sdk.Coins
	}

	appTestCase struct {
		desc             string
		expSimPass       bool
		expPass          bool
		msgs             []sdk.Msg
		accNums          []uint64
		accSeqs          []uint64
		privKeys         []cryptotypes.PrivKey
		expectedBalances []expectedBalance
		expInError       []string
	}
)

var (
	priv1 = secp256k1.GenPrivKey()
	addr1 = sdk.AccAddress(priv1.PubKey().Address())
	priv2 = secp256k1.GenPrivKey()
	addr2 = sdk.AccAddress(priv2.PubKey().Address())
	priv3 = secp256k1.GenPrivKey()
	addr3 = sdk.AccAddress(priv3.PubKey().Address())

	defaultFeeAmount = big.NewInt(10).Exp(big.NewInt(10), big.NewInt(15), nil).Int64()

	coins     = sdk.Coins{sdk.NewInt64Coin("matic", 10*defaultFeeAmount)}
	halfCoins = sdk.Coins{sdk.NewInt64Coin("matic", 5*defaultFeeAmount)}

	sendMsg1 = types.NewMsgSend(addr1, addr2, coins)

	multiSendMsg1 = &types.MsgMultiSend{
		Inputs:  []types.Input{types.NewInput(addr1, coins)},
		Outputs: []types.Output{types.NewOutput(addr2, coins)},
	}
	multiSendMsg2 = &types.MsgMultiSend{
		Inputs: []types.Input{types.NewInput(addr1, coins)},
		Outputs: []types.Output{
			types.NewOutput(addr2, halfCoins),
			types.NewOutput(addr3, halfCoins),
		},
	}
	multiSendMsg3 = &types.MsgMultiSend{
		Inputs: []types.Input{types.NewInput(addr2, coins)},
		Outputs: []types.Output{
			types.NewOutput(addr1, coins),
		},
	}
	multiSendMsg4 = &types.MsgMultiSend{
		Inputs: []types.Input{types.NewInput(addr1, coins)},
		Outputs: []types.Output{
			types.NewOutput(moduleAccAddr, coins),
		},
	}
	invalidMultiSendMsg = &types.MsgMultiSend{
		Inputs:  []types.Input{types.NewInput(addr1, coins), types.NewInput(addr2, coins)},
		Outputs: []types.Output{},
	}
)

type suite struct {
	BankKeeper         bankkeeper.Keeper
	AccountKeeper      types.AccountKeeper
	DistributionKeeper distrkeeper.Keeper
	App                *simapp.SimApp // use simapp instead for tests since depinject is not supported yet for heimdall app initialization
}

func createTestSuite(t *testing.T, genesisAccounts []authtypes.GenesisAccount) suite {
	res := suite{}

	var genAccounts []authtypes.GenesisAccount
	for _, acc := range genesisAccounts {
		genAccounts = append(genAccounts, acc)
	}

	// create validator set with single validator
	valSet, err := simtestutil.CreateRandomValidatorSet()
	require.NoError(t, err)
	app := simapp.SetupWithGenesisValSet(t, valSet, genAccounts)
	res.App = app
	res.AccountKeeper = app.AccountKeeper
	res.BankKeeper = app.BankKeeper
	res.DistributionKeeper = app.DistrKeeper

	return res
}

// CheckBalance checks the balance of an account.
func checkBalance(t *testing.T, baseApp *baseapp.BaseApp, addr sdk.AccAddress, balances sdk.Coins, keeper bankkeeper.Keeper) {
	ctxCheck := baseApp.NewContext(true)
	keeperBalances := keeper.GetAllBalances(ctxCheck, addr)
	require.True(t, balances.Equal(keeperBalances))
}

func TestSendNotEnoughBalance(t *testing.T) {
	acc1 := &authtypes.BaseAccount{
		Address: addr1.String(),
	}

	acc3 := &authtypes.BaseAccount{
		Address: addr3.String(),
	}
	genAccs := []authtypes.GenesisAccount{acc1, acc3}
	s := createTestSuite(t, genAccs)
	baseApp := s.App.BaseApp
	ctx := baseApp.NewContext(false)

	require.NoError(t, testutil.FundAccount(ctx, s.BankKeeper, addr1, sdk.NewCoins(sdk.NewInt64Coin("matic", 67*defaultFeeAmount))))
	require.NoError(t, testutil.FundAccount(ctx, s.BankKeeper, addr3, sdk.NewCoins(sdk.NewInt64Coin("matic", defaultFeeAmount-1))))
	_, err := baseApp.FinalizeBlock(&abci.RequestFinalizeBlock{Height: baseApp.LastBlockHeight() + 1})
	require.NoError(t, err)
	_, err = baseApp.Commit()
	require.NoError(t, err)

	testCases := []struct {
		name              string
		sender            sdk.AccAddress
		privKey           *secp256k1.PrivKey
		account           *authtypes.BaseAccount
		balance           sdk.Coins
		shouldSeqIncrease bool
	}{
		{
			"enough balance to pay for fees but not for transfer",
			addr1,
			priv1,
			acc1,
			sdk.Coins{sdk.NewInt64Coin("matic", 67*defaultFeeAmount)},
			true,
		},
		{
			"not enough balance to pay for fees",
			addr3,
			priv3,
			acc3,
			sdk.Coins{sdk.NewInt64Coin("matic", defaultFeeAmount-1)},
			false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			res1 := s.AccountKeeper.GetAccount(ctx, tc.sender)
			require.NotNil(t, res1)
			_, ok := res1.(*authtypes.BaseAccount)
			require.True(t, ok)

			origAccNum := res1.GetAccountNumber()
			origSeq := res1.GetSequence()

			sendMsg := types.NewMsgSend(tc.sender, addr2, sdk.Coins{sdk.NewInt64Coin("matic", 100*defaultFeeAmount)})
			header := cmtproto.Header{Height: baseApp.LastBlockHeight() + 1}
			txConfig := moduletestutil.MakeTestTxConfig()
			_, _, err := simtestutil.SignCheckDeliver(t, txConfig, baseApp, header, []sdk.Msg{sendMsg}, "", []uint64{origAccNum}, []uint64{origSeq}, false, false, tc.privKey)
			require.Error(t, err)

			if tc.shouldSeqIncrease {
				checkBalance(t, baseApp, tc.sender, tc.balance.Sub(sdk.NewInt64Coin("matic", defaultFeeAmount)), s.BankKeeper)
			} else {
				checkBalance(t, baseApp, tc.sender, tc.balance, s.BankKeeper)
			}

			ctx2 := baseApp.NewContext(true)
			res2 := s.AccountKeeper.GetAccount(ctx2, tc.sender)
			require.NotNil(t, res2)

			require.Equal(t, origAccNum, res2.GetAccountNumber())
			if tc.shouldSeqIncrease {
				require.Equal(t, origSeq+1, res2.GetSequence())
			} else {
				require.Equal(t, origSeq, res2.GetSequence())
			}

		})
	}

}

func TestMsgMultiSendWithAccounts(t *testing.T) {
	acc := &authtypes.BaseAccount{
		Address: addr1.String(),
	}

	genAccs := []authtypes.GenesisAccount{acc}
	s := createTestSuite(t, genAccs)
	baseApp := s.App.BaseApp
	ctx := baseApp.NewContext(false)

	require.NoError(t, testutil.FundAccount(ctx, s.BankKeeper, addr1, sdk.NewCoins(sdk.NewInt64Coin("matic", 68*defaultFeeAmount))))
	_, err := baseApp.FinalizeBlock(&abci.RequestFinalizeBlock{Height: baseApp.LastBlockHeight() + 1})
	require.NoError(t, err)
	_, err = baseApp.Commit()
	require.NoError(t, err)

	res1 := s.AccountKeeper.GetAccount(ctx, addr1)
	require.NotNil(t, res1)
	require.Equal(t, acc, res1.(*authtypes.BaseAccount))

	testCases := []appTestCase{
		{
			desc:       "make a valid tx",
			msgs:       []sdk.Msg{multiSendMsg1},
			accNums:    []uint64{0},
			accSeqs:    []uint64{0},
			expSimPass: true,
			expPass:    true,
			privKeys:   []cryptotypes.PrivKey{priv1},
			expectedBalances: []expectedBalance{
				{addr1, sdk.Coins{sdk.NewInt64Coin("matic", 57*defaultFeeAmount)}},
				{addr2, sdk.Coins{sdk.NewInt64Coin("matic", 10*defaultFeeAmount)}},
			},
		},
		{
			desc:       "wrong accNum should pass Simulate, but not Deliver",
			msgs:       []sdk.Msg{multiSendMsg1, multiSendMsg2},
			accNums:    []uint64{1}, // wrong account number
			accSeqs:    []uint64{1},
			expSimPass: true, // doesn't check signature
			expPass:    false,
			privKeys:   []cryptotypes.PrivKey{priv1},
		},
		{
			desc:       "wrong accSeq should not pass Simulate",
			msgs:       []sdk.Msg{multiSendMsg4},
			accNums:    []uint64{0},
			accSeqs:    []uint64{0}, // wrong account sequence
			expSimPass: false,
			expPass:    false,
			privKeys:   []cryptotypes.PrivKey{priv1},
		},
		{
			desc:       "multiple inputs not allowed",
			msgs:       []sdk.Msg{invalidMultiSendMsg},
			accNums:    []uint64{0},
			accSeqs:    []uint64{0},
			expSimPass: false,
			expPass:    false,
			privKeys:   []cryptotypes.PrivKey{priv1},
		},
	}

	for _, tc := range testCases {
		t.Logf("testing %s", tc.desc)
		header := cmtproto.Header{Height: baseApp.LastBlockHeight() + 1}
		txConfig := moduletestutil.MakeTestTxConfig()
		_, _, err := simtestutil.SignCheckDeliver(t, txConfig, baseApp, header, tc.msgs, "", tc.accNums, tc.accSeqs, tc.expSimPass, tc.expPass, tc.privKeys...)
		if tc.expPass {
			require.NoError(t, err)
		} else {
			require.Error(t, err)
		}

		for _, eb := range tc.expectedBalances {
			checkBalance(t, baseApp, eb.addr, eb.coins, s.BankKeeper)
		}
	}
}

func TestMsgMultiSendMultipleOut(t *testing.T) {
	acc1 := &authtypes.BaseAccount{
		Address: addr1.String(),
	}
	acc2 := &authtypes.BaseAccount{
		Address: addr2.String(),
	}

	genAccs := []authtypes.GenesisAccount{acc1, acc2}
	s := createTestSuite(t, genAccs)
	baseApp := s.App.BaseApp
	ctx := baseApp.NewContext(false)

	require.NoError(t, testutil.FundAccount(ctx, s.BankKeeper, addr1, sdk.NewCoins(sdk.NewInt64Coin("matic", 43*defaultFeeAmount))))
	require.NoError(t, testutil.FundAccount(ctx, s.BankKeeper, addr2, sdk.NewCoins(sdk.NewInt64Coin("matic", 42*defaultFeeAmount))))
	_, err := baseApp.FinalizeBlock(&abci.RequestFinalizeBlock{Height: baseApp.LastBlockHeight() + 1})
	require.NoError(t, err)
	_, err = baseApp.Commit()
	require.NoError(t, err)

	testCases := []appTestCase{
		{
			msgs:       []sdk.Msg{multiSendMsg2},
			accNums:    []uint64{0},
			accSeqs:    []uint64{0},
			expSimPass: true,
			expPass:    true,
			privKeys:   []cryptotypes.PrivKey{priv1},
			expectedBalances: []expectedBalance{
				{addr1, sdk.Coins{sdk.NewInt64Coin("matic", 32*defaultFeeAmount)}},
				{addr2, sdk.Coins{sdk.NewInt64Coin("matic", 47*defaultFeeAmount)}},
				{addr3, sdk.Coins{sdk.NewInt64Coin("matic", 5*defaultFeeAmount)}},
			},
		},
	}

	for _, tc := range testCases {
		header := cmtproto.Header{Height: baseApp.LastBlockHeight() + 1}
		txConfig := moduletestutil.MakeTestTxConfig()
		_, _, err := simtestutil.SignCheckDeliver(t, txConfig, baseApp, header, tc.msgs, "", tc.accNums, tc.accSeqs, tc.expSimPass, tc.expPass, tc.privKeys...)
		require.NoError(t, err)

		for _, eb := range tc.expectedBalances {
			checkBalance(t, baseApp, eb.addr, eb.coins, s.BankKeeper)
		}
	}
}

func TestMsgMultiSendDependent(t *testing.T) {
	acc1 := authtypes.NewBaseAccountWithAddress(addr1)
	acc2 := authtypes.NewBaseAccountWithAddress(addr2)
	err := acc2.SetAccountNumber(1)
	require.NoError(t, err)

	genAccs := []authtypes.GenesisAccount{acc1, acc2}
	s := createTestSuite(t, genAccs)
	baseApp := s.App.BaseApp
	ctx := baseApp.NewContext(false)

	require.NoError(t, testutil.FundAccount(ctx, s.BankKeeper, addr1, sdk.NewCoins(sdk.NewInt64Coin("matic", 43*defaultFeeAmount))))
	require.NoError(t, testutil.FundAccount(ctx, s.BankKeeper, addr2, sdk.NewCoins(sdk.NewInt64Coin("matic", defaultFeeAmount))))
	_, err = baseApp.FinalizeBlock(&abci.RequestFinalizeBlock{Height: baseApp.LastBlockHeight() + 1})
	require.NoError(t, err)
	_, err = baseApp.Commit()
	require.NoError(t, err)

	testCases := []appTestCase{
		{
			msgs:       []sdk.Msg{multiSendMsg1},
			accNums:    []uint64{0},
			accSeqs:    []uint64{0},
			expSimPass: true,
			expPass:    true,
			privKeys:   []cryptotypes.PrivKey{priv1},
			expectedBalances: []expectedBalance{
				{addr1, sdk.Coins{sdk.NewInt64Coin("matic", 32*defaultFeeAmount)}},
				{addr2, sdk.Coins{sdk.NewInt64Coin("matic", 11*defaultFeeAmount)}},
			},
		},
		{
			msgs:       []sdk.Msg{multiSendMsg3},
			accNums:    []uint64{1},
			accSeqs:    []uint64{0},
			expSimPass: true,
			expPass:    true,
			privKeys:   []cryptotypes.PrivKey{priv2},
			expectedBalances: []expectedBalance{
				{addr1, sdk.Coins{sdk.NewInt64Coin("matic", 42*defaultFeeAmount)}},
			},
		},
	}

	for _, tc := range testCases {
		header := cmtproto.Header{Height: baseApp.LastBlockHeight() + 1}
		txConfig := moduletestutil.MakeTestTxConfig()
		_, _, err := simtestutil.SignCheckDeliver(t, txConfig, baseApp, header, tc.msgs, "", tc.accNums, tc.accSeqs, tc.expSimPass, tc.expPass, tc.privKeys...)
		require.NoError(t, err)

		for _, eb := range tc.expectedBalances {
			checkBalance(t, baseApp, eb.addr, eb.coins, s.BankKeeper)
		}
	}
}

func TestMsgSetSendEnabled(t *testing.T) {
	t.Skip("skipping test as not relevant to Heimdall (MsgSetSendEnabled is not required as the only denom supported is matic)")
	acc1 := authtypes.NewBaseAccountWithAddress(addr1)

	genAccs := []authtypes.GenesisAccount{acc1}
	s := createTestSuite(t, genAccs)

	ctx := s.App.BaseApp.NewContext(false)
	require.NoError(t, testutil.FundAccount(ctx, s.BankKeeper, addr1, sdk.NewCoins(sdk.NewInt64Coin("matic", 101))))
	require.NoError(t, testutil.FundAccount(ctx, s.BankKeeper, addr1, sdk.NewCoins(sdk.NewInt64Coin("matic", 100000))))
	addr1Str := addr1.String()
	govAddr := s.BankKeeper.GetAuthority()
	goodGovProp, err := govv1.NewMsgSubmitProposal(
		[]sdk.Msg{
			types.NewMsgSetSendEnabled(govAddr, nil, nil),
		},
		sdk.Coins{{Denom: "matic", Amount: sdkmath.NewInt(100000)}},
		addr1Str,
		"set default send enabled to true",
		"Change send enabled",
		"Modify send enabled and set to true",
		false,
	)
	require.NoError(t, err, "making goodGovProp")

	testCases := []appTestCase{
		{
			desc:       "wrong authority",
			expSimPass: false,
			expPass:    false,
			msgs: []sdk.Msg{
				types.NewMsgSetSendEnabled(addr1Str, nil, nil),
			},
			accSeqs: []uint64{0},
			expInError: []string{
				"invalid authority",
				"cosmos10d07y265gmmuvt4z0w9aw880jnsr700j6zn9kn",
				addr1Str,
				"expected gov account as only signer for proposal message",
			},
		},
		{
			desc:       "right authority wrong signer",
			expSimPass: false,
			expPass:    false,
			msgs: []sdk.Msg{
				types.NewMsgSetSendEnabled(govAddr, nil, nil),
			},
			accSeqs: []uint64{1}, // wrong signer, so this sequence doesn't actually get used.
			expInError: []string{
				"pubKey does not match signer address",
				govAddr,
				"with signer index: 0",
				"invalid pubkey",
			},
		},
		{
			desc:       "submitted good as gov prop",
			expSimPass: true,
			expPass:    true,
			msgs: []sdk.Msg{
				goodGovProp,
			},
			accSeqs:    []uint64{1},
			expInError: nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.desc, func(tt *testing.T) {
			header := cmtproto.Header{Height: s.App.LastBlockHeight() + 1}
			txGen := moduletestutil.MakeTestTxConfig()
			_, _, err = simtestutil.SignCheckDeliver(tt, txGen, s.App.BaseApp, header, tc.msgs, "", []uint64{0}, tc.accSeqs, tc.expSimPass, tc.expPass, priv1)
			if len(tc.expInError) > 0 {
				require.Error(tt, err)
				for _, exp := range tc.expInError {
					assert.ErrorContains(tt, err, exp)
				}
			} else {
				require.NoError(tt, err)
			}
		})
	}
}
