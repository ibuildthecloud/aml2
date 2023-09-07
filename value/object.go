package value

import (
	"fmt"
	"sort"
)

type Object struct {
	Entries  []Entry
	Contract ObjectContract
}

type ObjectContract interface {
	RequiredKeys() ([]string, error)
	LookupValue(key string) (Value, bool, error)
}

func NewObject(data map[string]any) *Object {
	o := &Object{}

	keys := make([]string, 0, len(data))
	for k := range data {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, key := range keys {
		o.Entries = append(o.Entries, Entry{
			Key:   key,
			Value: NewValue(data[key]),
		})
	}

	return o
}

func (n *Object) LookupValue(key string) (Value, bool, error) {
	var (
		entry Entry
		found bool
	)
	for _, e := range n.Entries {
		if e.Key == key {
			entry = e
			found = true
		}
	}

	return entry.Value, found, nil
}

func (n *Object) Close(contract ObjectContract) {
	n.Contract = contract
}

func (n *Object) TargetKind() Kind {
	return ObjectKind
}

func (n *Object) Kind() Kind {
	if n.Contract != nil {
		return SchemaKind
	}
	return ObjectKind
}

func (n *Object) NativeValue() any {
	result := map[string]any{}
	for _, entry := range n.Entries {
		if entry.Value.Kind() == FuncKind || entry.Value.Kind() == SchemaKind {
			continue
		}
		result[entry.Key] = entry.Value.NativeValue()
	}
	return result
}

func (n *Object) Keys() ([]string, error) {
	result := make([]string, 0, len(n.Entries))
	for _, entry := range n.Entries {
		result = append(result, entry.Key)
	}
	return result, nil
}

func (n *Object) lookup(key string) (Value, bool, error) {
	if n.Contract != nil {
		return n.Contract.LookupValue(key)
	}
	return n.LookupValue(key)
}

func (n *Object) requiredKeys() (result []string, _ error) {
	if n.Contract == nil {
		for _, entry := range n.Entries {
			result = append(result, entry.Key)
		}
		return result, nil
	}

	return n.Contract.RequiredKeys()
}

func (n *Object) Merge(right Value) (Value, error) {
	if merged, err := mergeNull(n, right); merged != nil || err != nil {
		return merged, err
	}

	var (
		head []Entry
		tail []Entry
	)

	requiredKeys, err := n.requiredKeys()
	if err != nil {
		return nil, err
	}

	keys, err := Keys(right)
	if err != nil {
		return nil, err
	}

	keysSeen := map[string]struct{}{}

	for _, key := range keys {
		rightValue, ok, err := Lookup(right, key)
		if err != nil {
			return nil, err
		}
		if !ok {
			continue
		}
		keysSeen[key] = struct{}{}

		schemaValue, ok, err := n.lookup(key)
		if err != nil {
			return nil, err
		}
		if ok {
			rightValue, err = schemaValue.Merge(rightValue)
			if err != nil {
				return nil, err
			}
		} else if n.Kind() == SchemaKind {
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
		def, ok, err := n.lookup(k)
		if err != nil {
			return nil, err
		}
		if n.Kind() != SchemaKind {
			if ok {
				head = append(head, Entry{
					Key:   k,
					Value: def,
				})
			}
		} else if def, hasDefault := DefaultValue(def); ok && hasDefault {
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

type Entry struct {
	Key   string
	Value Value
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
