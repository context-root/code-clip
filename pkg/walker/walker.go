package walker

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/boyter/gocodewalker"
)

// Options holds configuration for the code-clip directory traversal.
type Options struct {
	// Paths to walk
	Paths []string

	// MaxDepth limits the recursion depth. 0 means infinite.
	MaxDepth int

	// Ignore defined custom ignore patterns.
	Ignore []string

	// Format specifies the output format (markdown, plain)
	Format string
}

// Result holds information about a single file found during the walk.
// It is intended to be passed to the formatter.
type Result struct {
	RelativePath string
	Extension    string
	Content      string
	Err          error
}

type rootInfo struct {
	path  string
	isDir bool
}

// Walk traverses the given paths, respecting .gitignore, hidden files,
// and custom options. It returns a channel of Results.
func Walk(opts Options) (<-chan Result, error) {
	roots := buildRootInfos(opts.Paths)

	var dirs []string
	for _, p := range opts.Paths {
		absPath, err := filepath.Abs(p)
		if err != nil {
			absPath = p
		}

		info, err := os.Stat(absPath)
		if err != nil {
			if os.IsNotExist(err) {
				return nil, fmt.Errorf("path %q does not exist", p)
			}
			return nil, fmt.Errorf("invalid path %q: %w", p, err)
		}

		if info.IsDir() {
			dirs = append(dirs, absPath)
		}
	}

	out := make(chan Result)

	go func() {
		defer close(out)

		for _, p := range opts.Paths {
			absPath, err := filepath.Abs(p)
			if err != nil {
				absPath = p
			}

			info, _ := os.Stat(absPath)
			if !info.IsDir() {
				if res, ok := buildResult(absPath, roots); ok {
					out <- res
				}
			}
		}

		if len(dirs) == 0 {
			return
		}

		fileChan := make(chan *gocodewalker.File)
		errChan := make(chan error, 1)

		// Configure gocodewalker for normal directory traversal
		cw := gocodewalker.NewParallelFileWalker(dirs, fileChan)
		cw.IgnoreGitIgnore = false
		cw.IgnoreIgnoreFile = false
		cw.IncludeHidden = false // Hidden files ignored by default if false

		globalExcludes := collectIgnorePatterns(opts.Paths)
		if len(opts.Ignore) > 0 {
			globalExcludes = append(globalExcludes, opts.Ignore...)
		}
		if len(globalExcludes) > 0 {
			cw.LocationExcludePattern = globalExcludes
		}

		cw.SetErrorHandler(func(e error) bool {
			// Don't halt on error, just emit it.
			out <- Result{Err: e}
			return true
		})

		go func() {
			errChan <- cw.Start()
		}()

		for f := range fileChan {
			if opts.MaxDepth > 0 && isTooDeep(f.Location, roots, opts.MaxDepth) {
				continue
			}

			if res, ok := buildResult(f.Location, roots); ok {
				out <- res
			}
		}

		// Wait for cw.Start() to finish
		if err := <-errChan; err != nil {
			out <- Result{Err: fmt.Errorf("walker error: %w", err)}
		}
	}()

	return out, nil
}


func buildRootInfos(paths []string) []rootInfo {
	roots := make([]rootInfo, 0, len(paths))
	for _, p := range paths {
		absPath, err := filepath.Abs(p)
		if err != nil {
			absPath = p
		}

		info, err := os.Stat(absPath)
		roots = append(roots, rootInfo{
			path:  absPath,
			isDir: err == nil && info.IsDir(),
		})
	}

	return roots
}

func collectIgnorePatterns(paths []string) []string {
	ignoreFileNames := []string{".gitignore", ".ignore", ".cursorignore"}
	seen := make(map[string]struct{})
	var globalExcludes []string

	for _, p := range paths {
		absPath, err := filepath.Abs(p)
		if err != nil {
			continue
		}

		dir := absPath
		if info, err := os.Stat(dir); err == nil && !info.IsDir() {
			dir = filepath.Dir(dir)
		}

		for {
			for _, ignoreName := range ignoreFileNames {
				potentialFile := filepath.Join(dir, ignoreName)
				if info, err := os.Stat(potentialFile); err == nil && !info.IsDir() {
					content, err := os.ReadFile(potentialFile)
					if err != nil {
						continue
					}

					lines := strings.Split(string(content), "\n")
					for _, line := range lines {
						line = strings.TrimSpace(line)
						if line == "" || strings.HasPrefix(line, "#") {
							continue
						}
						if _, ok := seen[line]; ok {
							continue
						}
						seen[line] = struct{}{}
						globalExcludes = append(globalExcludes, line)
					}
				}
			}

			parent := filepath.Dir(dir)
			if parent == dir {
				break
			}
			dir = parent
		}
	}

	return globalExcludes
}

func isTooDeep(location string, roots []rootInfo, maxDepth int) bool {
	for _, root := range roots {
		if !root.isDir {
			continue
		}

		rel, ok := relFromRoot(root.path, location)
		if !ok || rel == "." {
			continue
		}

		if pathDepth(rel) > maxDepth {
			return true
		}
	}

	return false
}

func pathDepth(rel string) int {
	if rel == "" || rel == "." {
		return 0
	}

	return strings.Count(filepath.ToSlash(rel), "/") + 1
}

func buildResult(location string, roots []rootInfo) (Result, bool) {
	contentBytes, err := os.ReadFile(location)
	if err != nil {
		return Result{}, false
	}

	return Result{
		RelativePath: relativePath(location, roots),
		Extension:    filepath.Ext(location),
		Content:      string(contentBytes),
	}, true
}

func relativePath(location string, roots []rootInfo) string {
	for _, root := range roots {
		rel, ok := relFromRoot(root.path, location)
		if !ok {
			continue
		}

		if root.isDir {
			return rel
		}

		if rel == "." {
			return filepath.Base(location)
		}
	}

	return location
}

func relFromRoot(root, location string) (string, bool) {
	rel, err := filepath.Rel(root, location)
	if err != nil {
		return "", false
	}
	if rel == "." {
		return rel, true
	}

	if rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return "", false
	}

	return rel, true
}
