package http

// splitSQL splits a SQL migration file into individual statements separated by ';'.
func splitSQL(data string) []string {
	var stmts []string
	var cur []byte
	for i := 0; i < len(data); i++ {
		if data[i] == ';' {
			s := trimBytes(cur)
			if len(s) > 0 {
				stmts = append(stmts, string(s))
			}
			cur = nil
			continue
		}
		cur = append(cur, data[i])
	}
	if s := trimBytes(cur); len(s) > 0 {
		stmts = append(stmts, string(s))
	}
	return stmts
}

func trimBytes(b []byte) []byte {
	start := 0
	for start < len(b) && (b[start] == ' ' || b[start] == '\t' || b[start] == '\n' || b[start] == '\r') {
		start++
	}
	end := len(b)
	for end > start && (b[end-1] == ' ' || b[end-1] == '\t' || b[end-1] == '\n' || b[end-1] == '\r') {
		end--
	}
	return b[start:end]
}
