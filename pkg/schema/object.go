package schema

type Object struct {
	Path         string  `json:"path,omitempty"`
	Description  string  `json:"description,omitempty"`
	Fields       []Field `json:"fields,omitempty"`
	AllowNewKeys bool    `json:"allowNewKeys,omitempty"`
}

func (o *Object) GetFields() []Field {
	return o.Fields
}

type Field struct {
	Name        string    `json:"name,omitempty"`
	Description string    `json:"description,omitempty"`
	Type        FieldType `json:"type,omitempty"`
	Match       bool      `json:"match,omitempty"`
	Optional    bool      `json:"optional,omitempty"`
}

func (f *Field) GetFields() []Field {
	return []Field{*f}
}

type FieldType struct {
	Kind        string       `json:"kind,omitempty"`
	Object      *Object      `json:"object,omitempty"`
	Constraint  []Constraint `json:"constraint,omitempty"`
	Default     any          `json:"default,omitempty"`
	Alternative *FieldType   `json:"alternative,omitempty"`
}

type Constraint struct {
	Description string `json:"description,omitempty"`
	Op          string `json:"op,omitempty"`
	Left        any    `json:"left,omitempty"`
	Right       any    `json:"right,omitempty"`
}
