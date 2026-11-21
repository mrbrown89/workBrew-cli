package main

import (
	"bytes"
	"errors"
	"strings"
	"testing"
)

func TestOutputFormatSet(t *testing.T) {
	tests := []struct {
		in      string
		want    outputFormat
		wantErr bool
	}{
		{"table", formatTable, false},
		{"json", formatJSON, false},
		{"JSON", formatJSON, false},
		{" json ", formatJSON, false},
		{"yaml", "", true},
		{"", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.in, func(t *testing.T) {
			var f outputFormat
			err := f.Set(tt.in)
			if (err != nil) != tt.wantErr {
				t.Fatalf("Set(%q) error = %v, wantErr %v", tt.in, err, tt.wantErr)
			}
			if f != tt.want {
				t.Errorf("Set(%q) = %q, want %q", tt.in, f, tt.want)
			}
		})
	}
}

func TestPrintJSON(t *testing.T) {
	var buf bytes.Buffer
	if err := printJSON(&buf, map[string]any{"items": []string{}}); err != nil {
		t.Fatal(err)
	}

	want := "{\n  \"items\": []\n}\n"
	if buf.String() != want {
		t.Errorf("printJSON = %q, want %q", buf.String(), want)
	}
}

// failingWriter accepts limit bytes, then fails every write.
type failingWriter struct {
	limit  int
	writes int
}

var errDiskFull = errors.New("no space left on device")

func (w *failingWriter) Write(b []byte) (int, error) {
	w.writes++
	if len(b) > w.limit {
		n := w.limit
		w.limit = 0
		return n, errDiskFull
	}
	w.limit -= len(b)
	return len(b), nil
}

func TestPrinterKeepsFirstError(t *testing.T) {
	w := &failingWriter{limit: 5}
	p := newPrinter(w)

	p.Println("hello world") // fails part-way
	p.Printf("%d\n", 42)     // skipped
	p.Print("more")          // skipped

	if !errors.Is(p.Err(), errDiskFull) {
		t.Fatalf("Err() = %v, want %v", p.Err(), errDiskFull)
	}
	if !strings.HasPrefix(p.Err().Error(), "writing output: ") {
		t.Errorf("Err() = %q, want \"writing output: ...\"", p.Err())
	}
	if w.writes != 1 {
		t.Errorf("underlying writes = %d, want 1 (later writes skipped)", w.writes)
	}

	if err := newPrinter(&bytes.Buffer{}).Err(); err != nil {
		t.Errorf("Err() with no writes = %v, want nil", err)
	}
}

func TestTableAligns(t *testing.T) {
	var buf bytes.Buffer
	p := newPrinter(&buf)

	table := newTable(p)
	table.Row("Name", "Count")
	table.Row("a-much-longer-name", 3)
	table.Flush()

	want := "Name                Count\na-much-longer-name  3\n"
	if buf.String() != want {
		t.Errorf("table =\n%q\nwant\n%q", buf.String(), want)
	}
	if p.Err() != nil {
		t.Errorf("Err() = %v", p.Err())
	}
}
