package secp256k1

import (
	"bytes"
	"crypto/sha256"
	"crypto/subtle"
	"fmt"
	"io"
	"math/big"

	"github.com/cometbft/cometbft/crypto"
	"github.com/decred/dcrd/dcrec/secp256k1/v4"
	ethCrypto "github.com/ethereum/go-ethereum/crypto"

	errorsmod "cosmossdk.io/errors"

	"github.com/cosmos/cosmos-sdk/codec"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/cosmos/cosmos-sdk/types/errors"
)

var (
	_ cryptotypes.PrivKey  = &PrivKey{}
	_ cryptotypes.PrivKey  = &PrivKeyOld{}
	_ codec.AminoMarshaler = &PrivKey{}
	_ codec.AminoMarshaler = &PrivKeyOld{}
)

const (
	PrivKeySize    = 32
	keyType        = "secp256k1"
	PrivKeyNameOld = "tendermint/PrivKeySecp256k1"
	PubKeyNameOld  = "tendermint/PubKeySecp256k1"
	PrivKeyName    = "cometbft/PrivKeySecp256k1eth"
	PubKeyName     = "cometbft/PubKeySecp256k1eth"
)

func (privKey *PrivKeyOld) Bytes() []byte {
	return privKey.Key
}

func (privKey *PrivKeyOld) PubKey() cryptotypes.PubKey {
	privateObject, err := ethCrypto.ToECDSA(privKey.Key)
	if err != nil {
		panic(err)
	}

	pk := ethCrypto.FromECDSAPub(&privateObject.PublicKey)
	return &PubKey{Key: pk}
}

func (privKey *PrivKeyOld) Equals(other cryptotypes.LedgerPrivKey) bool {
	return privKey.Type() == other.Type() && subtle.ConstantTimeCompare(privKey.Bytes(), other.Bytes()) == 1
}

func (privKey *PrivKeyOld) Type() string {
	return keyType
}

func (privKey *PrivKeyOld) Sign(msg []byte) ([]byte, error) {
	privateObject, err := ethCrypto.ToECDSA(privKey.Key)
	if err != nil {
		return nil, err
	}

	return ethCrypto.Sign(ethCrypto.Keccak256(msg), privateObject)
}

// MarshalAmino overrides Amino binary marshaling.
func (privKey PrivKeyOld) MarshalAmino() ([]byte, error) {
	return privKey.Key, nil
}

// UnmarshalAmino overrides Amino binary marshaling.
func (privKey *PrivKeyOld) UnmarshalAmino(bz []byte) error {
	if len(bz) != PrivKeySize {
		return fmt.Errorf("invalid privkey size")
	}
	privKey.Key = bz

	return nil
}

// MarshalAminoJSON overrides Amino JSON marshaling.
func (privKey PrivKeyOld) MarshalAminoJSON() ([]byte, error) {
	// When we marshal to Amino JSON, we don't marshal the "key" field itself,
	// just its contents (i.e. the key bytes).
	return privKey.MarshalAmino()
}

// UnmarshalAminoJSON overrides Amino JSON marshaling.
func (privKey *PrivKeyOld) UnmarshalAminoJSON(bz []byte) error {
	return privKey.UnmarshalAmino(bz)
}

// Bytes returns the byte representation of the Private Key.
func (privKey *PrivKey) Bytes() []byte {
	return privKey.Key
}

// PubKey performs the point-scalar multiplication from the privKey on the
// generator point to get the pubkey.
func (privKey *PrivKey) PubKey() cryptotypes.PubKey {
	privateObject, err := ethCrypto.ToECDSA(privKey.Key)
	if err != nil {
		panic(err)
	}

	pk := ethCrypto.FromECDSAPub(&privateObject.PublicKey)
	return &PubKey{Key: pk}
}

// Equals - you probably don't need to use this.
// Runs in constant time based on length of the
func (privKey *PrivKey) Equals(other cryptotypes.LedgerPrivKey) bool {
	return privKey.Type() == other.Type() && subtle.ConstantTimeCompare(privKey.Bytes(), other.Bytes()) == 1
}

func (privKey *PrivKey) Type() string {
	return keyType
}

// MarshalAmino overrides Amino binary marshaling.
func (privKey PrivKey) MarshalAmino() ([]byte, error) {
	return privKey.Key, nil
}

// UnmarshalAmino overrides Amino binary marshaling.
func (privKey *PrivKey) UnmarshalAmino(bz []byte) error {
	if len(bz) != PrivKeySize {
		return fmt.Errorf("invalid privkey size")
	}
	privKey.Key = bz

	return nil
}

// MarshalAminoJSON overrides Amino JSON marshaling.
func (privKey PrivKey) MarshalAminoJSON() ([]byte, error) {
	// When we marshal to Amino JSON, we don't marshal the "key" field itself,
	// just its contents (i.e. the key bytes).
	return privKey.MarshalAmino()
}

// UnmarshalAminoJSON overrides Amino JSON marshaling.
func (privKey *PrivKey) UnmarshalAminoJSON(bz []byte) error {
	return privKey.UnmarshalAmino(bz)
}

// GenPrivKey generates a new ECDSA private key on curve secp256k1 private key.
// It uses OS randomness to generate the private key.
func GenPrivKey() *PrivKey {
	return &PrivKey{Key: genPrivKey(crypto.CReader())}
}

// genPrivKey generates a new secp256k1 private key using the provided reader.
func genPrivKey(rand io.Reader) []byte {
	var privKeyBytes [PrivKeySize]byte
	d := new(big.Int)
	for {
		privKeyBytes = [PrivKeySize]byte{}
		_, err := io.ReadFull(rand, privKeyBytes[:])
		if err != nil {
			panic(err)
		}

		d.SetBytes(privKeyBytes[:])
		// break if we found a valid point (i.e. > 0 and < N == curverOrder)
		isValidFieldElement := 0 < d.Sign() && d.Cmp(secp256k1.S256().N) < 0
		if isValidFieldElement {
			break
		}
	}

	return privKeyBytes[:]
}

var one = new(big.Int).SetInt64(1)

// GenPrivKeyFromSecret hashes the secret with SHA2, and uses
// that 32 byte output to create the private key.
//
// It makes sure the private key is a valid field element by setting:
//
// c = sha256(secret)
// k = (c mod (n − 1)) + 1, where n = curve order.
//
// NOTE: secret should be the output of a KDF like bcrypt,
// if it's derived from user input.
func GenPrivKeyFromSecret(secret []byte) *PrivKey {
	secHash := sha256.Sum256(secret)
	// to guarantee that we have a valid field element, we use the approach of:
	// "Suite B Implementer’s Guide to FIPS 186-3", A.2.1
	// https://apps.nsa.gov/iaarchive/library/ia-guidance/ia-solutions-for-classified/algorithm-guidance/suite-b-implementers-guide-to-fips-186-3-ecdsa.cfm
	// see also https://github.com/golang/go/blob/0380c9ad38843d523d9c9804fe300cb7edd7cd3c/src/crypto/ecdsa/ecdsa.go#L89-L101
	fe := new(big.Int).SetBytes(secHash[:])
	n := new(big.Int).Sub(secp256k1.S256().N, one)
	fe.Mod(fe, n)
	fe.Add(fe, one)

	feB := fe.Bytes()
	privKey32 := make([]byte, PrivKeySize)
	// copy feB over to fixed 32 byte privKey32 and pad (if necessary)
	copy(privKey32[32-len(feB):32], feB)

	return &PrivKey{Key: privKey32}
}

//-------------------------------------

var (
	_ cryptotypes.PubKey   = &PubKey{}
	_ cryptotypes.PubKey   = &PubKeyOld{}
	_ codec.AminoMarshaler = &PubKey{}
	_ codec.AminoMarshaler = &PubKeyOld{}
)

// PubKeySize (uncompressed) is comprised of 65 bytes for two field elements (x and y)
// and a prefix byte (0x04) to indicate that it is uncompressed.
const PubKeySize = 65

// Address returns an Ethereum style addresses
func (pubKey *PubKey) Address() crypto.Address {
	if len(pubKey.Key) != PubKeySize {
		panic(fmt.Sprintf("length of pubkey is incorrect %d != %d", len(pubKey.Key), PubKeySize))
	}

	return crypto.Address(ethCrypto.Keccak256(pubKey.Key[1:])[12:])
}

// Bytes returns the pubkey byte format.
func (pubKey *PubKey) Bytes() []byte {
	return pubKey.Key
}

func (pubKey *PubKey) String() string {
	return fmt.Sprintf("PubKeySecp256k1{%X}", pubKey.Key)
}

func (pubKey *PubKey) Type() string {
	return keyType
}

func (pubKey *PubKey) Equals(other cryptotypes.PubKey) bool {
	return pubKey.Type() == other.Type() && bytes.Equal(pubKey.Bytes(), other.Bytes())
}

// MarshalAmino overrides Amino binary marshaling.
func (pubKey PubKey) MarshalAmino() ([]byte, error) {
	return pubKey.Key, nil
}

// UnmarshalAmino overrides Amino binary marshaling.
func (pubKey *PubKey) UnmarshalAmino(bz []byte) error {
	if len(bz) != PubKeySize {
		return errorsmod.Wrap(errors.ErrInvalidPubKey, "invalid pubkey size")
	}
	pubKey.Key = bz

	return nil
}

// MarshalAminoJSON overrides Amino JSON marshaling.
func (pubKey PubKey) MarshalAminoJSON() ([]byte, error) {
	// When we marshal to Amino JSON, we don't marshal the "key" field itself,
	// just its contents (i.e. the key bytes).
	return pubKey.MarshalAmino()
}

// UnmarshalAminoJSON overrides Amino JSON marshaling.
func (pubKey *PubKey) UnmarshalAminoJSON(bz []byte) error {
	return pubKey.UnmarshalAmino(bz)
}

// Address returns an ethereum style addresses
func (pubKey *PubKeyOld) Address() crypto.Address {
	if len(pubKey.Key) != PubKeySize {
		panic(fmt.Sprintf("length of pubkey is incorrect %d != %d", len(pubKey.Key), PubKeySize))
	}

	return crypto.Address(ethCrypto.Keccak256(pubKey.Key[1:])[12:])
}

// Bytes returns the pubkey byte format.
func (pubKey *PubKeyOld) Bytes() []byte {
	return pubKey.Key
}

func (pubKey *PubKeyOld) String() string {
	return fmt.Sprintf("PubKeySecp256k1{%X}", pubKey.Key)
}

func (pubKey *PubKeyOld) Type() string {
	return keyType
}

func (pubKey *PubKeyOld) Equals(other cryptotypes.PubKey) bool {
	return pubKey.Type() == other.Type() && bytes.Equal(pubKey.Bytes(), other.Bytes())
}

func (pubKey *PubKeyOld) VerifySignature(msg, sigStr []byte) bool {
	if len(sigStr) != SigSize {
		return false
	}

	hash := ethCrypto.Keccak256(msg)
	return ethCrypto.VerifySignature(pubKey.Key, hash, sigStr[:64])
}

// MarshalAmino overrides Amino binary marshaling.
func (pubKey PubKeyOld) MarshalAmino() ([]byte, error) {
	return pubKey.Key, nil
}

// UnmarshalAmino overrides Amino binary marshaling.
func (pubKey *PubKeyOld) UnmarshalAmino(bz []byte) error {
	if len(bz) != PubKeySize {
		return errorsmod.Wrap(errors.ErrInvalidPubKey, "invalid pubkey size")
	}
	pubKey.Key = bz

	return nil
}

// MarshalAminoJSON overrides Amino JSON marshaling.
func (pubKey PubKeyOld) MarshalAminoJSON() ([]byte, error) {
	// When we marshal to Amino JSON, we don't marshal the "key" field itself,
	// just its contents (i.e. the key bytes).
	return pubKey.MarshalAmino()
}

// UnmarshalAminoJSON overrides Amino JSON marshaling.
func (pubKey *PubKeyOld) UnmarshalAminoJSON(bz []byte) error {
	return pubKey.UnmarshalAmino(bz)
}
