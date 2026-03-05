package walker_test

import (
	"os"
	"path/filepath"
	"testing"

	"code-clip/pkg/walker"
)

func TestWalkRespectsGitignore(t *testing.T) {
	// Setup a temporary directory structure
	tempDir, err := os.MkdirTemp("", "codeclip-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create files
	/*
		/tempDir
			.gitignore
			main.go
			secret/
				password.txt
	*/

	if err := os.WriteFile(filepath.Join(tempDir, ".gitignore"), []byte("secret/\n"), 0644); err != nil {
		t.Fatalf("Failed to write .gitignore: %v", err)
	}
	if err := os.WriteFile(filepath.Join(tempDir, ".ignore"), []byte("dist/\n"), 0644); err != nil {
		t.Fatalf("Failed to write .ignore: %v", err)
	}
	if err := os.WriteFile(filepath.Join(tempDir, "main.go"), []byte("package main\n"), 0644); err != nil {
		t.Fatalf("Failed to write main.go: %v", err)
	}

	secretDir := filepath.Join(tempDir, "secret")
	if err := os.MkdirAll(secretDir, 0755); err != nil {
		t.Fatalf("Failed to create secret dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(secretDir, "password.txt"), []byte("hunter2"), 0644); err != nil {
		t.Fatalf("Failed to write secret file: %v", err)
	}

	distDir := filepath.Join(tempDir, "dist")
	if err := os.MkdirAll(distDir, 0755); err != nil {
		t.Fatalf("Failed to create dist dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(distDir, "bundle.js"), []byte("console.log('test')"), 0644); err != nil {
		t.Fatalf("Failed to write dist file: %v", err)
	}

	opts := walker.Options{
		Paths:    []string{tempDir},
		MaxDepth: 0,
		Format:   "plain",
	}

	outChan, err := walker.Walk(opts)
	if err != nil {
		t.Fatalf("Walk failed to start: %v", err)
	}

	var results []walker.Result
	for res := range outChan {
		if res.Err != nil {
			t.Fatalf("Error during walk: %v", res.Err)
		}
		results = append(results, res)
	}

	// Should only find .gitignore, .ignore, and main.go. secret/password.txt and dist/bundle.js MUST be ignored.
	foundMain := false
	foundSecret := false
	foundDist := false

	for _, res := range results {
		if res.RelativePath == "main.go" || filepath.Base(res.RelativePath) == "main.go" {
			foundMain = true
		}
		if res.RelativePath == "secret/password.txt" || filepath.Base(res.RelativePath) == "password.txt" {
			foundSecret = true
		}
		if res.RelativePath == "dist/bundle.js" || filepath.Base(res.RelativePath) == "bundle.js" {
			foundDist = true
		}
	}

	if !foundMain {
		t.Errorf("Expected to find main.go, but didn't.")
	}
	if foundSecret {
		t.Errorf("Expected secret/password.txt to be ignored (via .gitignore), but it was found.")
	}
	if foundDist {
		t.Errorf("Expected dist/bundle.js to be ignored (via .ignore), but it was found.")
	}
}

func TestWalkRespectsAncestorIgnore(t *testing.T) {
	// Setup a temporary directory structure
	rootDir, err := os.MkdirTemp("", "codeclip-test-ancestor")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(rootDir)

	/*
		/rootDir
			.ignore (ignores "node_modules/")
			projectA/
				main.go
				node_modules/
					index.js
	*/

	if err := os.WriteFile(filepath.Join(rootDir, ".ignore"), []byte("node_modules/\n"), 0644); err != nil {
		t.Fatalf("Failed to write ancestor .ignore: %v", err)
	}

	projectDir := filepath.Join(rootDir, "projectA")
	if err := os.MkdirAll(projectDir, 0755); err != nil {
		t.Fatalf("Failed to create project dir: %v", err)
	}

	if err := os.WriteFile(filepath.Join(projectDir, "main.go"), []byte("package main"), 0644); err != nil {
		t.Fatalf("Failed to write main.go: %v", err)
	}

	nodeModulesDir := filepath.Join(projectDir, "node_modules")
	if err := os.MkdirAll(nodeModulesDir, 0755); err != nil {
		t.Fatalf("Failed to create node_modules dir: %v", err)
	}

	if err := os.WriteFile(filepath.Join(nodeModulesDir, "index.js"), []byte("console.log('test')"), 0644); err != nil {
		t.Fatalf("Failed to write index.js: %v", err)
	}

	// We start the walk from projectA, NOT rootDir!
	// It should ascend, find the .ignore in rootDir, and apply it.
	opts := walker.Options{
		Paths:    []string{projectDir},
		MaxDepth: 0,
		Format:   "plain",
	}

	outChan, err := walker.Walk(opts)
	if err != nil {
		t.Fatalf("Walk failed to start: %v", err)
	}

	foundMain := false
	foundNodeModules := false

	for res := range outChan {
		if res.Err != nil {
			continue
		}

		base := filepath.Base(res.RelativePath)
		if base == "main.go" {
			foundMain = true
		}
		if base == "index.js" {
			foundNodeModules = true
		}
	}

	if !foundMain {
		t.Errorf("Expected to find projectA/main.go")
	}
	if foundNodeModules {
		t.Errorf("Expected node_modules/index.js to be ignored by ancestor .ignore, but it was found.")
	}
}

func TestWalkMaxDepth(t *testing.T) {
	// Setup a temporary directory structure
	tempDir, err := os.MkdirTemp("", "codeclip-test-depth")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create files at depth 1, 2, 3
	/*
		/tempDir
			d1.txt (depth 1)
			nested1/
				d2.txt (depth 2)
				nested2/
					d3.txt (depth 3)
	*/

	if err := os.WriteFile(filepath.Join(tempDir, "d1.txt"), []byte("d1"), 0644); err != nil {
		t.Fatalf("Failed to write file: %v", err)
	}

	n1 := filepath.Join(tempDir, "nested1")
	if err := os.MkdirAll(n1, 0755); err != nil {
		t.Fatalf("Failed to create dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(n1, "d2.txt"), []byte("d2"), 0644); err != nil {
		t.Fatalf("Failed to write file: %v", err)
	}

	n2 := filepath.Join(n1, "nested2")
	if err := os.MkdirAll(n2, 0755); err != nil {
		t.Fatalf("Failed to create dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(n2, "d3.txt"), []byte("d3"), 0644); err != nil {
		t.Fatalf("Failed to write file: %v", err)
	}

	opts := walker.Options{
		Paths:    []string{tempDir},
		MaxDepth: 2, // Should see d1, and d2. Should NOT see d3.
		Format:   "plain",
	}

	outChan, err := walker.Walk(opts)
	if err != nil {
		t.Fatalf("Walk failed to start: %v", err)
	}

	foundD1 := false
	foundD2 := false
	foundD3 := false

	for res := range outChan {
		base := filepath.Base(res.RelativePath)
		if base == "d1.txt" {
			foundD1 = true
		}
		if base == "d2.txt" {
			foundD2 = true
		}
		if base == "d3.txt" {
			foundD3 = true
		}
	}

	if !foundD1 {
		t.Errorf("Expected to find d1.txt")
	}
	if !foundD2 {
		t.Errorf("Expected to find d2.txt")
	}
	if foundD3 {
		t.Errorf("Did not expect to find d3.txt (depth 3) with max depth 2")
	}
}

func TestWalkMultipleFiles(t *testing.T) {
	// Setup a temporary directory structure
	tempDir, err := os.MkdirTemp("", "codeclip-test-multiple")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	if err := os.WriteFile(filepath.Join(tempDir, "file1.txt"), []byte("file1"), 0644); err != nil {
		t.Fatalf("Failed to write file1: %v", err)
	}

	if err := os.WriteFile(filepath.Join(tempDir, "file2.txt"), []byte("file2"), 0644); err != nil {
		t.Fatalf("Failed to write file2: %v", err)
	}

	opts := walker.Options{
		Paths:    []string{filepath.Join(tempDir, "file1.txt"), filepath.Join(tempDir, "file2.txt")},
		MaxDepth: 0,
		Format:   "plain",
	}

	outChan, err := walker.Walk(opts)
	if err != nil {
		t.Fatalf("Walk failed to start: %v", err)
	}

	var results []walker.Result
	for res := range outChan {
		if res.Err != nil {
			t.Fatalf("Error during walk: %v", res.Err)
		}
		results = append(results, res)
	}

	if len(results) != 2 {
		t.Fatalf("Expected 2 results, got %d", len(results))
	}

	foundFile1 := false
	foundFile2 := false

	for _, res := range results {
		if res.RelativePath == "file1.txt" || filepath.Base(res.RelativePath) == "file1.txt" {
			foundFile1 = true
		}
		if res.RelativePath == "file2.txt" || filepath.Base(res.RelativePath) == "file2.txt" {
			foundFile2 = true
		}
	}

	if !foundFile1 {
		t.Errorf("Expected to find file1.txt")
	}
	if !foundFile2 {
		t.Errorf("Expected to find file2.txt")
	}
}

func TestWalkMixedPaths(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "codeclip-test-mixed")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	if err := os.WriteFile(filepath.Join(tempDir, "file1.txt"), []byte("file1"), 0644); err != nil {
		t.Fatalf("Failed to write file1: %v", err)
	}

	nestedDir := filepath.Join(tempDir, "nested")
	if err := os.MkdirAll(nestedDir, 0755); err != nil {
		t.Fatalf("Failed to create nested dir: %v", err)
	}

	if err := os.WriteFile(filepath.Join(nestedDir, "file2.txt"), []byte("file2"), 0644); err != nil {
		t.Fatalf("Failed to write file2: %v", err)
	}

	opts := walker.Options{
		Paths:    []string{filepath.Join(tempDir, "file1.txt"), nestedDir},
		MaxDepth: 0,
		Format:   "plain",
	}

	outChan, err := walker.Walk(opts)
	if err != nil {
		t.Fatalf("Walk failed to start: %v", err)
	}

	foundFile1 := false
	foundFile2 := false

	for res := range outChan {
		if res.Err != nil {
			t.Fatalf("Error during walk: %v", res.Err)
		}
		if res.RelativePath == "file1.txt" || filepath.Base(res.RelativePath) == "file1.txt" {
			foundFile1 = true
		}
		if res.RelativePath == "nested/file2.txt" || filepath.Base(res.RelativePath) == "file2.txt" {
			foundFile2 = true
		}
	}

	if !foundFile1 {
		t.Errorf("Expected to find file1.txt from direct file argument")
	}
	if !foundFile2 {
		t.Errorf("Expected to find file2.txt from directory traversal")
	}
}

func TestWalkNonExistentPaths(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "codeclip-test-nonexistent")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	if err := os.WriteFile(filepath.Join(tempDir, "file1.txt"), []byte("file1"), 0644); err != nil {
		t.Fatalf("Failed to write file1: %v", err)
	}

	// Deliberately use a non-existent path
	opts := walker.Options{
		Paths:    []string{filepath.Join(tempDir, "file1.txt"), filepath.Join(tempDir, "does-not-exist.txt")},
		MaxDepth: 0,
		Format:   "plain",
	}

	outChan, err := walker.Walk(opts)
	if err == nil {
		t.Fatalf("Expected Walk to return an error for does-not-exist.txt, but it didn't")
	}

	if outChan != nil {
		t.Fatalf("Expected outChan to be nil when an error is returned")
	}
}
