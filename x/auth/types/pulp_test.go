package types

import (
	"github.com/cosmos/cosmos-sdk/testutil/testdata"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	assert "github.com/stretchr/testify/require"
)

// TODO HV2 this is imported from heimdall, fix it (by implementing GetPulpHash properly) and unskip the test

func TestGetPulpHash(t *testing.T) {
	t.Skip()
	t.Parallel()

	tc := struct {
		in  sdk.Msg
		out []byte
	}{
		in:  testdata.NewTestMsg(nil),
		out: []byte{142, 88, 179, 79},
	}
	out := GetPulpHash(tc.in)
	assert.Equal(t, string(tc.out), string(out))
}
