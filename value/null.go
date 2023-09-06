package value

func NewNull() *Null {
	return &Null{}
}

type Null struct {
}

func (n *Null) Eq(right Value) (Value, error) {
	if right.Kind() == NullKind {
		return True, nil
	}
	return False, nil
}

func (n *Null) Ne(right Value) (Value, error) {
	if right.Kind() == NullKind {
		return False, nil
	}
	return True, nil
}

func (n *Null) Kind() Kind {
	return NullKind
}

func (n *Null) NativeValue() any {
	return nil
}

func (n *Null) Merge(val Value) (Value, error) {
	return mergeNative(n, val)
}
