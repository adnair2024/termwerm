package analyzer

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type Rating int
const (
	Efficient Rating = iota + 1
	Moderate
	Inefficient
)

type FunctionMetric struct {
	Rating                  Rating
	RatingTag, Complexity   string
	Language, Name, File    string
	Line, EndLine, MaxDepth int
	IsRecursive             bool
	Reason, Tip, Snippet    string
}

var langExts = map[string]string{
	".go": "Go", ".py": "Python", ".js": "JavaScript", ".jsx": "JavaScript",
	".ts": "TypeScript", ".tsx": "TypeScript", ".rs": "Rust", ".c": "C", ".cpp": "C++",
	".cc": "C++", ".h": "C/C++", ".hpp": "C++", ".java": "Java", ".cs": "C#", ".lua": "Lua",
}

func ScanDirectory(root string) ([]FunctionMetric, error) {
	var metrics []FunctionMetric
	err := filepath.Walk(root, func(p string, info os.FileInfo, err error) error {
		if err != nil || info == nil { return nil }
		if info.IsDir() {
			if n := info.Name(); (strings.HasPrefix(n, ".") && n != ".") || n == "node_modules" || n == "vendor" || n == "dist" || n == "build" || n == "target" { return filepath.SkipDir }
			return nil
		}
		ext := strings.ToLower(filepath.Ext(p))
		lang, ok := langExts[ext]
		if !ok { return nil }
		rel, err := filepath.Rel(root, p)
		if err != nil { rel = p }
		m, _ := scanFile(p, rel, lang, ext)
		metrics = append(metrics, m...)
		return nil
	})
	return metrics, err
}

func scanFile(path, displayPath, lang, ext string) ([]FunctionMetric, error) {
	b, err := os.ReadFile(path)
	if err != nil { return nil, err }
	lines := strings.Split(string(b), "\n")
	switch ext {
	case ".py": return scanPython(lines, displayPath, lang), nil
	case ".lua": return scanLua(lines, displayPath), nil
	default: return scanBraces(lines, displayPath, lang), nil
	}
}

func extractFuncName(line, kw string) string {
	idx := strings.Index(line, kw)
	if idx == -1 { return "" }
	rest := strings.TrimSpace(line[idx+len(kw):])
	if kw == "func " && strings.HasPrefix(rest, "(") {
		if cIdx := strings.Index(rest, ")"); cIdx != -1 { rest = strings.TrimSpace(rest[cIdx+1:]) }
	}
	if pIdx := strings.Index(rest, "("); pIdx != -1 {
		if f := strings.Fields(strings.TrimSpace(rest[:pIdx])); len(f) > 0 { return strings.Trim(f[len(f)-1], "*&") }
	}
	if eq := strings.LastIndex(line[:idx], "="); eq != -1 {
		if f := strings.Fields(strings.TrimSpace(line[:eq])); len(f) > 0 { return strings.Trim(f[len(f)-1], "*&") }
	}
	return ""
}

func isLoop(s string) bool { return strings.HasPrefix(s, "for ") || strings.HasPrefix(s, "for(") || strings.HasPrefix(s, "while ") || strings.HasPrefix(s, "while(") }
func isSelfCall(line, name string) bool {
	idx := strings.Index(line, name+"(")
	if idx < 0 { return false }
	if idx > 0 {
		p := line[idx-1]
		if p == '.' || p == ':' || (p >= 'a' && p <= 'z') || (p >= 'A' && p <= 'Z') || (p >= '0' && p <= '9') || p == '_' { return false }
	}
	return true
}

func scanBraces(lines []string, file, lang string) []FunctionMetric {
	var metrics []FunctionMetric
	headers := []string{"func ", "function ", "fn ", "void ", "int ", "def "}
	for i, n := 0, len(lines); i < n; i++ {
		trimmed, name := strings.TrimSpace(lines[i]), ""
		for _, kw := range headers {
			if (strings.HasPrefix(trimmed, kw) || strings.Contains(lines[i], " "+kw)) && func() bool { name = extractFuncName(lines[i], kw); return name != "" }() { break }
		}
		if name == "" { continue }
		start, depth, maxDepth, loopDepths, pending, recur, body := i+1, 0, 0, []int{}, false, false, false
		for j := i; j < n; j++ {
			ct := strings.TrimSpace(lines[j])
			if j > i && !recur && isSelfCall(ct, name) { recur = true }
			if isLoop(ct) { pending = true }
			for _, ch := range lines[j] {
				if ch == '{' {
					depth++
					body = true
					if pending {
						loopDepths, pending = append(loopDepths, depth), false
						if len(loopDepths) > maxDepth { maxDepth = len(loopDepths) }
					}
				} else if ch == '}' {
					if len(loopDepths) > 0 && loopDepths[len(loopDepths)-1] >= depth { loopDepths = loopDepths[:len(loopDepths)-1] }
					depth--
				}
			}
			if body && depth <= 0 {
				metrics = append(metrics, evaluate(name, file, lang, start, j+1, maxDepth, recur, lines[i:j+1]))
				i = j
				break
			}
		}
	}
	return metrics
}

func scanPython(lines []string, file, lang string) []FunctionMetric {
	var metrics []FunctionMetric
	for i, n := 0, len(lines); i < n; i++ {
		trimmed := strings.TrimSpace(lines[i])
		if !strings.HasPrefix(trimmed, "def ") { continue }
		name := extractFuncName(trimmed, "def ")
		if name == "" { continue }
		fIndent, start, maxDepth, loopIndents, recur, end := len(lines[i])-len(strings.TrimLeft(lines[i], " ")), i+1, 0, []int{}, false, i+1
		for j := i + 1; j < n; j++ {
			ct := strings.TrimSpace(lines[j])
			if ct == "" || strings.HasPrefix(ct, "#") { continue }
			indent := len(lines[j]) - len(strings.TrimLeft(lines[j], " "))
			if indent <= fIndent { end = j; break }
			end = j + 1
			if !recur && isSelfCall(ct, name) { recur = true }
			for len(loopIndents) > 0 && indent <= loopIndents[len(loopIndents)-1] { loopIndents = loopIndents[:len(loopIndents)-1] }
			if isLoop(ct) {
				loopIndents = append(loopIndents, indent)
				if len(loopIndents) > maxDepth { maxDepth = len(loopIndents) }
			}
		}
		metrics = append(metrics, evaluate(name, file, lang, start, end, maxDepth, recur, lines[i:end]))
		i = end - 1
	}
	return metrics
}

func scanLua(lines []string, file string) []FunctionMetric {
	var metrics []FunctionMetric
	for i, n := 0, len(lines); i < n; i++ {
		trimmed := strings.TrimSpace(lines[i])
		if !strings.Contains(trimmed, "function") || strings.HasPrefix(trimmed, "--") { continue }
		name := extractFuncName(trimmed, "function")
		if name == "" { continue }
		start, depth, maxDepth, loopDepths, recur := i+1, 1, 0, []int{}, false
		for j := i + 1; j < n; j++ {
			ct := strings.TrimSpace(lines[j])
			if strings.HasPrefix(ct, "--") { continue }
			if !recur && isSelfCall(ct, name) { recur = true }
			for _, w := range strings.Fields(ct) {
				w = strings.Trim(w, ",;()")
				if w == "if" || w == "for" || w == "while" || w == "function" {
					depth++
					if w == "for" || w == "while" {
						loopDepths = append(loopDepths, depth)
						if len(loopDepths) > maxDepth { maxDepth = len(loopDepths) }
					}
				} else if w == "end" {
					if len(loopDepths) > 0 && loopDepths[len(loopDepths)-1] >= depth { loopDepths = loopDepths[:len(loopDepths)-1] }
					depth--
				}
			}
			if depth <= 0 {
				metrics = append(metrics, evaluate(name, file, "Lua", start, j+1, maxDepth, recur, lines[i:j+1]))
				i = j
				break
			}
		}
	}
	return metrics
}

func evaluate(name, file, lang string, start, end, depth int, recur bool, lines []string) FunctionMetric {
	m := FunctionMetric{Language: lang, Name: name, File: file, Line: start, EndLine: end, MaxDepth: depth, IsRecursive: recur, Snippet: strings.Join(lines, "\n"), Rating: Efficient, RatingTag: "[▲ Efficient]", Complexity: "O(1) / O(n)", Reason: "Fast linear or constant execution.", Tip: "No nested loops found. Great efficiency!"}
	if recur {
		m.Rating, m.RatingTag, m.Complexity, m.Reason, m.Tip = Inefficient, "[▼ Not Efficient]", "O(2ⁿ)", fmt.Sprintf("Recursive call detected in %s().", name), "Refactor recursion to iteration or memoization."
	} else if depth >= 3 {
		m.Rating, m.RatingTag, m.Complexity, m.Reason, m.Tip = Inefficient, "[▼ Not Efficient]", "O(n³+)", fmt.Sprintf("Found %d nested loops.", depth), "Flatten nested loops or pre-index data."
	} else if depth == 2 {
		m.Rating, m.RatingTag, m.Complexity, m.Reason, m.Tip = Moderate, "[● Moderate]", "O(n²)", "Found 2 nested loops.", "Consider using a hash map or set lookup to replace the inner loop."
	}
	return m
}
