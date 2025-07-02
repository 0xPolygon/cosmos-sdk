package main

import (
	"testing"

	"cosmossdk.io/tools/cosmovisor"
	"github.com/stretchr/testify/assert"
)

func TestGetHelpText(t *testing.T) {
	expectedPieces := []string{
		"Cosmovisor",
		cosmovisor.EnvName, cosmovisor.EnvHome,
		"https://docs.cosmos.network/main/tooling/cosmovisor",
	}

	actual := GetHelpText()
	for _, piece := range expectedPieces {
		assert.Contains(t, actual, piece)
	}
}
