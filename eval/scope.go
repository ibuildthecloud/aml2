package eval

import (
	"errors"

	"github.com/acorn-io/aml/value"
)

type Data map[string]any

func (m Data) Get(key string) (value.Value, bool, error) {
	obj, ok := m[key]
	if !ok {
		return nil, ok, nil
	}
	return value.NewValue(obj), true, nil
}

func (m Data) Push(lookup ValueLookup) Scope {
	return nested{
		parent: m,
		lookup: lookup,
	}
}

type nested struct {
	parent Scope
	lookup ValueLookup
}

func (n nested) Get(key string) (ret value.Value, _ bool, _ error) {
	v, ok, err := n.lookup.Lookup(n, key)
	if e := (*ErrPathNotFound)(nil); errors.As(err, &e) {
		ok = false
	} else if err != nil {
		return nil, false, err
	}
	if ok {
		return v, ok, nil
	}
	return n.parent.Get(key)
}

func (n nested) Push(lookup ValueLookup) Scope {
	return nested{
		parent: n,
		lookup: lookup,
	}
}
