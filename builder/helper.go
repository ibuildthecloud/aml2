package builder

import (
	"fmt"
	"strings"
	"unicode"

	"github.com/acorn-io/aml/ast"
	"github.com/acorn-io/aml/eval"
)

func getComments(node ast.Node) (result eval.Comments) {
	for _, cg := range ast.Comments(node) {
		var group []string

		if cg != nil {
			for _, c := range cg.List {
				l := strings.TrimLeftFunc(strings.TrimPrefix(c.Text, "//"), unicode.IsSpace)
				group = append(group, l)
			}
		}
		result.Comments = append(result.Comments, group)
	}
	return
}

type ErrUnknownError struct {
	Node ast.Node
}

func (e *ErrUnknownError) Error() string {
	return fmt.Sprintf("unknown node %T encountered at %s", e.Node, e.Node.Pos())
}

func NewErrUnknownError(node ast.Node) error {
	return &ErrUnknownError{
		Node: node,
	}
}
