package v2

import (
	"fmt"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"

	storetypes "cosmossdk.io/store/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/address"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/group"
	"github.com/cosmos/cosmos-sdk/x/group/internal/orm"
)

const (
	ModuleName = "group"

	// Group Policy Table
	GroupPolicyTablePrefix    byte = 0x20
	GroupPolicyTableSeqPrefix byte = 0x21
)

// Migrate migrates the x/group module state from the consensus version 1 to version 2.
// Specifically, it changes the group policy account from module account to base account.
func Migrate(
	ctx sdk.Context,
	storeKey storetypes.StoreKey,
	accountKeeper group.AccountKeeper,
	groupPolicySeq orm.Sequence,
	groupPolicyTable orm.PrimaryKeyTable,
) error {
	store := ctx.KVStore(storeKey)
	curAccVal := groupPolicySeq.CurVal(store)
	groupPolicyAccountPubKey := make(map[string]cryptotypes.PubKey, 0)
	groupPolicyAccountDerivationKey := make(map[string][]byte, 0)
	policyKey := []byte{GroupPolicyTablePrefix}
	for i := uint64(0); i <= curAccVal; i++ {
		pubKey := secp256k1.GenPrivKey().PubKey()
		derivationKey := pubKey.Address()
		groupPolicyAcc := sdk.AccAddress(address.Module(group.ModuleName, policyKey, derivationKey))
		groupPolicyAccountDerivationKey[groupPolicyAcc.String()] = derivationKey
		groupPolicyAccountPubKey[groupPolicyAcc.String()] = pubKey
	}

	// get all group policies
	var groupPolicies []*group.GroupPolicyInfo
	if _, err := groupPolicyTable.Export(store, &groupPolicies); err != nil {
		return fmt.Errorf("failed to get group policies: %w", err)
	}

	for _, policy := range groupPolicies {
		addr, err := accountKeeper.AddressCodec().StringToBytes(policy.Address)
		if err != nil {
			return fmt.Errorf("failed to convert group policy account address: %w", err)
		}

		// get the account address by acc id
		oldAcc := accountKeeper.GetAccount(ctx, addr)
		// remove the old account
		accountKeeper.RemoveAccount(ctx, oldAcc)

		// create the group policy account
		derivationKey, ok := groupPolicyAccountDerivationKey[policy.Address]
		if !ok {
			// should never happen
			panic(fmt.Errorf("group policy account %s derivation key not found", policy.Address))
		}

		_, err = authtypes.NewModuleCredential(group.ModuleName, []byte{GroupPolicyTablePrefix}, derivationKey)
		if err != nil {
			return err
		}
		baseAccount, err := authtypes.NewBaseAccountWithPubKey(groupPolicyAccountPubKey[policy.Address])
		if err != nil {
			return fmt.Errorf("failed to create new group policy account: %w", err)
		}

		// set account number
		err = baseAccount.SetAccountNumber(oldAcc.GetAccountNumber())
		if err != nil {
			return err
		}

		// NOTE: we do not call NewAccount because we do not want to bump the account number

		// set new account
		// because we have only changed the account type, so we can use:
		//   - the same account number
		//   - the same address
		accountKeeper.SetAccount(ctx, baseAccount)
	}

	return nil
}
