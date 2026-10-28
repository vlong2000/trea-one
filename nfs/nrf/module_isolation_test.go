package nrf_test

import (
	"os"
	"strings"
	"testing"
)

func TestGoModDoesNotDependOnOtherNFs(t *testing.T) {
	data, err := os.ReadFile("go.mod")
	if err != nil {
		t.Fatal(err)
	}
	text := string(data)
	blocked := []string{
		"nfs/amf", "nfs/smf", "nfs/upf", "nfs/ausf", "nfs/udm",
		"nfs/udr", "nfs/pcf", "nfs/nssf", "github.com/free5gc/amf",
		"github.com/free5gc/smf",
	}
	for _, item := range blocked {
		if strings.Contains(text, item) {
			t.Fatalf("nrf go.mod depends on other NF: %s", item)
		}
	}
	if !strings.Contains(text, "module github.com/vlong2000/trea-one/nfs/nrf") {
		t.Fatal("unexpected module path")
	}
}
