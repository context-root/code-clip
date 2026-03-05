package formatter_test

import (
	"bytes"
	"errors"
	"testing"

	"code-clip/pkg/formatter"
	"code-clip/pkg/walker"
)

func TestMarkdownFormatter_WriteResult(t *testing.T) {
	fmtr := formatter.GetFormatter("markdown", 2)
	var buf bytes.Buffer

	// Test normal output
	res := walker.Result{
		RelativePath: "main.go",
		Extension:    ".go",
		Content:      "package main\n",
	}
	fmtr.WriteResult(&buf, res)
	expected := "## `main.go`\n\n```go\npackage main\n\n```\n\n"
	if buf.String() != expected {
		t.Errorf("expected %q, got %q", expected, buf.String())
	}

	buf.Reset()
	// Test error output
	resErr := walker.Result{Err: errors.New("read error")}
	fmtr.WriteResult(&buf, resErr)
	if !bytes.Contains(buf.Bytes(), []byte("Error reading path: read error\n")) {
		t.Errorf("expected error output, got %q", buf.String())
	}
}

func TestMarkdownFormatter_NoExtension(t *testing.T) {
	fmtr := formatter.GetFormatter("markdown", 1)
	var buf bytes.Buffer
	res := walker.Result{
		RelativePath: "Makefile",
		Extension:    "",
		Content:      "all: build",
	}
	fmtr.WriteResult(&buf, res)
	expected := "# `Makefile`\n\n```txt\nall: build\n```\n\n"
	if buf.String() != expected {
		t.Errorf("expected %q, got %q", expected, buf.String())
	}
}

func TestPlainFormatter_WriteResult(t *testing.T) {
	fmtr := formatter.GetFormatter("plain", 1)
	var buf bytes.Buffer
	res := walker.Result{
		RelativePath: "test.txt",
		Content:      "hello",
	}
	fmtr.WriteResult(&buf, res)
	expected := "--- test.txt ---\nhello\n\n"
	if buf.String() != expected {
		t.Errorf("expected %q, got %q", expected, buf.String())
	}

	buf.Reset()
	resErr := walker.Result{Err: errors.New("test error")}
	fmtr.WriteResult(&buf, resErr)
	if !bytes.Contains(buf.Bytes(), []byte("Error reading path: test error\n")) {
		t.Errorf("expected error output, got %q", buf.String())
	}
}

func TestXMLFormatter_WriteResult(t *testing.T) {
	fmtr := formatter.GetFormatter("xml", 1)
	var buf bytes.Buffer

	fmtr.WriteHeader(&buf)
	res := walker.Result{
		RelativePath: "data.xml",
		Content:      "<data>val</data>",
	}
	fmtr.WriteResult(&buf, res)
	fmtr.WriteFooter(&buf)

	expected := "<documents>\n<document>\n<source>data.xml</source>\n<document_content>\n<data>val</data>\n</document_content>\n</document>\n\n</documents>\n"
	if buf.String() != expected {
		t.Errorf("expected %q, got %q", expected, buf.String())
	}

	buf.Reset()
	fmtr.WriteResult(&buf, walker.Result{Err: errors.New("xml error")})
	if !bytes.Contains(buf.Bytes(), []byte("Error reading path: xml error\n")) {
		t.Errorf("expected error output, got %q", buf.String())
	}
}
