package ast

import (
	"strings"
	"unicode/utf8"

	"github.com/acorn-io/aml/errors"
)

func isAllowedCharacter(ch rune) bool {
	return 'a' <= ch && ch <= 'z' || 'A' <= ch && ch <= 'Z'
}

func isAllowedDigit(ch rune) bool {
	return '0' <= ch && ch <= '9'
}

// IsValidIdent reports whether str is a valid identifier.
func IsValidIdent(ident string) bool {
	if ident == "" {
		return false
	}

	consumed := false
	if strings.HasPrefix(ident, "#") {
		ident = ident[1:]
		// Note: _#0 is not allowed by the spec, although _0 is.
		// TODO: set consumed to true here to allow #0.
		consumed = false
	}

	if !consumed {
		if r, _ := utf8.DecodeRuneInString(ident); isAllowedDigit(r) {
			return false
		}
	}

	for _, r := range ident {
		if isAllowedCharacter(r) || isAllowedDigit(r) || r == '_' || r == '$' {
			continue
		}
		return false
	}
	return true
}

// ErrIsExpression reports whether a label is an expression.
// This error is never returned directly. Use errors.Is.
var ErrIsExpression = errors.New("not a concrete label")
