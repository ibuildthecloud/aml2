package value

import (
	"errors"
	"fmt"

	"github.com/acorn-io/aml/pkg/schema"
)

type TypeSchema struct {
	KindValue    Kind
	Constraints  []Checker
	Alternate    *TypeSchema
	DefaultValue Value
}

func NewDefault(v Value) Value {
	return &TypeSchema{
		KindValue:    TargetKind(v),
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

func checkerToConstraint(checker Checker) (result schema.Constraint, _ bool, _ error) {
	right, ok, err := checker.RightNative()
	if err != nil {
		return result, false, err
	} else if !ok {
		right = nil
	}

	left, ok, err := checker.LeftNative()
	if err != nil {
		return result, ok, err
	} else if !ok {
		left = nil
	}

	return schema.Constraint{
		Description: checker.Description(),
		Op:          checker.OpString(),
		Left:        left,
		Right:       right,
	}, true, nil
}

func typeSchemaToFieldType(n *TypeSchema) (result schema.FieldType, _ bool, _ error) {
	result.Kind = string(n.KindValue)

	if n.DefaultValue != nil {
		def, ok, err := NativeValue(n.DefaultValue)
		if err != nil || !ok {
			return result, ok, err
		}
		result.Default = def
	}

	for _, checker := range n.Constraints {
		constraint, ok, err := checkerToConstraint(checker)
		if err != nil {
			return result, false, err
		} else if !ok {
			continue
		}

		result.Constraint = append(result.Constraint, constraint)
	}

	if n.Alternate != nil {
		alt, ok, err := typeSchemaToFieldType(n.Alternate)
		if err != nil || !ok {
			return result, ok, err
		}
		result.Alternative = &alt
	}

	return result, true, nil
}

func (n *TypeSchema) Schema(seen map[string]struct{}) (any, bool, error) {
	//	if n.DefaultValue != nil {
	//		return n.DefaultValue, true, nil
	//	}
	//	if n.Alternate != nil {
	//		return n.Alternate.NativeValue()
	//	}
	//	return nil, false, fmt.Errorf("unspecified value of kind %s", n.KindValue)
	//}
	//
	//func (n *TypeSchema) Schema() (any, bool, error) {
	field := &schema.Field{}

	fieldType, ok, err := typeSchemaToFieldType(n)
	if err != nil || !ok {
		return nil, ok, err
	}

	field.Type = fieldType
	return field, false, nil
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

func TargetKind(v Value) Kind {
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
	cp.Alternate = addOr(&cp, rightSchema.Alternate)
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

func addOr(left, right *TypeSchema) *TypeSchema {
	cp := *left
	if cp.Alternate == nil {
		cp.Alternate = right
	} else {
		cp.Alternate = addOr(left.Alternate, right)
	}
	return &cp
}

func (n *TypeSchema) Or(right Value) (Value, error) {
	rightSchema, ok := right.(*TypeSchema)
	if !ok {
		rightSchema = &TypeSchema{
			KindValue:    TargetKind(right),
			DefaultValue: right,
		}
	}

	return addOr(n, rightSchema), nil
}

func (n *TypeSchema) Default() (Value, bool) {
	return n.DefaultValue, n.DefaultValue != nil
}

func checkType(schema *TypeSchema, right Value) (Value, error) {
	var errs []error

	if schema.TargetKind() != right.Kind() {
		errs = append(errs, fmt.Errorf("expected kind %s but got %s", schema.TargetKind(), right.Kind()))
	}

	if err := Constraints(schema.Constraints).Check(right); err != nil {
		errs = append(errs, err)
	}

	if schema.DefaultValue != nil && !IsSimpleKind(right.Kind()) {
		v, err := Merge(schema.DefaultValue, right)
		if err == nil {
			right = v
		} else {
			errs = append(errs, err)
		}
	}

	if len(errs) == 0 {
		return right, nil
	}

	if schema.Alternate != nil {
		ret, newErrs := checkType(schema.Alternate, right)
		if newErrs == nil {
			return ret, nil
		}
		errs = append(errs, newErrs)
	}

	return right, errors.Join(errs...)
}

func (n *TypeSchema) Merge(right Value) (Value, error) {
	if right.Kind() == SchemaKind {
		return And(n, right)
	}
	return checkType(n, right)
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