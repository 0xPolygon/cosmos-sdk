package keyring

import (
	"errors"

	signingv1beta1 "cosmossdk.io/api/cosmos/tx/signing/v1beta1"

	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
)

var _ Keyring = NoKeyring{}

var errNoKeyring = errors.New("no keyring configured")

type NoKeyring struct{}

func (k NoKeyring) List() ([]string, error) {
	return nil, errNoKeyring
}

func (k NoKeyring) LookupAddressByKeyName(_ string) ([]byte, error) {
	return nil, errNoKeyring
}

func (k NoKeyring) GetPubKey(_ string) (cryptotypes.PubKey, error) {
	return nil, errNoKeyring
}

func (k NoKeyring) Sign(_ string, _ []byte, _ signingv1beta1.SignMode) ([]byte, error) {
	return nil, errNoKeyring
}
