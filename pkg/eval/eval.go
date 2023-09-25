package eval

import (
	"context"

	"github.com/acorn-io/aml/pkg/value"
)

func Eval(ctx context.Context, expr Expression) (value.Value, bool, error) {
	scope := Builtin.Push(nil, ScopeOption{
		Context: ctx,
	})
	return expr.ToValue(scope)
}
