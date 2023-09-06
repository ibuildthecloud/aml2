package value

var (
	True  = &Boolean{Boolean: true}
	False = &Boolean{}
)

type Boolean struct {
	Boolean bool
}

func (n *Boolean) Kind() Kind {
	return BoolKind
}

func (n *Boolean) NativeValue() any {
	return n.Boolean
}

func (n *Boolean) Eq(right Value) (Value, error) {
	if right.Kind() == BoolKind {
		return NewValue(n.Boolean == right.NativeValue().(bool)), nil
	}
	return False, nil
}

func (n *Boolean) Ne(right Value) (Value, error) {
	if right.Kind() == BoolKind {
		return NewValue(n.Boolean != right.NativeValue().(bool)), nil
	}
	return True, nil
}

func (n *Boolean) And(right Value) (Value, error) {
	b, err := ToBool(right)
	if err != nil {
		return nil, err
	}
	return NewValue(n.Boolean && b), nil
}

func (n *Boolean) Or(right Value) (Value, error) {
	b, err := ToBool(right)
	if err != nil {
		return nil, err
	}
	return NewValue(n.Boolean || b), nil
}

func (n *Boolean) Merge(val Value) (Value, error) {
	return mergeNative(n, val)
}
