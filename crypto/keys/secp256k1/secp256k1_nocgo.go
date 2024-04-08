//go:build !libsecp256k1_sdk
// +build !libsecp256k1_sdk

package secp256k1

import (
	"errors"
	ethCrypto "github.com/ethereum/go-ethereum/crypto"

	secp256k1 "github.com/decred/dcrd/dcrec/secp256k1/v4"
	"github.com/decred/dcrd/dcrec/secp256k1/v4/ecdsa"
)

// SigSize is the size of the ECDSA signature.
const SigSize = 65

// Sign creates an ECDSA signature on curve Secp256k1, using SHA256 on the msg.
// The returned signature will be of the form R || S (in lower-S form).
func (privKey *PrivKey) Sign(msg []byte) ([]byte, error) {
	privateObject, err := ethCrypto.ToECDSA(privKey.Key)
	if err != nil {
		return nil, err
	}

	return ethCrypto.Sign(ethCrypto.Keccak256(msg), privateObject)
}

// VerifySignature verifies a signature of the form R || S || V.
// It rejects signatures which are not in lower-S form.
func (pubKey *PubKey) VerifySignature(msg []byte, sigStr []byte) bool {
	if len(sigStr) != SigSize {
		return false
	}

	hash := ethCrypto.Keccak256(msg)
	return ethCrypto.VerifySignature(pubKey.Key, hash, sigStr[:64])
}

// Read Signature struct from R || S. Caller needs to ensure
// that len(sigStr) == 64.
// Rejects malleable signatures (if S value if it is over half order).
func signatureFromBytes(sigStr []byte) (*ecdsa.Signature, error) {
	var r secp256k1.ModNScalar
	r.SetByteSlice(sigStr[:32])
	var s secp256k1.ModNScalar
	s.SetByteSlice(sigStr[32:64])
	if s.IsOverHalfOrder() {
		return nil, errors.New("signature is not in lower-S form")
	}

	return ecdsa.NewSignature(&r, &s), nil
}
