package value

import (
	"fmt"
	"testing"

	"github.com/hexops/autogold/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUnary(t *testing.T) {
	tests := []struct {
		op     string
		val    any
		expect autogold.Value
	}{
		{op: "+", val: 1, expect: autogold.Expect(Number("1"))},
		{op: "-", val: 1, expect: autogold.Expect(Number("-1"))},
		{op: "-", val: Number("4"), expect: autogold.Expect(Number("-4"))},
		{op: "!", val: false, expect: autogold.Expect(true)},
	}

	for i, test := range tests {
		t.Run(fmt.Sprintf("%s%d", t.Name(), i), func(t *testing.T) {
			v, err := UnaryOperation(test.op, NewValue(test.val))
			require.NoError(t, err)
			test.expect.Equal(t, v.NativeValue())
		})
	}
}

func TestIndex(t *testing.T) {
	v, ok, err := Index(NewValue([]any{
		"key", "key2",
	}), 1)
	require.NoError(t, err)
	assert.True(t, ok)
	assert.Equal(t, v.NativeValue(), "key2")
}

func TestSlice(t *testing.T) {
	v, ok, err := Slice(NewValue([]any{
		"key", "key2", "key3",
	}), 1, 3)
	require.NoError(t, err)
	assert.True(t, ok)
	assert.Equal(t, v.NativeValue(), []any{"key2", "key3"})
}

func TestLookup(t *testing.T) {
	v, ok, err := Lookup(NewValue(map[string]any{
		"key": "value",
	}), "key")
	require.NoError(t, err)
	assert.True(t, ok)
	assert.Equal(t, v.NativeValue(), "value")
}

func TestBinary(t *testing.T) {
	tests := []struct {
		op     string
		left   any
		right  any
		expect autogold.Value
	}{
		{op: "*", left: 2, right: 3, expect: autogold.Expect(Number("6"))},
		{op: "*", left: 2.0, right: 3, expect: autogold.Expect(Number("6"))},
		{op: "*", left: 0.1, right: 30, expect: autogold.Expect(Number("3"))},
		{op: "/", left: 6, right: 2, expect: autogold.Expect(Number("3"))},
		{op: "&&", left: false, right: true, expect: autogold.Expect(false)},
		{op: "||", left: false, right: true, expect: autogold.Expect(true)},
		{op: "<", left: 3, right: 4, expect: autogold.Expect(true)},
		{op: "<=", left: 4, right: 4, expect: autogold.Expect(true)},
		{op: ">", left: 3, right: 4, expect: autogold.Expect(false)},
		{op: ">=", left: 4, right: 4, expect: autogold.Expect(true)},
		{op: "==", left: 1, right: 1, expect: autogold.Expect(true)},
		{op: "==", left: true, right: true, expect: autogold.Expect(true)},
		{op: "==", left: "x", right: "x", expect: autogold.Expect(true)},
		{op: "==", left: nil, right: nil, expect: autogold.Expect(true)},
		{op: "!=", left: 1, right: 1, expect: autogold.Expect(false)},
		{op: "!=", left: true, right: true, expect: autogold.Expect(false)},
		{op: "!=", left: "x", right: "x", expect: autogold.Expect(false)},
		{op: "!=", left: nil, right: nil, expect: autogold.Expect(false)},
	}

	for i, test := range tests {
		t.Run(fmt.Sprintf("%s%d", t.Name(), i), func(t *testing.T) {
			v, err := BinaryOperation(test.op, NewValue(test.left), NewValue(test.right))
			require.NoError(t, err)
			test.expect.Equal(t, v.NativeValue())
		})
	}
}
