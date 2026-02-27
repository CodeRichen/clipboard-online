package utils

import (
	"strconv"
	"strings"
)

// groupState tracks per-group RTF parser state
type groupState struct {
	suppress bool // suppress text output in this group
	ucSkip   int  // substitute byte count after \uN
}

// ExtractTextFromRTF parses an RTF document and returns the plain text content.
// Uses a stack-based approach to properly handle nested groups.
func ExtractTextFromRTF(data []byte) string {
	content := string(data)
	var sb strings.Builder
	var stack []groupState
	cur := groupState{suppress: false, ucSkip: 0}

	i := 0
	for i < len(content) {
		ch := content[i]

		switch ch {
		case '{':
			stack = append(stack, cur)
			i++
			continue
		case '}':
			if len(stack) > 0 {
				cur = stack[len(stack)-1]
				stack = stack[:len(stack)-1]
			}
			i++
			continue
		}

		if ch == '\r' || ch == '\n' {
			i++
			continue
		}

		if ch != '\\' {
			if !cur.suppress {
				sb.WriteByte(ch)
			}
			i++
			continue
		}

		// Backslash — control word or escape
		i++ // skip '\'
		if i >= len(content) {
			break
		}
		next := content[i]

		// Literal escaped chars
		if next == '\\' || next == '{' || next == '}' {
			if !cur.suppress {
				sb.WriteByte(next)
			}
			i++
			continue
		}

		// \'XX — ANSI hex char, skip
		if next == '\'' {
			i += 3
			continue
		}

		// \* — ignorable destination: suppress this group
		if next == '*' {
			cur.suppress = true
			i++
			continue
		}

		// line break after backslash
		if next == '\n' || next == '\r' {
			if !cur.suppress {
				sb.WriteByte('\n')
			}
			i++
			continue
		}

		// Read control word name
		j := i
		for j < len(content) && rtfIsAlpha(content[j]) {
			j++
		}
		word := content[i:j]
		i = j

		// Read optional signed number
		numStart := i
		if i < len(content) && content[i] == '-' {
			i++
		}
		for i < len(content) && rtfIsDigit(content[i]) {
			i++
		}
		numStr := content[numStart:i]

		// Consume ONE trailing space (RTF delimiter)
		if i < len(content) && content[i] == ' ' {
			i++
		}

		switch word {
		// Destination groups — suppress text inside
		case "fonttbl", "colortbl", "stylesheet", "info",
			"pict", "object", "fldinst",
			"header", "footer", "headerl", "headerr", "headerf",
			"footerl", "footerr", "footerf":
			cur.suppress = true

		case "uc":
			if n, err := strconv.Atoi(numStr); err == nil {
				cur.ucSkip = n
			}

		case "u":
			n, err := strconv.Atoi(numStr)
			if err == nil {
				if n < 0 {
					n += 65536
				}
				if !cur.suppress {
					sb.WriteRune(rune(n))
				}
			}
			// Skip cur.ucSkip substitute chars
			for skip := 0; skip < cur.ucSkip && i < len(content); skip++ {
				if content[i] == '\\' {
					i++
					if i < len(content) {
						if content[i] == '\'' {
							i += 3
						} else {
							for i < len(content) && rtfIsAlpha(content[i]) {
								i++
							}
							for i < len(content) && rtfIsDigit(content[i]) {
								i++
							}
							if i < len(content) && content[i] == ' ' {
								i++
							}
						}
					}
				} else {
					i++
				}
			}

		case "par", "line":
			if !cur.suppress {
				sb.WriteByte('\n')
			}
		case "tab":
			if !cur.suppress {
				sb.WriteByte('\t')
			}
		// All other control words — silently ignore
		}
	}

	result := strings.TrimSpace(sb.String())
	for strings.Contains(result, "\n\n\n") {
		result = strings.ReplaceAll(result, "\n\n\n", "\n\n")
	}
	return result
}

func rtfIsAlpha(c byte) bool {
	return (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z')
}

func rtfIsDigit(c byte) bool {
	return c >= '0' && c <= '9'
}
