// Copyright 2018 The CUE Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// This file contains the exported entry points for invoking the

package parser

import (
	"io"

	"github.com/acorn-io/aml/pkg/ast"
	"github.com/acorn-io/aml/pkg/errors"
	"github.com/acorn-io/aml/pkg/token"
)

// Option specifies a parse option.
type Option func(p *parser)

var (
	// Trace causes parsing to print a trace of parsed productions.
	Trace    Option = traceOpt
	traceOpt        = func(p *parser) {
		p.mode |= traceMode
	}

	// AllErrors causes all errors to be reported (not just the first 10 on different lines).
	AllErrors Option = allErrors
	allErrors        = func(p *parser) {
		p.mode |= allErrorsMode
	}

	AllowMatch Option = allowMatch
	allowMatch        = func(p *parser) {
		p.mode |= allowMatchMode
	}
)

// A mode value is a set of flags (or 0).
// They control the amount of source code parsed and other optional
// parser functionality.
type mode uint

const (
	parseCommentsMode mode = 1 << iota // parse comments and add them to AST
	allowMatchMode
	traceMode     // print a trace of parsed productions
	allErrorsMode // report all errors (not just the first 10 on different lines)
)

func ParseFile(filename string, src io.Reader, mode ...Option) (f *ast.File, err error) {
	text, err := io.ReadAll(src)
	if err != nil {
		return nil, err
	}

	var pp parser
	defer func() {
		if pp.panicking {
			_ = recover()
		}

		// set result values
		if f == nil {
			// source is not a valid Go source file - satisfy
			// ParseFile API and return a valid (but) empty
			// *File
			f = &ast.File{
				// Scope: NewScope(nil),
			}
		}

		err = errors.Sanitize(pp.errors)
	}()

	// parse source
	pp.init(filename, text, mode)
	f = pp.parseFile()
	if f == nil {
		return nil, pp.errors
	}
	f.Filename = filename

	return f, pp.errors
}

func ParseExpr(filename string, src io.Reader, mode ...Option) (ast.Expr, error) {
	text, err := io.ReadAll(src)
	if err != nil {
		return nil, err
	}

	var p parser
	defer func() {
		if p.panicking {
			_ = recover()
		}
		err = errors.Sanitize(p.errors)
	}()

	// parse expr
	p.init(filename, text, mode)
	// Set up pkg-level scopes to avoid nil-pointer errors.
	// This is not needed for a correct expression x as the
	// parser will be ok with a nil topScope, but be cautious
	// in case of an erroneous x.
	e := p.parseRHS()

	// If a comma was inserted, consume it;
	// report an error if there's more tokens.
	if p.tok == token.COMMA && p.lit == "\n" {
		p.next()
	}

	p.expect(token.EOF)
	return e, p.errors
}
