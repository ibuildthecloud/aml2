package eval

import (
	"fmt"
	"strings"

	"github.com/acorn-io/aml/pkg/value"
)

type Parens struct {
	Comments Comments
	Expr     Expression
}

type Default struct {
	Comments Comments
	Expr     Expression
	Pos      Position
}

func (d *Default) ToValue(scope Scope) (value.Value, bool, error) {
	v, ok, err := d.Expr.ToValue(scope.Push(nil, ScopeOption{
		Default: true,
	}))
	if err != nil || !ok {
		return nil, ok, err
	}
	return value.NewDefault(v), true, nil
}

func (p *Parens) ToValue(scope Scope) (value.Value, bool, error) {
	return p.Expr.ToValue(scope.Push(nil, ScopeOption{
		Default: true,
	}))
}

type Op struct {
	Unary    bool
	Comments Comments
	Operator value.Operator
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
		newValue, err := value.UnaryOperation(o.Operator, left)
		return newValue, true, err
	}

	right, ok, err := o.Right.ToValue(scope)
	if err != nil || !ok {
		return nil, ok, err
	}

	newValue, err := value.BinaryOperation(o.Operator, left, right)
	return newValue, true, err
}

type Lookup struct {
	Comments Comments
	Pos      Position
	Key      string

	evaluating bool
}

func (l *Lookup) ToValue(scope Scope) (value.Value, bool, error) {
	if l.evaluating {
		return value.Undefined{Pos: value.Position(l.Pos)}, true, nil
	}
	l.evaluating = true
	defer func() { l.evaluating = false }()

	v, ok, err := scope.Get(l.Key)
	if err != nil {
		return nil, false, err
	}
	if !ok {
		return nil, false, &ErrPathNotFound{
			Key: value.NewValue(l.Key),
			Pos: l.Pos,
		}
	}
	return v, true, nil
}

type ErrPathNotFound struct {
	Key value.Value
	Pos Position
}

func (c *ErrPathNotFound) Error() string {
	return fmt.Sprintf("path not found: %s %s", c.Key, c.Pos)
}

type Selector struct {
	Comments Comments
	Pos      Position
	Base     Expression
	Key      Expression
}

func (s *Selector) ToValue(scope Scope) (value.Value, bool, error) {
	key, ok, err := s.Key.ToValue(scope)
	if err != nil || !ok {
		return nil, ok, err
	}

	v, ok, err := s.Base.ToValue(scope)
	if err != nil {
		return nil, false, err
	}
	if !ok {
		return nil, false, nil
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
	base, ok, err := i.Base.ToValue(scope)
	if err != nil || !ok {
		return nil, ok, err
	}

	indexValue, ok, err := i.Index.ToValue(scope)
	if err != nil || !ok {
		return nil, ok, err
	}

	if indexValue.Kind() == value.StringKind {
		return value.Lookup(base, indexValue)
	}

	return value.Index(base, indexValue)
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
		start, end value.Value
	)

	v, ok, err := s.Base.ToValue(scope)
	if err != nil || !ok {
		return nil, ok, err
	}

	if s.Start != nil {
		start, ok, err = s.Start.ToValue(scope)
		if err != nil || !ok {
			return nil, ok, err
		}
	}

	if s.End != nil {
		end, ok, err = s.End.ToValue(scope)
		if err != nil || !ok {
			return nil, ok, err
		}
	}

	newValue, ok, err := value.Slice(v, start, end)
	if err != nil || !ok {
		return nil, ok, err
	}

	return newValue, true, nil
}

type Call struct {
	Comments Comments
	Pos      Position
	Func     Expression
	Args     []Field
}

func (c *Call) ToValue(scope Scope) (value.Value, bool, error) {
	v, ok, err := c.Func.ToValue(scope)
	if err != nil || !ok {
		return nil, ok, err
	}

	var args []value.CallArgument
	for _, field := range c.Args {
		var arg value.CallArgument
		if posArg, ok := field.(IsPositionalArgument); ok {
			arg.Positional = posArg.IsPositionalArgument()
		}
		v, ok, err := field.ToValue(scope)
		if err != nil {
			return nil, false, err
		} else if !ok {
			continue
		}
		arg.Value = v
		args = append(args, arg)
	}

	return value.Call(v, args...)
}

type If struct {
	Comments  Comments
	Condition Expression
	Value     Expression
	Else      Expression
}

func (i *If) ToValue(scope Scope) (value.Value, bool, error) {
	v, ok, err := i.Condition.ToValue(scope)
	if err != nil || !ok {
		return nil, ok, err
	}

	if v.Kind() == value.UndefinedKind {
		return v, false, err
	}

	b, err := value.ToBool(v)
	if err != nil {
		return nil, false, err
	}
	if !b {
		if i.Else != nil {
			return i.Else.ToValue(scope)
		}
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
			if val.Kind() == value.UndefinedKind {
				return val, true, nil
			}
			nv, ok, err := value.NativeValue(val)
			if err != nil {
				return nil, false, err
			}
			if !ok {
				continue
			}
			result = append(result, value.Escape(fmt.Sprint(nv)))
		}
	}
	s, err := value.Unquote(strings.Join(result, ""))
	return value.NewValue(s), true, err
}

type For struct {
	Comments   Comments
	Key        string
	Value      string
	Collection Expression
	Body       Expression
	Merge      bool
}

type entry struct {
	Key   value.Value
	Value value.Value
}

func toList(v value.Value) (result []entry, _ error) {
	if v.Kind() == value.ArrayKind {
		list, err := value.ToValueArray(v)
		if err != nil {
			return nil, err
		}
		for i, item := range list {
			result = append(result, entry{
				Key:   value.NewValue(i),
				Value: item,
			})
		}
		return
	} else if v.Kind() == value.ObjectKind {
		keys, err := value.Keys(v)
		if err != nil {
			return nil, err
		}
		for _, key := range keys {
			v, ok, err := value.Lookup(v, value.NewValue(key))
			if err != nil {
				return nil, err
			}
			if !ok {
				continue
			}
			result = append(result, entry{
				Key:   value.NewValue(key),
				Value: v,
			})
		}
	} else {
		result = append(result, entry{
			Key:   value.NewValue(0),
			Value: v,
		})
	}

	return
}

func (f *For) ToValue(scope Scope) (value.Value, bool, error) {
	collection, ok, err := f.Collection.ToValue(scope)
	if err != nil || !ok {
		return nil, ok, err
	}

	list, err := toList(collection)
	if err != nil {
		return nil, false, err
	}

	array := value.Array{}

	for _, item := range list {
		data := map[string]any{}
		if f.Key != "" {
			data[f.Key] = item.Key
		}
		if f.Value != "" {
			data[f.Value] = item.Value
		}

		newValue, ok, err := f.Body.ToValue(scope.Push(ScopeData(data)))
		if err != nil {
			return nil, false, err
		}
		if !ok {
			continue
		}

		array = append(array, newValue)

		if newValue.Kind() == value.ObjectKind {
			scope = scope.Push(ValueScopeLookup{
				Value: newValue,
			})
		}
	}

	if f.Merge {
		vals := array.ToValues()
		if len(vals) == 0 {
			return value.NewObject(nil), true, nil
		}
		v, err := value.Merge(vals...)
		return v, true, err
	}

	return array, true, nil
}

type Expression interface {
	ToValue(scope Scope) (value.Value, bool, error)
}

type IsPositionalArgument interface {
	IsPositionalArgument() bool
}

type Value struct {
	Value value.Value
}

func (v Value) ToValue(_ Scope) (value.Value, bool, error) {
	return v.Value, true, nil
}
