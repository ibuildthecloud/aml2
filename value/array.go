package value

type Array struct {
	Array []Value
}

func NewArray(objs []any) *Array {
	a := &Array{
		Array: make([]Value, 0, len(objs)),
	}
	for _, obj := range objs {
		a.Array = append(a.Array, NewValue(obj))
	}
	return a
}

func (a *Array) Slice(start, end int64) (Value, bool, error) {
	if int(start) >= len(a.Array) || int(end) > len(a.Array) || start < 0 || end < 0 || start > end {
		return nil, false, nil
	}
	return &Array{
		Array: a.Array[start:end],
	}, true, nil
}

func (a *Array) Index(idx int64) (Value, bool, error) {
	if int(idx) >= len(a.Array) || idx < 0 {
		return nil, false, nil
	}
	return a.Array[idx], true, nil
}

func (a *Array) ToValues() []Value {
	return a.Array
}

func (a *Array) Kind() Kind {
	return ArrayKind
}

func (a *Array) NativeValue() any {
	result := make([]any, 0, len(a.Array))
	for _, v := range a.Array {
		result = append(result, v.NativeValue())
	}
	return result
}

func (a *Array) Len() (int64, error) {
	return int64(len(a.Array)), nil
}

func (a *Array) Merge(val Value) (Value, error) {
	return mergeNative(a, val)
}
