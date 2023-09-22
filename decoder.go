package aml

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"

	"github.com/acorn-io/aml/pkg/ast"
	"github.com/acorn-io/aml/pkg/eval"
	"github.com/acorn-io/aml/pkg/parser"
	"github.com/acorn-io/aml/pkg/schema"
	"github.com/acorn-io/aml/pkg/value"
)

type Option struct {
	PositionalArgs []any
	Args           map[string]any
	Profiles       []string
	SourceName     string
}

func (o Option) Complete() Option {
	if o.SourceName == "" {
		o.SourceName = "<inline>"
	}
	return o
}

type Options []Option

func (o Options) Merge() (result Option) {
	for _, opt := range o {
		result.PositionalArgs = append(result.PositionalArgs, opt.PositionalArgs...)
		result.Profiles = append(result.Profiles, opt.Profiles...)
		if opt.SourceName != "" {
			result.SourceName = opt.SourceName
		}
		if len(opt.Args) > 0 && result.Args == nil {
			result.Args = map[string]any{}
		}
		for k, v := range opt.Args {
			result.Args[k] = v
		}
	}
	return
}

type Decoder struct {
	opts  Option
	input io.Reader
}

func NewDecoder(input io.Reader, opts ...Option) *Decoder {
	return &Decoder{
		opts:  Options(opts).Merge().Complete(),
		input: input,
	}
}

func (d *Decoder) Decode(out any) error {
	parsed, err := parser.ParseFile(d.opts.SourceName, d.input)
	if err != nil {
		return err
	}

	switch n := out.(type) {
	case *ast.File:
		*n = *parsed
		return nil
	}

	file, err := eval.Build(parsed, eval.BuildOption{
		PositionalArgs: d.opts.PositionalArgs,
		Args:           d.opts.Args,
		Profiles:       d.opts.Profiles,
	})

	switch n := out.(type) {
	case *eval.File:
		*n = *file
		return nil
	}

	switch n := out.(type) {
	case *schema.File:
		fileSchema, err := file.ToSchema()
		if err != nil {
			return err
		}
		*n = *fileSchema
		return nil
	}

	val, ok, err := file.ToValue(eval.Builtin)
	if err != nil {
		return err
	} else if !ok {
		return fmt.Errorf("source <%s> did not produce a value", d.opts.SourceName)
	}

	switch n := out.(type) {
	case *value.Value:
		*n = val
		return nil
	}

	nv, ok, err := value.NativeValue(val)
	if err != nil {
		return err
	} else if !ok {
		return fmt.Errorf("value kind %s from source %s did not produce a native value", val.Kind(), d.opts.SourceName)
	}

	buf := &bytes.Buffer{}
	if err := json.NewEncoder(buf).Encode(nv); err != nil {
		return err
	}

	return json.NewDecoder(buf).Decode(out)
}

func Unmarshal(data []byte, v any, opts ...Option) error {
	return NewDecoder(bytes.NewReader(data), opts...).Decode(v)
}
