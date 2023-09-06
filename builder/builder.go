package builder

import (
	"errors"
	"fmt"
	"strings"

	"github.com/acorn-io/aml/ast"
	"github.com/acorn-io/aml/eval"
	"github.com/acorn-io/aml/token"
	"github.com/acorn-io/aml/value"
)

func Build(file *ast.File) (*eval.Struct, error) {
	return fileToObject(file)
}

func fileToObject(file *ast.File) (*eval.Struct, error) {
	fields, err := declsToFields(file.Decls)
	if err != nil {
		return nil, err
	}

	return &eval.Struct{
		Comments: getComments(file),
		Fields:   fields,
	}, err
}

func declsToFields(decls []ast.Decl) (result []eval.Field, err error) {
	var (
		errs   []error
		fields []eval.Field
	)

	for _, decl := range decls {
		field, err := declToField(decl)
		if err != nil {
			errs = append(errs, err)
			continue
		}
		fields = append(fields, field)
	}

	return fields, errors.Join(errs...)
}

func declToField(decl ast.Decl) (_ eval.Field, err error) {
	switch v := decl.(type) {
	case *ast.Field:
		var result eval.KeyValue
		result.Comments = getComments(decl)
		result.Optional = v.Constraint == token.OPTION
		result.Key, err = labelToKey(v.Label)
		if err != nil {
			return &result, err
		}

		result.Value, err = exprToExpression(v.Value)
		return &result, err
	case *ast.EmbedDecl:
		var result eval.Embedded
		result.Comments = getComments(decl)
		result.Expression, err = exprToExpression(v.Expr)
		return &result, err
	case *ast.LetClause:
		var result eval.KeyValue
		result.Comments = getComments(decl)
		result.Local = true
		result.Key, err = labelToKey(v.Ident)
		if err != nil {
			return nil, err
		}

		result.Value, err = exprToExpression(v.Expr)
		return &result, err
	case *ast.Comprehension:
		var result eval.Embedded
		result.Comments = getComments(decl)
		result.Expression, err = comprehensionToExpression(v, true)
		return &result, err
	default:
		return nil, NewErrUnknownError(decl)
	}
}

func interpolationToExpression(comp *ast.Interpolation) (eval.Expression, error) {
	result := &eval.Interpolation{}

	for i := range comp.Elts {
		switch {
		case i == 0:
			lit := *comp.Elts[i].(*ast.BasicLit)
			lit.Value = strings.TrimSuffix(lit.Value, "\\(")
			result.Parts = append(result.Parts, lit.Value)
		case i == len(comp.Elts)-1:
			lit := *comp.Elts[i].(*ast.BasicLit)
			lit.Value = strings.TrimPrefix(lit.Value, ")")
			result.Parts = append(result.Parts, lit.Value)
		case i%2 == 0:
			lit := *comp.Elts[i].(*ast.BasicLit)
			lit.Value = strings.TrimPrefix(lit.Value, ")")
			lit.Value = strings.TrimSuffix(lit.Value, "\\(")
			result.Parts = append(result.Parts, lit.Value)
		case i%2 == 1:
			expr, err := exprToExpression(comp.Elts[i])
			if err != nil {
				return nil, err
			}
			result.Parts = append(result.Parts, expr)
		}
	}

	return result, nil
}

func comprehensionToExpression(comp *ast.Comprehension, field bool) (eval.Expression, error) {
	value, err := exprToExpression(comp.Value)
	if err != nil {
		return nil, err
	}

	switch c := comp.Clauses[0].(type) {
	case *ast.IfClause:
		condition, err := exprToExpression(c.Condition)
		if err != nil {
			return nil, err
		}
		return &eval.If{
			Comments:  getComments(c),
			Condition: condition,
			Value:     value,
		}, nil
	case *ast.ForClause:
		e, err := forToFor(c, value)
		if err != nil {
			return nil, err
		}
		if field {
			return &eval.MergeObjectArray{
				Array: e,
			}, nil
		}
		return e, nil
	default:
		return nil, NewErrUnknownError(comp.Clauses[0])
	}
}

func forToFor(comp *ast.ForClause, value eval.Expression) (*eval.For, error) {
	var (
		result = &eval.For{
			Comments: getComments(comp),
			Body:     value,
		}
		err error
	)

	if comp.Key != nil {
		result.Key, err = labelToExpression(comp.Key)
		if err != nil {
			return nil, err
		}
	}

	result.Value, err = labelToExpression(comp.Value)
	if err != nil {
		return nil, err
	}

	result.List, err = exprToExpression(comp.Source)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func basicListToValue(lit *ast.BasicLit) (eval.Expression, error) {
	switch lit.Kind {
	case token.INT, token.FLOAT:
		return eval.Value{
			Value: value.Number(lit.Value),
		}, nil
	case token.STRING:
		s, err := value.Unquote(lit.Value)
		if err != nil {
			return nil, err
		}
		return eval.Value{
			Value: &value.String{
				String: s,
			},
		}, nil
	case token.TRUE:
		return eval.Value{
			Value: &value.Boolean{
				Boolean: true,
			},
		}, nil
	case token.FALSE:
		return eval.Value{
			Value: &value.Boolean{
				Boolean: false,
			},
		}, nil
	case token.NULL:
		return eval.Value{
			Value: &value.Null{},
		}, nil
	default:
		return nil, fmt.Errorf("unknown literal kind %s, value %s at %s", lit.Kind.String(), lit.Value, lit.Pos())
	}
}

func structToExpression(s *ast.StructLit) (eval.Expression, error) {
	fields, err := declsToFields(s.Elts)
	if err != nil {
		return nil, err
	}
	return &eval.Struct{
		Comments: getComments(s),
		Fields:   fields,
	}, err
}

func listToExpression(list *ast.ListLit) (eval.Expression, error) {
	exprs, err := exprsToExpressions(list.Elts)
	if err != nil {
		return nil, err
	}
	return &eval.Array{
		Comments: getComments(list),
		Items:    exprs,
	}, nil
}

func exprsToExpressions(exprs []ast.Expr) (result []eval.Expression, _ error) {
	var errs []error
	for _, expr := range exprs {
		newExpr, err := exprToExpression(expr)
		if err != nil {
			errs = append(errs, err)
			continue
		}
		result = append(result, newExpr)
	}
	return result, errors.Join(errs...)
}

func unaryToExpression(bin *ast.UnaryExpr) (eval.Expression, error) {
	left, err := exprToExpression(bin.X)
	if err != nil {
		return nil, err
	}

	return &eval.Op{
		Comments: getComments(bin),
		Unary:    true,
		Operator: bin.Op.String(),
		Left:     left,
		Pos:      pos(bin.OpPos),
	}, nil
}

func pos(t token.Pos) eval.Position {
	return eval.Position(t.Position())
}

func binaryToExpression(bin *ast.BinaryExpr) (eval.Expression, error) {
	left, err := exprToExpression(bin.X)
	if err != nil {
		return nil, err
	}

	right, err := exprToExpression(bin.Y)
	if err != nil {
		return nil, err
	}

	return &eval.Op{
		Comments: getComments(bin),
		Operator: bin.Op.String(),
		Left:     left,
		Right:    right,
		Pos:      pos(bin.OpPos),
	}, nil
}

func parensToExpression(parens *ast.ParenExpr) (eval.Expression, error) {
	expr, err := exprToExpression(parens.X)
	return &eval.Parens{
		Comments: getComments(parens),
		Expr:     expr,
	}, err
}

func identToExpression(ident *ast.Ident) (eval.Expression, error) {
	key, err := value.Unquote(ident.Name)
	if err != nil {
		return nil, err
	}
	return &eval.Lookup{
		Comments: getComments(ident),
		Pos:      pos(ident.NamePos),
		Key:      key,
	}, nil
}

func selectorToExpression(sel *ast.SelectorExpr) (eval.Expression, error) {
	key, err := labelToExpression(sel.Sel)
	if err != nil {
		return nil, err
	}

	selExpr, err := exprToExpression(sel.X)
	if err != nil {
		return nil, err
	}

	return &eval.Selector{
		Comments: getComments(sel),
		Pos:      pos(sel.Pos()),
		Base:     selExpr,
		Key:      key,
	}, nil
}

func indexToExpression(indexExpr *ast.IndexExpr) (eval.Expression, error) {
	base, err := exprToExpression(indexExpr.X)
	if err != nil {
		return nil, err
	}

	index, err := exprToExpression(indexExpr.Index)
	if err != nil {
		return nil, err
	}

	return &eval.Index{
		Comments: getComments(indexExpr),
		Pos:      pos(indexExpr.Pos()),
		Base:     base,
		Index:    index,
	}, nil
}

func sliceToExpression(sliceExpr *ast.SliceExpr) (eval.Expression, error) {
	base, err := exprToExpression(sliceExpr.X)
	if err != nil {
		return nil, err
	}

	low, err := exprToExpression(sliceExpr.Low)
	if err != nil {
		return nil, err
	}

	high, err := exprToExpression(sliceExpr.High)
	if err != nil {
		return nil, err
	}

	return &eval.Slice{
		Comments: getComments(sliceExpr),
		Pos:      pos(sliceExpr.Lbrack),
		Base:     base,
		Start:    low,
		End:      high,
	}, nil
}

func callToExpression(callExpr *ast.CallExpr) (eval.Expression, error) {
	f, err := exprToExpression(callExpr.Fun)
	if err != nil {
		return nil, err
	}

	args, err := exprsToExpressions(callExpr.Args)
	if err != nil {
		return nil, err
	}

	return &eval.Call{
		Comments: getComments(callExpr),
		Pos:      pos(callExpr.Lparen),
		Func:     f,
		Args:     args,
	}, nil
}

func exprToExpression(expr ast.Expr) (eval.Expression, error) {
	if expr == nil {
		return nil, nil
	}

	switch n := expr.(type) {
	case *ast.BasicLit:
		return basicListToValue(n)
	case *ast.StructLit:
		return structToExpression(n)
	case *ast.ListLit:
		return listToExpression(n)
	case *ast.BinaryExpr:
		return binaryToExpression(n)
	case *ast.UnaryExpr:
		return unaryToExpression(n)
	case *ast.ParenExpr:
		return parensToExpression(n)
	case *ast.Ident:
		return identToExpression(n)
	case *ast.SelectorExpr:
		return selectorToExpression(n)
	case *ast.IndexExpr:
		return indexToExpression(n)
	case *ast.SliceExpr:
		return sliceToExpression(n)
	case *ast.CallExpr:
		return callToExpression(n)
	case *ast.Comprehension:
		return comprehensionToExpression(n, false)
	case *ast.Interpolation:
		return interpolationToExpression(n)
	default:
		return nil, NewErrUnknownError(n)
	}
}

func labelToKey(label ast.Label) (eval.Key, error) {
	expr, err := labelToExpression(label)
	if err != nil {
		return eval.Key{}, err
	}
	return eval.Key{
		Value: expr,
	}, nil
}

func labelToExpression(expr ast.Label) (eval.Expression, error) {
	if expr == nil {
		return nil, nil
	}

	switch n := expr.(type) {
	case *ast.BasicLit:
		s, err := value.Unquote(n.Value)
		if err != nil {
			return nil, err
		}
		return eval.Value{
			Value: &value.String{
				String: s,
			},
		}, nil
	case *ast.Ident:
		s, err := value.Unquote(n.Name)
		if err != nil {
			return nil, err
		}
		return eval.Value{
			Value: &value.String{
				String: s,
			},
		}, nil
	case *ast.Interpolation:
		return exprToExpression(n)
	default:
		return nil, NewErrUnknownError(n)
	}
}
