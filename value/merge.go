package value

import "fmt"

func mergeNative(left, right Value) (Value, error) {
	if merged, err := mergeNull(left, right); merged != nil || err != nil {
		return merged, err
	}
	return right, nil
}

func mergeNull(left, right Value) (Value, error) {
	if left.Kind() == SchemaKind && right.Kind() == ObjectKind {
		return nil, nil
	}
	if left.Kind() == NullKind {
		return right, nil
	}
	if right.Kind() == NullKind {
		return left, nil
	}
	if left.Kind() != right.Kind() {
		return nil, fmt.Errorf("can not override field kind %s with kind %s", left.Kind(), right.Kind())
	}
	return nil, nil
}
