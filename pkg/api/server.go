package api

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/keitahigaki/tfdrift-falco/pkg/api/broadcaster"
	"github.com/keitahigaki/tfdrift-falco/pkg/api/handlers"
	apimiddleware "github.com/keitahigaki/tfdrift-falco/pkg/api/middleware"
	"github.com/keitahigaki/tfdrift-falco/pkg/config"
	"github.com/keitahigaki/tfdrift-falco/pkg/detector"
	"github.com/keitahigaki/tfdrift-falco/pkg/graph"
	log "github.com/sirupsen/logrus"
)

// Server represents the API server
type Server struct {
	cfg         *config.Config
	detector    *detector.Detector
	broadcaster *broadcaster.Broadcaster
	graphStore  *graph.Store
	router      *chi.Mux
	port        int
	version     string
}

// NewServer creates a new API server
func NewServer(cfg *config.Config, det *detector.Detector, port int, version string) *Server {
	s := &Server{
		cfg:         cfg,
		detector:    det,
		broadcaster: broadcaster.NewBroadcaster(),
		graphStore:  graph.NewStore(),
		port:        port,
		version:     version,
	}

	s.setupRouter()

	// Populate sample data for testing
	s.graphStore.PopulateSampleData()
	log.Info("Populated sample graph data for testing")

	return s
}

// setupRouter configures all routes and middleware
func (s *Server) setupRouter() {
	r := chi.NewRouter()

	// Middleware
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(apimiddleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(apimiddleware.NewCORS().Handler)

	// Timeout for all requests
	r.Use(middleware.Timeout(60 * time.Second))

	// Health check (no /api/v1 prefix for simplicity)
	healthHandler := handlers.NewHealthHandler(s.version)
	r.Get("/health", healthHandler.GetHealth)

	// API v1 routes
	r.Route("/api/v1", func(r chi.Router) {
		// Health check also available under /api/v1
		r.Get("/health", healthHandler.GetHealth)

		// Graph endpoints
		graphHandler := handlers.NewGraphHandler(s.graphStore)
		r.Get("/graph", graphHandler.GetGraph)
		r.Get("/graph/nodes", graphHandler.GetNodes)
		r.Get("/graph/edges", graphHandler.GetEdges)

		// State endpoints (Phase 2)
		// r.Get("/state", stateHandler.GetState)
		// r.Get("/state/resources", stateHandler.GetResources)

		// Events endpoints (Phase 2)
		// r.Get("/events", eventsHandler.GetEvents)

		// Drifts endpoints (Phase 2)
		// r.Get("/drifts", driftsHandler.GetDrifts)

		// Stats endpoints (Phase 2)
		// r.Get("/stats", statsHandler.GetStats)

		// SSE endpoint (Phase 4)
		// r.Get("/stream", sseHandler.HandleSSE)
	})

	// WebSocket endpoint (Phase 3)
	// r.Get("/ws", wsHandler.HandleWebSocket)

	s.router = r
}

// Start starts the API server
func (s *Server) Start(ctx context.Context) error {
	addr := fmt.Sprintf(":%d", s.port)
	server := &http.Server{
		Addr:         addr,
		Handler:      s.router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	log.Infof("Starting API server on %s", addr)
	log.Infof("Health check: http://localhost%s/health", addr)
	log.Infof("API base URL: http://localhost%s/api/v1", addr)

	// Start the server in a goroutine
	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("API server error: %v", err)
		}
	}()

	// Wait for context cancellation
	<-ctx.Done()
	log.Info("Shutting down API server...")

	// Graceful shutdown with timeout
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Errorf("API server shutdown error: %v", err)
		return err
	}

	log.Info("API server stopped")
	return nil
}

// GetBroadcaster returns the broadcaster for detector integration
func (s *Server) GetBroadcaster() *broadcaster.Broadcaster {
	return s.broadcaster
}

// GetGraphStore returns the graph store for detector integration
func (s *Server) GetGraphStore() *graph.Store {
	return s.graphStore
}
