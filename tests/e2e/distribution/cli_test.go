//go:build e2e
// +build e2e

package distribution

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

func TestE2ETestSuite(t *testing.T) {
	t.Skip("skipping test for HV2, see https://polygon.atlassian.net/browse/POS-2540")
	suite.Run(t, new(E2ETestSuite))
}

func TestGRPCQueryTestSuite(t *testing.T) {
	t.Skip("skipping test for HV2, see https://polygon.atlassian.net/browse/POS-2540")
	suite.Run(t, new(GRPCQueryTestSuite))
}

func TestWithdrawAllSuite(t *testing.T) {
	t.Skip("skipping test for HV2, see https://polygon.atlassian.net/browse/POS-2540")
	suite.Run(t, new(WithdrawAllTestSuite))
}
