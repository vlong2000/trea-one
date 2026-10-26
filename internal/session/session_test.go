package session

import (
	"bufio"
	"bytes"
	"context"
	"io"
	"net"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"mml-server/internal/mapping"
)

func loadTable(t *testing.T) *mapping.Table {
	t.Helper()
	path := filepath.Join(t.TempDir(), "m.txt")
	if err := os.WriteFile(path, []byte("PING\tPONG\nHELLO\tWORLD\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	tbl, err := mapping.Load(path)
	if err != nil {
		t.Fatal(err)
	}
	return tbl
}

func startSession(t *testing.T, tbl *mapping.Table) (client net.Conn, cancel context.CancelFunc, done chan struct{}) {
	t.Helper()
	c, s := net.Pipe()
	ctx, cancel := context.WithCancel(context.Background())
	done = make(chan struct{})
	go func() {
		defer close(done)
		Handle(ctx, s, tbl)
	}()
	t.Cleanup(func() {
		cancel()
		c.Close()
		select {
		case <-done:
		case <-time.After(2 * time.Second):
			t.Error("session did not exit")
		}
	})
	return c, cancel, done
}

func readLine(t *testing.T, r io.Reader) string {
	t.Helper()
	buf := make([]byte, 256)
	n, err := r.Read(buf)
	if err != nil {
		t.Fatal(err)
	}
	return string(buf[:n])
}

func TestHandleLF(t *testing.T) {
	c, _, _ := startSession(t, loadTable(t))
	if _, err := c.Write([]byte("PING\n")); err != nil {
		t.Fatal(err)
	}
	if got := readLine(t, c); got != "PONG\r\n" {
		t.Fatalf("got %q", got)
	}
}

func TestHandleCRLF(t *testing.T) {
	c, _, _ := startSession(t, loadTable(t))
	if _, err := c.Write([]byte("PING\r\n")); err != nil {
		t.Fatal(err)
	}
	if got := readLine(t, c); got != "PONG\r\n" {
		t.Fatalf("got %q", got)
	}
}

func TestHandleTwoLines(t *testing.T) {
	c, _, _ := startSession(t, loadTable(t))
	if _, err := c.Write([]byte("PING\nHELLO\n")); err != nil {
		t.Fatal(err)
	}
	data := make([]byte, 0, 32)
	deadline := time.Now().Add(2 * time.Second)
	for len(data) < len("PONG\r\nWORLD\r\n") {
		c.SetReadDeadline(deadline)
		buf := make([]byte, 32)
		n, err := c.Read(buf)
		if err != nil {
			t.Fatal(err)
		}
		data = append(data, buf[:n]...)
	}
	if string(data) != "PONG\r\nWORLD\r\n" {
		t.Fatalf("got %q", data)
	}
}

func TestHandleUnknown(t *testing.T) {
	c, _, _ := startSession(t, loadTable(t))
	if _, err := c.Write([]byte("NOPE\n")); err != nil {
		t.Fatal(err)
	}
	if got := readLine(t, c); got != "ERROR: unknown command\r\n" {
		t.Fatalf("got %q", got)
	}
}

func TestHandleIAC(t *testing.T) {
	c, _, _ := startSession(t, loadTable(t))
	payload := append([]byte{0xFF, 251, 1}, []byte("PING\n")...)
	if _, err := c.Write(payload); err != nil {
		t.Fatal(err)
	}
	if got := readLine(t, c); got != "PONG\r\n" {
		t.Fatalf("got %q", got)
	}
}

func TestHandleLineTooLong(t *testing.T) {
	c, _, done := startSession(t, loadTable(t))
	long := bytes.Repeat([]byte("A"), 4097)
	go func() { _, _ = c.Write(long) }()
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("expected session close on oversize line")
	}
}

func TestReadRequestLineVariants(t *testing.T) {
	cases := []struct {
		in   string
		want string
	}{
		{"CMD\n", "CMD"},
		{"CMD\r\n", "CMD"},
	}
	for _, tc := range cases {
		r := bufio.NewReader(strings.NewReader(tc.in))
		got, err := readRequestLine(r)
		if err != nil {
			t.Fatalf("%q: %v", tc.in, err)
		}
		if got != tc.want {
			t.Fatalf("%q: got %q want %q", tc.in, got, tc.want)
		}
	}
}
