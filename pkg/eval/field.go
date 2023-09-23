package eval

import (
	"github.com/acorn-io/aml/pkg/schema"
	"github.com/acorn-io/aml/pkg/value"
)

// type assertions
var (
	_ Field = (*KeyValue)(nil)
)

type Field interface {
	Expression
	Keys(scope Scope) ([]string, error)
	DescribeFields(ctx value.SchemaContext, scope Scope) ([]schema.Field, error)
	// ToValueForKey should return the value (right hand side) for this key. If the key evaluates to undefined
	// then (nil, false, nil) should be returned. If the value (right hand side) evaluates to undefined, undefined
	// should be returned
	ToValueForKey(scope Scope, key string) (value.Value, bool, error)
}

type KeyValue struct {
	Comments Comments
	Key      FieldKey
	Value    Expression
	Pos      Position
	Local    bool
	Optional bool
}

func (k *KeyValue) DescribeFields(ctx value.SchemaContext, scope Scope) ([]schema.Field, error) {
	key, ok, err := k.Key.ToString(scope)
	if err != nil {
		return nil, err
	} else if !ok {
		return nil, nil
	}

	v, ok, err := k.getValueValue(scope, key)
	if err != nil {
		return nil, err
	} else if !ok {
		return nil, nil
	}

	ft, err := value.DescribeFieldType(ctx, v)
	if err != nil {
		return nil, err
	}

	return []schema.Field{
		{
			Name:        key,
			Match:       k.Key.IsMatch(),
			Description: k.Comments.Last(),
			Optional:    k.Optional,
			Type:        ft,
		},
	}, nil
}

func (k *KeyValue) ToValueForKey(scope Scope, key string) (value.Value, bool, error) {
	if ok, err := k.Key.Matches(scope, key); err != nil || !ok {
		return nil, ok, err
	}
	return k.getValueValue(scope, key)
}

func (k *KeyValue) Keys(scope Scope) ([]string, error) {
	if k.Optional || k.Local || k.Key.IsMatch() {
		return nil, nil
	}
	s, ok, err := k.Key.ToString(scope)
	if err != nil || !ok {
		return nil, err
	}
	return []string{s}, nil
}

func (k *KeyValue) getValueValue(scope Scope, key string) (ret value.Value, _ bool, _ error) {
	scope = scope.Push(nil, ScopeOption{
		Path: key,
	})
	v, ok, err := k.Value.ToValue(scope)
	if err != nil || !ok {
		return nil, ok, err
	}
	if value.IsSimpleKind(v.Kind()) && scope.IsSchema() {
		return value.NewDefault(v), true, nil
	}
	return v, true, nil
}

func (k *KeyValue) IsArgumentDefinition() bool {
	if v, ok := k.Key.Key.(Value); ok {
		if s, ok := v.Value.(value.String); ok {
			return string(s) == "args" || string(s) == "profiles"
		}
	}
	return false
}

func (k *KeyValue) ToValue(scope Scope) (value.Value, bool, error) {
	if k.Local || k.Key.IsMatch() {
		return nil, false, nil
	}

	var (
		v   value.Value
		ok  bool
		err error
	)

	keyValue, ok, err := k.Key.Key.ToValue(scope)
	if err != nil || !ok || keyValue.Kind() == value.UndefinedKind {
		return keyValue, ok, err
	}

	key, err := value.ToString(keyValue)

	v, ok, err = k.getValueValue(scope, key)
	if err != nil || !ok {
		return nil, ok, err
	}

	return &value.Object{
		Entries: []value.Entry{{
			Key:   key,
			Value: v,
		}},
	}, true, nil
}

func FieldsToValues(scope Scope, fields []Field) (result []value.Value, _ error) {
	for _, field := range fields {
		v, ok, err := field.ToValue(scope)
		if err != nil {
			return nil, err
		} else if !ok {
			continue
		}
		result = append(result, v)
	}
	return
}

type FieldKey struct {
	Match Expression
	Key   Expression
	Pos   Position
}

func (k *FieldKey) IsMatch() bool {
	return k.Match != nil
}

func (k *FieldKey) ToString(scope Scope) (string, bool, error) {
	source := k.Key
	if k.IsMatch() {
		source = k.Match
	}

	v, ok, err := source.ToValue(scope)
	if err != nil || !ok {
		return "", ok, err
	}

	s, err := value.ToString(v)
	return s, true, err
}

func (k *FieldKey) Matches(scope Scope, key string) (_ bool, returnErr error) {
	if k.IsMatch() {
		v, ok, err := k.Match.ToValue(scope)
		if err != nil || !ok {
			return ok, err
		}
		return value.Match(v, value.NewValue(key))
	}

	v, ok, err := k.Key.ToValue(scope)
	if err != nil || !ok || v.Kind() == value.UndefinedKind {
		return false, err
	}

	keyPattern, err := value.ToString(v)
	if err != nil || !ok {
		return ok, err
	}

	return keyPattern == key, nil
}
