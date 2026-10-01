package tui

import (
	"fmt"
	"strings"
	"testing"
)

func TestTableRendersAllRowsWithoutTruncation(t *testing.T) {
	rows := make([][]string, 25)
	for i := range rows {
		rows[i] = []string{fmt.Sprintf("item-%02d", i), "任务", "ready"}
	}
	table, err := NewTable([]string{"NAME", "DETAIL", "STATUS"}, rows)
	if err != nil {
		t.Fatal(err)
	}
	view := table.View()
	for _, want := range []string{"NAME", "DETAIL", "STATUS", "item-00", "item-24", "任务", "ready"} {
		if !strings.Contains(view, want) {
			t.Errorf("table view is missing %q: %q", want, view)
		}
	}
	if got := len(strings.Split(view, "\n")); got != len(rows)+1 {
		t.Errorf("table has %d lines, want %d", got, len(rows)+1)
	}
	if strings.Contains(view, "\x1b") {
		t.Errorf("static table contains terminal escape sequences: %q", view)
	}
}

func TestTableRequiresRectangularRows(t *testing.T) {
	if _, err := NewTable(nil, nil); err == nil {
		t.Fatal("table without columns succeeded")
	}
	if _, err := NewTable([]string{"A", "B"}, [][]string{{"one"}}); err == nil {
		t.Fatal("table with a short row succeeded")
	}
	if _, err := NewTable([]string{"A"}, [][]string{{"one", "two"}}); err == nil {
		t.Fatal("table with a long row succeeded")
	}
}

func TestTableEscapesControlCharacters(t *testing.T) {
	table, err := NewTable([]string{"NAME"}, [][]string{{"line\nnext"}, {"\x1b[31mred"}})
	if err != nil {
		t.Fatal(err)
	}
	view := table.View()
	if !strings.Contains(view, `"line\nnext"`) || !strings.Contains(view, `"\x1b[31mred"`) {
		t.Fatalf("control characters were not escaped: %q", view)
	}
	if got := len(strings.Split(view, "\n")); got != 3 {
		t.Fatalf("table has %d lines, want 3", got)
	}
}
