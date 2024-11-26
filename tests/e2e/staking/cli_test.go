//go:build e2e
// +build e2e

package testutil

import (
	"testing"

	"cosmossdk.io/simapp"
	"github.com/cosmos/cosmos-sdk/testutil/network"
	"github.com/stretchr/testify/suite"
)

func TestE2ETestSuite(t *testing.T) {
	t.Skip("In HV2 we have own stake module")
	cfg := network.DefaultConfig(simapp.NewTestNetworkFixture)
	cfg.NumValidators = 2
	suite.Run(t, NewE2ETestSuite(cfg))
}
