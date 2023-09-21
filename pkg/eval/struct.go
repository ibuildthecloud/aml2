package eval

import (
	"strings"

	"github.com/acorn-io/aml/pkg/schema"
	"github.com/acorn-io/aml/pkg/value"
)

type Struct struct {
	Comments Comments
	Fields   []Field
}

func (s *Struct) FieldsSchema(scope Scope, seen map[string]struct{}) (result []schema.Field, error error) {
	scope = scope.Push(s)
	for _, field := range s.Fields {
		schema, err := field.Schema(scope, seen)
		if err != nil {
			return nil, err
		}
		result = append(result, schema...)
	}
	return
}

func (s *Struct) ScopeLookup(scope Scope, key string) (value.Value, bool, error) {
	var values []value.Value
	scope = scope.Push(s)

	for _, field := range s.Fields {
		val, ok, err := field.ToValueForKey(scope, key)
		if err != nil {
			return nil, false, err
		}
		if !ok {
			continue
		}
		values = append(values, val)
	}

	result, err := value.Merge(values...)
	return result, result != nil, err
}

func (s *Struct) ToValue(scope Scope) (value.Value, bool, error) {
	if scope.IsSchema() {
		return value.NewValue(&contract{
			s:     s,
			scope: scope,
		}), true, nil
	}

	scope = scope.Push(s)
	values, err := FieldsToValues(scope.Push(s), s.Fields)
	if err != nil {
		return nil, false, err
	}

	result, err := value.Merge(values...)
	if err != nil {
		return nil, false, err
	}
	if result == nil {
		return value.NewObject(nil), true, nil
	}
	return result, true, nil
}

type Comments struct {
	Comments [][]string
}

func (c Comments) Last() string {
	if len(c.Comments) == 0 {
		return ""
	}
	return strings.Join(c.Comments[len(c.Comments)-1], "\n")
}
