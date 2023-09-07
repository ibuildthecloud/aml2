package value

type Array []Value

func NewArray(objs []any) Array {
	a := make([]Value, 0, len(objs))
	for _, obj := range objs {
		a = append(a, NewValue(obj))
	}
	return a
}

func (a Array) Slice(start, end int64) (Value, bool, error) {
	if int(start) >= len(a) || int(end) > len(a) || start < 0 || end < 0 || start > end {
		return nil, false, nil
	}
	return a[start:end], true, nil
}

func (a Array) Index(idx int64) (Value, bool, error) {
	if int(idx) >= len(a) || idx < 0 {
		return nil, false, nil
	}
	return a[idx], true, nil
}

func (a Array) ToValues() []Value {
	return a
}

func (a Array) Kind() Kind {
	return ArrayKind
}

func (a Array) NativeValue() any {
	result := make([]any, 0, len(a))
	for _, v := range a {
		result = append(result, v.NativeValue())
	}
	return result
}

func (a Array) Len() (int64, error) {
	return int64(len(a)), nil
}

func (a Array) Merge(val Value) (Value, error) {
	return mergeNative(a, val)
}
