# termwerm

[![Go Version](https://img.shields.io/github/go-mod/go-version/adnair2024/termwerm)](https://github.com/adnair2024/termwerm)
[![Go Report Card](https://goreportcard.com/badge/github.com/adnair2024/termwerm)](https://goreportcard.com/report/github.com/adnair2024/termwerm)
[![License](https://img.shields.io/github/license/adnair2024/termwerm)](LICENSE)

Zero-bloat, zero-CGO terminal UI tool in Go to audit function algorithmic efficiency across codebases. Fast, keyboard-driven, and plain-English.

## Features
- **Multi-Language Support:** Scans Go, Python, Lua, JavaScript, TypeScript, Rust, C, C++, Java, C# files.
- **Scope & Recursion Engine:** Accurately detects nested loop depths and recursive calls without external compilers.
- **3-Tier Rating:**
  - `[▲ Efficient]` — $O(1)$ or $O(n)$ linear / constant execution.
  - `[● Moderate]` — $O(n^2)$ nested loops.
  - `[▼ Not Efficient]` — $O(n^{3+})$ or $O(2^n)$ deep loops or recursion.
- **Side-by-Side TUI:** Left table + right details pane with formatted code snippet.
- **One-Key Markdown Export:** Press `m` or run `--export` to generate `termwerm-audit.md`.

## Installation & Running

```bash
# One-liner install
go install github.com/adnair2024/termwerm@latest

# Run directly with Go
go run . [path]

# Or build the binary
go build .
./termwerm [path]

# Headless Markdown export
./termwerm --export [path]
```

## Keybindings
- `j` / `k` (or `↑` / `↓`): Navigate functions
- `1` / `2` / `3`: Filter by rating tier (`0` to reset filter)
- `tab`: Toggle focus between function list and snippet viewport for scrolling
- `m`: Export audit report to `termwerm-audit.md`
- `q`: Quit
