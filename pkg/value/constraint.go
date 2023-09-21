package value

import (
	"fmt"
)

type Checker interface {
	Check(left Value) error
	Description() string
	OpString() string
	LeftNative() (any, bool, error)
	RightNative() (any, bool, error)
}

type Constraints []Checker

func (c Constraints) Check(left Value) error {
	for _, checker := range c {
		err := checker.Check(left)
		if err != nil {
			return err
		}
	}
	return nil
}

type Constraint struct {
	Op    string
	Right Value
}

func (c *Constraint) Description() string {
	return ""
}

func (c *Constraint) OpString() string {
	return c.Op
}

func (c *Constraint) LeftNative() (any, bool, error) {
	return nil, false, nil
}

func (c *Constraint) RightNative() (any, bool, error) {
	return NativeValue(c.Right)
}

func (c *Constraint) check(op Operator, left, right Value) error {
	v, err := BinaryOperation(op, left, right)
	if err != nil {
		return err
	}
	b, err := ToBool(v)
	if err != nil {
		return err
	}
	if !b {
		return fmt.Errorf("unmatched constraint %s %s %s", left, c.Op, right)
	}
	return nil
}

func (c *Constraint) Check(left Value) error {
	switch Operator(c.Op) {
	case GtOp, GeOp, LtOp, LeOp, EqOp, NeqOp:
		return c.check(Operator(c.Op), left, c.Right)
	default:
		return fmt.Errorf("unknown operator for constraint: %s", c.Op)
	}
}
