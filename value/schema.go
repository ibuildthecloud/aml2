package value

import (
	"errors"
	"fmt"
)

type Schema struct {
	KindValue    Kind
	Constraints  []Checker
	UnionTypes   []Schema
	DefaultValue Value
}

type Condition func(val Value) (Value, error)

func (n *Schema) Kind() Kind {
	return SchemaKind
}

func (n *Schema) TargetKind() Kind {
	return n.KindValue
}

func (n *Schema) GetConditions() []Condition {
	return n.GetConditions()
}

func (n *Schema) NativeValue() any {
	return n.KindValue
}

func (n *Schema) Gt(right Value) (Value, error) {
	result := *n
	result.Constraints = append(result.Constraints, &Constraint{
		Op:    ">",
		Right: right,
	})
	return &result, nil
}

func (n *Schema) Ge(right Value) (Value, error) {
	result := *n
	result.Constraints = append(result.Constraints, &Constraint{
		Op:    ">=",
		Right: right,
	})
	return &result, nil
}

func (n *Schema) Le(right Value) (Value, error) {
	result := *n
	result.Constraints = append(result.Constraints, &Constraint{
		Op:    "<=",
		Right: right,
	})
	return &result, nil
}

func (n *Schema) Lt(right Value) (Value, error) {
	result := *n
	result.Constraints = append(result.Constraints, &Constraint{
		Op:    "<",
		Right: right,
	})
	return &result, nil
}

func targetKind(v Value) Kind {
	if tk, ok := v.(interface {
		TargetKind() Kind
	}); ok {
		return tk.TargetKind()
	}
	return v.Kind()
}

func (n *Schema) Or(right Value) (Value, error) {
	result := *n
	schema, ok := right.(*Schema)
	if !ok {
		result.UnionTypes = append(result.UnionTypes, Schema{
			KindValue:    targetKind(right),
			DefaultValue: right,
		})
		return &result, nil
	}

	if n.TargetKind() == schema.TargetKind() {
		result.Constraints = []Checker{
			&OrConstraint{
				Left:  n.Constraints,
				Right: schema.Constraints,
			},
		}
		result.UnionTypes = append(result.UnionTypes, schema.UnionTypes...)
	} else {
		cp := *schema
		cp.UnionTypes = nil

		result.UnionTypes = append(result.UnionTypes, cp)
		result.UnionTypes = append(result.UnionTypes, schema.UnionTypes...)
	}

	return &result, nil
}

func (n *Schema) Default() (Value, bool) {
	return n.DefaultValue, n.DefaultValue != nil
}

func checkTypes(schemas []Schema, right Value) (Value, error) {
	var errs []error
	for _, schema := range schemas {
		if schema.TargetKind() != right.Kind() {
			errs = append(errs, fmt.Errorf("expected kind %s but got %s", schema.TargetKind(), right.Kind()))
			continue
		}
		if err := Constraints(schema.Constraints).Check(right); err != nil {
			errs = append(errs, err)
			continue
		}
		if schema.DefaultValue != nil {
			v, err := schema.DefaultValue.Merge(right)
			if err != nil {
				errs = append(errs, err)
				continue
			}
			return v, nil
		}
		return right, nil
	}
	return right, errors.Join(errs...)
}

func (n *Schema) Merge(right Value) (Value, error) {
	schemas := []Schema{*n}
	schemas = append(schemas, n.UnionTypes...)
	return checkTypes(schemas, right)
}

type Defaulter interface {
	Default() (Value, bool)
}

func DefaultValue(v Value) (Value, bool) {
	if v == nil {
		return nil, false
	}
	if v, ok := v.(Defaulter); ok {
		return v.Default()
	}
	if v.Kind() == SchemaKind {
		return nil, false
	}
	return v, true
}
