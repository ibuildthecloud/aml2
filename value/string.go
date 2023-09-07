package value

type String string

func (s String) Kind() Kind {
	return StringKind
}

func (s String) NativeValue() any {
	return (string)(s)
}

func (s String) Eq(right Value) (Value, error) {
	if right.Kind() == StringKind {
		return NewValue((string)(s) == s.NativeValue().(string)), nil
	}
	return False, nil
}

func (s String) Ne(right Value) (Value, error) {
	if right.Kind() == StringKind {
		return NewValue((string)(s) != s.NativeValue().(string)), nil
	}
	return True, nil
}

func (s String) Merge(val Value) (Value, error) {
	return mergeNative(s, val)
}
