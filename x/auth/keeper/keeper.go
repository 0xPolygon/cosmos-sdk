package keeper

import (
	"context"
	"errors"
	"fmt"

	"cosmossdk.io/collections"
	"cosmossdk.io/collections/indexes"
	"cosmossdk.io/core/address"
	"cosmossdk.io/core/store"
	errorsmod "cosmossdk.io/errors"
	"cosmossdk.io/log"

	"github.com/cosmos/cosmos-sdk/codec"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/auth/types"
)

// AccountKeeperI is the interface contract that x/auth's keeper implements.
type AccountKeeperI interface {
	// Return a new account with the next account number and the specified address. Does not save the new account to the store.
	NewAccountWithAddress(context.Context, sdk.HeimdallAddress) sdk.AccountI

	// Return a new account with the next account number. Does not save the new account to the store.
	NewAccount(context.Context, sdk.AccountI) sdk.AccountI

	// Check if an account exists in the store.
	HasAccount(context.Context, sdk.HeimdallAddress) bool

	// Retrieve an account from the store.
	GetAccount(context.Context, sdk.HeimdallAddress) sdk.AccountI

	// Set an account in the store.
	SetAccount(context.Context, sdk.AccountI)

	// Remove an account from the store.
	RemoveAccount(context.Context, sdk.AccountI)

	// Iterate over all accounts, calling the provided function. Stop iteration when it returns true.
	IterateAccounts(context.Context, func(sdk.AccountI) bool)

	// Fetch the public key of an account at a specified address
	GetPubKey(context.Context, sdk.HeimdallAddress) (cryptotypes.PubKey, error)

	// Fetch the sequence of an account at a specified address.
	GetSequence(context.Context, sdk.HeimdallAddress) (uint64, error)

	// Fetch the next account number, and increment the internal counter.
	NextAccountNumber(context.Context) uint64

	// GetModulePermissions fetches per-module account permissions
	GetModulePermissions() map[string]types.PermissionsForAddress

	// AddressCodec returns the account address codec.
	AddressCodec() address.Codec
}

func NewAccountIndexes(sb *collections.SchemaBuilder) AccountsIndexes {
	return AccountsIndexes{
		Number: indexes.NewUnique(
			sb, types.AccountNumberStoreKeyPrefix, "account_by_number", collections.Uint64Key, sdk.AccAddressKey,
			func(_ sdk.HeimdallAddress, v sdk.AccountI) (uint64, error) {
				return v.GetAccountNumber(), nil
			},
		),
	}
}

type AccountsIndexes struct {
	// Number is a unique index that indexes accounts by their account number.
	Number *indexes.Unique[uint64, sdk.HeimdallAddress, sdk.AccountI]
}

func (a AccountsIndexes) IndexesList() []collections.Index[sdk.HeimdallAddress, sdk.AccountI] {
	return []collections.Index[sdk.HeimdallAddress, sdk.AccountI]{
		a.Number,
	}
}

// AccountKeeper encodes/decodes accounts using the go-amino (binary)
// encoding/decoding library.
type AccountKeeper struct {
	addressCodec address.Codec

	storeService store.KVStoreService
	cdc          codec.BinaryCodec
	permAddrs    map[string]types.PermissionsForAddress
	// TODO CHECK HEIMDALL-V2 bech32Prefix?
	bech32Prefix string

	// The prototypical AccountI constructor.
	proto func() sdk.AccountI

	// the address capable of executing a MsgUpdateParams message. Typically, this
	// should be the x/gov module account.
	authority string

	// State
	Schema        collections.Schema
	Params        collections.Item[types.Params]
	AccountNumber collections.Sequence
	Accounts      *collections.IndexedMap[sdk.HeimdallAddress, sdk.AccountI, AccountsIndexes]
}

var _ AccountKeeperI = &AccountKeeper{}

// NewAccountKeeper returns a new AccountKeeperI that uses go-amino to
// (binary) encode and decode concrete sdk.Accounts.
// `maccPerms` is a map that takes accounts' addresses as keys, and their respective permissions as values. This map is used to construct
// types.PermissionsForAddress and is used in keeper.ValidatePermissions. Permissions are plain strings,
// and don't have to fit into any predefined structure. This auth module does not use account permissions internally, though other modules
// may use auth.Keeper to access the accounts permissions map.
func NewAccountKeeper(
	cdc codec.BinaryCodec, storeService store.KVStoreService, proto func() sdk.AccountI,
	maccPerms map[string][]string, ac address.Codec, bech32Prefix, authority string,
) AccountKeeper {
	permAddrs := make(map[string]types.PermissionsForAddress)
	for name, perms := range maccPerms {
		permAddrs[name] = types.NewPermissionsForAddress(name, perms)
	}

	sb := collections.NewSchemaBuilder(storeService)

	ak := AccountKeeper{
		addressCodec:  ac,
		bech32Prefix:  bech32Prefix,
		storeService:  storeService,
		proto:         proto,
		cdc:           cdc,
		permAddrs:     permAddrs,
		authority:     authority,
		Params:        collections.NewItem(sb, types.ParamsKey, "params", codec.CollValue[types.Params](cdc)),
		AccountNumber: collections.NewSequence(sb, types.GlobalAccountNumberKey, "account_number"),
		Accounts:      collections.NewIndexedMap(sb, types.AddressStoreKeyPrefix, "accounts", sdk.AccAddressKey, codec.CollInterfaceValue[sdk.AccountI](cdc), NewAccountIndexes(sb)),
	}
	schema, err := sb.Build()
	if err != nil {
		panic(err)
	}
	ak.Schema = schema
	return ak
}

// GetAuthority returns the x/auth module's authority.
func (ak AccountKeeper) GetAuthority() string {
	return ak.authority
}

// AddressCodec returns the x/auth account address codec.
// x/auth is tied to bech32 encoded user accounts
func (ak AccountKeeper) AddressCodec() address.Codec {
	return ak.addressCodec
}

// Logger returns a module-specific logger.
func (ak AccountKeeper) Logger(ctx context.Context) log.Logger {
	return sdk.UnwrapSDKContext(ctx).Logger().With("module", "x/"+types.ModuleName)
}

// GetPubKey Returns the PubKey of the account at address
func (ak AccountKeeper) GetPubKey(ctx context.Context, addr sdk.HeimdallAddress) (cryptotypes.PubKey, error) {
	acc := ak.GetAccount(ctx, addr)
	if acc == nil {
		return nil, errorsmod.Wrapf(sdkerrors.ErrUnknownAddress, "account %s does not exist", addr)
	}

	return acc.GetPubKey(), nil
}

// GetSequence Returns the Sequence of the account at address
func (ak AccountKeeper) GetSequence(ctx context.Context, addr sdk.HeimdallAddress) (uint64, error) {
	acc := ak.GetAccount(ctx, addr)
	if acc == nil {
		return 0, errorsmod.Wrapf(sdkerrors.ErrUnknownAddress, "account %s does not exist", addr)
	}

	return acc.GetSequence(), nil
}

// NextAccountNumber returns and increments the global account number counter.
// If the global account number is not set, it initializes it with value 0.
func (ak AccountKeeper) NextAccountNumber(ctx context.Context) uint64 {
	n, err := ak.AccountNumber.Next(ctx)
	if err != nil {
		panic(err)
	}
	return n
}

// GetModulePermissions fetches per-module account permissions.
func (ak AccountKeeper) GetModulePermissions() map[string]types.PermissionsForAddress {
	return ak.permAddrs
}

// ValidatePermissions validates that the module account has been granted
// permissions within its set of allowed permissions.
func (ak AccountKeeper) ValidatePermissions(macc sdk.ModuleAccountI) error {
	permAddr := ak.permAddrs[macc.GetName()]
	for _, perm := range macc.GetPermissions() {
		if !permAddr.HasPermission(perm) {
			return fmt.Errorf("invalid module permission %s", perm)
		}
	}

	return nil
}

// GetModuleAddress returns an address based on the module name
func (ak AccountKeeper) GetModuleAddress(moduleName string) sdk.HeimdallAddress {
	permAddr, ok := ak.permAddrs[moduleName]
	if !ok {
		return sdk.HeimdallAddress{}
	}

	return sdk.AccAddressToHeimdallAddress(permAddr.GetAddress())
}

// GetModuleAddressAndPermissions returns an address and permissions based on the module name
func (ak AccountKeeper) GetModuleAddressAndPermissions(moduleName string) (addr sdk.HeimdallAddress, permissions []string) {
	permAddr, ok := ak.permAddrs[moduleName]
	if !ok {
		return addr, permissions
	}

	return sdk.AccAddressToHeimdallAddress(permAddr.GetAddress()), permAddr.GetPermissions()
}

// GetModuleAccountAndPermissions gets the module account from the auth account store and its
// registered permissions
func (ak AccountKeeper) GetModuleAccountAndPermissions(ctx context.Context, moduleName string) (sdk.ModuleAccountI, []string) {
	addr, perms := ak.GetModuleAddressAndPermissions(moduleName)
	if addr.Empty() {
		return nil, []string{}
	}

	acc := ak.GetAccount(ctx, addr)
	if acc != nil {
		macc, ok := acc.(sdk.ModuleAccountI)
		if !ok {
			panic("account is not a module account")
		}
		return macc, perms
	}

	// create a new module account
	macc := types.NewEmptyModuleAccount(moduleName, perms...)
	maccI := (ak.NewAccount(ctx, macc)).(sdk.ModuleAccountI) // set the account number
	ak.SetModuleAccount(ctx, maccI)

	return maccI, perms
}

// GetModuleAccount gets the module account from the auth account store, if the account does not
// exist in the AccountKeeper, then it is created.
func (ak AccountKeeper) GetModuleAccount(ctx context.Context, moduleName string) sdk.ModuleAccountI {
	acc, _ := ak.GetModuleAccountAndPermissions(ctx, moduleName)
	return acc
}

// SetModuleAccount sets the module account to the auth account store
func (ak AccountKeeper) SetModuleAccount(ctx context.Context, macc sdk.ModuleAccountI) {
	ak.SetAccount(ctx, macc)
}

// add getter for bech32Prefix
func (ak AccountKeeper) getBech32Prefix() (string, error) {
	return ak.bech32Prefix, nil
}

// GetParams gets the auth module's parameters.
func (ak AccountKeeper) GetParams(ctx context.Context) (params types.Params) {
	params, err := ak.Params.Get(ctx)
	if err != nil && !errors.Is(err, collections.ErrNotFound) {
		panic(err)
	}
	return params
}

// GetBlockProposer returns block proposer
func (ak AccountKeeper) GetBlockProposer(ctx sdk.Context) (sdk.HeimdallAddress, bool) {
	// TODO CHECK HEIMDALL-V2 are these implementations equivalent for GetBlockProposer
	kvStore := ak.storeService.OpenKVStore(ctx)
	isProposerPresent, _ := kvStore.Has(types.ProposerKey()) // TODO CHECK HEIMDALL-V2 handle error?
	if !isProposerPresent {
		return sdk.HeimdallAddress{}, false
	}
	blockProposerBytes, _ := kvStore.Get(types.ProposerKey()) // TODO CHECK HEIMDALL-V2 handle error?
	return sdk.BytesToHeimdallAddress(blockProposerBytes), true

	//store := ctx.KVStore(ak.key)
	//if !store.Has(types.ProposerKey()) {
	//	return hmTypes.HeimdallAddress{}, false
	//}
	//bz := store.Get(types.ProposerKey())
	//return hmTypes.BytesToHeimdallAddress(bz), true
}

// SetBlockProposer sets block proposer
func (ak AccountKeeper) SetBlockProposer(ctx sdk.Context, addr sdk.HeimdallAddress) {
	// TODO CHECK HEIMDALL-V2 are these implementations equivalent for SetBlockProposer
	kvStore := ak.storeService.OpenKVStore(ctx)
	kvStore.Set(types.ProposerKey(), addr.Bytes()) // TODO CHECK HEIMDALL-V2 handle error?
	//store := ctx.KVStore(ak.key)
	//store.Set(types.ProposerKey(), addr.Bytes())
}

// RemoveBlockProposer removes block proposer from store
func (ak AccountKeeper) RemoveBlockProposer(ctx sdk.Context) {
	// TODO CHECK HEIMDALL-V2 are these implementations equivalent for RemoveBlockProposer
	kvStore := ak.storeService.OpenKVStore(ctx)
	kvStore.Delete(types.ProposerKey()) // TODO CHECK HEIMDALL-V2 handle error?
	//store := ctx.KVStore(ak.key)
	//store.Delete(types.ProposerKey())
}
