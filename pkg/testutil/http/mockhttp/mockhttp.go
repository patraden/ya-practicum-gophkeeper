package mockhttp

import (
	"context"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

const (
	ContentType     = "Content-Type"
	ContentTypeText = "text/plain"
	ContentTypeJSON = "application/json"
)

// MockHTTPServer represents a TLS-enabled HTTP server.
// It wraps http.Server and holds TLS certificate paths.
// Designed primarily for testing or mock purposes but can serve real traffic.
type Server struct {
	httpServer  *http.Server
	ServerAddr  string // ServerAddr is the address the server listens on (e.g. ":8443").
	tlsCertPath string // tlsCertPath is the file path to the TLS certificate.
	tlskeyPath  string // tlskeyPath is the file path to the TLS private key.
}

// NewServer creates a new MockHTTPServer with the given address, TLS certificate, key paths, and HTTP handler.
// The handler typically is a router that registers routes and middleware.
func NewServer(serverAddr, tlsCertPath, tlskeyPath string, handler http.Handler) *Server {
	return &Server{
		httpServer: &http.Server{
			Addr:              serverAddr,
			Handler:           handler,
			ReadHeaderTimeout: time.Second,
		},
		tlsCertPath: tlsCertPath,
		tlskeyPath:  tlskeyPath,
		ServerAddr:  serverAddr,
	}
}

// Run starts the HTTPS server.
// It blocks until the server exits or encounters an error.
//
//nolint:wrapcheck // reason : ignoring error wrap in testutils package.
func (s *Server) Run() error {
	return s.httpServer.ListenAndServeTLS(s.tlsCertPath, s.tlskeyPath)
}

// Shutdown gracefully shuts down the HTTP server without interrupting active connections.
// It accepts a context to set timeout/deadline for the shutdown.
//
//nolint:wrapcheck // reason : ignoring error wrap in testutils package.
func (s *Server) Shutdown(ctx context.Context) error {
	return s.httpServer.Shutdown(ctx)
}

// Handler defines an interface for registering routes on a chi.Router.
// Implementations should attach routes and middlewares within RegisterRoutes.
type Handler interface {
	RegisterRoutes(router chi.Router)
}

// NewMockedRouter creates a new chi router with standard middleware (request ID, logging, recovery).
func NewMockedRouter(handlers ...Handler) http.Handler {
	router := chi.NewRouter()

	// Standard middlewares
	router.Use(middleware.RequestID) // Attach unique request IDs to each request
	router.Use(middleware.Logger)    // Log HTTP requests via chi's default logger
	router.Use(middleware.Recoverer) // Recover from panics gracefully and respond with 500

	// Register application-specific handlers
	for _, h := range handlers {
		h.RegisterRoutes(router)
	}

	// Custom 404 Not Found handler responds with 400 Bad Request and message
	router.NotFound(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "path not found", http.StatusBadRequest)
	}))

	return router
}
