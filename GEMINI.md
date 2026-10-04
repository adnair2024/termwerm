# termwerm - Minimalist Code Efficiency TUI

## Objective
A lightweight, zero-CGO terminal tool in Go that scans any codebase, detects nested loops and recursion to estimate algorithmic efficiency, and presents findings in a dead-simple, 3-tier rating interface ([▲ Efficient], [● Moderate], [▼ Not Efficient]) with one-key Markdown export.

---

## 1. Design Principles
- Zero CGO / Zero External Compilers: Must run with a bare `go run .` anywhere.
- Readability First: Keep the analyzer under 200 lines of pure Go.
- Zero Jargon: Anyone should understand the findings in 5 seconds.

---

## 2. Architecture & Engine

### Analyzer (internal/analyzer)
Use a clean token-and-scope scanner that works across languages (.go, .py, .js, .ts, .rs, .c, .cpp, .java):
1. Detect Functions: Identify function headers (`func `, `def `, `function `, `fn `, `void `, `int `).
2. Track Scope & Loop Depth:
   - For brace languages (`{ ... }`), track `{` and `}` nesting while counting `for` and `while` keywords.
   - For indentation languages (Python), track indentation levels while counting `for` and `while` lines.
3. Recursion: Check if the function name is called inside its own body.

### 3-Tier Rating Rules
- 0 to 1 loop depth -> [▲ Efficient] | O(1) or O(n) | "Fast linear or constant execution."
- 2 nested loops -> [● Moderate] | O(n²) | "2 levels of loops. Consider a map or set lookup."
- 3+ nested loops OR recursion -> [▼ Not Efficient] | O(n³+) / O(2ⁿ) | "Deeply nested or recursive. Needs refactoring."

---

## 3. Minimalist TUI & Usage

### Layout
- Left (40% width): Clean table showing `Rating | Lang | Function | File:Line`.
- Right (60% width):
  - Big Rating Tag: `[▲ Efficient]`, `[● Moderate]`, or `[▼ Not Efficient]`
  - Plain-English Reason: e.g. "Found a loop inside another loop at line 48."
  - Simple Tip: e.g. "Try pre-grouping items into a map instead of searching each time."
  - Code Snippet: Read-only viewport with line numbers.

### Keybindings (Only 4 keys needed)
- `j` / `k` (or arrows): Select function
- `1` / `2` / `3`: Filter by tier (`0` to reset)
- `m`: Save report to `termwerm-audit.md`
- `q`: Quit

---

## 4. Markdown Export (m)
Pressing `m` creates `termwerm-audit.md` in the current folder:
- Header with scan date and totals for ▲, ●, and ▼.
- Grouped list of all [▼ Not Efficient] and [● Moderate] functions with their file location, explanation, tip, and code block.

---

## 5. Agent Implementation Steps

### Step 1: Init Project
go mod init termwerm
go get github.com/charmbracelet/bubbletea
go get github.com/charmbracelet/lipgloss
go get github.com/charmbracelet/bubbles

### Step 2: Analyzer (internal/analyzer/scanner.go)
- Recursively walk target dir (skip .git, node_modules, vendor).
- Scan lines to extract functions, nested loop depth, and self-calls.
- Output a slice of `FunctionMetric` objects.

### Step 3: TUI (internal/ui/app.go)
- Initialize Bubbletea table (left) and viewport (right).
- Color codes:
  - Green for [▲ Efficient]
  - Amber for [● Moderate]
  - Red for [▼ Not Efficient]
- Handle `m` keypress to write `termwerm-audit.md` and show "Saved to termwerm-audit.md" in footer.

### Step 4: CLI Entrypoint (main.go)
- Accept optional path argument: `termwerm [path]` (defaults to `.`).
- If `--export` is passed, write markdown and exit without opening TUI.
