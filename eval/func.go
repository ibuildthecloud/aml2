package eval

import (
	"fmt"

	"github.com/acorn-io/aml/value"
)

type FunctionDefinition struct {
	Comments Comments
	Pos      Position
	Body     *Struct
}

func (f *FunctionDefinition) ToValue(scope Scope) (value.Value, bool, error) {
	argsFields, bodyFields := f.splitFields()
	argNames, argsSchema, err := f.toArgsSchema(scope, argsFields)
	if err != nil {
		return nil, false, err
	}
	return &Function{
		Scope: scope,
		Body: &Struct{
			Fields: bodyFields,
		},
		ArgsSchema: argsSchema,
		ArgNames:   argNames,
	}, true, nil
}

func (f *FunctionDefinition) toArgsSchema(scope Scope, argDefs []Field) ([]string, value.Value, error) {
	s := Schema{
		Struct: &Struct{
			Fields: argDefs,
		},
	}
	v, _, err := s.ToValue(scope)
	if err != nil {
		return nil, nil, err
	}

	args, ok, err := value.Lookup(v, value.NewValue("args"))
	if err != nil || !ok {
		return nil, v, err
	}

	keys, err := value.Keys(args)
	return keys, v, err
}

func (f *FunctionDefinition) splitFields() (argFields []Field, bodyFields []Field) {
	for _, field := range f.Body.Fields {
		arg, ok := field.(IsArgumentDefinition)
		if ok && arg.IsArgumentDefinition() {
			argFields = append(argFields, field)
			continue
		}
		bodyFields = append(bodyFields, field)
	}
	return
}

type IsArgumentDefinition interface {
	IsArgumentDefinition() bool
}

type Function struct {
	Scope      Scope
	Body       Expression
	ArgsSchema value.Value
	ArgNames   []string
}

func (c *Function) Kind() value.Kind {
	return value.FuncKind
}

func (c *Function) Merge(val value.Value) (value.Value, error) {
	//TODO implement me
	panic("implement me")
}

func (c *Function) callArgumentToValue(args []value.CallArgument) (value.Value, error) {
	var argValues []value.Value
	for i, arg := range args {
		if arg.Positional {
			if i >= len(c.ArgNames) {
				return nil, fmt.Errorf("invalid arg index %d, args len %d", i, len(c.ArgNames))
			}
			argValues = append(argValues, value.NewObject(map[string]any{
				c.ArgNames[i]: arg.Value,
			}))
		} else if arg.Value.Kind() != value.ObjectKind {
			return nil, fmt.Errorf("invalid argument kind %s (index %d)", arg.Value.Kind(), i)
		} else {
			argValues = append(argValues, arg.Value)
		}
	}

	argValue, err := value.Merge(argValues...)
	if err != nil {
		return nil, err
	}

	return value.Merge(c.ArgsSchema, value.NewValue(map[string]any{
		"args": argValue,
	}))
}

func (c *Function) Call(args []value.CallArgument) (value.Value, bool, error) {
	argsValue, err := c.callArgumentToValue(args)
	if err != nil {
		return nil, false, err
	}

	ret, ok, err := c.Body.ToValue(c.Scope.Push(ValueScopeLookup{
		Value: argsValue,
	}))
	if err != nil || !ok {
		return nil, ok, err
	}
	return value.Lookup(ret, value.NewValue("return"))
}
