package types

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/cometbft/cometbft/crypto"
	"github.com/cometbft/cometbft/crypto/secp256k1"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

var (
	_ sdk.AccountI                       = (*BaseAccount)(nil)
	_ GenesisAccount                     = (*BaseAccount)(nil)
	_ codectypes.UnpackInterfacesMessage = (*BaseAccount)(nil)
	_ GenesisAccount                     = (*ModuleAccount)(nil)
	_ sdk.ModuleAccountI                 = (*ModuleAccount)(nil)
)

// NewBaseAccount creates a new BaseAccount object.
func NewBaseAccount(address sdk.HeimdallAddress, pubKey cryptotypes.PubKey, accountNumber, sequence uint64) *BaseAccount {
	acc := &BaseAccount{
		Address:       address.String(),
		AccountNumber: accountNumber,
		Sequence:      sequence,
	}

	err := acc.SetPubKey(pubKey)
	if err != nil {
		panic(err)
	}

	return acc
}

// ProtoBaseAccount - a prototype function for BaseAccount
func ProtoBaseAccount() sdk.AccountI {
	return &BaseAccount{}
}

// NewBaseAccountWithAddress - returns a new base account with a given address
// leaving AccountNumber and Sequence to zero.
func NewBaseAccountWithAddress(addr sdk.HeimdallAddress) *BaseAccount {
	return &BaseAccount{
		Address: addr.String(),
	}
}

// GetAddress - Implements sdk.AccountI.
func (acc BaseAccount) GetAddress() sdk.HeimdallAddress {
	// TODO HV2 removed Bech32 related logic
	// addr, _ := sdk.AccAddressFromBech32(acc.Address)
	return sdk.HeimdallAddress{}
}

// SetAddress - Implements sdk.AccountI.
func (acc *BaseAccount) SetAddress(addr sdk.HeimdallAddress) error {
	if len(acc.Address) != 0 {
		return errors.New("cannot override BaseAccount address")
	}

	acc.Address = addr.String()
	return nil
}

// GetPubKey - Implements sdk.AccountI.
func (acc BaseAccount) GetPubKey() (pk cryptotypes.PubKey) {
	if acc.PubKey == nil {
		return nil
	}
	content, ok := acc.PubKey.GetCachedValue().(cryptotypes.PubKey)
	if !ok {
		return nil
	}
	return content
}

// SetPubKey - Implements sdk.AccountI.
func (acc *BaseAccount) SetPubKey(pubKey cryptotypes.PubKey) error {
	if pubKey == nil {
		acc.PubKey = nil
		return nil
	}
	any, err := codectypes.NewAnyWithValue(pubKey)
	if err == nil {
		acc.PubKey = any
	}
	return err
}

// GetAccountNumber - Implements AccountI
func (acc BaseAccount) GetAccountNumber() uint64 {
	return acc.AccountNumber
}

// SetAccountNumber - Implements AccountI
func (acc *BaseAccount) SetAccountNumber(accNumber uint64) error {
	acc.AccountNumber = accNumber
	return nil
}

// GetSequence - Implements sdk.AccountI.
func (acc BaseAccount) GetSequence() uint64 {
	return acc.Sequence
}

// SetSequence - Implements sdk.AccountI.
func (acc *BaseAccount) SetSequence(seq uint64) error {
	acc.Sequence = seq
	return nil
}

// TODO HV2 It's should not be needed, but verify (proto implements it in `auth.pb.go`)
// String implements fmt.Stringer
func (acc BaseAccount) String() string {
	var pubkey string

	if acc.PubKey != nil {
		// pubkey = sdk.MustBech32ifyAccPub(acc.PubKey)

		var pubObject secp256k1.PubKey

		cdc.MustUnmarshalBinaryBare(acc.PubKey.Bytes(), &pubObject)

		pubkey = "0x" + hex.EncodeToString(pubObject[:])
	}

	return fmt.Sprintf(`Account:
  Address:       %s
  Pubkey:        %s
  AccountNumber: %d
  Sequence:      %d`,
		acc.Address, pubkey, acc.AccountNumber, acc.Sequence)
}

// Validate checks for errors on the account fields
func (acc BaseAccount) Validate() error {
	if acc.Address == "" || acc.PubKey == nil {
		return nil
	}

	// TODO HV2 removed Bech32 related logic
	//accAddr, err := sdk.HeimdallAddressFromBech32(acc.Address)
	//if err != nil {
	//	return err
	//}

	if !bytes.Equal(acc.GetPubKey().Address().Bytes(), acc.Address.Bytes()) {
		return errors.New("account address and pubkey address do not match")
	}

	return nil
}

// UnpackInterfaces implements UnpackInterfacesMessage.UnpackInterfaces
func (acc BaseAccount) UnpackInterfaces(unpacker codectypes.AnyUnpacker) error {
	if acc.PubKey == nil {
		return nil
	}
	var pubKey cryptotypes.PubKey
	return unpacker.UnpackAny(acc.PubKey, &pubKey)
}

// NewModuleAddressOrBech32Address NewModuleAddressOrAddress gets an input string and returns an AccAddress.
// If the input is a valid address, it returns the address.
// If the input is a module name, it returns the module address.
// TODO HV2 removed Bech32 related logic
//func NewModuleAddressOrBech32Address(input string) sdk.HeimdallAddress {
//	if addr, err := sdk.HeimdallAddressFromBech32(input); err == nil {
//		return addr
//	}
//
//	return NewModuleAddress(input)
//}

// NewModuleAddress creates an AccAddress from the hash of the module's name
func NewModuleAddress(name string) sdk.HeimdallAddress {
	return sdk.BytesToHeimdallAddress(crypto.AddressHash([]byte(name)).Bytes())
}

// NewEmptyModuleAccount creates a empty ModuleAccount from a string
func NewEmptyModuleAccount(name string, permissions ...string) *ModuleAccount {
	moduleAddress := NewModuleAddress(name)
	baseAcc := NewBaseAccountWithAddress(moduleAddress)

	if err := validatePermissions(permissions...); err != nil {
		panic(err)
	}

	return &ModuleAccount{
		BaseAccount: baseAcc,
		Name:        name,
		Permissions: permissions,
	}
}

// NewModuleAccount creates a new ModuleAccount instance
func NewModuleAccount(ba *BaseAccount, name string, permissions ...string) *ModuleAccount {
	if err := validatePermissions(permissions...); err != nil {
		panic(err)
	}

	return &ModuleAccount{
		BaseAccount: ba,
		Name:        name,
		Permissions: permissions,
	}
}

// HasPermission returns whether or not the module account has permission.
func (ma ModuleAccount) HasPermission(permission string) bool {
	for _, perm := range ma.Permissions {
		if perm == permission {
			return true
		}
	}
	return false
}

// GetName returns the name of the holder's module
func (ma ModuleAccount) GetName() string {
	return ma.Name
}

// GetPermissions returns permissions granted to the module account
func (ma ModuleAccount) GetPermissions() []string {
	return ma.Permissions
}

// SetPubKey - Implements AccountI
func (ma ModuleAccount) SetPubKey(pubKey cryptotypes.PubKey) error {
	return fmt.Errorf("not supported for module accounts")
}

// Validate checks for errors on the account fields
func (ma ModuleAccount) Validate() error {
	if strings.TrimSpace(ma.Name) == "" {
		return errors.New("module account name cannot be blank")
	}

	if ma.BaseAccount == nil {
		return errors.New("uninitialized ModuleAccount: BaseAccount is nil")
	}

	if ma.Address != sdk.BytesToHeimdallAddress(crypto.AddressHash([]byte(ma.Name))).String() {
		return fmt.Errorf("address %s cannot be derived from the module name '%s'", ma.Address, ma.Name)
	}

	return ma.BaseAccount.Validate()
}

type moduleAccountPretty struct {
	Address       sdk.HeimdallAddress `json:"address"`
	PubKey        string              `json:"public_key"`
	AccountNumber uint64              `json:"account_number"`
	Sequence      uint64              `json:"sequence"`
	Name          string              `json:"name"`
	Permissions   []string            `json:"permissions"`
}

// MarshalJSON returns the JSON representation of a ModuleAccount.
func (ma ModuleAccount) MarshalJSON() ([]byte, error) {
	// TODO HV2 removed Bech32 related logic
	//accAddr, err := sdk.HeimdallAddressFromBech32(ma.Address)
	//if err != nil {
	//	return nil, err
	//}

	return json.Marshal(moduleAccountPretty{
		Address:       ma.Address,
		PubKey:        "",
		AccountNumber: ma.AccountNumber,
		Sequence:      ma.Sequence,
		Name:          ma.Name,
		Permissions:   ma.Permissions,
	})
}

// UnmarshalJSON unmarshals raw JSON bytes into a ModuleAccount.
func (ma *ModuleAccount) UnmarshalJSON(bz []byte) error {
	var alias moduleAccountPretty
	if err := json.Unmarshal(bz, &alias); err != nil {
		return err
	}

	ma.BaseAccount = NewBaseAccount(alias.Address, nil, alias.AccountNumber, alias.Sequence)
	ma.Name = alias.Name
	ma.Permissions = alias.Permissions

	return nil
}

// AccountI is an interface used to store coins at a given address within state.
// It presumes a notion of sequence numbers for replay protection,
// a notion of account numbers for replay protection for previously pruned accounts,
// and a pubkey for authentication purposes.
//
// Many complex conditions can be used in the concrete struct which implements AccountI.
//
// Deprecated: Use `AccountI` from types package instead.
type AccountI interface {
	sdk.AccountI
}

// ModuleAccountI defines an account interface for modules that hold tokens in
// an escrow.
//
// Deprecated: Use `ModuleAccountI` from types package instead.
type ModuleAccountI interface {
	sdk.ModuleAccountI
}

// GenesisAccounts defines a slice of GenesisAccount objects
type GenesisAccounts []GenesisAccount

// Contains returns true if the given address exists in a slice of GenesisAccount
// objects.
func (ga GenesisAccounts) Contains(addr sdk.HeimdallAddress) bool {
	for _, acc := range ga {
		if acc.GetAddress().Equals(addr) {
			return true
		}
	}

	return false
}

// GenesisAccount defines a genesis account that embeds an AccountI with validation capabilities.
type GenesisAccount interface {
	sdk.AccountI

	Validate() error
}

// AccountProcessor is an interface to process account as per module
type AccountProcessor func(*GenesisAccount, *BaseAccount) sdk.AccountI
