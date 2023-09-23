package value

import (
	"errors"
	"fmt"
	"strings"

	"github.com/acorn-io/aml/pkg/schema"
)

type DescribeObjecter interface {
	DescribeObject(ctx SchemaContext) (*schema.Object, bool, error)
}

func DescribeObject(ctx SchemaContext, val Value) (*schema.Object, error) {
	if err := assertType(val, SchemaKind); err != nil {
		return nil, err
	}
	if s, ok := val.(DescribeObjecter); ok {
		schema, ok, err := s.DescribeObject(ctx)
		if err != nil {
			return nil, err
		}
		if !ok {
			return nil, fmt.Errorf("value kind %s did not provide a schema description", val.Kind())
		}
		return schema, nil
	}
	return nil, fmt.Errorf("value kind %s can not be converted to schema description", val.Kind())
}

type ObjectSchema struct {
	Contract Contract
}

type Contract interface {
	Path() string
	Description() string
	Fields(ctx SchemaContext) ([]schema.Field, error)
	RequiredKeys() ([]string, error)
	LookupValue(key string) (Value, bool, error)
	AllowNewKeys() bool
}

func NewObjectSchema(contract Contract) *ObjectSchema {
	return &ObjectSchema{
		Contract: contract,
	}
}

func (n *ObjectSchema) TargetKind() Kind {
	return ObjectKind
}

func (n *ObjectSchema) Kind() Kind {
	return SchemaKind
}

func (n *ObjectSchema) Fields(ctx SchemaContext) (result []schema.Field, _ error) {
	fields, err := n.Contract.Fields(ctx)
	if err != nil {
		return nil, err
	}

	var (
		fieldNames   = map[string]int{}
		mergedFields []schema.Field
	)

	for _, field := range fields {
		if i, ok := fieldNames[field.Name]; ok {
			mergedFields[i] = mergedFields[i].Merge(field)
		} else {
			fieldNames[field.Name] = len(mergedFields)
			mergedFields = append(mergedFields, field)
		}
	}

	return mergedFields, nil
}

func (n *ObjectSchema) DescribeObject(ctx SchemaContext) (*schema.Object, bool, error) {
	if ctx.haveSeen(n.Contract.Path()) {
		return &schema.Object{
			Description:  n.Contract.Description(),
			Path:         n.Contract.Path(),
			Reference:    true,
			AllowNewKeys: n.Contract.AllowNewKeys(),
		}, true, nil
	}

	ctx.addSeen(n.Contract.Path())

	fields, err := n.Fields(ctx)
	if err != nil {
		return nil, false, err
	}

	return &schema.Object{
		Description:  n.Contract.Description(),
		Path:         n.Contract.Path(),
		Fields:       fields,
		AllowNewKeys: n.Contract.AllowNewKeys(),
	}, true, nil
}

func (n *ObjectSchema) Keys() ([]string, error) {
	return n.Contract.RequiredKeys()
}

func (n *ObjectSchema) LookupValue(key Value) (Value, bool, error) {
	s, err := ToString(key)
	if err != nil {
		return nil, false, err
	}
	return n.Contract.LookupValue(s)
}

func (n *ObjectSchema) Merge(right Value) (Value, error) {
	var (
		head []Entry
		tail []Entry
	)

	if schema, ok := right.(*ObjectSchema); ok {
		return NewValue(&mergedContract{
			Left:  n.Contract,
			Right: schema.Contract,
		}), nil
	}

	if err := assertType(right, ObjectKind); err != nil {
		// This is a check that the schema doesn't have an invalid embeeded
		_, _, serr := n.DescribeObject(SchemaContext{})
		if serr != nil {
			return nil, errors.Join(err, serr)
		}
		return nil, err
	}

	requiredKeys, err := n.Contract.RequiredKeys()
	if err != nil {
		return nil, err
	}

	keys, err := Keys(right)
	if err != nil {
		return nil, err
	}

	keysSeen := map[string]struct{}{}

	for _, key := range keys {
		rightValue, ok, err := Lookup(right, NewValue(key))
		if err != nil {
			return nil, err
		}
		if !ok {
			continue
		}
		keysSeen[key] = struct{}{}

		schemaValue, ok, err := n.Contract.LookupValue(key)
		if err != nil {
			return nil, err
		}
		if ok {
			rightValue, err = Merge(schemaValue, rightValue)
			if err != nil {
				return nil, err
			}
		} else if !n.Contract.AllowNewKeys() {
			return nil, &ErrUnknownField{
				Key: key,
			}
		}

		tail = append(tail, Entry{
			Key:   key,
			Value: rightValue,
		})
	}

	var missingKeys []string
	for _, k := range requiredKeys {
		if _, seen := keysSeen[k]; seen {
			continue
		}
		def, ok, err := n.Contract.LookupValue(k)
		if err != nil {
			return nil, err
		}
		if def, hasDefault := DefaultValue(def); ok && hasDefault {
			head = append(head, Entry{
				Key:   k,
				Value: def,
			})
		} else {
			missingKeys = append(missingKeys, k)
		}
	}

	if len(missingKeys) > 0 {
		return nil, &ErrMissingRequiredKeys{
			Keys: missingKeys,
		}
	}

	return &Object{
		Entries: append(head, tail...),
	}, nil
}

type mergedContract struct {
	Left, Right Contract
}

func (m *mergedContract) Description() string {
	left, right := m.Left.Description(), m.Right.Description()
	var parts []string
	if left != "" {
		parts = append(parts, left)
	}
	if right != "" {
		parts = append(parts, right)
	}
	return strings.Join(parts, "\n")
}

func (m *mergedContract) Fields(ctx SchemaContext) ([]schema.Field, error) {
	leftFields, err := m.Left.Fields(ctx)
	if err != nil {
		return nil, err
	}

	rightFields, err := m.Right.Fields(ctx)
	if err != nil {
		return nil, err
	}

	return append(leftFields, rightFields...), nil
}

func (m *mergedContract) Path() string {
	return m.Left.Path()
}

func (m *mergedContract) RequiredKeys() ([]string, error) {
	result, err := m.Left.RequiredKeys()
	if err != nil {
		return nil, err
	}
	seen := map[string]struct{}{}
	for _, key := range result {
		seen[key] = struct{}{}
	}

	rightKeys, err := m.Right.RequiredKeys()
	if err != nil {
		return nil, err
	}

	for _, key := range rightKeys {
		if _, ok := seen[key]; !ok {
			result = append(result, key)
			seen[key] = struct{}{}
		}
	}

	return result, nil
}

func (m *mergedContract) LookupValue(key string) (Value, bool, error) {
	leftValue, ok, err := m.Left.LookupValue(key)
	if err != nil {
		return nil, false, err
	}

	if !ok {
		return m.Right.LookupValue(key)
	}

	rightValue, ok, err := m.Right.LookupValue(key)
	if err != nil {
		return nil, false, err
	}

	if !ok {
		return leftValue, true, nil
	}

	result, err := Merge(leftValue, rightValue)
	return result, true, err
}

func (m *mergedContract) AllowNewKeys() bool {
	return m.Left.AllowNewKeys() || m.Right.AllowNewKeys()
}

type ErrUnknownField struct {
	Key string
}

func (e *ErrUnknownField) Error() string {
	return fmt.Sprintf("unknown field: %s", e.Key)
}

type ErrMissingRequiredKeys struct {
	Keys []string
}

func (e *ErrMissingRequiredKeys) Error() string {
	return fmt.Sprintf("missing required keys: %v", e.Keys)
}
