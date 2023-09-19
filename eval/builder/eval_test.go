package builder

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/acorn-io/aml/eval"
	"github.com/acorn-io/aml/parser"
	"github.com/acorn-io/aml/value"
	"github.com/hexops/autogold/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSuccessfulEval(t *testing.T) {
	dir := fmt.Sprintf("testdata/%s", t.Name())
	files, err := os.ReadDir(dir)
	require.Nil(t, err)

	for _, file := range files {
		if !strings.HasSuffix(file.Name(), ".acorn") {
			continue
		}
		t.Run(strings.TrimSuffix(file.Name(), ".acorn"), func(t *testing.T) {
			data, err := os.ReadFile(filepath.Join(dir, file.Name()))
			require.NoError(t, err)

			ast, err := parser.ParseFile(file.Name(), bytes.NewReader(data))
			require.NoError(t, err)

			result, err := Build(ast)
			require.NoError(t, err)

			v, ok, err := result.ToValue(eval.Builtin)
			if err == nil {
				assert.True(t, ok)

				var data []byte
				nv, ok, err := value.NativeValue(v)
				if err == nil {
					require.True(t, ok)
					data, err = json.MarshalIndent(nv, "", "  ")
				}
				require.NoError(t, err)
				autogold.ExpectFile(t, string(data))
			} else {
				autogold.ExpectFile(t, err)
			}
		})
	}
}