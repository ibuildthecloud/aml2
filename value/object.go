package value

import (
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

func (n *Object) Kind() Kind {
	return ObjectKind
}

func (n *Object) NativeValue() any {
	result := map[string]any{}
	for _, entry := range n.Entries {
		if entry.Value.Kind() == FuncKind {
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

func (n *Object) Merge(right Value) (Value, error) {
	if merged, err := mergeNull(n, right); merged != nil || err != nil {
		return merged, err
	}
	if len(n.Entries) == 0 {
		return right, nil
	}

	keys, err := Keys(right)
	if err != nil {
		return nil, err
	}

	seen := map[string]int{}
	var result []Entry
	for i, entry := range n.Entries {
		seen[entry.Key] = i
		result = append(result, entry)
	}

	for _, key := range keys {
		newValue, ok, err := Lookup(right, key)
		if err != nil {
			return nil, err
		}
		if !ok {
			continue
		}
		exiting, ok := seen[key]
		if ok {
			merged, err := result[exiting].Value.Merge(newValue)
			if err != nil {
				return nil, err
			}
			result[exiting] = Entry{
				Key:   key,
				Value: merged,
			}
		} else {
			result = append(result, Entry{
				Key:   key,
				Value: newValue,
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
