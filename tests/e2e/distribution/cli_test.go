//go:build e2e
// +build e2e

package distribution

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

func TestE2ETestSuite(t *testing.T) {
	t.Skip("In HV2 we dont use distribution module")
	suite.Run(t, new(E2ETestSuite))
}

func TestGRPCQueryTestSuite(t *testing.T) {
	t.Skip("In HV2 we dont use distribution module")
	suite.Run(t, new(GRPCQueryTestSuite))
}

func TestWithdrawAllSuite(t *testing.T) {
	t.Skip("In HV2 we dont use distribution module")
	suite.Run(t, new(WithdrawAllTestSuite))
}
