package value

import (
	"fmt"
	"strings"

	"github.com/acorn-io/aml/pkg/schema"
)

type ObjectSchema struct {
	Contract Contract
}

type Contract interface {
	Path() string
	Description() string
	Fields(seen map[string]struct{}) ([]schema.Field, error)
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

func (n *ObjectSchema) Schema(seen map[string]struct{}) (any, bool, error) {
	if _, ok := seen[n.Contract.Path()]; ok {
		return nil, false, nil
	}

	seen[n.Contract.Path()] = struct{}{}
	fields, err := n.Contract.Fields(seen)
	if err != nil {
		return nil, false, err
	}

	return &schema.Object{
		Description:  n.Contract.Description(),
		Path:         n.Contract.Path(),
		Fields:       fields,
		AllowNewKeys: n.Contract.AllowNewKeys(),
	}, false, nil
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

func (m *mergedContract) Fields(seen map[string]struct{}) ([]schema.Field, error) {
	leftFields, err := m.Left.Fields(seen)
	if err != nil {
		return nil, err
	}

	rightFields, err := m.Right.Fields(seen)
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
