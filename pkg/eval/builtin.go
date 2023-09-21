package eval

import "github.com/acorn-io/aml/pkg/value"

var Builtin Scope

func init() {
	data := map[string]any{}
	for _, kind := range value.Kinds {
		if kind == value.UndefinedKind {
			continue
		}
		data[string(kind)] = &value.TypeSchema{
			KindValue: kind,
		}
	}
	Builtin = EmptyScope{}.Push(ScopeData(data))
}
