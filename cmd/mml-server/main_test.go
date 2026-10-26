package main

import (
	"net"
	"os"
	"path/filepath"
	"strconv"
	"testing"
)

func TestRunInvalidPort(t *testing.T) {
	if code := run([]string{"--port", "0"}); code != 1 {
		t.Fatalf("port 0: exit %d", code)
	}
	if code := run([]string{"--port", "70000"}); code != 1 {
		t.Fatalf("port 70000: exit %d", code)
	}
}

func TestRunMissingMapping(t *testing.T) {
	path := filepath.Join(t.TempDir(), "nope.txt")
	if code := run([]string{"--mapping", path, "--host", "127.0.0.1", "--port", "2323"}); code != 1 {
		t.Fatalf("missing mapping: exit %d", code)
	}
}

func TestRunListenInUse(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer ln.Close()
	_, port, err := net.SplitHostPort(ln.Addr().String())
	if err != nil {
		t.Fatal(err)
	}
	if _, err := strconv.Atoi(port); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(t.TempDir(), "m.txt")
	if err := os.WriteFile(path, []byte("PING\tPONG\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	code := run([]string{"--host", "127.0.0.1", "--port", port, "--mapping", path})
	if code != 1 {
		t.Fatalf("listen in use: exit %d", code)
	}
}
