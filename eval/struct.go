package eval

import (
	"github.com/acorn-io/aml/value"
)

type Struct struct {
	Comments Comments
	Fields   []Field
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
