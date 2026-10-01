package unknownproto

import (
	"testing"

	"github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/testutil/testdata"
	"github.com/stretchr/testify/require"
)

// nestedTestVersion2 chains depth self-referential wraps via TestVersion2.C (a repeated
// message field of the same type), yielding a message whose traversal recursion depth is
// exactly depth.
func nestedTestVersion2(depth int) *testdata.TestVersion2 {
	cur := &testdata.TestVersion2{}
	for range depth {
		cur = &testdata.TestVersion2{C: []*testdata.TestVersion2{cur}}
	}
	return cur
}

// nestedTestVersion2ViaAny chains depth self-referential wraps via TestVersion2.G, a real
// google.protobuf.Any field (unlike nestedTestVersion2's ordinary C field) -- directly
// exercising the self-chained-Any traversal path RejectUnknownFieldsWithRecursionLimit's own
// doc comment names as its motivating use case, which nestedTestVersion2 alone does not touch.
func nestedTestVersion2ViaAny(depth int) *testdata.TestVersion2 {
	cur := &testdata.TestVersion2{}
	for range depth {
		cur = &testdata.TestVersion2{G: &types.Any{
			TypeUrl: "/testpb.TestVersion2",
			Value:   mustMarshal(cur),
		}}
	}
	return cur
}

// TestRejectUnknownFieldsWithRecursionLimit pins the caller-supplied recursion-limit
// behavior RejectUnknownFieldsWithRecursionLimit adds on top of RejectUnknownFields/
// RejectUnknownFieldsStrict (both of which keep using the unchanged package default,
// verified by every other test in this package still passing): a limit below the actual
// traversal depth stops it with the exported ErrRecursionLimitReached sentinel, distinguishable
// via errors.Is from any other error this function can return, while a limit at or above the
// actual depth succeeds exactly like the unbounded-default entry points would.
func TestRejectUnknownFieldsWithRecursionLimit(t *testing.T) {
	shallow := mustMarshal(nestedTestVersion2(2))
	deep := mustMarshal(nestedTestVersion2(5))

	_, err := RejectUnknownFieldsWithRecursionLimit(shallow, new(testdata.TestVersion2), false, DefaultAnyResolver{}, 3)
	require.NoError(t, err, "a limit above the actual recursion depth must not be tripped")

	hasUnknownNonCriticals, err := RejectUnknownFieldsWithRecursionLimit(deep, new(testdata.TestVersion2), false, DefaultAnyResolver{}, 3)
	require.Error(t, err, "a limit below the actual recursion depth must stop traversal")
	require.ErrorIs(t, err, ErrRecursionLimitReached)
	require.False(t, hasUnknownNonCriticals, "hitting the recursion limit carries no information about non-critical fields")

	// The exact same bytes must still succeed under RejectUnknownFields' unbounded-by-default
	// limit -- the new entry point only tightens the bound for callers that opt in.
	_, err = RejectUnknownFields(deep, new(testdata.TestVersion2), false, DefaultAnyResolver{})
	require.NoError(t, err, "RejectUnknownFields' own default limit must remain unchanged")
}

// TestRejectUnknownFieldsWithRecursionLimit_ExactBoundary pins the tightest possible margin
// on both sides: a limit exactly equal to the actual recursion depth must still succeed (not
// just a limit comfortably above it), and one less than that must still fail. This closes a
// gap the wider-margin cases above leave open -- they'd both still pass under an off-by-one
// mutant (e.g. comparing recursionLimit == 1 instead of <= 0) that only misclassifies inputs
// exactly at this boundary.
func TestRejectUnknownFieldsWithRecursionLimit_ExactBoundary(t *testing.T) {
	atDepth3 := mustMarshal(nestedTestVersion2(3))

	_, err := RejectUnknownFieldsWithRecursionLimit(atDepth3, new(testdata.TestVersion2), false, DefaultAnyResolver{}, 3)
	require.NoError(t, err, "a limit exactly equal to the actual recursion depth must succeed")

	_, err = RejectUnknownFieldsWithRecursionLimit(atDepth3, new(testdata.TestVersion2), false, DefaultAnyResolver{}, 2)
	require.ErrorIs(t, err, ErrRecursionLimitReached, "one less than the actual recursion depth must fail")
}

// TestRejectUnknownFieldsWithRecursionLimit_AnyChain exercises the self-chained-Any path
// directly (nestedTestVersion2ViaAny), rather than only the ordinary-message-field path
// nestedTestVersion2 covers -- proving the recursion limit is enforced identically whether a
// level is reached via an ordinary field or via resolving a further google.protobuf.Any, which
// is the scenario this function's own doc comment names as its motivating use case.
func TestRejectUnknownFieldsWithRecursionLimit_AnyChain(t *testing.T) {
	shallow := mustMarshal(nestedTestVersion2ViaAny(2))
	deep := mustMarshal(nestedTestVersion2ViaAny(5))

	_, err := RejectUnknownFieldsWithRecursionLimit(shallow, new(testdata.TestVersion2), false, DefaultAnyResolver{}, 3)
	require.NoError(t, err, "a limit above the actual Any-chain depth must not be tripped")

	_, err = RejectUnknownFieldsWithRecursionLimit(deep, new(testdata.TestVersion2), false, DefaultAnyResolver{}, 3)
	require.ErrorIs(t, err, ErrRecursionLimitReached, "a limit below the actual Any-chain depth must stop traversal")
}

// TestRejectUnknownFieldsWithRecursionLimit_NonPositiveLimit is the regression test for a
// real bug: doRejectUnknownFields' cap used to check recursionLimit == 0 specifically, so a
// caller passing a negative limit (e.g. from a config value or arithmetic bug) would never
// hit it -- recursionLimit only ever decreases, so it steps past zero without ever landing on
// it exactly, silently falling through to unbounded traversal on non-empty input. Fixed by
// checking recursionLimit <= 0. Zero and a representative negative value must both reject
// non-empty input immediately, matching "no recursion budget" rather than "no bound at all".
func TestRejectUnknownFieldsWithRecursionLimit_NonPositiveLimit(t *testing.T) {
	shallow := mustMarshal(nestedTestVersion2(1))

	for _, limit := range []int{0, -1, -1000} {
		_, err := RejectUnknownFieldsWithRecursionLimit(shallow, new(testdata.TestVersion2), false, DefaultAnyResolver{}, limit)
		require.ErrorIs(t, err, ErrRecursionLimitReached, "limit=%d", limit)
	}
}
