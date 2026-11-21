package main

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"text/tabwriter"

	"github.com/spf13/cobra"
)

// outputFormat is the value of the -o/--output flag. It implements
// pflag.Value so an invalid format is rejected while flags are parsed.
type outputFormat string

const (
	formatTable outputFormat = "table"
	formatJSON  outputFormat = "json"
)

func (f *outputFormat) String() string { return string(*f) }

func (f *outputFormat) Type() string { return "format" }

func (f *outputFormat) Set(value string) error {
	switch format := outputFormat(strings.ToLower(strings.TrimSpace(value))); format {
	case formatTable, formatJSON:
		*f = format
		return nil
	}

	return fmt.Errorf("must be %q or %q", formatTable, formatJSON)
}

// addOutputFlag registers the shared -o/--output flag on cmd and its subcommands.
func addOutputFlag(cmd *cobra.Command, format *outputFormat) {
	*format = formatTable
	cmd.PersistentFlags().VarP(format, "output", "o", "Output format: table or json")
}

// printJSON writes v to w as indented JSON followed by a newline.
func printJSON(w io.Writer, v any) error {
	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")
	return encoder.Encode(v)
}

// printer writes text to an io.Writer and remembers the first write error,
// so a command can print line by line and check once at the end (for
// example, stdout redirected to a full disk). Writes after an error are
// skipped.
//
// Commands return Err() for stdout. For stderr (prompts, warnings, errors)
// it is not checked: there is nowhere left to report the failure.
type printer struct {
	w   io.Writer
	err error
}

func newPrinter(w io.Writer) *printer {
	return &printer{w: w}
}

// Write implements io.Writer, recording the first error.
func (p *printer) Write(b []byte) (int, error) {
	if p.err != nil {
		return 0, p.err
	}

	n, err := p.w.Write(b)
	if err != nil {
		p.err = err
	}
	return n, err
}

// The fmt errors below come from p.Write, which records them in p.err.

func (p *printer) Print(a ...any)                 { _, _ = fmt.Fprint(p, a...) }
func (p *printer) Println(a ...any)               { _, _ = fmt.Fprintln(p, a...) }
func (p *printer) Printf(format string, a ...any) { _, _ = fmt.Fprintf(p, format, a...) }

// Err returns the first write error, if any.
func (p *printer) Err() error {
	if p.err != nil {
		return fmt.Errorf("writing output: %w", p.err)
	}
	return nil
}

// Heading prints title, an underline and a blank line.
func (p *printer) Heading(title string) {
	p.Println(title)
	p.Println(strings.Repeat("-", len(title)))
	p.Println()
}

// table aligns tab-separated rows with text/tabwriter. Write errors end up
// in the printer it was created from.
type table struct {
	tw *tabwriter.Writer
}

func newTable(p *printer) *table {
	return &table{tw: tabwriter.NewWriter(p, 0, 0, 2, ' ', 0)}
}

// Row writes one row of cells.
func (t *table) Row(cells ...any) {
	line := make([]string, len(cells))
	for i, cell := range cells {
		line[i] = fmt.Sprint(cell)
	}
	// tabwriter only fails when the underlying printer does, which records it.
	_, _ = io.WriteString(t.tw, strings.Join(line, "\t")+"\n")
}

// Flush writes the aligned rows.
func (t *table) Flush() {
	_ = t.tw.Flush() // recorded by the printer, as in Row
}
