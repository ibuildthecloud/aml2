package eval

import (
	"context"
	"errors"
	"fmt"

	"github.com/acorn-io/aml/pkg/value"
)

type ScopeOption struct {
	Schema       bool
	AllowNewKeys bool
	Default      bool
	Path         string
	Context      context.Context
}

func combine(opts []ScopeOption) (result ScopeOption) {
	for _, opt := range opts {
		if opt.Schema {
			result.Schema = true
		}
		if opt.AllowNewKeys {
			result.Schema = true
			result.AllowNewKeys = true
		}
		if opt.Default {
			result.Default = opt.Default
		}
		if opt.Path != "" {
			result.Path = opt.Path
		}
		if opt.Context != nil {
			result.Context = opt.Context
		}
	}
	return
}

type Scope interface {
	Context() context.Context
	Path() string
	Get(key string) (value.Value, bool, error)
	Push(lookup ScopeLookuper, opts ...ScopeOption) Scope
	IsSchema() bool
	AllowNewKeys() bool
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

type EmptyScope struct {
}

func (e EmptyScope) Path() string {
	return ""
}

func (e EmptyScope) Get(key string) (value.Value, bool, error) {
	return nil, false, nil
}

func (a EmptyScope) Context() context.Context {
	return context.Background()
}

func (e EmptyScope) Push(lookup ScopeLookuper, opts ...ScopeOption) Scope {
	return scopePush(e, lookup, opts...)
}

func (e EmptyScope) IsSchema() bool {
	return false
}

func (e EmptyScope) AllowNewKeys() bool {
	return false
}

type nested struct {
	path   string
	parent Scope
	lookup ScopeLookuper
	opts   ScopeOption
}

func (n nested) AllowNewKeys() bool {
	if n.opts.Default {
		return false
	}
	if n.opts.Schema {
		return n.opts.AllowNewKeys
	}
	return n.parent.AllowNewKeys()
}

func (n nested) IsSchema() bool {
	if n.opts.Default {
		return false
	}
	if n.opts.Schema {
		return true
	}
	return n.parent.IsSchema()
}

func (n nested) Context() context.Context {
	if n.opts.Context != nil {
		return n.opts.Context
	}
	return n.parent.Context()
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

func (n nested) Path() string {
	return n.path
}

func scopePush(n Scope, lookup ScopeLookuper, opts ...ScopeOption) Scope {
	if lookup == nil {
		lookup = ScopeData(nil)
	}
	o := combine(opts)
	newPath := appendPath(n.Path(), o.Path)
	if len(newPath) > 500 {
		panic("stack depth too deep " + fmt.Sprint(len(newPath)))
	}
	return nested{
		path:   newPath,
		parent: n,
		lookup: lookup,
		opts:   o,
	}
}

func (n nested) Push(lookup ScopeLookuper, opts ...ScopeOption) Scope {
	return scopePush(n, lookup, opts...)
}

func appendPath(current, next string) string {
	if next == "" {
		return current
	} else if current == "" {
		return next
	}
	return current + "." + next
}

type ValueScopeLookup struct {
	Value value.Value
}

func (v ValueScopeLookup) ScopeLookup(_ Scope, key string) (value.Value, bool, error) {
	return value.Lookup(v.Value, value.NewValue(key))
}
