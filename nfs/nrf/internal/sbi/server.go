package sbi

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/vlong2000/trea-one/nfs/nrf/internal/store"
	"github.com/vlong2000/trea-one/nfs/nrf/pkg/factory"
)

type Server struct {
	cfg    *factory.Config
	store  *store.Store
	http   *http.Server
	logger *log.Logger
}

func New(cfg *factory.Config, st *store.Store, logger *log.Logger) *Server {
	if logger == nil {
		logger = log.Default()
	}
	s := &Server{cfg: cfg, store: st, logger: logger}
	mux := http.NewServeMux()
	mux.HandleFunc("GET /nnrf-nfm/v1/", s.handleNfmIndex)
	mux.HandleFunc("PUT /nnrf-nfm/v1/nf-instances/{nfInstanceID}", s.handleRegister)
	mux.HandleFunc("GET /nnrf-nfm/v1/nf-instances/{nfInstanceID}", s.handleGetInstance)
	mux.HandleFunc("DELETE /nnrf-nfm/v1/nf-instances/{nfInstanceID}", s.handleDeregister)
	mux.HandleFunc("GET /nnrf-nfm/v1/nf-instances", s.handleListInstances)
	mux.HandleFunc("GET /nnrf-disc/v1/", s.handleDiscIndex)
	mux.HandleFunc("GET /nnrf-disc/v1/nf-instances", s.handleDiscover)
	mux.HandleFunc("GET /bootstrapping", s.handleBootstrapping)
	s.http = &http.Server{
		Addr:              cfg.BindingAddr(),
		Handler:           logging(logger, mux),
		ReadHeaderTimeout: 10 * time.Second,
	}
	return s
}

func (s *Server) ListenAndServe() error {
	ln, err := net.Listen("tcp", s.http.Addr)
	if err != nil {
		return fmt.Errorf("nrf: listen %s: %w", s.http.Addr, err)
	}
	s.logger.Printf("NRF SBI listening on %s", s.http.Addr)
	if err := s.http.Serve(ln); err != nil && err != http.ErrServerClosed {
		return err
	}
	return nil
}

func (s *Server) Shutdown(ctx context.Context) error {
	return s.http.Shutdown(ctx)
}

func (s *Server) Addr() string {
	return s.http.Addr
}

func (s *Server) handleNfmIndex(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, "nrf")
}

func (s *Server) handleDiscIndex(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, "nrf")
}

func (s *Server) handleBootstrapping(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{
		"status":            "OPERATIVE",
		"nrfInstanceId":     "standalone-nrf",
		"oauth2Required":    false,
		"serviceNameList":   s.cfg.Configuration.ServiceNameList,
		"nrfServiceFeatures": map[string]any{},
	})
}

func (s *Server) handleRegister(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("nfInstanceID")
	if id == "" {
		writeProblem(w, http.StatusBadRequest, "MANDATORY_IE_MISSING", "nfInstanceID is required")
		return
	}
	body, err := io.ReadAll(io.LimitReader(r.Body, 1<<20))
	if err != nil {
		writeProblem(w, http.StatusBadRequest, "UNSPECIFIED", "read body failed")
		return
	}
	profile := store.Profile{}
	if len(body) > 0 {
		if err := json.Unmarshal(body, &profile); err != nil {
			writeProblem(w, http.StatusBadRequest, "UNSPECIFIED", "invalid json")
			return
		}
	}
	if asString(profile["nfType"]) == "" || asString(profile["nfStatus"]) == "" {
		writeProblem(w, http.StatusBadRequest, "MANDATORY_IE_MISSING", "nfType and nfStatus are required")
		return
	}
	if pid := asString(profile["nfInstanceId"]); pid != "" && pid != id {
		writeProblem(w, http.StatusBadRequest, "UNSPECIFIED", "nfInstanceId does not match path")
		return
	}
	profile["nfInstanceId"] = id
	existed := s.store.Put(id, profile)
	location := s.cfg.SbiURI() + factory.NfmPrefix + "/nf-instances/" + id
	w.Header().Set("Location", location)
	if existed {
		writeJSON(w, http.StatusOK, profile)
		return
	}
	writeJSON(w, http.StatusCreated, profile)
}

func (s *Server) handleGetInstance(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("nfInstanceID")
	p, ok := s.store.Get(id)
	if !ok {
		writeProblem(w, http.StatusNotFound, "RESOURCE_URI_STRUCTURE_NOT_FOUND", "NFProfile not found")
		return
	}
	writeJSON(w, http.StatusOK, p)
}

func (s *Server) handleDeregister(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("nfInstanceID")
	if !s.store.Delete(id) {
		writeProblem(w, http.StatusNotFound, "RESOURCE_URI_STRUCTURE_NOT_FOUND", "NFProfile not found")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) handleListInstances(w http.ResponseWriter, r *http.Request) {
	nfType := r.URL.Query().Get("nf-type")
	items := []map[string]string{}
	for _, p := range s.store.List(nfType) {
		id := asString(p["nfInstanceId"])
		items = append(items, map[string]string{
			"href": s.cfg.SbiURI() + factory.NfmPrefix + "/nf-instances/" + id,
		})
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"_links": map[string]any{"item": items},
		"nfType": nfType,
	})
}

func (s *Server) handleDiscover(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	target := q.Get("target-nf-type")
	requester := q.Get("requester-nf-type")
	if target == "" || requester == "" {
		writeProblem(w, http.StatusBadRequest, "MANDATORY_IE_MISSING", "target-nf-type and requester-nf-type are required")
		return
	}
	if !validNfType(target) || !validNfType(requester) {
		writeProblem(w, http.StatusBadRequest, "INVALID_QUERY_PARAM", "unknown nf type")
		return
	}
	instances := s.store.Discover(target, requester)
	writeJSON(w, http.StatusOK, map[string]any{
		"validityPeriod": 100,
		"nfInstances":    instances,
	})
}

func validNfType(t string) bool {
	switch strings.ToUpper(t) {
	case "NRF", "UDM", "AMF", "SMF", "AUSF", "NEF", "PCF", "SMSF", "NSSF", "UDR",
		"LMF", "GMLC", "5G_EIR", "SEPP", "UPF", "N3IWF", "AF", "UDSF", "BSF", "CHF", "NWDAF":
		return true
	default:
		return false
	}
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeProblem(w http.ResponseWriter, status int, cause, detail string) {
	writeJSON(w, status, map[string]any{
		"status": status,
		"cause":  cause,
		"detail": detail,
		"title":  http.StatusText(status),
	})
}

func asString(v any) string {
	s, _ := v.(string)
	return s
}

func logging(logger *log.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logger.Printf("%s %s", r.Method, r.URL.Path)
		next.ServeHTTP(w, r)
	})
}
