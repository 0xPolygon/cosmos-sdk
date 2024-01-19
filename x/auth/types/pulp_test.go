package types

import (
	"testing"

	"github.com/cosmos/cosmos-sdk/testutil/testdata"
	sdk "github.com/cosmos/cosmos-sdk/types"
	assert "github.com/stretchr/testify/require"
)

// TODO HV2: this is imported from heimdall, but not used anywhere (not heimdall, not cosmos).
//  Can we delete it? Also the test fails because of the missing implementation of the GetPulpHash function

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
