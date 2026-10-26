package server

import (
	"context"
	"net"
	"os"
	"path/filepath"
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

func startServer(t *testing.T) (addr string, cancel context.CancelFunc) {
	t.Helper()
	ctx, cancel := context.WithCancel(context.Background())
	srv := New("127.0.0.1:0", loadTable(t))
	errCh := make(chan error, 1)
	go func() { errCh <- srv.ListenAndServe(ctx) }()
	addr = srv.Addr()
	t.Cleanup(func() {
		cancel()
		select {
		case err := <-errCh:
			if err != nil {
				t.Errorf("ListenAndServe: %v", err)
			}
		case <-time.After(3 * time.Second):
			t.Error("server did not stop")
		}
	})
	return addr, cancel
}

func sendRecv(t *testing.T, conn net.Conn, req string) string {
	t.Helper()
	conn.SetDeadline(time.Now().Add(2 * time.Second))
	if _, err := conn.Write([]byte(req)); err != nil {
		t.Fatal(err)
	}
	buf := make([]byte, 256)
	n, err := conn.Read(buf)
	if err != nil {
		t.Fatal(err)
	}
	return string(buf[:n])
}

func TestPingLF(t *testing.T) {
	addr, _ := startServer(t)
	c, err := net.Dial("tcp", addr)
	if err != nil {
		t.Fatal(err)
	}
	defer c.Close()
	if got := sendRecv(t, c, "PING\n"); got != "PONG\r\n" {
		t.Fatalf("got %q", got)
	}
}

func TestPingCRLF(t *testing.T) {
	addr, _ := startServer(t)
	c, err := net.Dial("tcp", addr)
	if err != nil {
		t.Fatal(err)
	}
	defer c.Close()
	if got := sendRecv(t, c, "PING\r\n"); got != "PONG\r\n" {
		t.Fatalf("got %q", got)
	}
}

func TestUnknownCommand(t *testing.T) {
	addr, _ := startServer(t)
	c, err := net.Dial("tcp", addr)
	if err != nil {
		t.Fatal(err)
	}
	defer c.Close()
	if got := sendRecv(t, c, "FOO\n"); got != "ERROR: unknown command\r\n" {
		t.Fatalf("got %q", got)
	}
}

func TestTwoClients(t *testing.T) {
	addr, _ := startServer(t)
	a, err := net.Dial("tcp", addr)
	if err != nil {
		t.Fatal(err)
	}
	defer a.Close()
	b, err := net.Dial("tcp", addr)
	if err != nil {
		t.Fatal(err)
	}
	defer b.Close()
	if got := sendRecv(t, a, "PING\n"); got != "PONG\r\n" {
		t.Fatalf("a: %q", got)
	}
	if got := sendRecv(t, b, "HELLO\n"); got != "WORLD\r\n" {
		t.Fatalf("b: %q", got)
	}
	if got := sendRecv(t, a, "HELLO\n"); got != "WORLD\r\n" {
		t.Fatalf("a2: %q", got)
	}
}

func TestPeerDisconnect(t *testing.T) {
	addr, _ := startServer(t)
	a, err := net.Dial("tcp", addr)
	if err != nil {
		t.Fatal(err)
	}
	b, err := net.Dial("tcp", addr)
	if err != nil {
		t.Fatal(err)
	}
	defer b.Close()
	a.Close()
	if got := sendRecv(t, b, "PING\n"); got != "PONG\r\n" {
		t.Fatalf("got %q", got)
	}
}

func TestListenAddressInUse(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer ln.Close()
	srv := New(ln.Addr().String(), loadTable(t))
	err = srv.ListenAndServe(context.Background())
	if err == nil {
		t.Fatal("expected listen error")
	}
}
