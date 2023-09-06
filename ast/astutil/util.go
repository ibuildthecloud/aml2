// Copyright 2019 CUE Authors
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

package astutil

import (
	"path"
	"strings"

	"github.com/acorn-io/aml/ast"
)

// ImportPathName derives the package name from the given import path.
//
// Examples:
//
//	string           string
//	foo.com/bar      bar
//	foo.com/bar:baz  baz
func ImportPathName(id string) string {
	name := path.Base(id)
	if p := strings.LastIndexByte(name, ':'); p > 0 {
		name = name[p+1:]
	}
	return name
}

// ImportInfo describes the information contained in an ImportSpec.
type ImportInfo struct {
	Ident   string // identifier used to refer to the import
	PkgName string // name of the package
	ID      string // full import path, including the name
	Dir     string // import path, excluding the name
}

// CopyComments associates comments of one node with another.
// It may change the relative position of comments.
func CopyComments(to, from ast.Node) {
	if from == nil {
		return
	}
	ast.SetComments(to, ast.Comments(from))
}

// CopyPosition sets the position of one node to another.
func CopyPosition(to, from ast.Node) {
	if from == nil {
		return
	}
	ast.SetPos(to, from.Pos())
}

// CopyMeta copies comments and position information from one node to another.
// It returns the destination node.
func CopyMeta(to, from ast.Node) ast.Node {
	if from == nil {
		return to
	}
	ast.SetComments(to, ast.Comments(from))
	ast.SetPos(to, from.Pos())
	return to
}
