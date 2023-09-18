package eval

import "github.com/acorn-io/aml/value"

type Schema struct {
	Comments Comments
	Struct   *Struct
}

func (s *Schema) ToValue(scope Scope) (value.Value, bool, error) {
	return s.Struct.ToValue(scope.Push(ScopeData(nil), ScopeOption{
		Schema: true,
	}))
}

type contract struct {
	s     *Struct
	scope Scope
}

func (c *contract) RequiredKeys() (result []string, _ error) {
	keySeen := map[string]struct{}{}
	for _, field := range c.s.Fields {
		keys, err := field.Keys(c.scope)
		if err != nil {
			return nil, err
		}
		for _, key := range keys {
			if _, ok := keySeen[key]; ok {
				continue
			}
			keySeen[key] = struct{}{}
			result = append(result, key)
		}
	}
	return
}

func (c *contract) LookupValue(key string) (value.Value, bool, error) {
	return c.s.ScopeLookup(c.scope, key)
}
