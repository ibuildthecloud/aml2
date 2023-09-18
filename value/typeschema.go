package value

import (
	"errors"
	"fmt"
)

type TypeSchema struct {
	KindValue    Kind
	Constraints  []Checker
	UnionTypes   []TypeSchema
	DefaultValue Value
}

func NewDefault(v Value) Value {
	return &TypeSchema{
		KindValue:    targetKind(v),
		DefaultValue: v,
	}
}

type Condition func(val Value) (Value, error)

func (n *TypeSchema) Kind() Kind {
	return SchemaKind
}

func (n *TypeSchema) TargetKind() Kind {
	return n.KindValue
}

func (n *TypeSchema) GetConditions() []Condition {
	return n.GetConditions()
}

func (n *TypeSchema) NativeValue() any {
	return n.KindValue
}

func (n *TypeSchema) Eq(right Value) (Value, error) {
	result := *n
	result.Constraints = append(result.Constraints, &Constraint{
		Op:    "==",
		Right: right,
	})
	return &result, nil
}

func (n *TypeSchema) Neq(right Value) (Value, error) {
	result := *n
	result.Constraints = append(result.Constraints, &Constraint{
		Op:    "!=",
		Right: right,
	})
	return &result, nil
}

func (n *TypeSchema) Gt(right Value) (Value, error) {
	result := *n
	result.Constraints = append(result.Constraints, &Constraint{
		Op:    ">",
		Right: right,
	})
	return &result, nil
}

func (n *TypeSchema) Ge(right Value) (Value, error) {
	result := *n
	result.Constraints = append(result.Constraints, &Constraint{
		Op:    ">=",
		Right: right,
	})
	return &result, nil
}

func (n *TypeSchema) Le(right Value) (Value, error) {
	result := *n
	result.Constraints = append(result.Constraints, &Constraint{
		Op:    "<=",
		Right: right,
	})
	return &result, nil
}

func (n *TypeSchema) Lt(right Value) (Value, error) {
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

func (n *TypeSchema) And(right Value) (Value, error) {
	rightSchema, ok := right.(*TypeSchema)
	if !ok {
		return nil, fmt.Errorf("expected kind %s, got %s", n.Kind(), right.Kind())
	}
	if n.TargetKind() != rightSchema.TargetKind() {
		return nil, fmt.Errorf("invalid schema condition %s && %s incompatible", n.TargetKind(), rightSchema.TargetKind())
	}

	cp := *n
	cp.UnionTypes = append(cp.UnionTypes, rightSchema.UnionTypes...)
	cp.Constraints = append(cp.Constraints, rightSchema.Constraints...)
	if cp.DefaultValue == nil {
		cp.DefaultValue = rightSchema.DefaultValue
	} else if rightSchema.DefaultValue != nil {
		eq, err := Eq(cp.DefaultValue, rightSchema.DefaultValue)
		if err != nil {
			return nil, err
		}
		b, err := ToBool(eq)
		if err != nil {
			return nil, err
		}
		if !b {
			return nil, fmt.Errorf("can not have two default values for schema kind %s, %s and %s", cp.TargetKind(), cp.DefaultValue, rightSchema.DefaultValue)
		}
	}

	return &cp, nil
}

func (n *TypeSchema) Or(right Value) (Value, error) {
	result := *n
	schema, ok := right.(*TypeSchema)
	if !ok {
		result.UnionTypes = append(result.UnionTypes, TypeSchema{
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

func (n *TypeSchema) Default() (Value, bool) {
	return n.DefaultValue, n.DefaultValue != nil
}

func checkTypes(schemas []TypeSchema, right Value) (Value, error) {
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
		if schema.DefaultValue != nil && !IsSimpleKind(right.Kind()) {
			v, err := Merge(schema.DefaultValue, right)
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

func (n *TypeSchema) Merge(right Value) (Value, error) {
	if right.Kind() == SchemaKind {
		return And(n, right)
	}
	schemas := []TypeSchema{*n}
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
