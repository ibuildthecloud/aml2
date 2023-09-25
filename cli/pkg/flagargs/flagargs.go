package flagargs

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/acorn-io/aml"
	"github.com/acorn-io/aml/pkg/schema"
	"github.com/acorn-io/aml/pkg/value"
	"github.com/spf13/pflag"
	"gopkg.in/yaml.v3"
)

type Flags struct {
	FlagSet    *pflag.FlagSet
	fieldFlags map[string]fieldFlag
	profile    *[]string
	Usage      func()
	argsFile   string
}

type fieldFlag struct {
	Field       schema.Field
	String      *string
	StringSlice *[]string
	Bool        *bool
}

func ParseArgs(argsFile, acornFile string, args []string) (map[string]any, []string, error) {
	f, err := os.Open(acornFile)
	if err != nil {
		return nil, nil, err
	}

	var file schema.File
	if err := aml.NewDecoder(f).Decode(&file); err != nil {
		return nil, nil, err
	}

	flags := New(argsFile, filepath.Base(acornFile), file.ProfileNames, file.Args)
	return flags.Parse(args)
}

func New(argsFile, filename string, profiles schema.Names, args schema.Object) *Flags {
	var (
		flagSet    = pflag.NewFlagSet(filename, pflag.ContinueOnError)
		fieldFlags = map[string]fieldFlag{}
		profile    *[]string
	)

	if len(profiles) == 0 {
		var empty []string
		profile = &empty
	} else {
		desc := strings.Builder{}
		desc.WriteString("Available profiles (")
		startLen := desc.Len()
		for _, name := range profiles {
			val := name.Value
			if name.Description != "" {
				val += ": " + name.Description
			}
			if desc.Len() > startLen {
				desc.WriteString(", ")
			}
			desc.WriteString(val)
		}
		desc.WriteString(")")
		profile = flagSet.StringSlice("profile", nil, desc.String())
	}

	for _, field := range args.Fields {
		flag := fieldFlag{
			Field: field,
		}
		if profile != nil && field.Name == "profile" {
			continue
		}
		if field.Type.Kind == schema.BoolKind {
			flag.Bool = flagSet.Bool(field.Name, false, field.Description)
		} else if field.Type.Kind == schema.ArrayKind {
			flag.StringSlice = flagSet.StringSlice(field.Name, nil, field.Description)
		} else {
			flag.String = flagSet.String(field.Name, "", field.Description)
		}
		fieldFlags[field.Name] = flag
	}

	return &Flags{
		fieldFlags: fieldFlags,
		profile:    profile,
		FlagSet:    flagSet,
		argsFile:   argsFile,
	}
}

func isYAMLFilename(v string) bool {
	for _, suffix := range []string{".yaml", ".yml"} {
		if strings.HasSuffix(strings.ToLower(v), suffix) {
			return true
		}
	}
	return false
}

func parseValue(v string, isNumber bool) (any, error) {
	if !strings.HasPrefix(v, "@") {
		if isNumber {
			return value.Number(v), nil
		}
		return v, nil
	}

	v = v[1:]
	data := map[string]any{}
	if strings.HasPrefix(v, "{") {
		if err := aml.Unmarshal([]byte(v), &data); err != nil {
			return nil, err
		}
		return data, nil
	}

	f, err := os.Open(v)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	if isYAMLFilename(v) {
		return data, yaml.NewDecoder(f).Decode(data)
	}

	return data, aml.NewDecoder(f).Decode(data)
}

func (f *Flags) readArgsFile() (map[string]any, error) {
	result := map[string]any{}

	if f.argsFile == "" {
		return result, nil
	}

	input, err := os.Open(f.argsFile)
	if os.IsNotExist(err) {
		return result, nil
	}

	if err := aml.NewDecoder(input).Decode(result); err != nil {
		return nil, err
	}

	return result, nil
}

func (f *Flags) Parse(args []string) (map[string]any, []string, error) {
	result, err := f.readArgsFile()
	if err != nil {
		return nil, nil, err
	}

	if f.Usage != nil {
		f.FlagSet.Usage = func() {
			f.Usage()
			f.FlagSet.PrintDefaults()
		}
	}

	if err := f.FlagSet.Parse(args); err != nil {
		return nil, nil, err
	}

	for name, field := range f.fieldFlags {
		flag := f.FlagSet.Lookup(name)

		switch {
		case !flag.Changed:
		case field.Bool != nil:
			result[name] = *field.Bool
		case field.StringSlice != nil:
			vals := []any{}
			for _, str := range *field.StringSlice {
				val, err := parseValue(str, field.Field.Type.Array.Items.Kind == schema.NumberKind)
				if err != nil {
					return nil, nil, err
				}
				vals = append(vals, val)
			}
			result[name] = vals
		default:
			result[name], err = parseValue(*field.String, field.Field.Type.Kind == schema.NumberKind)
			if err != nil {
				return nil, nil, err
			}
		}
	}

	return result, *f.profile, nil
}

func (f *Flags) flagChanged(name string) bool {
	return f.FlagSet.Lookup(name).Changed
}