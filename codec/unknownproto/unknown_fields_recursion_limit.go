package unknownproto

import (
	"errors"

	"github.com/cosmos/gogoproto/jsonpb"
	"github.com/cosmos/gogoproto/proto"
)

// defaultRecursionLimit matches https://github.com/protocolbuffers/protobuf-go/blob/v1.35.2/encoding/protowire/wire.go#L28.
const defaultRecursionLimit = 10_000

// ErrRecursionLimitReached is returned when traversal hits its recursion limit before
// finishing, whether that limit is the package default or a caller-supplied bound via
// RejectUnknownFieldsWithRecursionLimit. Exported as a sentinel (rather than only a formatted
// string) so callers that intentionally pass a tight limit -- to bound this function's own
// traversal cost below its unbounded-by-default 10,000, e.g. as a cheap pre-check for
// pathologically self-chained google.protobuf.Any nesting before paying the real cost of a
// full decode -- can distinguish "traversal was cut short by the bound" from every other
// unknown-field/type error this function returns, via errors.Is.
var ErrRecursionLimitReached = errors.New("recursion limit reached")

// RejectUnknownFieldsWithRecursionLimit behaves exactly like RejectUnknownFields, except the
// caller supplies the recursion limit instead of the package default (10,000). Every
// consensus-affecting caller in this SDK should keep using RejectUnknownFields/
// RejectUnknownFieldsStrict -- this entry point exists for callers that need this function's
// own real traversal (the exact code path a full decode already runs, resolving each
// google.protobuf.Any via the same resolver and recursing through message-typed fields
// exactly as the schema declares) bounded well below that default, as a cheap, tightly-capped
// pre-check ahead of the real, unbounded-by-default decode -- e.g. rejecting a transaction
// whose Any nesting (self-chained directly or via an intervening ordinary message field) would
// otherwise cost this function's own uncapped recursion to reject, before paying that cost.
// Returns ErrRecursionLimitReached specifically when recursionLimit is what stopped traversal,
// distinguishable via errors.Is from every other error this function can return.
//
// x/tx/decode carries an independent, protoreflect-based reimplementation of this same
// unknown-field/Any-traversal algorithm with its own hardcoded, unpatched 10,000 limit -- this
// function does not touch it. Callers relying on a tight recursion bound as a DoS pre-check
// must confirm their app's TxConfig actually routes through x/auth/tx's DefaultTxDecoder (this
// package), not x/tx/decode's Decoder; if that ever changes, the same unbounded-recursion cost
// this function exists to bound reopens silently through that other path.
//
// recursionLimit <= 0 rejects immediately (ErrRecursionLimitReached) on any non-empty input,
// matching the intent of "no recursion budget" rather than silently falling through to
// unbounded traversal. Choose recursionLimit generously enough for every legitimate message
// shape your application can produce: too tight a bound rejects real, valid traffic exactly
// like it rejects an attack, with no way for a caller to tell the two apart from this
// function's return value alone.
func RejectUnknownFieldsWithRecursionLimit(bz []byte, msg proto.Message, allowUnknownNonCriticals bool, resolver jsonpb.AnyResolver, recursionLimit int) (hasUnknownNonCriticals bool, err error) {
	return doRejectUnknownFields(bz, msg, allowUnknownNonCriticals, resolver, recursionLimit)
}
