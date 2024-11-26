package keeper_test

import (
	sdkmath "cosmossdk.io/math"

	topupTypes "github.com/0xPolygon/heimdall-v2/x/topup/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"
	"github.com/cosmos/cosmos-sdk/x/bank/types"
)

func (suite *KeeperTestSuite) TestExportGenesis() {
	ctx := suite.ctx

	expectedMetadata := suite.getTestMetadata()
	expectedBalances, expTotalSupply := suite.getTestBalancesAndSupply()

	// Adding genesis supply to the expTotalSupply
	genesisSupply, _, err := suite.bankKeeper.GetPaginatedTotalSupply(suite.ctx, &query.PageRequest{Limit: query.PaginationMaxLimit})
	suite.Require().NoError(err)
	expTotalSupply = expTotalSupply.Add(genesisSupply...)

	for i := range []int{1, 2} {
		suite.bankKeeper.SetDenomMetaData(ctx, expectedMetadata[i])
		accAddr, err1 := sdk.AccAddressFromHex(expectedBalances[i].Address)
		if err1 != nil {
			panic(err1)
		}
		// set balances via mint and send
		suite.mockMintCoins(mintAcc)
		suite.
			Require().
			NoError(suite.bankKeeper.MintCoins(ctx, topupTypes.ModuleName, expectedBalances[i].Coins))
		suite.mockSendCoinsFromModuleToAccount(mintAcc, accAddr)
		suite.
			Require().
			NoError(suite.bankKeeper.SendCoinsFromModuleToAccount(ctx, topupTypes.ModuleName, accAddr, expectedBalances[i].Coins))
	}

	suite.Require().NoError(suite.bankKeeper.SetParams(ctx, types.DefaultParams()))

	exportGenesis := suite.bankKeeper.ExportGenesis(ctx)

	suite.Require().Len(exportGenesis.Params.SendEnabled, 0) //nolint:staticcheck // SA1019: types.DefaultParams().SendEnabled is deprecated: Use DefaultSendEnabled instead.  (staticcheck)
	suite.Require().Equal(types.DefaultParams().DefaultSendEnabled, exportGenesis.Params.DefaultSendEnabled)
	suite.Require().Equal(expTotalSupply, exportGenesis.Supply)
	suite.Require().Subset(exportGenesis.Balances, expectedBalances)
	suite.Require().Equal(expectedMetadata, exportGenesis.DenomMetadata)
}

func (suite *KeeperTestSuite) getTestBalancesAndSupply() ([]types.Balance, sdk.Coins) {
	addr2, _ := sdk.AccAddressFromHex("0xdeadbeefdeadbeefdeadbeefdeadbeefdeadbeef")
	addr1, _ := sdk.AccAddressFromHex("0xde0b295669a9fd93d5f28d9ec85e40f4cb697bae")
	addr1Balance := sdk.Coins{sdk.NewInt64Coin("testcoin3", 10)}
	addr2Balance := sdk.Coins{sdk.NewInt64Coin("testcoin1", 32), sdk.NewInt64Coin("testcoin2", 34)}

	totalSupply := addr1Balance
	totalSupply = totalSupply.Add(addr2Balance...)

	return []types.Balance{
		{Address: addr2.String(), Coins: addr2Balance},
		{Address: addr1.String(), Coins: addr1Balance},
	}, totalSupply
}

func (suite *KeeperTestSuite) TestInitGenesis() {
	m := types.Metadata{Description: sdk.DefaultBondDenom, Base: sdk.DefaultBondDenom, Display: sdk.DefaultBondDenom}
	g := types.DefaultGenesisState()
	g.DenomMetadata = []types.Metadata{m}
	bk := suite.bankKeeper
	bk.InitGenesis(suite.ctx, g)

	m2, found := bk.GetDenomMetaData(suite.ctx, m.Base)
	suite.Require().True(found)
	suite.Require().Equal(m, m2)
}

func (suite *KeeperTestSuite) TestTotalSupply() {
	// Prepare some test data.
	defaultGenesis := types.DefaultGenesisState()
	balances := []types.Balance{
		{Coins: sdk.NewCoins(sdk.NewCoin("foocoin", sdkmath.NewInt(1))), Address: "0xdeadbeefdeadbeefdeadbeefdeadbeefdeadbeef"},
		{Coins: sdk.NewCoins(sdk.NewCoin("barcoin", sdkmath.NewInt(1))), Address: "0xde0b295669a9fd93d5f28d9ec85e40f4cb697bae"},
		{Coins: sdk.NewCoins(sdk.NewCoin("foocoin", sdkmath.NewInt(10)), sdk.NewCoin("barcoin", sdkmath.NewInt(20))), Address: "0xd00df00dd00df00dd00df00dd00df00dd00df00d"},
	}
	totalSupply := sdk.NewCoins(sdk.NewCoin("foocoin", sdkmath.NewInt(11)), sdk.NewCoin("barcoin", sdkmath.NewInt(21)))

	genesisSupply, _, err := suite.bankKeeper.GetPaginatedTotalSupply(suite.ctx, &query.PageRequest{Limit: query.PaginationMaxLimit})
	suite.Require().NoError(err)

	testcases := []struct {
		name        string
		genesis     *types.GenesisState
		expSupply   sdk.Coins
		expPanic    bool
		expPanicMsg string
	}{
		{
			"calculation NOT matching genesis Supply field",
			types.NewGenesisState(defaultGenesis.Params, balances, sdk.NewCoins(sdk.NewCoin("wrongcoin", sdkmath.NewInt(1))), defaultGenesis.DenomMetadata, defaultGenesis.SendEnabled),
			nil, true, "genesis supply is incorrect, expected 1wrongcoin, got 21barcoin,11foocoin",
		},
		{
			"calculation matches genesis Supply field",
			types.NewGenesisState(defaultGenesis.Params, balances, totalSupply, defaultGenesis.DenomMetadata, defaultGenesis.SendEnabled),
			totalSupply, false, "",
		},
		{
			"calculation is correct, empty genesis Supply field",
			types.NewGenesisState(defaultGenesis.Params, balances, nil, defaultGenesis.DenomMetadata, defaultGenesis.SendEnabled),
			totalSupply, false, "",
		},
	}

	for _, tc := range testcases {
		tc := tc
		suite.Run(tc.name, func() {
			if tc.expPanic {
				suite.PanicsWithError(tc.expPanicMsg, func() { suite.bankKeeper.InitGenesis(suite.ctx, tc.genesis) })
			} else {
				suite.bankKeeper.InitGenesis(suite.ctx, tc.genesis)
				totalSupply, _, err := suite.bankKeeper.GetPaginatedTotalSupply(suite.ctx, &query.PageRequest{Limit: query.PaginationMaxLimit})
				suite.Require().NoError(err)

				// adding genesis supply to expected supply
				expected := tc.expSupply.Add(genesisSupply...)
				suite.Require().Equal(expected, totalSupply)
			}
		})
	}
}
