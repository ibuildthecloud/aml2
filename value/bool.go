package value

var (
	True  = Boolean(true)
	False = Boolean(false)
)

type Boolean bool

func (n Boolean) IsTrue() bool {
	return bool(n)
}

func (n Boolean) Kind() Kind {
	return BoolKind
}

func (n Boolean) NativeValue() (any, bool, error) {
	return (bool)(n), true, nil
}

func (n Boolean) Eq(right Value) (Value, error) {
	if err := assertType(right, BoolKind); err != nil {
		return nil, err
	}
	rightBool, err := ToBool(right)
	if err != nil {
		return nil, err
	}
	return NewValue(bool(n) == rightBool), nil
}

func (n Boolean) Ne(right Value) (Value, error) {
	if err := assertType(right, BoolKind); err != nil {
		return nil, err
	}
	rightBool, err := ToBool(right)
	if err != nil {
		return nil, err
	}
	return NewValue(bool(n) != rightBool), nil
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
