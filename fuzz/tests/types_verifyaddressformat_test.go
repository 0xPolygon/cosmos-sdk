//go:build gofuzz || go1.18

package tests

import (
	"testing"

	"github.com/cosmos/cosmos-sdk/types"
)

func FuzzTypesVerifyAddressFormat(f *testing.F) {
	f.Fuzz(func(_ *testing.T, data []byte) {
		_ = types.VerifyAddressFormat(data)
	})
}
