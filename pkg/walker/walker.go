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

// Walk traverses the given paths, respecting .gitignore, hidden files,
// and custom options. It returns a channel of Results.
func Walk(opts Options) (<-chan Result, error) {
	out := make(chan Result)

	go func() {
		defer close(out)

		fileChan := make(chan *gocodewalker.File)
		errChan := make(chan error, 1)

		// If the user specified exactly one file (not a directory), gocodewalker will fail.
		// We should handle that cleanly by just yielding the file immediately.
		if len(opts.Paths) == 1 {
			info, err := os.Stat(opts.Paths[0])
			if err == nil && !info.IsDir() {
				// We must yield to the channel in a goroutine because it's unbuffered,
				// and the consumer loop is further down in THIS goroutine.
				go func() {
					fileChan <- &gocodewalker.File{
						Location: opts.Paths[0],
						Filename: filepath.Base(opts.Paths[0]),
					}
					close(fileChan)
				}()
				// Unblock the errChan listener at the end of the Walk function
				errChan <- nil
				// We don't return here! We still need to let the loop below process `fileChan`
				// into the `out` channel, so it actually gets sent to the caller.
				goto ProcessLoop
			}
		}

		{
			// Configure gocodewalker for normal directory traversal
			cw := gocodewalker.NewParallelFileWalker(opts.Paths, fileChan)
			cw.IgnoreGitIgnore = false
			cw.IgnoreIgnoreFile = false
			cw.IncludeHidden = false // Hidden files ignored by default if false

			// 1. Manually ascend the directory tree of each input path to root,
			// looking for `.gitignore`, `.ignore`, or `.cursorignore`. 
			var globalExcludes []string
			ignoreFileNames := []string{".gitignore", ".ignore", ".cursorignore"}
			
			for _, p := range opts.Paths {
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
							if err == nil {
								lines := strings.Split(string(content), "\n")
								for _, line := range lines {
									line = strings.TrimSpace(line)
									if line != "" && !strings.HasPrefix(line, "#") {
										globalExcludes = append(globalExcludes, line)
									}
								}
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

			// Map custom ignore patterns if any
			if len(opts.Ignore) > 0 {
				globalExcludes = append(globalExcludes, opts.Ignore...)
			}
			
			if len(globalExcludes) > 0 {
				cw.LocationExcludePattern = globalExcludes
			}

			cw.SetErrorHandler(func(e error) bool {
				// Don't halt on error, just log/emit it.
				out <- Result{Err: e}
				return true
			})

			go func() {
				errChan <- cw.Start()
			}()
		}

	ProcessLoop:
		for f := range fileChan {
			// Handle max depth manually since the library doesn't expose it directly yet
			// f.Location is the absolute path. We need to find its depth relative to the
			// input path it was found under.

			// We skip directories in output by default anyway, but let's ensure
			// we don't output files deeper than MaxDepth.
			if opts.MaxDepth > 0 {
				isTooDeep := false
				for _, root := range opts.Paths {
					absRoot, err := filepath.Abs(root)
					if err != nil {
						continue
					}

					if strings.HasPrefix(f.Location, absRoot) {
						rel, err := filepath.Rel(absRoot, f.Location)
						if err == nil {
							// Count path separators. "." is depth 0. "foo" is depth 1.
							// "foo/bar" is depth 2.
							if rel == "." {
								// Root itself, depth 0. (Unlikely to be a file returned by walker, but just in case)
								break
							}

							depth := strings.Count(filepath.ToSlash(rel), "/") + 1
							if depth > opts.MaxDepth {
								isTooDeep = true
								break
							}
						}
					}
				}

				if isTooDeep {
					continue
				}
			}

			// Read file content and formulate result
			// We defer content loading to the formatter or load it here?
			// Since the goal is printing, loading it here is fine.

			// To match the rust version, we try to read to string. If it fails (e.g., binary),
			// we could skip or emit error. For now, let's keep it simple.
			contentBytes, err := os.ReadFile(f.Location)
			if err != nil {
				// We don't error out entirely, just skip or record it.
				continue
			}
			content := string(contentBytes)

			// Compute relative path for better display
			relPath := f.Location
			for _, root := range opts.Paths {
				absRoot, _ := filepath.Abs(root)
				if strings.HasPrefix(f.Location, absRoot) {
					rel, _ := filepath.Rel(absRoot, f.Location)
					relPath = rel
					break
				}
			}

			out <- Result{
				RelativePath: relPath,
				Extension:    filepath.Ext(f.Location),
				Content:      content,
				Err:          nil,
			}
		}

		// Wait for cw.Start() to finish
		if err := <-errChan; err != nil {
			out <- Result{Err: fmt.Errorf("walker error: %w", err)}
		}
	}()

	return out, nil
}
