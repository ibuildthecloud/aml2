package eval

import (
	"fmt"

	"github.com/acorn-io/aml/value"
)

type Struct struct {
	Comments Comments
	Fields   []Field
}

// type assertions
var (
	_ Field = (*Embedded)(nil)
	_ Field = (*KeyValue)(nil)
)

type Field interface {
	ValueLookup
	Expression
}

type ValueLookup interface {
	Lookup(scope Scope, key string) (value.Value, bool, error)
}

type Scope interface {
	Get(key string) (value.Value, bool, error)
	Push(lookup ValueLookup) Scope
}

type MapLookup map[string]value.Value

func (m MapLookup) Lookup(_ Scope, key string) (value.Value, bool, error) {
	ret, ok := m[key]
	return ret, ok, nil
}

type NativeMapLookup map[string]any

func (m NativeMapLookup) Lookup(_ Scope, key string) (value.Value, bool, error) {
	ret, ok := m[key]
	return value.NewValue(ret), ok, nil
}

func (s *Struct) Lookup(scope Scope, key string) (value.Value, bool, error) {
	var (
		last  value.Value
		found bool
	)

	scope = scope.Push(s)
	for _, field := range s.Fields {
		val, ok, err := field.Lookup(scope, key)
		if err != nil {
			return nil, false, err
		}
		if !ok {
			continue
		}
		found = true
		if last == nil {
			last = val
		} else {
			last, err = val.Merge(val)
			if err != nil {
				return nil, false, err
			}
		}
	}

	return last, found, nil
}

func (s *Struct) ToValue(scope Scope) (value.Value, bool, error) {
	var (
		last  value.Value = &value.Null{}
		found             = false
	)

	scope = scope.Push(s)
	for _, field := range s.Fields {
		val, ok, err := field.ToValue(scope)
		if err != nil {
			return nil, false, err
		}
		if !ok {
			continue
		}
		found = true
		if last == nil {
			last = val
		} else {
			last, err = last.Merge(val)
			if err != nil {
				return nil, false, err
			}
		}
	}

	if !found {
		return &value.Object{}, true, nil
	}

	return last, true, nil
}

type Embedded struct {
	Comments   Comments
	Expression Expression

	evaluating bool
}

func (e *Embedded) Lookup(scope Scope, key string) (value.Value, bool, error) {
	if e.evaluating {
		return nil, false, nil
	}
	e.evaluating = true
	defer func() {
		e.evaluating = false
	}()
	v, ok, err := e.Expression.ToValue(scope)
	if err != nil || !ok {
		return nil, ok, err
	}

	return value.Lookup(v, key)
}

func (e *Embedded) ToValue(scope Scope) (value.Value, bool, error) {
	if e.evaluating {
		return nil, false, nil
	}
	e.evaluating = true
	defer func() {
		e.evaluating = false
	}()
	return e.Expression.ToValue(scope)
}

type Key struct {
	Value Expression
}

func (k *Key) ToString(scope Scope) (string, error) {
	return exprToString(k.Value, scope)
}

func (k *Key) Matches(scope Scope, key string) (_ bool, returnErr error) {
	keyPattern, err := k.ToString(scope)
	if err != nil {
		return false, err
	}

	return keyPattern == key, nil
}

type KeyValue struct {
	Local    bool
	Comments Comments
	Key      Key
	Value    Expression
	Optional bool

	evaluating bool
}

func (k *KeyValue) Lookup(scope Scope, key string) (value.Value, bool, error) {
	if k.evaluating {
		return nil, false, nil
	}
	k.evaluating = true
	defer func() {
		k.evaluating = false
	}()
	if ok, err := k.Key.Matches(scope, key); err != nil {
		return nil, false, err
	} else if !ok {
		return nil, false, nil
	}
	return k.Value.ToValue(scope)
}

func (k *KeyValue) ToValue(scope Scope) (value.Value, bool, error) {
	if k.Local || k.evaluating {
		return nil, false, nil
	}

	k.evaluating = true
	defer func() {
		k.evaluating = false
	}()

	v, ok, err := k.Value.ToValue(scope)
	if err != nil || !ok {
		return nil, ok, err
	}

	key, err := k.Key.ToString(scope)
	if err != nil {
		return nil, false, err
	}

	return &value.Object{
		Entries: []value.Entry{{
			Key:   key,
			Value: v,
		}},
	}, true, nil
}

type Comments struct {
	Comments [][]string
}

type Position struct {
	Filename string
	Offset   int
	Line     int
	Column   int
}

func (p Position) String() string {
	if p.Filename == "" {
		return fmt.Sprintf("%d:%d", p.Line, p.Column)
	}
	return fmt.Sprintf("%s:%d:%d", p.Filename, p.Line, p.Column)
}
