package factory

import (
	"os"
	"path/filepath"
	"testing"
)

func writeTempConfig(t *testing.T, body string) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "nrfcfg.yaml")
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}
	return path
}

func validYAML() string {
	return `info:
  version: 1.0.2
  description: test
configuration:
  sbi:
    scheme: http
    registerIPv4: 127.0.0.1
    bindingIPv4: 127.0.0.1
    port: 8000
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
`
}

func TestReadConfigSuccess(t *testing.T) {
	path := writeTempConfig(t, validYAML())
	cfg, err := ReadConfig(path)
	if err != nil {
		t.Fatalf("ReadConfig: %v", err)
	}
	if cfg.BindingAddr() != "127.0.0.1:8000" {
		t.Fatalf("BindingAddr=%s", cfg.BindingAddr())
	}
	if cfg.SbiURI() != "http://127.0.0.1:8000" {
		t.Fatalf("SbiURI=%s", cfg.SbiURI())
	}
}

func TestReadConfigMissingFile(t *testing.T) {
	_, err := ReadConfig(filepath.Join(t.TempDir(), "missing.yaml"))
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestReadConfigInvalidYAML(t *testing.T) {
	path := writeTempConfig(t, "::::")
	_, err := ReadConfig(path)
	if err == nil {
		t.Fatal("expected parse error")
	}
}

func TestReadConfigInvalidPort(t *testing.T) {
	path := writeTempConfig(t, `info:
  version: 1.0.2
configuration:
  sbi:
    scheme: http
    bindingIPv4: 127.0.0.1
    port: 70000
  DefaultPlmnId:
    mcc: 208
    mnc: 93
  serviceNameList:
    - nnrf-nfm
logger:
  enable: true
  level: info
`)
	_, err := ReadConfig(path)
	if err == nil {
		t.Fatal("expected invalid port")
	}
}

func TestReadConfigInvalidService(t *testing.T) {
	path := writeTempConfig(t, `info:
  version: 1.0.2
configuration:
  sbi:
    scheme: http
    bindingIPv4: 127.0.0.1
    port: 8000
  DefaultPlmnId:
    mcc: 208
    mnc: 93
  serviceNameList:
    - namf-comm
logger:
  enable: true
  level: info
`)
	_, err := ReadConfig(path)
	if err == nil {
		t.Fatal("expected invalid service")
	}
}
