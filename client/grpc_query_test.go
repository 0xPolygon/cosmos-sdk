package client_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/suite"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"

	"cosmossdk.io/math"
	hApp "github.com/0xPolygon/heimdall-v2/app"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/testutil/testdata"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/bank/types"
)

type IntegrationTestSuite struct {
	suite.Suite

	ctx                   sdk.Context
	cdc                   codec.Codec
	genesisAccount        *authtypes.BaseAccount
	bankClient            types.QueryClient
	testClient            testdata.QueryClient
	genesisAccountBalance int64
}

func (s *IntegrationTestSuite) SetupSuite() {
	s.T().Log("setting up integration test suite")

	setUpAppResult := hApp.SetupApp(s.T(), 1)
	app := setUpAppResult.App
	hApp.RequestFinalizeBlock(s.T(), app, app.LastBlockHeight()+1)

	s.ctx = app.BaseApp.NewContext(false)
	s.cdc = app.AppCodec()
	queryHelper := baseapp.NewQueryServerTestHelper(s.ctx, app.InterfaceRegistry())
	types.RegisterQueryServer(queryHelper, app.BankKeeper)
	testdata.RegisterQueryServer(queryHelper, testdata.QueryImpl{})
	s.bankClient = types.NewQueryClient(queryHelper)
	s.testClient = testdata.NewQueryClient(queryHelper)

	accounts := app.AccountKeeper.GetAllAccounts(s.ctx)
	for _, acc := range accounts {
		if acc.GetPubKey() != nil {
			s.genesisAccount = acc.(*authtypes.BaseAccount)
			break
		}
	}
	genesisCoins := app.BankKeeper.GetBalance(s.ctx, s.genesisAccount.GetAddress(), sdk.DefaultBondDenom)
	s.genesisAccountBalance = genesisCoins.Amount.Int64()
}

func (s *IntegrationTestSuite) TearDownSuite() {
	s.T().Log("tearing down integration test suite")
}

func (s *IntegrationTestSuite) TestGRPCQuery() {
	denom := sdk.DefaultBondDenom

	// gRPC query to test service should work
	testRes, err := s.testClient.Echo(context.Background(), &testdata.EchoRequest{Message: "hello"})
	s.Require().NoError(err)
	s.Require().Equal("hello", testRes.Message)

	// gRPC query to bank service should work
	var header metadata.MD
	res, err := s.bankClient.Balance(
		context.Background(),
		&types.QueryBalanceRequest{Address: s.genesisAccount.GetAddress().String(), Denom: denom},
		grpc.Header(&header), // Also fetch grpc header
	)
	s.Require().NoError(err)
	bal := res.GetBalance()
	s.Equal(sdk.NewCoin(denom, math.NewInt(s.genesisAccountBalance)), *bal)
}

func TestIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(IntegrationTestSuite))
}
