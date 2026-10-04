package analyzer

import (
	"os"
	"path/filepath"
	"testing"
)

func TestScanner(t *testing.T) {
	tempDir := t.TempDir()

	goCode := `package main

func simpleFunc() {
	println("hello")
}

func nestedTwo() {
	for i := 0; i < 10; i++ {
		for j := 0; j < 10; j++ {
			println(i, j)
		}
	}
}

func nestedThree() {
	for i := 0; i < 10; i++ {
		for j := 0; j < 10; j++ {
			for k := 0; k < 10; k++ {
				println(i, j, k)
			}
		}
	}
}

func recurse(n int) int {
	if n <= 1 {
		return 1
	}
	return recurse(n - 1)
}
`

	pyCode := `def py_simple():
    print("linear")

def py_nested():
    for x in range(10):
        for y in range(10):
            print(x, y)

def py_recursive(n):
    if n <= 1:
        return 1
    return py_recursive(n - 1)
`

	luaCode := `local M = {}

function M.setup()
    print("setting up plugin")
end

local function process_items(tbl)
    for i = 1, #tbl do
        for j = 1, #tbl[i] do
            print(tbl[i][j])
        end
    end
end

function M.recurse_nodes(node)
    if not node then return end
    M.recurse_nodes(node.next)
end

return M
`

	if err := os.WriteFile(filepath.Join(tempDir, "main.go"), []byte(goCode), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(tempDir, "script.py"), []byte(pyCode), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(tempDir, "plugin.lua"), []byte(luaCode), 0644); err != nil {
		t.Fatal(err)
	}

	metrics, err := ScanDirectory(tempDir)
	if err != nil {
		t.Fatalf("ScanDirectory failed: %v", err)
	}

	findMetric := func(name string) *FunctionMetric {
		for _, m := range metrics {
			if m.Name == name {
				return &m
			}
		}
		return nil
	}

	simple := findMetric("simpleFunc")
	if simple == nil || simple.Rating != Efficient {
		t.Errorf("expected simpleFunc to be Efficient, got %v", simple)
	}

	n2 := findMetric("nestedTwo")
	if n2 == nil || n2.Rating != Moderate || n2.MaxDepth != 2 {
		t.Errorf("expected nestedTwo to be Moderate (depth 2), got %v", n2)
	}

	n3 := findMetric("nestedThree")
	if n3 == nil || n3.Rating != Inefficient || n3.MaxDepth != 3 {
		t.Errorf("expected nestedThree to be Inefficient (depth 3), got %v", n3)
	}

	rec := findMetric("recurse")
	if rec == nil || rec.Rating != Inefficient || !rec.IsRecursive {
		t.Errorf("expected recurse to be Inefficient (recursive), got %v", rec)
	}

	pySimp := findMetric("py_simple")
	if pySimp == nil || pySimp.Rating != Efficient {
		t.Errorf("expected py_simple to be Efficient, got %v", pySimp)
	}

	pyNest := findMetric("py_nested")
	if pyNest == nil || pyNest.Rating != Moderate || pyNest.MaxDepth != 2 {
		t.Errorf("expected py_nested to be Moderate (depth 2), got %v", pyNest)
	}

	pyRec := findMetric("py_recursive")
	if pyRec == nil || pyRec.Rating != Inefficient || !pyRec.IsRecursive {
		t.Errorf("expected py_recursive to be Inefficient (recursive), got %v", pyRec)
	}

	luaSetup := findMetric("M.setup")
	if luaSetup == nil || luaSetup.Rating != Efficient {
		t.Errorf("expected M.setup to be Efficient, got %v", luaSetup)
	}

	luaNest := findMetric("process_items")
	if luaNest == nil || luaNest.Rating != Moderate || luaNest.MaxDepth != 2 {
		t.Errorf("expected process_items to be Moderate (depth 2), got %v", luaNest)
	}

	luaRec := findMetric("M.recurse_nodes")
	if luaRec == nil || luaRec.Rating != Inefficient || !luaRec.IsRecursive {
		t.Errorf("expected M.recurse_nodes to be Inefficient (recursive), got %v", luaRec)
	}
}
