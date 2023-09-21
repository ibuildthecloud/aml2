package value

import "regexp"

type String string

func (s String) Kind() Kind {
	return StringKind
}

func (s String) NativeValue() (any, bool, error) {
	return (string)(s), true, nil
}

func (s String) Eq(right Value) (Value, error) {
	if err := assertType(right, StringKind); err != nil {
		return nil, err
	}
	rightString, err := ToString(right)
	if err != nil {
		return nil, err
	}
	return NewValue(string(s) == rightString), nil
}

func (s String) Neq(right Value) (Value, error) {
	if err := assertType(right, StringKind); err != nil {
		return nil, err
	}
	rightString, err := ToString(right)
	if err != nil {
		return nil, err
	}
	return NewValue(string(s) != rightString), nil
}

func (s String) Match(right Value) (bool, error) {
	if err := assertType(right, StringKind); err != nil {
		return false, err
	}

	rightString, err := ToString(right)
	if err != nil {
		return false, err
	}

	re, err := regexp.Compile(string(s))
	if err != nil {
		return false, err
	}

	m := re.FindStringIndex(rightString)
	// regexp must fully match string, not a subset of it
	return m != nil && m[0] == 0 && m[1] == len(rightString), nil
}
