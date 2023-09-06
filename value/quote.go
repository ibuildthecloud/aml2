package value

import (
	"strconv"
	"strings"
)

func Escape(s string) string {
	s = strconv.Quote(s)
	return s[1 : len(s)-1]
}

func Unquote(s string) (string, error) {
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
		return strconv.Unquote(s)
	}
	return s, nil
}
