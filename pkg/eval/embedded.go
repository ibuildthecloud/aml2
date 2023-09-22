package eval

import (
	"github.com/acorn-io/aml/pkg/schema"
	"github.com/acorn-io/aml/pkg/value"
)

// type assertions
var (
	_ Field = (*Embedded)(nil)
)

type Embedded struct {
	Comments   Comments
	Expression Expression
}

func (e *Embedded) GetFields(ctx value.SchemaContext, scope Scope) ([]schema.Field, error) {
	v, ok, err := e.ToValue(scope)
	if err != nil {
		return nil, err
	} else if !ok {
		return nil, nil
	}

	return getFields(ctx, v)
}

func (e *Embedded) IsPositionalArgument() bool {
	return true
}

func (e *Embedded) ToValueForKey(scope Scope, key string) (value.Value, bool, error) {
	v, ok, err := e.ToValue(scope)
	if err != nil || !ok {
		return nil, ok, err
	} else if v.Kind() == value.UndefinedKind {
		return nil, false, nil
	}
	return value.Lookup(v, value.NewValue(key))
}

func (e *Embedded) Keys(scope Scope) ([]string, error) {
	v, ok, err := e.ToValue(scope)
	if err != nil {
		return nil, err
	} else if !ok {
		return nil, nil
	}
	return value.Keys(v)
}

func (e *Embedded) ToValue(scope Scope) (value.Value, bool, error) {
	return e.Expression.ToValue(scope)
}
