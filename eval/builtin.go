package eval

import "github.com/acorn-io/aml/value"

var Builtin Data

func init() {
	Builtin = Data{}

	for _, kind := range value.Kinds {
		Builtin[string(kind)] = &value.Schema{
			KindValue: kind,
		}
	}
}
