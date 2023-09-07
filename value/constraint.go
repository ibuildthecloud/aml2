package value

import (
	"fmt"
)

type Checker interface {
	Check(left Value) error
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

func (c *Constraint) check(f func(left, right Value) (Value, error), left, right Value) error {
	v, err := f(left, right)
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
	switch c.Op {
	case ">":
		return c.check(Gt, left, c.Right)
	case ">=":
		return c.check(Ge, left, c.Right)
	case "<":
		return c.check(Lt, left, c.Right)
	case "<=":
		return c.check(Le, left, c.Right)
	default:
		return fmt.Errorf("unknown operator for constraint: %s", c.Op)
	}
}

type OrConstraint struct {
	Left  Constraints
	Right Constraints
}

func (o *OrConstraint) Check(left Value) error {
	if err := o.Left.Check(left); err == nil {
		return nil
	}
	return o.Right.Check(left)
}
