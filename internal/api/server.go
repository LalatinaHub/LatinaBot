package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/LalatinaHub/LatinaBot/internal/service/proxycheck"
	"github.com/LalatinaHub/LatinaBot/pkg/logger"
)

// Server represents the internal HTTP server.
type Server struct {
	httpServer   *http.Server
	proxyChecker *proxycheck.Checker
}

// NewServer creates a new HTTP server.
func NewServer(addr string, proxyChecker *proxycheck.Checker) *Server {
	s := &Server{
		proxyChecker: proxyChecker,
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/api/v1/proxy/check", s.handleProxyCheck)
	mux.HandleFunc("/", s.handleRoot)

	s.httpServer = &http.Server{
		Addr:              addr,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	return s
}

// Handler returns the HTTP handler for testing.
func (s *Server) Handler() http.Handler {
	return s.httpServer.Handler
}

// Start runs the HTTP server.
func (s *Server) Start() error {
	logger.Info().Str("addr", s.httpServer.Addr).Msg("HTTP server is listening")
	return s.httpServer.ListenAndServe()
}

// Shutdown stops the HTTP server gracefully.
func (s *Server) Shutdown(ctx context.Context) error {
	return s.httpServer.Shutdown(ctx)
}

func (s *Server) handleProxyCheck(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	proxyIP := r.URL.Query().Get("ip")
	if proxyIP == "" {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{}`))
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	result := s.proxyChecker.CheckIP(ctx, proxyIP)

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(result); err != nil {
		http.Error(w, fmt.Sprintf("Error %s", err.Error()), http.StatusInternalServerError)
	}
}

func (s *Server) handleRoot(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("Welcome to LatinaBot (Go 1.24+)!"))
}
