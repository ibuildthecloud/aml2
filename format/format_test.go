package format

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/hexops/autogold/v2"
	"github.com/stretchr/testify/require"
)

func TestFormat(t *testing.T) {
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

			out, err := Format(bytes.NewReader(data))
			if err != nil {
				autogold.ExpectFile(t, err)
			} else {
				autogold.ExpectFile(t, autogold.Raw(out))
			}
		})
	}
}
