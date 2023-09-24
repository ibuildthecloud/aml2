package eval

import (
	"context"
	"fmt"
	"strings"

	"github.com/acorn-io/aml/pkg/value"
)

type FunctionDefinition struct {
	Comments   Comments
	Pos        Position
	Body       *Struct
	ReturnBody bool
	AssignRoot bool
}

func (f *FunctionDefinition) ToValue(scope Scope) (value.Value, bool, error) {
	argsFields, bodyFields := f.splitFields()
	argNames, argsSchema, err := f.toSchema(scope, argsFields, "args", false)
	if err != nil {
		return nil, false, err
	}
	profileNames, profileSchema, err := f.toSchema(scope, argsFields, "profiles", true)
	if err != nil {
		return nil, false, err
	}
	body := &Struct{
		Fields: bodyFields,
	}
	return &Function{
		Scope:          scope,
		Body:           body,
		ArgsSchema:     argsSchema,
		ArgNames:       argNames,
		ProfileNames:   profileNames,
		ProfilesSchema: profileSchema,
		ReturnBody:     f.ReturnBody,
		AssignRoot:     f.AssignRoot,
	}, true, nil
}

func (f *FunctionDefinition) toSchema(scope Scope, argDefs []Field, fieldName string, allowNewFields bool) ([]string, value.Value, error) {
	s := Schema{
		AllowNewFields: allowNewFields,
		Struct: &Struct{
			Fields: argDefs,
		},
	}
	v, _, err := s.ToValue(scope)
	if err != nil {
		return nil, nil, err
	}

	args, ok, err := value.Lookup(v, value.NewValue(fieldName))
	if err != nil || !ok {
		return nil, v, err
	}

	keys, err := value.Keys(args)
	return keys, args, err
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
	Scope          Scope
	Body           Expression
	ArgsSchema     value.Value
	ArgNames       []string
	ProfilesSchema value.Value
	ProfileNames   []string
	ReturnBody     bool
	AssignRoot     bool
}

func (c *Function) Kind() value.Kind {
	return value.FuncKind
}

func (c *Function) getProfiles(v value.Value) (profiles []value.Value, _ bool, _ error) {
	v, ok, err := value.Lookup(v, value.NewValue("profiles"))
	if err != nil || !ok {
		return nil, ok, err
	} else if v.Kind() == value.UndefinedKind {
		return []value.Value{v}, true, nil
	}

	if v.Kind() != value.ArrayKind {
		return nil, false, fmt.Errorf("profiles type should be an array")
	}

	profileNames, err := value.ToValueArray(v)
	if err != nil {
		return nil, false, err
	}

	for _, profileName := range profileNames {
		profile, ok, err := value.Lookup(c.ProfilesSchema, profileName)
		if err != nil {
			return nil, false, err
		} else if !ok {
			if strings.HasSuffix(fmt.Sprint(profileName), "?") {
				continue
			}
			return nil, false, fmt.Errorf("missing profile: %s", profileName)
		} else {
			profiles = append(profiles, profile)
		}
	}

	return profiles, true, nil
}

func (c *Function) callArgumentToValue(args []value.CallArgument) (value.Value, error) {
	var (
		argValues []value.Value
		profiles  []value.Value
	)

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
		} else if profile, profilesSet, err := c.getProfiles(arg.Value); err != nil {
			return nil, err
		} else if profilesSet {
			profiles = append(profiles, profile...)
		} else {
			argValues = append(argValues, arg.Value)
		}
	}

	argValue, err := value.Merge(argValues...)
	if err != nil {
		return nil, err
	}

	if argValue == nil {
		argValue = value.NewObject(nil)
	}

	for i := len(profiles) - 1; i >= 0; i-- {
		argValue, err = value.Merge(profiles[i], argValue)
		if err != nil {
			return nil, err
		}
	}

	return value.Merge(c.ArgsSchema, argValue)
}

type rootLookup struct {
	f *Function
}

func (r rootLookup) ScopeLookup(scope Scope, key string) (value.Value, bool, error) {
	if key == "$" {
		return r.f.Body.ToValue(scope)
	}
	return nil, false, nil
}

type depthKey struct{}

const MaxCallDepth = 100

func (c *Function) Call(ctx context.Context, args []value.CallArgument) (value.Value, bool, error) {
	argsValue, err := c.callArgumentToValue(args)
	if err != nil {
		return nil, false, err
	}

	depth, _ := ctx.Value(depthKey{}).(int)
	if depth > MaxCallDepth {
		return nil, false, fmt.Errorf("exceed max call depth %d > %d", depth, MaxCallDepth)
	}
	ctx = context.WithValue(ctx, depthKey{}, depth+1)

	select {
	case <-ctx.Done():
		return nil, false, fmt.Errorf("context is closed: %w", ctx.Err())
	default:
	}

	var path string
	if c.Scope.Path() != "" {
		path = "()"
	}

	scope := c.Scope.Push(ScopeData(map[string]any{
		"args": argsValue,
	}), ScopeOption{
		Path:    path,
		Context: ctx,
	})
	if c.AssignRoot {
		scope = scope.Push(rootLookup{
			f: c,
		})
	}

	ret, ok, err := c.Body.ToValue(scope)
	if err != nil || !ok {
		return nil, ok, err
	}
	if c.ReturnBody {
		return ret, true, nil
	}
	return value.Lookup(ret, value.NewValue("return"))
}
