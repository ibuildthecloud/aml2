package value

import (
	"fmt"
)

type ObjectSchema struct {
	Contract Contract
}

type Contract interface {
	RequiredKeys() ([]string, error)
	LookupValue(key string) (Value, bool, error)
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

func (n *ObjectSchema) NativeValue() any {
	panic("!!!!")
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
		} else {
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
