package value

import "fmt"

type CallArgument struct {
	Positional bool
	Value      Value
}

type Caller interface {
	Call(args []CallArgument) (Value, bool, error)
}

func Call(value Value, args ...CallArgument) (Value, bool, error) {
	if value.Kind() == UndefinedKind {
		return value, true, nil
	}
	if caller, ok := value.(Caller); ok {
		return caller.Call(args)
	}
	return nil, false, fmt.Errorf("kind %s is not callable", value.Kind())
}
