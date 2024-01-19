package types

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"

	"github.com/ethereum/go-ethereum/common"
	"github.com/hashicorp/golang-lru/simplelru"
	"sigs.k8s.io/yaml"

	errorsmod "cosmossdk.io/errors"

	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

const (

	// TODO HV2 prefixes are now all empty strings. Do we need Purpose, CoinType, FullFundraiserPath ?

	// Constants defined here are the defaults value for address.
	// You can use the specific values for your project.
	// Add the follow lines to the `main()` of your server.
	//
	//	config := sdk.GetConfig()
	//	config.SetBech32PrefixForAccount(yourBech32PrefixAccAddr, yourBech32PrefixAccPub)
	//	config.SetBech32PrefixForValidator(yourBech32PrefixValAddr, yourBech32PrefixValPub)
	//	config.SetBech32PrefixForConsensusNode(yourBech32PrefixConsAddr, yourBech32PrefixConsPub)
	//	config.SetPurpose(yourPurpose)
	//	config.SetCoinType(yourCoinType)
	//	config.Seal()

	// Bech32MainPrefix defines the main SDK Bech32 prefix of an account's address
	Bech32MainPrefix = ""

	// Purpose is the ATOM purpose as defined in SLIP44 (https://github.com/satoshilabs/slips/blob/master/slip-0044.md)
	Purpose = 44

	// CoinType is the ATOM coin type as defined in SLIP44 (https://github.com/satoshilabs/slips/blob/master/slip-0044.md)
	CoinType = 118

	// FullFundraiserPath is the parts of the BIP44 HD path that are fixed by
	// what we used during the ATOM fundraiser.
	FullFundraiserPath = "m/44'/118'/0'/0/0"

	// PrefixAccount is the prefix for account keys
	PrefixAccount = ""
	// PrefixValidator is the prefix for validator keys
	PrefixValidator = ""
	// PrefixConsensus is the prefix for consensus keys
	PrefixConsensus = ""
	// PrefixPublic is the prefix for public keys
	PrefixPublic = ""
	// PrefixOperator is the prefix for operator keys
	PrefixOperator = ""

	// PrefixAddress is the prefix for addresses
	PrefixAddress = ""

	// Bech32PrefixAccAddr defines the Bech32 prefix of an account's address
	Bech32PrefixAccAddr = Bech32MainPrefix
	// Bech32PrefixAccPub defines the Bech32 prefix of an account's public key
	Bech32PrefixAccPub = Bech32MainPrefix + PrefixPublic
	// Bech32PrefixValAddr defines the Bech32 prefix of a validator's operator address
	Bech32PrefixValAddr = Bech32MainPrefix + PrefixValidator + PrefixOperator
	// Bech32PrefixValPub defines the Bech32 prefix of a validator's operator public key
	Bech32PrefixValPub = Bech32MainPrefix + PrefixValidator + PrefixOperator + PrefixPublic
	// Bech32PrefixConsAddr defines the Bech32 prefix of a consensus node address
	Bech32PrefixConsAddr = Bech32MainPrefix + PrefixValidator + PrefixConsensus
	// Bech32PrefixConsPub defines the Bech32 prefix of a consensus node public key
	Bech32PrefixConsPub = Bech32MainPrefix + PrefixValidator + PrefixConsensus + PrefixPublic
)

// cache variables
var (
	// AccAddress.String() is expensive and if unoptimized dominantly showed up in profiles,
	// yet has no mechanisms to trivially cache the result given that AccAddress is a []byte type.
	accAddrMu     sync.Mutex
	accAddrCache  *simplelru.LRU
	consAddrMu    sync.Mutex
	consAddrCache *simplelru.LRU
	valAddrMu     sync.Mutex
	valAddrCache  *simplelru.LRU

	isCachingEnabled atomic.Bool
)

// sentinel errors
var (
	ErrEmptyHexAddress = errors.New("decoding address from hex string failed: empty address")
	ErrNotHexAddress   = errors.New("decoding address from hex string failed: not valid address")
)

func init() {
	var err error
	SetAddrCacheEnabled(true)

	// in total the cache size is 61k entries. Key is 32 bytes and value is around 50-70 bytes.
	// That will make around 92 * 61k * 2 (LRU) bytes ~ 11 MB
	if accAddrCache, err = simplelru.NewLRU(60000, nil); err != nil {
		panic(err)
	}
	if consAddrCache, err = simplelru.NewLRU(500, nil); err != nil {
		panic(err)
	}
	if valAddrCache, err = simplelru.NewLRU(500, nil); err != nil {
		panic(err)
	}
}

// SetAddrCacheEnabled enables or disables accAddrCache, consAddrCache, and valAddrCache. By default, caches are enabled.
func SetAddrCacheEnabled(enabled bool) {
	isCachingEnabled.Store(enabled)
}

// IsAddrCacheEnabled returns if the address caches are enabled.
func IsAddrCacheEnabled() bool {
	return isCachingEnabled.Load()
}

// Address is a common interface for different types of addresses used by the SDK
type Address interface {
	Equals(Address) bool
	Empty() bool
	Marshal() ([]byte, error)
	MarshalJSON() ([]byte, error)
	Bytes() []byte
	String() string
	Format(s fmt.State, verb rune)
}

// Ensure that different address types implement the interface
var (
	_ Address = AccAddress{}
	_ Address = ValAddress{}
	_ Address = ConsAddress{}
)

// ----------------------------------------------------------------------------
// account
// ----------------------------------------------------------------------------

// AccAddress a wrapper around bytes meant to represent an account address.
// When marshaled to a string or JSON, it uses hex.
type AccAddress []byte

// AccAddressFromHex creates an AccAddress from a HEX-encoded string.
func AccAddressFromHex(address string) (addr AccAddress, err error) {
	bz, err := addressBytesFromHexString(address)
	return AccAddress(bz), err
}

// VerifyAddressFormat verifies that the provided bytes form a valid address
func VerifyAddressFormat(bz []byte) error {
	if len(bz) == 0 {
		return errorsmod.Wrap(sdkerrors.ErrUnknownAddress, "addresses cannot be empty")
	}

	if !common.IsHexAddress(common.BytesToAddress(bz).String()) {
		return errorsmod.Wrapf(sdkerrors.ErrUnknownAddress, "invalid address")
	}

	return nil
}

// MustAccAddressFromHex calls AccAddressFromHex and panics on error.
func MustAccAddressFromHex(address string) AccAddress {
	addr, err := AccAddressFromHex(address)
	if err != nil {
		panic(err)
	}

	return addr
}

// Equals Returns boolean for whether two AccAddresses are Equal
func (aa AccAddress) Equals(aa2 Address) bool {
	if aa.Empty() && aa2.Empty() {
		return true
	}

	return bytes.Equal(aa.Bytes(), aa2.Bytes())
}

// Empty Returns boolean for whether an AccAddress is empty
func (aa AccAddress) Empty() bool {
	return len(aa) == 0
}

// Marshal returns the raw address bytes. It is needed for protobuf
// compatibility.
func (aa AccAddress) Marshal() ([]byte, error) {
	return aa, nil
}

// Unmarshal sets the address to the given data. It is needed for protobuf
// compatibility.
func (aa *AccAddress) Unmarshal(data []byte) error {
	*aa = data
	return nil
}

// MarshalJSON marshals to JSON using hex.
func (aa AccAddress) MarshalJSON() ([]byte, error) {
	return json.Marshal(aa.String())
}

// MarshalYAML marshals to YAML using hex.
func (aa AccAddress) MarshalYAML() (interface{}, error) {
	return aa.String(), nil
}

// UnmarshalJSON unmarshals from JSON assuming hex encoding.
func (aa *AccAddress) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}

	*aa = common.HexToAddress(s).Bytes()
	return nil
}

// UnmarshalYAML unmarshals from JSON assuming hex encoding.
func (aa *AccAddress) UnmarshalYAML(data []byte) error {
	var s string
	if err := yaml.Unmarshal(data, &s); err != nil {
		return err
	}

	*aa = common.HexToAddress(s).Bytes()

	return nil
}

// Bytes returns the raw address bytes.
func (aa AccAddress) Bytes() []byte {
	return aa[:]
}

// String implements the Stringer interface.
func (aa AccAddress) String() string {
	return "0x" + common.Bytes2Hex(aa.Bytes())
}

// Format implements the fmt.Formatter interface.

func (aa AccAddress) Format(s fmt.State, verb rune) {
	switch verb {
	case 's':
		s.Write([]byte(aa.String()))
	case 'p':
		s.Write([]byte(fmt.Sprintf("%p", aa)))
	default:
		s.Write([]byte(fmt.Sprintf("%X", aa.Bytes())))
	}
}

// ----------------------------------------------------------------------------
// validator operator
// ----------------------------------------------------------------------------

// ValAddress defines a wrapper around bytes meant to present a validator's
// operator. When marshaled to a string or JSON, it uses hex.
type ValAddress []byte

// ValAddressFromHex creates a ValAddress from a hex string.
func ValAddressFromHex(address string) (addr ValAddress, err error) {
	bz, err := addressBytesFromHexString(address)
	return ValAddress(bz), err
}

// Equals Returns boolean for whether two ValAddresses are Equal
func (va ValAddress) Equals(va2 Address) bool {
	if va.Empty() && va2.Empty() {
		return true
	}

	return bytes.Equal(va.Bytes(), va2.Bytes())
}

// Empty Returns boolean for whether an ValAddress is empty
func (va ValAddress) Empty() bool {
	return len(va) == 0
}

// Marshal returns the raw address bytes. It is needed for protobuf
// compatibility.
func (va ValAddress) Marshal() ([]byte, error) {
	return va, nil
}

// Unmarshal sets the address to the given data. It is needed for protobuf
// compatibility.
func (va *ValAddress) Unmarshal(data []byte) error {
	*va = data
	return nil
}

// MarshalJSON marshals to JSON using hex.
func (va ValAddress) MarshalJSON() ([]byte, error) {
	return json.Marshal(va.String())
}

// MarshalYAML marshals to YAML using hex.
func (va ValAddress) MarshalYAML() (interface{}, error) {
	return va.String(), nil
}

// UnmarshalJSON unmarshals from JSON assuming hex encoding.
func (va *ValAddress) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}

	*va = common.HexToAddress(s).Bytes()

	return nil
}

// UnmarshalYAML unmarshals from YAML assuming hex encoding.
func (va *ValAddress) UnmarshalYAML(data []byte) error {
	var s string
	if err := yaml.Unmarshal(data, &s); err != nil {
		return err
	}

	*va = common.HexToAddress(s).Bytes()

	return nil
}

// Bytes returns the raw address bytes.
func (va ValAddress) Bytes() []byte {
	return va[:]
}

// String implements the Stringer interface.
func (va ValAddress) String() string {
	// TODO HV2: does this need to be done for all Tx related types? Like TxHash. Is there a way to do it with proto?
	return "0x" + common.Bytes2Hex(va.Bytes())
}

// Format implements the fmt.Formatter interface.
func (va ValAddress) Format(s fmt.State, verb rune) {
	switch verb {
	case 's':
		s.Write([]byte(va.String()))
	case 'p':
		s.Write([]byte(fmt.Sprintf("%p", va)))
	default:
		s.Write([]byte(fmt.Sprintf("%X", va.Bytes())))
	}
}

// ----------------------------------------------------------------------------
// consensus node
// ----------------------------------------------------------------------------

// ConsAddress defines a wrapper around bytes meant to present a consensus node.
// When marshaled to a string or JSON, it uses hex.
type ConsAddress []byte

// ConsAddressFromHex creates a ConsAddress from a hex string.
func ConsAddressFromHex(address string) (addr ConsAddress, err error) {
	bz, err := addressBytesFromHexString(address)
	return ConsAddress(bz), err
}

// GetConsAddress get ConsAddress from pubkey
func GetConsAddress(pubkey cryptotypes.PubKey) ConsAddress {
	return ConsAddress(pubkey.Address())
}

// Equals Returns boolean for whether two ConsAddress are Equal
func (ca ConsAddress) Equals(ca2 Address) bool {
	if ca.Empty() && ca2.Empty() {
		return true
	}

	return bytes.Equal(ca.Bytes(), ca2.Bytes())
}

// Empty Returns boolean for whether an ConsAddress is empty
func (ca ConsAddress) Empty() bool {
	return len(ca) == 0
}

// Marshal returns the raw address bytes. It is needed for protobuf
// compatibility.
func (ca ConsAddress) Marshal() ([]byte, error) {
	return ca, nil
}

// Unmarshal sets the address to the given data. It is needed for protobuf
// compatibility.
func (ca *ConsAddress) Unmarshal(data []byte) error {
	*ca = data
	return nil
}

// MarshalJSON marshals to JSON using hex.
func (ca ConsAddress) MarshalJSON() ([]byte, error) {
	return json.Marshal(ca.String())
}

// MarshalYAML marshals to YAML using hex.
func (ca ConsAddress) MarshalYAML() (interface{}, error) {
	return ca.String(), nil
}

// UnmarshalJSON unmarshals from JSON assuming hex encoding.
func (ca *ConsAddress) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}

	*ca = common.HexToAddress(s).Bytes()

	return nil
}

// UnmarshalYAML unmarshals from YAML assuming hex encoding.
func (ca *ConsAddress) UnmarshalYAML(data []byte) error {
	var s string
	if err := yaml.Unmarshal(data, &s); err != nil {
		return err
	}

	*ca = common.HexToAddress(s).Bytes()

	return nil
}

// Bytes returns the raw address bytes.
func (ca ConsAddress) Bytes() []byte {
	return ca[:]
}

// String implements the Stringer interface.
func (ca ConsAddress) String() string {
	return "0x" + common.Bytes2Hex(ca.Bytes())
}

// HexifyAddressBytes returns a hex representation of address bytes.
// Returns an empty string if the byte slice is 0-length. Returns an error if the hex conversion
// fails or the prefix is empty.
func HexifyAddressBytes(_ string, bs []byte) (string, error) {
	if len(bs) == 0 {
		return "", nil
	}
	return common.Bytes2Hex(bs), nil
}

// MustHexifyAddressBytes returns a hex representation of address bytes.
// Returns an empty sting if the byte slice is 0-length. It panics if the hex conversion
// fails or the prefix is empty.
func MustHexifyAddressBytes(prefix string, bs []byte) string {
	s, err := HexifyAddressBytes(prefix, bs)
	if err != nil {
		panic(err)
	}
	return s
}

// Format implements the fmt.Formatter interface.

func (ca ConsAddress) Format(s fmt.State, verb rune) {
	switch verb {
	case 's':
		s.Write([]byte(ca.String()))
	case 'p':
		s.Write([]byte(fmt.Sprintf("%p", ca)))
	default:
		s.Write([]byte(fmt.Sprintf("%X", ca.Bytes())))
	}
}

// ----------------------------------------------------------------------------
// auxiliary
// ----------------------------------------------------------------------------

var errHexEmptyAddress = errors.New("decoding hex address failed: must provide a non empty address")

// GetFromHex decodes a bytestring from a hex encoded string.
func GetFromHex(hexStr, _ string) ([]byte, error) {
	if len(hexStr) == 0 {
		return nil, errHexEmptyAddress
	}

	bz := common.FromHex(hexStr)

	return bz, nil
}

func addressBytesFromHexString(address string) ([]byte, error) {
	if len(address) == 0 {
		return nil, ErrEmptyHexAddress
	}
	if !common.IsHexAddress(address) {
		return nil, ErrNotHexAddress
	}

	return common.FromHex(address), nil
}
