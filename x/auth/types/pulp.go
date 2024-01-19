package types

import (
	"encoding/hex"
	"errors"
	"reflect"

	"github.com/ethereum/go-ethereum/rlp"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth/migrations/legacytx"
)

// TODO HV2: this is imported from heimdall, but not used anywhere (not heimdall, not cosmos). Can we delete it?
const (
	// PulpHashLength pulp hash length
	PulpHashLength int = 4
)

// Pulp codec for RLP
type Pulp struct {
	typeInfos map[string]reflect.Type
}

// GetPulpHash returns string hash
func GetPulpHash(msg sdk.Msg) []byte {
	// TODO HV2: msg.Route() and msg.Type() unavailable in cosmos.
	//  Anyway this function is not used anywhere in heimdall, hence I believe it can be deleted.
	// return crypto.Keccak256([]byte(fmt.Sprintf("%s::%s", msg.Route(), msg.Type())))[:PulpHashLength]
	return nil
}

// RegisterConcrete should be used to register concrete types that will appear in
// interface fields/elements to be encoded/decoded by pulp.
func (p *Pulp) RegisterConcrete(msg sdk.Msg) {
	rtype := reflect.TypeOf(msg)
	p.typeInfos[hex.EncodeToString(GetPulpHash(msg))] = rtype
}

// GetMsgTxInstance get new instance associated with base tx
func (p *Pulp) GetMsgTxInstance(hash []byte) interface{} {
	rtype := p.typeInfos[hex.EncodeToString(hash[:PulpHashLength])]

	return reflect.New(rtype).Elem().Interface().(sdk.Msg)
}

// EncodeToBytes encodes msg to bytes
func (p *Pulp) EncodeToBytes(tx legacytx.StdTx) ([]byte, error) {
	msg := tx.GetMsgs()[0]

	txBytes, err := rlp.EncodeToBytes(tx)
	if err != nil {
		return nil, err
	}

	return append(GetPulpHash(msg), txBytes[:]...), nil
}

// DecodeBytes decodes bytes to msg
func (p *Pulp) DecodeBytes(data []byte) (interface{}, error) {
	var txRaw legacytx.StdTxRaw

	if len(data) <= PulpHashLength {
		return nil, errors.New("Invalid data length, should be greater than PulpPrefix")
	}

	if err := rlp.DecodeBytes(data[PulpHashLength:], &txRaw); err != nil {
		return nil, err
	}

	rtype := p.typeInfos[hex.EncodeToString(data[:PulpHashLength])]
	newMsg := reflect.New(rtype).Interface()

	if err := rlp.DecodeBytes(txRaw.Msg[:], newMsg); err != nil {
		return nil, err
	}

	// change pointer to non-pointer
	vptr := reflect.New(reflect.TypeOf(newMsg).Elem()).Elem()
	vptr.Set(reflect.ValueOf(newMsg).Elem())
	// return vptr.Interface(), nil

	return legacytx.StdTx{
		Msgs:       []sdk.Msg{vptr.Interface().(sdk.Msg)},
		Signatures: []legacytx.StdSignature{txRaw.Signature},
		Memo:       txRaw.Memo,
	}, nil
}
