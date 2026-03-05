package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"

	"code-clip/pkg/formatter"
	"code-clip/pkg/walker"
)

// stringSlice is a custom flag type for capturing multiple instances of a flag (e.g., -i).
type stringSlice []string

func (i *stringSlice) String() string {
	return strings.Join(*i, ", ")
}

func (i *stringSlice) Set(value string) error {
	*i = append(*i, value)
	return nil
}

type cliOptions struct {
	options       walker.Options
	markdownDepth int
	printSummary  bool
}

func usage(w io.Writer) {
	fmt.Fprintf(w, "code-clip is a CLI tool for recursively printing directory contents for LLM context.\n\n")
	fmt.Fprintf(w, "Usage: code-clip [OPTIONS] [PATHS...]\n\n")
	fmt.Fprintf(w, "Options:\n")
	fmt.Fprintf(w, "  -o, --format string          Output format (markdown, plain, xml) (default \"markdown\")\n")
	fmt.Fprintf(w, "      --output-format string   Deprecated alias for --format\n")
	fmt.Fprintf(w, "  -m, --markdown-depth int     Markdown heading depth (default 1)\n")
	fmt.Fprintf(w, "  -d, --max-depth int          Maximum depth to recurse (0 = infinite)\n")
	fmt.Fprintf(w, "  -i, --ignore string          Files/patterns to ignore\n")
	fmt.Fprintf(w, "      --no-token-count         Disable token count estimation\n")
	fmt.Fprintf(w, "  -h, --help                   Show help\n")
}

func parseArgs(args []string) (cliOptions, error) {
	var (
		format        string
		maxDepth      int
		markdownDepth int
		ignores       stringSlice
		noTokenCount  bool
		help          bool
	)

	fs := flag.NewFlagSet("code-clip", flag.ContinueOnError)
	fs.SetOutput(io.Discard)

	fs.BoolVar(&help, "h", false, "Show help")
	fs.BoolVar(&help, "help", false, "Show help")
	fs.StringVar(&format, "format", "markdown", "Output format (markdown, plain, xml)")
	fs.StringVar(&format, "output-format", "markdown", "Deprecated alias for --format")
	fs.StringVar(&format, "o", "markdown", "Output format (markdown, plain, xml)")
	fs.IntVar(&maxDepth, "d", 0, "Maximum depth to recurse (0 = infinite)")
	fs.IntVar(&maxDepth, "max-depth", 0, "Maximum depth to recurse (0 = infinite)")
	fs.IntVar(&markdownDepth, "m", 1, "Markdown heading depth")
	fs.IntVar(&markdownDepth, "markdown-depth", 1, "Markdown heading depth")
	fs.Var(&ignores, "i", "Files/patterns to ignore")
	fs.Var(&ignores, "ignore", "Files/patterns to ignore")
	fs.BoolVar(&noTokenCount, "no-token-count", false, "Disable token count estimation")

	flagArgs, paths, err := splitArgs(fs, args)
	if err != nil {
		return cliOptions{}, err
	}

	if err := fs.Parse(flagArgs); err != nil {
		return cliOptions{}, err
	}

	if help {
		return cliOptions{}, flag.ErrHelp
	}

	if len(paths) == 0 {
		paths = []string{"."}
	}

	return cliOptions{
		options: walker.Options{
			Paths:    paths,
			MaxDepth: maxDepth,
			Ignore:   ignores,
			Format:   format,
		},
		markdownDepth: markdownDepth,
		printSummary:  !noTokenCount,
	}, nil
}

func splitArgs(fs *flag.FlagSet, args []string) ([]string, []string, error) {
	var flagArgs []string
	var paths []string

	for i := 0; i < len(args); i++ {
		arg := args[i]
		if arg == "--" {
			paths = append(paths, args[i+1:]...)
			break
		}

		if !strings.HasPrefix(arg, "-") || arg == "-" {
			paths = append(paths, arg)
			continue
		}

		name, hasValue := flagName(arg)
		f := fs.Lookup(name)
		if f == nil {
			return nil, nil, fmt.Errorf("unknown flag: %s", arg)
		}

		if isBoolFlag(f) || hasValue {
			flagArgs = append(flagArgs, arg)
			continue
		}

		if i+1 >= len(args) || args[i+1] == "--" {
			return nil, nil, fmt.Errorf("flag needs an argument: %s", arg)
		}

		flagArgs = append(flagArgs, arg, args[i+1])
		i++
	}

	return flagArgs, paths, nil
}

func flagName(arg string) (string, bool) {
	trimmed := strings.TrimLeft(arg, "-")
	name, _, hasValue := strings.Cut(trimmed, "=")
	return name, hasValue
}

func isBoolFlag(f *flag.Flag) bool {
	type boolFlag interface {
		IsBoolFlag() bool
	}

	bf, ok := f.Value.(boolFlag)
	return ok && bf.IsBoolFlag()
}

func main() {
	opts, err := parseArgs(os.Args[1:])
	if err != nil {
		if errors.Is(err, flag.ErrHelp) {
			usage(os.Stderr)
			return
		}
		fmt.Fprintf(os.Stderr, "Error: %v\n\n", err)
		usage(os.Stderr)
		os.Exit(1)
	}

	// Walk returns a channel of Results, which makes streaming to stdout immediate.
	resultsChan, err := walker.Walk(opts.options)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error initializing walker: %v\n", err)
		os.Exit(1)
	}

	// Instantiate the correct formatter
	fmtr := formatter.GetFormatter(opts.options.Format, opts.markdownDepth)

	// Print optional headers (e.g., <documents> for XML)
	fmtr.WriteHeader(os.Stdout)

	var totalFiles int
	var totalChars int

	// Loop over the channel until closed.
	for res := range resultsChan {
		if res.Err == nil {
			totalFiles++
			totalChars += len(res.Content)
		}
		fmtr.WriteResult(os.Stdout, res)
	}

	// Print optional footers (e.g., </documents> for XML)
	fmtr.WriteFooter(os.Stdout)

	// Print token estimation summary to stderr
	// Rough rule of thumb: 1 token ≈ 4 characters of English text/code
	if opts.printSummary {
		estimatedTokens := totalChars / 4
		fmt.Fprintf(os.Stderr, "[code-clip] Copied %d files (~%d tokens)\n", totalFiles, estimatedTokens)
	}
}
