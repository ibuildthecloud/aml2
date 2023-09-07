package eval

import (
	"fmt"
	"strings"

	"github.com/acorn-io/aml/value"
)

type Array struct {
	Comments Comments
	Items    []Expression
}

func (a *Array) ToValue(scope Scope) (value.Value, bool, error) {
	var objs []any

	for _, item := range a.Items {
		v, ok, err := item.ToValue(scope)
		if err != nil {
			return nil, false, err
		}
		if !ok {
			continue
		}
		_, isFor := item.(*For)
		if isFor && v.Kind() == value.ArrayKind {
			for _, item := range v.NativeValue().([]any) {
				objs = append(objs, value.NewValue(item))
			}
		} else {
			objs = append(objs, v)
		}
	}
	return value.NewArray(objs), true, nil
}

type Parens struct {
	Comments Comments
	Expr     Expression
}

func (p *Parens) ToValue(scope Scope) (value.Value, bool, error) {
	return p.Expr.ToValue(scope)
}

type Op struct {
	Schema   bool
	Unary    bool
	Comments Comments
	Operator string
	Left     Expression
	Right    Expression
	Pos      Position
}

func (o *Op) ToValue(scope Scope) (value.Value, bool, error) {
	left, ok, err := o.Left.ToValue(scope)
	if err != nil || !ok {
		return nil, ok, err
	}

	if o.Unary {
		newValue, err := value.Unary(o.Operator, left)
		return newValue, true, err
	}

	right, ok, err := o.Right.ToValue(scope)
	if err != nil || !ok {
		return nil, ok, err
	}

	newValue, err := value.Binary(o.Operator, left, right)
	return newValue, true, err
}

type Lookup struct {
	Comments Comments
	Pos      Position
	Key      string
}

func (l *Lookup) ToValue(scope Scope) (value.Value, bool, error) {
	v, ok, err := scope.Get(l.Key)
	if err != nil {
		return nil, false, err
	}
	if !ok {
		return nil, false, &ErrPathNotFound{
			Key: l.Key,
			Pos: l.Pos,
		}
	}
	return v, true, nil
}

type ErrPathNotFound struct {
	Key string
	Pos Position
}

func (c *ErrPathNotFound) Error() string {
	return fmt.Sprintf("path not found: %s %s", c.Key, c.Pos)
}

type ErrIndexNotFound struct {
	Index int64
	Pos   Position
}

func (c *ErrIndexNotFound) Error() string {
	return fmt.Sprintf("index not found: %d %s", c.Index, c.Pos)
}

type Selector struct {
	Comments Comments
	Pos      Position
	Base     Expression
	Key      Expression
}

func (s *Selector) ToValue(scope Scope) (value.Value, bool, error) {
	key, err := exprToString(s.Key, scope)
	if err != nil {
		return nil, false, err
	}

	v, ok, err := s.Base.ToValue(scope)
	if err != nil {
		return nil, false, err
	}
	if !ok {
		return nil, false, &ErrPathNotFound{
			Key: key,
			Pos: s.Pos,
		}
	}

	newValue, ok, err := value.Lookup(v, key)
	if err != nil {
		return nil, false, err
	}
	if !ok {
		return nil, false, &ErrPathNotFound{
			Key: key,
			Pos: s.Pos,
		}
	}

	return newValue, true, nil
}

type Index struct {
	Comments Comments
	Pos      Position
	Base     Expression
	Index    Expression
}

func (i *Index) ToValue(scope Scope) (value.Value, bool, error) {
	key, err := exprToInt(i.Index, scope)
	if err != nil {
		return nil, false, err
	}

	v, ok, err := i.Base.ToValue(scope)
	if err != nil {
		return nil, false, err
	}
	if !ok {
		return nil, false, &ErrIndexNotFound{
			Index: key,
			Pos:   i.Pos,
		}
	}

	newValue, ok, err := value.Index(v, key)
	if err != nil {
		return nil, false, err
	}
	if !ok {
		return nil, false, &ErrIndexNotFound{
			Index: key,
			Pos:   i.Pos,
		}
	}

	return newValue, true, nil
}

type Slice struct {
	Comments Comments
	Pos      Position
	Base     Expression
	Start    Expression
	End      Expression
}

func (s *Slice) ToValue(scope Scope) (value.Value, bool, error) {
	var (
		err   error
		end   = int64(0)
		start = int64(0)
	)

	v, ok, err := s.Base.ToValue(scope)
	if err != nil {
		return nil, false, err
	}
	if !ok {
		return nil, false, &ErrPathNotFound{
			Key: fmt.Sprint(start),
			Pos: s.Pos,
		}
	}

	if s.Start != nil {
		start, err = exprToInt(s.Start, scope)
		if err != nil {
			return nil, false, err
		}
	}

	if s.End == nil {
		end, err = value.Len(v)
		if err != nil {
			return nil, false, err
		}
	} else {
		end, err = exprToInt(s.End, scope)
		if err != nil {
			return nil, false, err
		}
	}

	newValue, ok, err := value.Slice(v, start, end)
	if !ok || err != nil {
		return nil, ok, err
	}

	return newValue, true, nil
}

type Call struct {
	Comments Comments
	Pos      Position
	Func     Expression
	Args     []Expression
}

func (c *Call) ToValue(scope Scope) (value.Value, bool, error) {
	panic("unsupported")
}

type If struct {
	Comments   Comments
	Condition  Expression
	Value      Expression
	evaluating bool
}

func (i *If) ToValue(scope Scope) (value.Value, bool, error) {
	if i.evaluating {
		return nil, false, nil
	}
	i.evaluating = true
	v, ok, err := i.Condition.ToValue(scope)
	i.evaluating = false
	if err != nil || !ok {
		return nil, ok, err
	}

	b, err := value.ToBool(v)
	if err != nil {
		return nil, false, err
	}

	if !b {
		return nil, false, nil
	}

	return i.Value.ToValue(scope)
}

type Interpolation struct {
	Parts []any
}

func (i *Interpolation) ToValue(scope Scope) (value.Value, bool, error) {
	var result []string
	for _, part := range i.Parts {
		switch v := part.(type) {
		case string:
			result = append(result, v)
		case Expression:
			val, ok, err := v.ToValue(scope)
			if err != nil {
				return nil, false, err
			}
			if !ok {
				continue
			}
			s := value.Escape(fmt.Sprint(val.NativeValue()))
			result = append(result, s)
		}
	}
	s, err := value.Unquote(strings.Join(result, ""))
	return value.NewValue(s), true, err
}

type For struct {
	Comments Comments
	Key      Expression
	Value    Expression
	List     Expression
	Body     Expression
}

func (f *For) ToValue(scope Scope) (value.Value, bool, error) {
	var (
		indexKey    string
		hasIndexKey bool
		valueKey    string
		err         error
	)

	if f.Key != nil {
		hasIndexKey = true
		indexKey, err = exprToString(f.Key, scope)
		if err != nil {
			return nil, false, err
		}
	}

	valueKey, err = exprToString(f.Value, scope)
	if err != nil {
		return nil, false, err
	}

	listValue, ok, err := f.List.ToValue(scope)
	if err != nil || !ok {
		return nil, ok, err
	}

	list, err := value.ToValueArray(listValue)
	if err != nil {
		return nil, false, err
	}

	array := value.Array{}

	for i, item := range list {
		data := map[string]value.Value{
			valueKey: item,
		}
		if hasIndexKey {
			data[indexKey] = value.NewValue(i)
		}

		newValue, ok, err := f.Body.ToValue(scope.Push(MapLookup(data)))
		if err != nil {
			return nil, false, err
		}
		if !ok {
			continue
		}

		array = append(array, newValue)

		if newValue.Kind() == value.ObjectKind {
			scope = scope.Push(NativeMapLookup(newValue.NativeValue().(map[string]any)))
		}
	}

	return array, true, nil
}

type SchemaAllowed interface {
	SchemaAllowed() bool
}

type MergeObjectArray struct {
	Array Expression
}

func (m *MergeObjectArray) ToValue(scope Scope) (value.Value, bool, error) {
	v, ok, err := m.Array.ToValue(scope)
	if err != nil || !ok {
		return nil, ok, err
	}

	valueArray, err := value.ToValueArray(v)
	if err != nil {
		return nil, false, err
	}

	var result value.Value
	for _, item := range valueArray {
		if result == nil {
			result = item
		} else {
			result, err = result.Merge(item)
			if err != nil {
				return nil, false, err
			}
		}
	}

	return result, true, nil
}

type Expression interface {
	ToValue(scope Scope) (value.Value, bool, error)
}

func exprToString(expr Expression, scope Scope) (string, error) {
	v, _, err := expr.ToValue(scope)
	if err != nil {
		return "", err
	}
	return value.ToString(v)
}

func exprToInt(expr Expression, scope Scope) (int64, error) {
	v, _, err := expr.ToValue(scope)
	if err != nil {
		return 0, err
	}
	return value.ToInt(v)
}

type Value struct {
	Value value.Value
}

func (v Value) ToValue(_ Scope) (value.Value, bool, error) {
	return v.Value, true, nil
}
