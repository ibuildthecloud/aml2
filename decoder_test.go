package aml

import (
	"testing"

	"github.com/acorn-io/aml/pkg/schema"
	"github.com/hexops/autogold/v2"
	"github.com/stretchr/testify/require"
)

const testDocument = `
args: {
// Foo
foo: 1
// Foo2
foo: string
}
args: two: 10
args: bar: 1
args: bar: number < 10
x: args.foo + args.bar + args.two
profiles: baz: two: 2
`

func TestUnmarshal(t *testing.T) {
	data := map[string]any{}

	err := Unmarshal([]byte(testDocument), &data, Option{
		PositionalArgs: []any{3},
		Args: map[string]any{
			"bar": 2,
		},
		Profiles: []string{"baz", "missing?"},
	})
	require.NoError(t, err)

	autogold.Expect(map[string]interface{}{
		"x": 7,
	}).Equal(t, data)
}

func TestSchemaUnmarshal(t *testing.T) {
	out := &schema.File{}
	err := Unmarshal([]byte(testDocument), out)
	require.NoError(t, err)

	autogold.Expect(nil).Equal(t, out)
}
