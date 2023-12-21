package address

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	jsoniter "github.com/json-iterator/go"
	"sigs.k8s.io/yaml"
	"sort"

	"github.com/cometbft/cometbft/crypto"

	"cosmossdk.io/errors"

	"github.com/cosmos/cosmos-sdk/internal/conv"
)

// Len is the length of base addresses
const Len = sha256.Size

// Addressable represents any type from which we can derive an address.
type Addressable interface {
	Address() []byte
}

// Hash creates a new address from address type and key.
// The functions should only be used by new types defining their own address function
// (eg public keys).
func Hash(typ string, key []byte) []byte {
	hasher := sha256.New()
	_, err := hasher.Write(conv.UnsafeStrToBytes(typ))
	// the error always nil, it's here only to satisfy the io.Writer interface
	errors.AssertNil(err)
	th := hasher.Sum(nil)

	hasher.Reset()
	_, err = hasher.Write(th)
	errors.AssertNil(err)
	_, err = hasher.Write(key)
	errors.AssertNil(err)
	return hasher.Sum(nil)
}

// Compose creates a new address based on sub addresses.
func Compose(typ string, subAddresses []Addressable) ([]byte, error) {
	as := make([][]byte, len(subAddresses))
	totalLen := 0
	var err error
	for i := range subAddresses {
		a := subAddresses[i].Address()
		as[i], err = LengthPrefix(a)
		if err != nil {
			return nil, fmt.Errorf("not compatible sub-adddress=%v at index=%d [%w]", a, i, err)
		}
		totalLen += len(as[i])
	}

	sort.Slice(as, func(i, j int) bool { return bytes.Compare(as[i], as[j]) <= 0 })
	key := make([]byte, totalLen)
	offset := 0
	for i := range as {
		copy(key[offset:], as[i])
		offset += len(as[i])
	}
	return Hash(typ, key), nil
}

// Module is a specialized version of a composed address for modules. Each module account
// is constructed from a module name and a sequence of derivation keys (at least one
// derivation key must be provided). The derivation keys must be unique
// in the module scope, and is usually constructed from some object id. Example, let's
// a x/dao module, and a new DAO object, it's address would be:
//
//	address.Module(dao.ModuleName, newDAO.ID)
func Module(moduleName string, derivationKeys ...[]byte) []byte {
	mKey := []byte(moduleName)
	if len(derivationKeys) == 0 { // fallback to the "traditional" ModuleAddress
		return crypto.AddressHash(mKey)
	}
	// need to append zero byte to avoid potential clash between the module name and the first
	// derivation key
	mKey = append(mKey, 0)
	addr := Hash("module", append(mKey, derivationKeys[0]...))
	for _, k := range derivationKeys[1:] {
		addr = Derive(addr, k)
	}
	return addr
}

// Derive derives a new address from the main `address` and a derivation `key`.
// This function is used to create a sub accounts. To create a module accounts use the
// `Module` function.
func Derive(address, key []byte) []byte {
	return Hash(conv.UnsafeBytesToStr(address), key)
}

// TODO HV2 move these types to heimdall?
// HeimdallHash represents heimdall address
type HeimdallHash common.Hash

// ZeroHeimdallHash represents zero address
var ZeroHeimdallHash = HeimdallHash{}

// EthHash get eth hash
func (aa HeimdallHash) EthHash() common.Hash {
	return common.Hash(aa)
}

// Equals returns boolean for whether two HeimdallHash are Equal
func (aa HeimdallHash) Equals(aa2 HeimdallHash) bool {
	if aa.Empty() && aa2.Empty() {
		return true
	}

	return bytes.Equal(aa.Bytes(), aa2.Bytes())
}

// Empty returns boolean for whether an AccAddress is empty
func (aa HeimdallHash) Empty() bool {
	return bytes.Equal(aa.Bytes(), ZeroHeimdallHash.Bytes())
}

// Marshal returns the raw address bytes. It is needed for protobuf
// compatibility.
func (aa HeimdallHash) Marshal() ([]byte, error) {
	return aa.Bytes(), nil
}

// Unmarshal sets the address to the given data. It is needed for protobuf
// compatibility.
func (aa *HeimdallHash) Unmarshal(data []byte) error {
	*aa = HeimdallHash(common.BytesToHash(data))
	return nil
}

// MarshalJSON marshals to JSON using Bech32.
func (aa HeimdallHash) MarshalJSON() ([]byte, error) {
	return jsoniter.ConfigFastest.Marshal(aa.String())
}

// MarshalYAML marshals to YAML using Bech32.
func (aa HeimdallHash) MarshalYAML() (interface{}, error) {
	return aa.String(), nil
}

// UnmarshalJSON unmarshals from JSON assuming Bech32 encoding.
func (aa *HeimdallHash) UnmarshalJSON(data []byte) error {
	var s string
	if err := jsoniter.ConfigFastest.Unmarshal(data, &s); err != nil {
		return err
	}

	*aa = HexToHeimdallHash(s)

	return nil
}

// UnmarshalYAML unmarshals from JSON assuming Bech32 encoding.
func (aa *HeimdallHash) UnmarshalYAML(data []byte) error {
	var s string
	if err := yaml.Unmarshal(data, &s); err != nil {
		return err
	}

	*aa = HexToHeimdallHash(s)

	return nil
}

// Bytes returns the raw address bytes.
func (aa HeimdallHash) Bytes() []byte {
	return aa[:]
}

// String implements the Stringer interface.
func (aa HeimdallHash) String() string {
	if aa.Empty() {
		return ""
	}

	return "0x" + hex.EncodeToString(aa.Bytes())
}

// Hex returns hex string
func (aa HeimdallHash) Hex() string {
	return aa.String()
}

// Format implements the fmt.Formatter interface.
// nolint: errcheck
func (aa HeimdallHash) Format(s fmt.State, verb rune) {
	switch verb {
	case 's':
		s.Write([]byte(aa.String()))
	case 'p':
		s.Write([]byte(fmt.Sprintf("%p", aa)))
	default:
		s.Write([]byte(fmt.Sprintf("%X", aa.Bytes())))
	}
}

//
// hash utils
//

// BytesToHeimdallHash returns Address with value b.
func BytesToHeimdallHash(b []byte) HeimdallHash {
	return HeimdallHash(common.BytesToHash(b))
}

// HexToHeimdallHash returns Address with value b.
func HexToHeimdallHash(b string) HeimdallHash {
	return HeimdallHash(common.HexToHash(b))
}
