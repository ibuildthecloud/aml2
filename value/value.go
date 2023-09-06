package value

const (
	NullKind   = Kind("null")
	StringKind = Kind("string")
	BoolKind   = Kind("bool")
	NumberKind = Kind("number")
	ArrayKind  = Kind("array")
	ObjectKind = Kind("object")
	FuncKind   = Kind("func")
)

type Kind string

type Value interface {
	Kind() Kind
	NativeValue() any
	Merge(val Value) (Value, error)
}
