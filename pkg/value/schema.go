package value

import (
	"fmt"

	"github.com/acorn-io/aml/pkg/schema"
)

type DescribeFieldTyper interface {
	DescribeFieldType(ctx SchemaContext) (schema.FieldType, error)
}

func DescribeFieldType(ctx SchemaContext, v Value) (result schema.FieldType, _ error) {
	switch TargetKind(v) {
	case ObjectKind:
		objSchema, err := DescribeObject(ctx, v)
		if err != nil {
			return result, err
		}
		return schema.FieldType{
			Kind:   string(ObjectKind),
			Object: objSchema,
		}, nil
	case ArrayKind:
		arraySchema, err := DescribeArray(ctx, v)
		if err != nil {
			return result, err
		}
		return schema.FieldType{
			Kind:  string(ArrayKind),
			Array: arraySchema,
		}, nil
	}

	if ft, ok := v.(DescribeFieldTyper); ok {
		return ft.DescribeFieldType(ctx)
	}

	return result, fmt.Errorf("failed to determine field type for kind %s", TargetKind(v))
}
