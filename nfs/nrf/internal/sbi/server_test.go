package sbi

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/vlong2000/trea-one/nfs/nrf/internal/store"
	"github.com/vlong2000/trea-one/nfs/nrf/pkg/factory"
)

func testServer() *Server {
	cfg := &factory.Config{
		Info: &factory.Info{Version: "1.0.2"},
		Configuration: &factory.Configuration{
			Sbi: &factory.Sbi{
				Scheme:       "http",
				RegisterIPv4: "127.0.0.1",
				BindingIPv4:  "127.0.0.1",
				Port:         8000,
			},
			ServiceNameList: []string{"nnrf-nfm", "nnrf-disc"},
		},
		Logger: &factory.Logger{Enable: true, Level: "info"},
	}
	return New(cfg, store.New(), log.New(io.Discard, "", 0))
}

func doJSON(t *testing.T, h http.Handler, method, path string, body any) *httptest.ResponseRecorder {
	t.Helper()
	var rdr io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			t.Fatal(err)
		}
		rdr = bytes.NewReader(b)
	}
	req := httptest.NewRequest(method, path, rdr)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)
	return rr
}

func TestRegisterGetDiscoverDeregister(t *testing.T) {
	s := testServer()
	h := s.http.Handler
	profile := map[string]any{
		"nfInstanceId": "amf-1",
		"nfType":       "AMF",
		"nfStatus":     "REGISTERED",
	}
	rr := doJSON(t, h, http.MethodPut, "/nnrf-nfm/v1/nf-instances/amf-1", profile)
	if rr.Code != http.StatusCreated {
		t.Fatalf("register code=%d body=%s", rr.Code, rr.Body.String())
	}
	if loc := rr.Header().Get("Location"); loc == "" {
		t.Fatal("missing Location")
	}

	rr = doJSON(t, h, http.MethodPut, "/nnrf-nfm/v1/nf-instances/amf-1", profile)
	if rr.Code != http.StatusOK {
		t.Fatalf("update code=%d", rr.Code)
	}

	rr = doJSON(t, h, http.MethodGet, "/nnrf-nfm/v1/nf-instances/amf-1", nil)
	if rr.Code != http.StatusOK {
		t.Fatalf("get code=%d", rr.Code)
	}

	rr = doJSON(t, h, http.MethodGet, "/nnrf-disc/v1/nf-instances?target-nf-type=AMF&requester-nf-type=SMF", nil)
	if rr.Code != http.StatusOK {
		t.Fatalf("discover code=%d body=%s", rr.Code, rr.Body.String())
	}
	var result struct {
		NfInstances []map[string]any `json:"nfInstances"`
	}
	if err := json.Unmarshal(rr.Body.Bytes(), &result); err != nil {
		t.Fatal(err)
	}
	if len(result.NfInstances) != 1 {
		t.Fatalf("discover count=%d", len(result.NfInstances))
	}

	rr = doJSON(t, h, http.MethodDelete, "/nnrf-nfm/v1/nf-instances/amf-1", nil)
	if rr.Code != http.StatusNoContent {
		t.Fatalf("deregister code=%d", rr.Code)
	}
	rr = doJSON(t, h, http.MethodGet, "/nnrf-nfm/v1/nf-instances/amf-1", nil)
	if rr.Code != http.StatusNotFound {
		t.Fatalf("get after delete code=%d", rr.Code)
	}
}

func TestRegisterMissingFields(t *testing.T) {
	s := testServer()
	rr := doJSON(t, s.http.Handler, http.MethodPut, "/nnrf-nfm/v1/nf-instances/x", map[string]any{"nfType": "AMF"})
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("code=%d", rr.Code)
	}
}

func TestDiscoverMissingQuery(t *testing.T) {
	s := testServer()
	rr := doJSON(t, s.http.Handler, http.MethodGet, "/nnrf-disc/v1/nf-instances", nil)
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("code=%d", rr.Code)
	}
}

func TestBootstrapping(t *testing.T) {
	s := testServer()
	rr := doJSON(t, s.http.Handler, http.MethodGet, "/bootstrapping", nil)
	if rr.Code != http.StatusOK {
		t.Fatalf("code=%d", rr.Code)
	}
}
