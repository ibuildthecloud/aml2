package aml

import (
	"testing"

	"github.com/acorn-io/aml/pkg/schema"
	"github.com/hexops/autogold/v2"
	"github.com/stretchr/testify/require"
)

func TestUnmarshal(t *testing.T) {
	data := map[string]any{}

	err := Unmarshal([]byte(`
args: foo: 1
args: two: 10
args: bar: 1
x: args.foo + args.bar + args.two
profiles: baz: two: 2
`), &data, Option{
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
	out := &schema.Object{}
	err := Unmarshal([]byte(`
// This is an object
{
	// This is a field
	foo: string != "" || number
}
`), out, Option{ParseAsSchema: true})
	require.NoError(t, err)

	autogold.Expect(&schema.Object{Fields: []schema.Field{
		{
			Name:        "foo",
			Description: "This is a field",
			Type: schema.FieldType{
				Kind: "string",
				Constraint: []schema.Constraint{{
					Op:    "!=",
					Right: "",
				}},
			},
			Union: []schema.FieldType{{Kind: "number"}},
		},
	}}).Equal(t, out)
}
