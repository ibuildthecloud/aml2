package value

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/acorn-io/aml/pkg/schema"
)

type DescribeArrayer interface {
	DescribeArray(ctx SchemaContext) (*schema.Array, bool, error)
}

func DescribeArray(ctx SchemaContext, val Value) (*schema.Array, error) {
	if err := assertType(val, SchemaKind); err != nil {
		return nil, err
	}
	if s, ok := val.(DescribeArrayer); ok {
		schema, ok, err := s.DescribeArray(ctx)
		if err != nil {
			return nil, err
		}
		if !ok {
			return nil, fmt.Errorf("array value did not provide a schema")
		}
		return schema, nil
	}
	return nil, fmt.Errorf("array value can not be converted to schema")
}

type ArraySchema []Value

func (a ArraySchema) Merge(right Value) (Value, error) {
	if ts, ok := right.(*TypeSchema); ok && ts.KindValue == ArrayKind {
		return a, nil
	}

	if err := assertType(right, ArrayKind); err != nil {
		return nil, err
	}

	if len(a) == 0 {
		return mergeNative(NewArray(nil), right)
	}

	var result []Value

	rightValues, err := ToValueArray(right)
	if err != nil {
		return nil, err
	}

outerLoop:
	for _, rightValue := range rightValues {
		var errs []error
		for _, schema := range a {
			newValue, err := Merge(schema, rightValue)
			if err == nil {
				result = append(result, newValue)
				continue outerLoop
			} else {
				errs = append(errs, err)
			}
		}
		if len(errs) > 0 {
			return nil, errors.Join(errs...)
		}
	}

	return Array(result), nil
}

func (a ArraySchema) DescribeArray(ctx SchemaContext) (*schema.Array, bool, error) {
	result := &schema.Array{}
	var last *schema.FieldType
	for i, v := range a {
		fieldType, err := DescribeFieldType(ctx, v)
		if err != nil {
			return nil, false, err
		}
		if i == 0 {
			result.Items = fieldType
			last = &result.Items
		} else {
			last.Alternate = &fieldType
			last = last.Alternate
		}
	}

	return result, true, nil
}

func (a ArraySchema) Slice(start, end int) (Value, bool, error) {
	if start >= len(a) || end > len(a) || start < 0 || end < 0 || start > end {
		return nil, false, nil
	}
	return a[start:end], true, nil
}

func (a ArraySchema) Index(idxValue Value) (Value, bool, error) {
	idx, err := ToInt(idxValue)
	if err != nil {
		return nil, false, err
	}
	if int(idx) >= len(a) || idx < 0 {
		return nil, false, nil
	}
	return a[idx], true, nil
}

func (a ArraySchema) ToValues() []Value {
	return a
}

func (a ArraySchema) Kind() Kind {
	return SchemaKind
}

func (a ArraySchema) TargetKind() Kind {
	return ArrayKind
}

func (a ArraySchema) String() string {
	data, _ := json.Marshal(a)
	return string(data)
}

func (a ArraySchema) Len() (Value, error) {
	return NewValue(len(a)), nil
}
