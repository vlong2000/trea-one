package service

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/vlong2000/trea-one/nfs/nrf/pkg/factory"
)

func freePort(t *testing.T) int {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	port := ln.Addr().(*net.TCPAddr).Port
	_ = ln.Close()
	return port
}

func writeCfg(t *testing.T, port int) string {
	t.Helper()
	body := fmt.Sprintf(`info:
  version: 1.0.2
  description: test
configuration:
  sbi:
    scheme: http
    registerIPv4: 127.0.0.1
    bindingIPv4: 127.0.0.1
    port: %d
    oauth: false
  DefaultPlmnId:
    mcc: 208
    mnc: 93
  serviceNameList:
    - nnrf-nfm
    - nnrf-disc
logger:
  enable: true
  level: info
  reportCaller: false
`, port)
	path := filepath.Join(t.TempDir(), "nrfcfg.yaml")
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestStartServesSBI(t *testing.T) {
	port := freePort(t)
	cfg, err := factory.ReadConfig(writeCfg(t, port))
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	errCh := make(chan error, 1)
	go func() { errCh <- New(cfg).Start(ctx) }()

	base := fmt.Sprintf("http://127.0.0.1:%d", port)
	deadline := time.Now().Add(3 * time.Second)
	var resp *http.Response
	for time.Now().Before(deadline) {
		resp, err = http.Get(base + "/bootstrapping")
		if err == nil {
			break
		}
		time.Sleep(20 * time.Millisecond)
	}
	if err != nil {
		t.Fatalf("bootstrapping: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status=%d", resp.StatusCode)
	}
	body, _ := io.ReadAll(resp.Body)
	var payload map[string]any
	if err := json.Unmarshal(body, &payload); err != nil {
		t.Fatal(err)
	}
	if payload["status"] != "OPERATIVE" {
		t.Fatalf("payload=%v", payload)
	}

	cancel()
	select {
	case err := <-errCh:
		if err != nil {
			t.Fatalf("start: %v", err)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("shutdown timeout")
	}
}

func TestBinaryHasOnlyNRFMain(t *testing.T) {
	if testing.Short() {
		t.Skip("short")
	}
	bin := filepath.Join(t.TempDir(), "nrf")
	cmd := exec.Command("go", "build", "-o", bin, "./cmd")
	cmd.Dir = moduleRoot(t)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("build: %v\n%s", err, out)
	}
	info, err := os.Stat(bin)
	if err != nil || info.Size() == 0 {
		t.Fatalf("binary missing")
	}
}

func moduleRoot(t *testing.T) string {
	t.Helper()
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("caller")
	}
	return filepath.Clean(filepath.Join(filepath.Dir(file), "..", ".."))
}
