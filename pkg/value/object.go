package value

import (
	"encoding/json"
	"fmt"
	"sort"
)

type Object struct {
	Entries []Entry
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

func (n *Object) LookupValue(key Value) (Value, bool, error) {
	for _, e := range n.Entries {
		b, err := Eq(key, NewValue(e.Key))
		if err != nil {
			return nil, false, err
		}

		if b, err := ToBool(b); err != nil || b {
			return e.Value, b, err
		}
	}

	return nil, false, nil
}

func (n *Object) Eq(right Value) (Value, error) {
	if right.Kind() != ObjectKind {
		return nil, fmt.Errorf("can not compare object with kind %s", right.Kind())
	}

	rightKeys, err := Keys(right)
	if err != nil {
		return nil, err
	}

	leftKeys, err := n.Keys()
	if err != nil {
		return nil, err
	}

	if len(rightKeys) != len(leftKeys) {
		return False, nil
	}

	sort.Strings(rightKeys)
	sort.Strings(leftKeys)

	for i, key := range rightKeys {
		if leftKeys[i] != key {
			return False, nil
		}

		leftValue, ok, err := n.LookupValue(NewValue(key))
		if err != nil || !ok {
			return False, err
		}

		rightValue, ok, err := Lookup(right, NewValue(key))
		if err != nil || !ok {
			return False, err
		}

		bValue, err := Eq(leftValue, rightValue)
		if err != nil {
			return nil, err
		}

		b, err := ToBool(bValue)
		if err != nil {
			return nil, err
		}
		if !b {
			return False, nil
		}
	}

	return True, nil
}

func (n *Object) Kind() Kind {
	return ObjectKind
}

func (n *Object) MarshalJSON() ([]byte, error) {
	result := map[string]any{}
	for _, entry := range n.Entries {
		result[entry.Key] = entry.Value
	}
	return json.Marshal(result)
}

func (n *Object) String() string {
	data, _ := n.MarshalJSON()
	return string(data)
}

func (n *Object) NativeValue() (any, bool, error) {
	result := map[string]any{}
	for _, entry := range n.Entries {
		nv, ok, err := NativeValue(entry.Value)
		if err != nil {
			return nil, false, err
		}
		if !ok {
			continue
		}
		result[entry.Key] = nv
	}
	return result, true, nil
}

func (n *Object) Keys() ([]string, error) {
	result := make([]string, 0, len(n.Entries))
	for _, entry := range n.Entries {
		result = append(result, entry.Key)
	}
	return result, nil
}

func (n *Object) Merge(right Value) (Value, error) {
	if err := assertKindsMatch(n, right); err != nil {
		return nil, err
	}

	var (
		result   []Entry
		keysSeen = map[string]int{}
	)

	for _, entry := range n.Entries {
		keysSeen[entry.Key] = len(result)
		result = append(result, entry)
	}

	keys, err := Keys(right)
	if err != nil {
		return nil, fmt.Errorf("failed to merge kind %s with %s: %w", ObjectKind, right.Kind(), err)
	}

	for _, key := range keys {
		rightValue, ok, err := Lookup(right, NewValue(key))
		if err != nil {
			return nil, err
		}
		if !ok {
			continue
		}

		if i, ok := keysSeen[key]; ok {
			rightValue, err = Merge(result[i].Value, rightValue)
			if err != nil {
				return nil, err
			}
			result[i].Value = rightValue
		} else {
			result = append(result, Entry{
				Key:   key,
				Value: rightValue,
			})
		}

	}

	return &Object{
		Entries: result,
	}, nil
}

type Entry struct {
	Key   string
	Value Value
}
