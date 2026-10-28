package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestMissingConfigExitsNonZero(t *testing.T) {
	bin := filepath.Join(t.TempDir(), "nrf")
	build := exec.Command("go", "build", "-o", bin, ".")
	build.Dir = "."
	if out, err := build.CombinedOutput(); err != nil {
		t.Fatalf("build: %v\n%s", err, out)
	}
	cmd := exec.Command(bin, "-c", filepath.Join(t.TempDir(), "missing.yaml"))
	cmd.Env = os.Environ()
	out, err := cmd.CombinedOutput()
	if err == nil {
		t.Fatalf("expected non-zero exit, out=%s", out)
	}
	if cmd.ProcessState.ExitCode() == 0 {
		t.Fatal("exit code 0")
	}
}
