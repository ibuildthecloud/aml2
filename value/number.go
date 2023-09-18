package value

import (
	"encoding/json"
	"fmt"
	"strconv"
)

type Number string

func (n Number) Kind() Kind {
	return NumberKind
}

func (n Number) NativeValue() (any, bool, error) {
	return n, true, nil
}

func toNum(n Value) (reti *int64, retf *float64, err error) {
	i, err := ToInt(n)
	if err == nil {
		reti = &i
	}

	f, err := ToFloat(n)
	if err == nil {
		retf = &f
	}

	if reti == nil && retf == nil {
		return nil, nil, fmt.Errorf("invalid number %s, not parsable as int or float", n)
	}

	return
}

func (n Number) binCompare(right Value, opName string, intFunc func(int64, int64) bool, floatFunc func(float64, float64) bool) (Value, error) {
	if right.Kind() != NumberKind {
		return nil, fmt.Errorf("can not compare (%s) number to invalid kind %s", opName, right.Kind())
	}

	li, lf, err := toNum(n)
	if err != nil {
		return nil, err
	}

	ri, rf, err := toNum(right)
	if err != nil {
		return nil, err
	}

	if li != nil && ri != nil {
		return NewValue(intFunc(*li, *ri)), nil
	} else if lf != nil && rf != nil {
		return NewValue(floatFunc(*lf, *rf)), nil
	} else {
		return nil, fmt.Errorf("can not compare (%s) incompatible numbers %s and %s", opName, n, right)
	}
}

func (n Number) binOp(right Value, opName string, intFunc func(int64, int64) int64, floatFunc func(float64, float64) float64) (Value, error) {
	if right.Kind() != NumberKind {
		return nil, fmt.Errorf("can not %s number to invalid kind %s", opName, right.Kind())
	}

	li, lf, err := toNum(n)
	if err != nil {
		return nil, err
	}

	ri, rf, err := toNum(right)
	if err != nil {
		return nil, err
	}

	if li != nil && ri != nil {
		return NewValue(intFunc(*li, *ri)), nil
	} else if lf != nil && rf != nil {
		return NewValue(floatFunc(*lf, *rf)), nil
	} else {
		return nil, fmt.Errorf("can not %s incompatible numbers %s and %s", opName, n, right)
	}
}

func (n Number) Sub(right Value) (Value, error) {
	return n.binOp(right, "subtract", func(i int64, i2 int64) int64 {
		return i - i2
	}, func(f float64, f2 float64) float64 {
		return f - f2
	})
}

func (n Number) Add(right Value) (Value, error) {
	return n.binOp(right, "add", func(i int64, i2 int64) int64 {
		return i + i2
	}, func(f float64, f2 float64) float64 {
		return f + f2
	})
}

func (n Number) Mul(right Value) (Value, error) {
	return n.binOp(right, "multiply", func(i int64, i2 int64) int64 {
		return i * i2
	}, func(f float64, f2 float64) float64 {
		return f * f2
	})
}

func (n Number) Div(right Value) (Value, error) {
	return n.binOp(right, "divide", func(i int64, i2 int64) int64 {
		return i / i2
	}, func(f float64, f2 float64) float64 {
		return f / f2
	})
}

func (n Number) Lt(right Value) (Value, error) {
	return n.binCompare(right, "less than", func(i int64, i2 int64) bool {
		return i < i2
	}, func(f float64, f2 float64) bool {
		return f < f2
	})
}

func (n Number) Gt(right Value) (Value, error) {
	return n.binCompare(right, "greater than", func(i int64, i2 int64) bool {
		return i > i2
	}, func(f float64, f2 float64) bool {
		return f > f2
	})
}

func (n Number) Le(right Value) (Value, error) {
	return n.binCompare(right, "less than equal", func(i int64, i2 int64) bool {
		return i <= i2
	}, func(f float64, f2 float64) bool {
		return f <= f2
	})
}

func (n Number) Ge(right Value) (Value, error) {
	return n.binCompare(right, "greater than equal", func(i int64, i2 int64) bool {
		return i >= i2
	}, func(f float64, f2 float64) bool {
		return f >= f2
	})
}

func (n Number) Eq(right Value) (Value, error) {
	if right.Kind() != NumberKind {
		return False, nil
	}
	return n.binCompare(right, "equals", func(i int64, i2 int64) bool {
		return i == i2
	}, func(f float64, f2 float64) bool {
		return f == f2
	})
}

func (n Number) Ne(right Value) (Value, error) {
	if right.Kind() != NumberKind {
		return False, nil
	}
	return n.binCompare(right, "not equals", func(i int64, i2 int64) bool {
		return i != i2
	}, func(f float64, f2 float64) bool {
		return f != f2
	})
}

func (n Number) ToInt() (int64, error) {
	return strconv.ParseInt(string(n), 10, 64)
}

func (n Number) ToFloat() (float64, error) {
	return strconv.ParseFloat(string(n), 64)
}

func (n Number) MarshalJSON() ([]byte, error) {
	return json.Marshal(json.Number(n))
}
