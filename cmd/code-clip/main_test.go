package main

import (
	"reflect"
	"testing"
)

func TestParseArgsFormatAfterPath(t *testing.T) {
	cfg, err := parseArgs([]string{"./path/to/file", "--format", "xml"})
	if err != nil {
		t.Fatalf("parseArgs returned error: %v", err)
	}

	if cfg.options.Format != "xml" {
		t.Fatalf("expected format xml, got %q", cfg.options.Format)
	}

	if !reflect.DeepEqual(cfg.options.Paths, []string{"./path/to/file"}) {
		t.Fatalf("unexpected paths: %v", cfg.options.Paths)
	}
}

func TestParseArgsTerminatorPreservesPaths(t *testing.T) {
	cfg, err := parseArgs([]string{"--format", "plain", "--", "--not-flag", "dir"})
	if err != nil {
		t.Fatalf("parseArgs returned error: %v", err)
	}

	if cfg.options.Format != "plain" {
		t.Fatalf("expected format plain, got %q", cfg.options.Format)
	}

	want := []string{"--not-flag", "dir"}
	if !reflect.DeepEqual(cfg.options.Paths, want) {
		t.Fatalf("unexpected paths: %v", cfg.options.Paths)
	}
}

func TestParseArgsMultiplePathsAndIgnore(t *testing.T) {
	cfg, err := parseArgs([]string{"./a", "--format", "plain", "./b", "-i", "node_modules", "./c"})
	if err != nil {
		t.Fatalf("parseArgs returned error: %v", err)
	}

	wantPaths := []string{"./a", "./b", "./c"}
	if !reflect.DeepEqual(cfg.options.Paths, wantPaths) {
		t.Fatalf("unexpected paths: %v", cfg.options.Paths)
	}

	if !reflect.DeepEqual(cfg.options.Ignore, []string{"node_modules"}) {
		t.Fatalf("unexpected ignores: %v", cfg.options.Ignore)
	}
}

func TestParseArgsOutputFormatAlias(t *testing.T) {
	cfg, err := parseArgs([]string{"--output-format", "xml"})
	if err != nil {
		t.Fatalf("parseArgs returned error: %v", err)
	}

	if cfg.options.Format != "xml" {
		t.Fatalf("expected format xml, got %q", cfg.options.Format)
	}
}

func TestParseArgsNoTokenCount(t *testing.T) {
	cfg, err := parseArgs([]string{"--no-token-count"})
	if err != nil {
		t.Fatalf("parseArgs returned error: %v", err)
	}

	if cfg.printSummary {
		t.Fatalf("expected printSummary to be false")
	}
}

func TestParseArgsUnknownFlag(t *testing.T) {
	_, err := parseArgs([]string{"--not-a-flag"})
	if err == nil {
		t.Fatalf("expected error for unknown flag")
	}
}
