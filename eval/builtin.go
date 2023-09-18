package eval

import "github.com/acorn-io/aml/value"

var Builtin ScopeData

func init() {
	Builtin = ScopeData{}

	for _, kind := range value.Kinds {
		if kind == value.UndefinedKind {
			continue
		}
		Builtin[string(kind)] = &value.TypeSchema{
			KindValue: kind,
		}
	}
}
