package v2

import (
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
)

// MigrateGenState accepts exported v0.46 x/auth genesis state and migrates it to
// v0.47 x/auth genesis state. The migration includes:
// - If the group module is enabled, replace group policy accounts from module accounts to base accounts.
func MigrateGenState(oldState *authtypes.GenesisState) *authtypes.GenesisState {
	newState := *oldState

	accounts, err := authtypes.UnpackAccounts(newState.Accounts)
	if err != nil {
		panic(err)
	}

	groupPolicyAccountCounter := uint64(0)
	for i, acc := range accounts {
		modAcc, ok := acc.(sdk.ModuleAccountI)
		if !ok {
			continue
		}

		if modAcc.GetName() != modAcc.GetAddress().String() {
			continue
		}

		// TODO HV2: this code should work, but I believe it breaks the intended functionality of the group module.
		//  This implementation is here only for compatibility and allow the tests to pass.
		//  Otherwise, NewBaseAccountWithPubKey(pubKey) would fail with any derivationKey.
		//  This is because it calls NewBaseAccountWithAddress, which validates the address to be ethereum hex compatible
		//  We are not going to use group module (as we don't use multisig or accounts policies).
		//  However, is there a better alternative which keeps the group module operational?
		pubKey := secp256k1.GenPrivKey().PubKey()
		derivationKey := pubKey.Address()

		_, err = authtypes.NewModuleCredential(ModuleName, []byte{GroupPolicyTablePrefix}, derivationKey)
		if err != nil {
			panic(err)
		}
		baseAccount, err := authtypes.NewBaseAccountWithPubKey(pubKey)
		if err != nil {
			panic(err)
		}

		if err := baseAccount.SetAccountNumber(modAcc.GetAccountNumber()); err != nil {
			panic(err)
		}
		accounts[i] = baseAccount
		groupPolicyAccountCounter++
	}

	packedAccounts, err := authtypes.PackAccounts(accounts)
	if err != nil {
		panic(err)
	}
	newState.Accounts = packedAccounts

	return &newState
}
