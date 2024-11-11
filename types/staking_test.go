package types_test

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

type stakingTestSuite struct {
	suite.Suite
}

func TestStakingTestSuite(t *testing.T) {
	suite.Run(t, new(stakingTestSuite))
}

func (s *stakingTestSuite) SetupSuite() {
	s.T().Parallel()
}

func (s *stakingTestSuite) TestTokensToConsensusPower() {
	s.T().Skip("skipping test for HV2, see https://polygon.atlassian.net/browse/POS-2540")
	s.Require().Equal(int64(0), sdk.TokensToConsensusPower(math.NewInt(999_999), sdk.DefaultPowerReduction))
	s.Require().Equal(int64(1), sdk.TokensToConsensusPower(math.NewInt(1_000_000), sdk.DefaultPowerReduction))
}
