package value

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

func Escape(s string) string {
	s = strconv.Quote(s)
	return s[1 : len(s)-1]
}

func unquoteRaw(s string) (string, error) {
	if strings.HasPrefix(s, "```") {
		if !strings.HasSuffix(s, "```") {
			return "", fmt.Errorf("raw string does not end with ```")
		}
		s = s[3 : len(s)-3]
	} else if strings.HasSuffix(s, "`") {
		s = s[1 : len(s)-1]
	} else {
		return "", fmt.Errorf("raw string does not end with `")
	}

	buf := strings.Builder{}
	buf.Grow(len(s))

	for i := 0; i < len(s); i++ {
		if i < len(s)-1 && s[i] == '\\' && s[i+1] == '`' {
			buf.WriteRune('`')
			i++
		} else {
			buf.WriteByte(s[i])
		}
	}
	return buf.String(), nil
}

func Unquote(s string) (string, error) {
	if strings.HasPrefix(s, "`") {
		return unquoteRaw(s)
	}
	if strings.HasPrefix(s, "\"\"\"") {
		s = strings.TrimPrefix(s, "\"\"")
		s = strings.TrimSuffix(s, "\"\"")
		lines := strings.Split(s, "\n")
		if lines[0] == "\"" && strings.TrimSpace(lines[len(lines)-1]) == "\"" {
			prefix := strings.TrimSuffix(lines[len(lines)-1], "\"")
			foundPrefix := true
			for _, line := range lines[1:] {
				if !strings.HasPrefix(line, prefix) {
					foundPrefix = false
					break
				}
			}
			if foundPrefix {
				lines = lines[1:]
				for i := range lines {
					lines[i] = strings.TrimPrefix(lines[i], prefix)
				}
				lines[0] = "\"" + lines[0]
			}
		}
		s = strings.Join(lines, "\\n")
	}
	if strings.HasPrefix(s, "\"") {
		ret, err := strconv.Unquote(s)
		if errors.Is(err, strconv.ErrSyntax) {
			err = fmt.Errorf("%w: invalid or missing escape (\\) sequence %s", err, s)
		} else if err != nil {
			err = fmt.Errorf("%w: %s", err, s)
		}
		return ret, err
	}
	return s, nil
}
