package formatter

import (
	"fmt"
	"io"
	"strings"

	"code-clip/pkg/walker"
)

// Formatter defines an interface for outputting formatting.
type Formatter interface {
	WriteHeader(w io.Writer)
	WriteResult(w io.Writer, res walker.Result)
	WriteFooter(w io.Writer)
}

// GetFormatter returns the appropriate formatter for the requested format.
func GetFormatter(format string, markdownDepth int) Formatter {
	switch format {
	case "xml":
		return &XMLFormatter{}
	case "plain":
		return &PlainFormatter{}
	case "markdown":
		fallthrough
	default:
		return &MarkdownFormatter{Depth: markdownDepth}
	}
}

// MarkdownFormatter formats output as standard Markdown.
type MarkdownFormatter struct {
	Depth int
}

func (f *MarkdownFormatter) WriteHeader(w io.Writer) {}
func (f *MarkdownFormatter) WriteFooter(w io.Writer) {}
func (f *MarkdownFormatter) WriteResult(w io.Writer, res walker.Result) {
	if res.Err != nil {
		fmt.Fprintf(w, "Error reading path: %v\n", res.Err)
		return
	}
	ext := strings.TrimPrefix(res.Extension, ".")
	if ext == "" {
		ext = "txt"
	}
	prefix := strings.Repeat("#", f.Depth)
	fmt.Fprintf(w, "%s `%s`\n\n```%s\n%s\n```\n\n", prefix, res.RelativePath, ext, res.Content)
}

// PlainFormatter formats output as plain text with simple dividers.
type PlainFormatter struct{}

func (f *PlainFormatter) WriteHeader(w io.Writer) {}
func (f *PlainFormatter) WriteFooter(w io.Writer) {}
func (f *PlainFormatter) WriteResult(w io.Writer, res walker.Result) {
	if res.Err != nil {
		fmt.Fprintf(w, "Error reading path: %v\n", res.Err)
		return
	}
	fmt.Fprintf(w, "--- %s ---\n%s\n\n", res.RelativePath, res.Content)
}

// XMLFormatter formats output specifically optimizing for LLM ingestion.
type XMLFormatter struct{}

func (f *XMLFormatter) WriteHeader(w io.Writer) {
	fmt.Fprintf(w, "<documents>\n")
}
func (f *XMLFormatter) WriteFooter(w io.Writer) {
	fmt.Fprintf(w, "</documents>\n")
}
func (f *XMLFormatter) WriteResult(w io.Writer, res walker.Result) {
	if res.Err != nil {
		fmt.Fprintf(w, "Error reading path: %v\n", res.Err)
		return
	}
	fmt.Fprintf(w, "<document>\n<source>%s</source>\n<document_content>\n%s\n</document_content>\n</document>\n\n", res.RelativePath, res.Content)
}
