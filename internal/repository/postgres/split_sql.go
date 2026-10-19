package postgres

import "strings"

// splitSQL splits a SQL migration file into individual statements separated by ';'.
// Empty statements and SQL comments (-- ... and /* ... */) are ignored.
func splitSQL(data string) []string {
	var stmts []string
	var cur strings.Builder
	inLine := false
	inBlock := false
	i := 0
	for i < len(data) {
		c := data[i]
		switch {
		case inLine:
			if c == '\n' {
				inLine = false
				cur.WriteByte(c)
			}
			i++
		case inBlock:
			if c == '*' && i+1 < len(data) && data[i+1] == '/' {
				inBlock = false
				i += 2
			} else {
				i++
			}
		case c == '-' && i+1 < len(data) && data[i+1] == '-':
			inLine = true
			i += 2
		case c == '/' && i+1 < len(data) && data[i+1] == '*':
			inBlock = true
			i += 2
		case c == ';':
			s := strings.TrimSpace(cur.String())
			if s != "" {
				stmts = append(stmts, s)
			}
			cur.Reset()
			i++
		default:
			cur.WriteByte(c)
			i++
		}
	}
	if s := strings.TrimSpace(cur.String()); s != "" {
		stmts = append(stmts, s)
	}
	return stmts
}
