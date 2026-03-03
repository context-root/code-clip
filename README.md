# code-clip

`code-clip` prints source files and formats them for LLM chat inputs (ChatGPT, Claude, etc.).

What it does:

```bash
~/my-project $ code-clip . | pbcopy
Copied 7 files (~4251 tokens)
```

It's designed for performance and with one thing in mind: to get the code you actually want to paste in to your LLM chat.

Use `-o xml` for Claude and ChatGPT.

```bash
~/my-project $ code-clip . -o xml | pbcopy
Copied 7 files (~4251 tokens)
```

For clipboard support, pipe the output to `pbcopy` (macOS), `xclip` (Linux), or `clip` (Windows).

## Features

* Recursively print directory contents
* Respects `.gitignore`, `.ignore`, and `.cursorignore` files, even in parent directories.
* Automatically respects `.gitignore`, `.ignore`, and `.cursorignore` files in the current directory and **all ancestor directories up to the root**.
  * Ignore custom files and folders using `-i` or `--ignore`
* Output format: Markdown or XML (`-o` or `--output-format`). **XML is highly recommended for Claude and ChatGPT.**
* Automatically prints a token-count estimation to `stderr` upon completion.
* Limit the traversal depth with `-d` or `--max-depth`
* Customize the Markdown heading depth using `-m` or `--markdown-depth` (e.g., `-m 3` will generate `### filename.go`).

## Installation

```bash
go install github.com/omarish/code-clip/cmd/code-clip@latest
```

*(Or clone this repository and run `make install`)*
*(Coming soon: homebrew and npx packages)*
