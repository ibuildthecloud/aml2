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
	return fileToObject(file, file.Schema)
}

func fileToObject(file *ast.File, schema bool) (*eval.Struct, error) {
	fields, err := declsToFields(file.Decls, schema)
	if err != nil {
		return nil, err
	}

	return &eval.Struct{
		Comments: getComments(file),
		Fields:   fields,
	}, err
}

func declsToFields(decls []ast.Decl, schema bool) (result []eval.Field, err error) {
	var (
		errs   []error
		fields []eval.Field
	)

	for _, decl := range decls {
		field, err := declToField(decl, schema)
		if err != nil {
			errs = append(errs, err)
			continue
		}
		fields = append(fields, field)
	}

	return fields, errors.Join(errs...)
}

func processKeyForSchema(f *eval.KeyValue) {
	str, ok := f.Key.Value.(eval.Value)
	if !ok {
		return
	}
	s, err := value.ToString(str.Value)
	if err != nil {
		return
	}
	if strings.HasPrefix(s, "#") {
		f.Schema = true
		f.Key.Value = &eval.Value{
			Value: value.NewValue(strings.TrimPrefix(s, "#")),
		}
	}
}

func declToField(decl ast.Decl, schema bool) (_ eval.Field, err error) {
	switch v := decl.(type) {
	case *ast.Field:
		var result eval.KeyValue
		result.Comments = getComments(decl)
		result.Optional = v.Constraint == token.OPTION
		result.Key, err = labelToKey(v.Label, schema)
		if err != nil {
			return &result, err
		}
		processKeyForSchema(&result)
		result.Value, err = exprToExpression(v.Value, schema || result.Schema)
		return &result, err
	case *ast.EmbedDecl:
		var result eval.Embedded
		result.Comments = getComments(decl)
		result.Expression, err = exprToExpression(v.Expr, schema)
		return &result, err
	case *ast.LetClause:
		var result eval.KeyValue
		result.Comments = getComments(decl)
		result.Local = true
		result.Key, err = labelToKey(v.Ident, schema)
		if err != nil {
			return nil, err
		}

		result.Value, err = exprToExpression(v.Expr, result.Schema)
		return &result, err
	case *ast.Comprehension:
		var result eval.Embedded
		result.Comments = getComments(decl)
		result.Expression, err = comprehensionToExpression(v, true, schema)
		return &result, err
	default:
		return nil, NewErrUnknownError(decl)
	}
}

func interpolationToExpression(comp *ast.Interpolation, schema bool) (eval.Expression, error) {
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
			expr, err := exprToExpression(comp.Elts[i], schema)
			if err != nil {
				return nil, err
			}
			result.Parts = append(result.Parts, expr)
		}
	}

	return result, nil
}

func comprehensionToExpression(comp *ast.Comprehension, field, schema bool) (eval.Expression, error) {
	value, err := exprToExpression(comp.Value, schema)
	if err != nil {
		return nil, err
	}

	switch c := comp.Clauses[0].(type) {
	case *ast.IfClause:
		condition, err := exprToExpression(c.Condition, schema)
		if err != nil {
			return nil, err
		}
		return &eval.If{
			Comments:  getComments(c),
			Condition: condition,
			Value:     value,
		}, nil
	case *ast.ForClause:
		e, err := forToFor(c, value, schema)
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

func forToFor(comp *ast.ForClause, value eval.Expression, schema bool) (*eval.For, error) {
	var (
		result = &eval.For{
			Comments: getComments(comp),
			Body:     value,
		}
		err error
	)

	if comp.Key != nil {
		result.Key, err = labelToExpression(comp.Key, schema)
		if err != nil {
			return nil, err
		}
	}

	result.Value, err = labelToExpression(comp.Value, schema)
	if err != nil {
		return nil, err
	}

	result.List, err = exprToExpression(comp.Source, schema)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func basicListToValue(lit *ast.BasicLit, schema bool) (eval.Expression, error) {
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
			Value: value.NewValue(s),
		}, nil
	case token.TRUE:
		return eval.Value{
			Value: value.True,
		}, nil
	case token.FALSE:
		return eval.Value{
			Value: value.False,
		}, nil
	case token.NULL:
		return eval.Value{
			Value: &value.Null{},
		}, nil
	default:
		return nil, fmt.Errorf("unknown literal kind %s, value %s at %s", lit.Kind.String(), lit.Value, lit.Pos())
	}
}

func structToExpression(s *ast.StructLit, schema bool) (eval.Expression, error) {
	fields, err := declsToFields(s.Elts, schema)
	if err != nil {
		return nil, err
	}
	return &eval.Struct{
		Comments: getComments(s),
		Fields:   fields,
		Schema:   schema,
	}, err
}

func listToExpression(list *ast.ListLit, schema bool) (eval.Expression, error) {
	exprs, err := exprsToExpressions(list.Elts, schema)
	if err != nil {
		return nil, err
	}
	return &eval.Array{
		Comments: getComments(list),
		Items:    exprs,
	}, nil
}

func exprsToExpressions(exprs []ast.Expr, schema bool) (result []eval.Expression, _ error) {
	var errs []error
	for _, expr := range exprs {
		newExpr, err := exprToExpression(expr, schema)
		if err != nil {
			errs = append(errs, err)
			continue
		}
		result = append(result, newExpr)
	}
	return result, errors.Join(errs...)
}

func unaryToExpression(bin *ast.UnaryExpr, schema bool) (eval.Expression, error) {
	left, err := exprToExpression(bin.X, schema)
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

func binaryToExpression(bin *ast.BinaryExpr, schema bool) (eval.Expression, error) {
	left, err := exprToExpression(bin.X, schema)
	if err != nil {
		return nil, err
	}

	right, err := exprToExpression(bin.Y, schema)
	if err != nil {
		return nil, err
	}

	return &eval.Op{
		Schema:   schema,
		Comments: getComments(bin),
		Operator: bin.Op.String(),
		Left:     left,
		Right:    right,
		Pos:      pos(bin.OpPos),
	}, nil
}

func parensToExpression(parens *ast.ParenExpr, schema bool) (eval.Expression, error) {
	expr, err := exprToExpression(parens.X, schema)
	return &eval.Parens{
		Comments: getComments(parens),
		Expr:     expr,
	}, err
}

func identToExpression(ident *ast.Ident, schema bool) (eval.Expression, error) {
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

func selectorToExpression(sel *ast.SelectorExpr, schema bool) (eval.Expression, error) {
	key, err := labelToExpression(sel.Sel, schema)
	if err != nil {
		return nil, err
	}

	selExpr, err := exprToExpression(sel.X, schema)
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

func indexToExpression(indexExpr *ast.IndexExpr, schema bool) (eval.Expression, error) {
	base, err := exprToExpression(indexExpr.X, schema)
	if err != nil {
		return nil, err
	}

	index, err := exprToExpression(indexExpr.Index, schema)
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

func sliceToExpression(sliceExpr *ast.SliceExpr, schema bool) (eval.Expression, error) {
	base, err := exprToExpression(sliceExpr.X, schema)
	if err != nil {
		return nil, err
	}

	low, err := exprToExpression(sliceExpr.Low, schema)
	if err != nil {
		return nil, err
	}

	high, err := exprToExpression(sliceExpr.High, schema)
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

func callToExpression(callExpr *ast.CallExpr, schema bool) (eval.Expression, error) {
	f, err := exprToExpression(callExpr.Fun, schema)
	if err != nil {
		return nil, err
	}

	args, err := exprsToExpressions(callExpr.Args, schema)
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

func exprToExpression(expr ast.Expr, schema bool) (eval.Expression, error) {
	if expr == nil {
		return nil, nil
	}

	switch n := expr.(type) {
	case *ast.BasicLit:
		return basicListToValue(n, schema)
	case *ast.StructLit:
		return structToExpression(n, schema)
	case *ast.ListLit:
		return listToExpression(n, schema)
	case *ast.BinaryExpr:
		return binaryToExpression(n, schema)
	case *ast.UnaryExpr:
		return unaryToExpression(n, schema)
	case *ast.ParenExpr:
		return parensToExpression(n, schema)
	case *ast.Ident:
		return identToExpression(n, schema)
	case *ast.SelectorExpr:
		return selectorToExpression(n, schema)
	case *ast.IndexExpr:
		return indexToExpression(n, schema)
	case *ast.SliceExpr:
		return sliceToExpression(n, schema)
	case *ast.CallExpr:
		return callToExpression(n, schema)
	case *ast.Comprehension:
		return comprehensionToExpression(n, false, schema)
	case *ast.Interpolation:
		return interpolationToExpression(n, schema)
	default:
		return nil, NewErrUnknownError(n)
	}
}

func labelToKey(label ast.Label, schema bool) (eval.Key, error) {
	expr, err := labelToExpression(label, schema)
	if err != nil {
		return eval.Key{}, err
	}
	return eval.Key{
		Value: expr,
	}, nil
}

func labelToExpression(expr ast.Label, schema bool) (eval.Expression, error) {
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
			Value: value.NewValue(s),
		}, nil
	case *ast.Ident:
		s, err := value.Unquote(n.Name)
		if err != nil {
			return nil, err
		}
		return eval.Value{
			Value: value.NewValue(s),
		}, nil
	case *ast.Interpolation:
		return exprToExpression(n, schema)
	default:
		return nil, NewErrUnknownError(n)
	}
}
