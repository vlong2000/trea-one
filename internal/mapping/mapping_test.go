package mapping

import (
	"os"
	"path/filepath"
	"testing"
)

func writeFile(t *testing.T, name, content string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), name)
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestLoadValidFile(t *testing.T) {
	path := writeFile(t, "ok.txt", "# comment\n\nPING\tPONG\nHELLO\tWORLD\n")
	tbl, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if got := tbl.Lookup("PING"); got != "PONG" {
		t.Fatalf("PING: got %q", got)
	}
	if got := tbl.Lookup("HELLO"); got != "WORLD" {
		t.Fatalf("HELLO: got %q", got)
	}
}

func TestLoadMissingFile(t *testing.T) {
	_, err := Load(filepath.Join(t.TempDir(), "missing.txt"))
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestLoadEmptyFile(t *testing.T) {
	path := writeFile(t, "empty.txt", "")
	_, err := Load(path)
	if err == nil {
		t.Fatal("expected error for empty file")
	}
}

func TestLoadCommentsOnly(t *testing.T) {
	path := writeFile(t, "comments.txt", "# a\n\n# b\n")
	_, err := Load(path)
	if err == nil {
		t.Fatal("expected error for no valid mappings")
	}
}

func TestLoadDuplicateKeyLastWins(t *testing.T) {
	path := writeFile(t, "dup.txt", "PING\tONE\nPING\tTWO\n")
	tbl, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if got := tbl.Lookup("PING"); got != "TWO" {
		t.Fatalf("got %q want TWO", got)
	}
}

func TestLoadSkipInvalidLines(t *testing.T) {
	path := writeFile(t, "skip.txt", "NOTAB\n\tEMPTYREQ\nPING\tPONG\n")
	tbl, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if got := tbl.Lookup("PING"); got != "PONG" {
		t.Fatalf("got %q", got)
	}
}

func TestLoadEmptyResponseAllowed(t *testing.T) {
	path := writeFile(t, "emptyresp.txt", "PING\t\n")
	tbl, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if got := tbl.Lookup("PING"); got != "" {
		t.Fatalf("got %q want empty", got)
	}
}

func TestLookupUnknown(t *testing.T) {
	path := writeFile(t, "ok.txt", "PING\tPONG\n")
	tbl, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if got := tbl.Lookup("NOPE"); got != UnknownCommand {
		t.Fatalf("got %q", got)
	}
	if got := tbl.Lookup(""); got != UnknownCommand {
		t.Fatalf("empty request got %q", got)
	}
}
