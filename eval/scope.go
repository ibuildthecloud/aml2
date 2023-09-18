package eval

import (
	"errors"

	"github.com/acorn-io/aml/value"
)

type ScopeOption struct {
	Schema bool
}

func combine(opts []ScopeOption) (result ScopeOption) {
	for _, opt := range opts {
		if opt.Schema {
			result.Schema = true
		}
	}
	return
}

type Scope interface {
	Get(key string) (value.Value, bool, error)
	Push(lookup ScopeLookuper, opts ...ScopeOption) Scope
	IsSchema() bool
}

type ScopeLookuper interface {
	ScopeLookup(scope Scope, key string) (value.Value, bool, error)
}

type ScopeData map[string]any

func (m ScopeData) ScopeLookup(_ Scope, key string) (value.Value, bool, error) {
	ret, ok := m[key]
	return value.NewValue(ret), ok, nil
}

func (m ScopeData) Get(key string) (value.Value, bool, error) {
	obj, ok := m[key]
	if !ok {
		return nil, ok, nil
	}
	return value.NewValue(obj), true, nil
}

func (m ScopeData) IsSchema() bool {
	return false
}

func (m ScopeData) Push(lookup ScopeLookuper, opts ...ScopeOption) Scope {
	return nested{
		parent: m,
		lookup: lookup,
		schema: combine(opts).Schema,
	}
}

type nested struct {
	parent Scope
	lookup ScopeLookuper
	schema bool
}

func (n nested) IsSchema() bool {
	if n.schema {
		return true
	}
	return n.parent.IsSchema()
}

func (n nested) Get(key string) (ret value.Value, ok bool, err error) {
	v, ok, err := n.lookup.ScopeLookup(n, key)
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

func (n nested) Push(lookup ScopeLookuper, opts ...ScopeOption) Scope {
	return nested{
		parent: n,
		lookup: lookup,
		schema: combine(opts).Schema,
	}
}

type ValueScopeLookup struct {
	Value value.Value
}

func (v ValueScopeLookup) ScopeLookup(_ Scope, key string) (value.Value, bool, error) {
	return value.Lookup(v.Value, value.NewValue(key))
}
