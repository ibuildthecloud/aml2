package value

var (
	True  = Boolean(true)
	False = Boolean(false)
)

type Boolean bool

func (n Boolean) Kind() Kind {
	return BoolKind
}

func (n Boolean) NativeValue() any {
	return (bool)(n)
}

func (n Boolean) Eq(right Value) (Value, error) {
	if right.Kind() == BoolKind {
		return NewValue((bool)(n) == right.NativeValue().(bool)), nil
	}
	return False, nil
}

func (n Boolean) Ne(right Value) (Value, error) {
	if right.Kind() == BoolKind {
		return NewValue((bool)(n) != right.NativeValue().(bool)), nil
	}
	return True, nil
}

func (n Boolean) And(right Value) (Value, error) {
	b, err := ToBool(right)
	if err != nil {
		return nil, err
	}
	return NewValue((bool)(n) && b), nil
}

func (n Boolean) Or(right Value) (Value, error) {
	b, err := ToBool(right)
	if err != nil {
		return nil, err
	}
	return NewValue((bool)(n) || b), nil
}

func (n Boolean) Merge(val Value) (Value, error) {
	return mergeNative(n, val)
}
