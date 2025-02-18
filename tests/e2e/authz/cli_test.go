//go:build e2e
// +build e2e

package authz

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"cosmossdk.io/simapp"

	"github.com/cosmos/cosmos-sdk/testutil/network"
)

func TestE2ETestSuite(t *testing.T) {
	t.Skip("In HV2 we dont use authz module")
	cfg := network.DefaultConfig(simapp.NewTestNetworkFixture)
	cfg.NumValidators = 1
	suite.Run(t, NewE2ETestSuite(cfg))
}
