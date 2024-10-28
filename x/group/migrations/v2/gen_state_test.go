package v2_test

import (
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	"testing"

	"github.com/stretchr/testify/require"

	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/group"
	v2 "github.com/cosmos/cosmos-sdk/x/group/migrations/v2"
)

func TestMigrateGenState(t *testing.T) {
	t.Skip("skipping test as not relevant for Heimdall")
	tests := []struct {
		name     string
		oldState *authtypes.GenesisState
		newState *authtypes.GenesisState
	}{
		{
			name: "group policy accounts are replaced by base accounts",
			oldState: authtypes.NewGenesisState(authtypes.DefaultParams(), authtypes.GenesisAccounts{
				&authtypes.ModuleAccount{
					BaseAccount: &authtypes.BaseAccount{
						Address:       "0x000000000000000000000000000000000000dead",
						AccountNumber: 3,
					},
					Name:        "distribution",
					Permissions: []string{},
				},
				&authtypes.ModuleAccount{
					BaseAccount: &authtypes.BaseAccount{
						Address:       "0x000000000000000000000000000000000000dead",
						AccountNumber: 8,
					},
					Name:        "0x000000000000000000000000000000000000dead",
					Permissions: []string{},
				},
			}),
			newState: authtypes.NewGenesisState(authtypes.DefaultParams(), authtypes.GenesisAccounts{
				&authtypes.ModuleAccount{
					BaseAccount: &authtypes.BaseAccount{
						Address:       "0x000000000000000000000000000000000000dead",
						AccountNumber: 3,
					},
					Name:        "distribution",
					Permissions: []string{},
				},
				func() *authtypes.BaseAccount {
					baseAccount := &authtypes.BaseAccount{
						Address:       "0x000000000000000000000000000000000000dead",
						AccountNumber: 8,
					}

					k := secp256k1.GenPrivKey().PubKey().Address()
					c, err := authtypes.NewModuleCredential(group.ModuleName, []byte{v2.GroupPolicyTablePrefix}, k)
					if err != nil {
						panic(err)
					}
					err = baseAccount.SetPubKey(c)
					if err != nil {
						panic(err)
					}

					return baseAccount
				}(),
			},
			),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			require.Error(t, authtypes.ValidateGenesis(*tc.oldState))
			actualState := v2.MigrateGenState(tc.oldState)
			require.Equal(t, tc.newState, actualState)
			require.NoError(t, authtypes.ValidateGenesis(*actualState))
		})
	}
}
