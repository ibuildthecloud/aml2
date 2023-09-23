package eval

import (
	"fmt"

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

func (e *Embedded) DescribeFields(ctx value.SchemaContext, scope Scope) ([]schema.Field, error) {
	scope = scope.Push(nil, ScopeOption{
		Path: fmt.Sprintf("embedded.%d", ctx.GetIndex()),
	})
	v, ok, err := e.ToValue(scope)
	if err != nil {
		return nil, err
	} else if !ok {
		return nil, nil
	}

	t := value.TargetKind(v)
	if t != value.ObjectKind {
		return nil, fmt.Errorf("embedded expressions in schemas must result in an object not %s", t)
	}

	obj, err := value.DescribeObject(ctx, v)
	if err != nil {
		return nil, err
	}

	return obj.Fields, nil
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
	v, ok, err := e.Expression.ToValue(scope)
	if err != nil || !ok {
		return nil, ok, err
	} else if t := value.TargetKind(v); scope.IsSchema() && t != value.ObjectKind && t != value.UndefinedKind {
		return nil, false, fmt.Errorf("in schemas embedded expressions must evaluate to kind object, not %s", t)
	}
	return v, true, nil
}
