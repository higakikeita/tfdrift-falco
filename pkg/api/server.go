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
	"github.com/keitahigaki/tfdrift-falco/pkg/api/sse"
	"github.com/keitahigaki/tfdrift-falco/pkg/api/websocket"
	"github.com/keitahigaki/tfdrift-falco/pkg/config"
	"github.com/keitahigaki/tfdrift-falco/pkg/detector"
	"github.com/keitahigaki/tfdrift-falco/pkg/graph"
	"github.com/keitahigaki/tfdrift-falco/pkg/terraform"
	log "github.com/sirupsen/logrus"
)

// Server represents the API server
type Server struct {
	cfg          *config.Config
	detector     *detector.Detector
	broadcaster  *broadcaster.Broadcaster
	graphStore   *graph.Store
	stateManager *terraform.StateManager
	wsHandler    *websocket.Handler
	sseHandler   *sse.Handler
	authMw       *apimiddleware.Auth
	rateLimiter  *apimiddleware.RateLimiter
	router       *chi.Mux
	port         int
	version      string
}

// NewServer creates a new API server
func NewServer(cfg *config.Config, det *detector.Detector, port int, version string) *Server {
	// Create broadcaster
	bc := broadcaster.NewBroadcaster()

	// Create WebSocket handler
	wsHandler := websocket.NewHandler(bc)

	// Create SSE handler
	sseHandler := sse.NewHandler(bc)

	// Create auth middleware
	authMw := apimiddleware.NewAuth(apimiddleware.AuthConfig{
		Enabled:   cfg.API.Auth.Enabled,
		JWTSecret: cfg.API.Auth.JWTSecret,
		JWTIssuer: cfg.API.Auth.JWTIssuer,
		JWTExpiry: cfg.API.Auth.JWTExpiry,
		APIKeys:   convertAPIKeys(cfg.API.Auth.APIKeys),
	})

	// Create rate limiter
	rateLimiter := apimiddleware.NewRateLimiter(apimiddleware.RateLimitConfig{
		Enabled:         cfg.API.RateLimit.Enabled,
		RequestsPerMin:  cfg.API.RateLimit.RequestsPerMin,
		BurstSize:       cfg.API.RateLimit.BurstSize,
		CleanupInterval: cfg.API.RateLimit.CleanupInterval,
	})

	s := &Server{
		cfg:          cfg,
		detector:     det,
		broadcaster:  bc,
		graphStore:   graph.NewStore(),
		stateManager: det.GetStateManager(),
		wsHandler:    wsHandler,
		sseHandler:   sseHandler,
		authMw:       authMw,
		rateLimiter:  rateLimiter,
		port:         port,
		version:      version,
	}

	s.setupRouter()

	// Connect detector to broadcaster and graph store
	det.SetBroadcaster(bc)
	det.SetGraphStore(s.graphStore)
	log.Info("Connected detector to broadcaster and graph store")

	// Connect StateManager to GraphStore for Terraform State-based graph building
	if s.stateManager != nil {
		s.graphStore.SetStateManager(s.stateManager)
		log.Info("Connected Terraform StateManager to GraphStore")
	}

	// Note: Graph will be populated automatically with Terraform State resources
	// and drift events will be overlaid as they are detected

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
	r.Use(s.rateLimiter.Middleware)

	// Health check (no /api/v1 prefix for simplicity)
	healthHandler := handlers.NewHealthHandler(s.version)
	r.Get("/health", healthHandler.GetHealth)

	// API v1 routes
	r.Route("/api/v1", func(r chi.Router) {
		// Public endpoints (no auth required)
		r.Get("/health", healthHandler.GetHealth)
		r.Get("/version", handlers.NewConfigHandler(s.cfg).GetVersion)

		// SSE endpoint - no timeout for streaming, auth applied
		r.Group(func(r chi.Router) {
			r.Use(s.authMw.Middleware)
			r.Get("/stream", s.sseHandler.HandleSSE)
		})

		// Protected API routes with auth + timeout
		r.Group(func(r chi.Router) {
			r.Use(s.authMw.Middleware)
			// Timeout for non-streaming API routes
			r.Use(middleware.Timeout(60 * time.Second))

			// Auth management endpoints
			authHandler := handlers.NewAuthHandler(s.authMw)
			r.Post("/auth/token", authHandler.GenerateToken)
			r.Get("/auth/api-keys", authHandler.ListAPIKeys)
			r.Post("/auth/api-keys", authHandler.CreateAPIKey)
			r.Delete("/auth/api-keys", authHandler.RevokeAPIKey)

			// Graph endpoints
			graphHandler := handlers.NewGraphHandler(s.graphStore)
			r.Get("/graph", graphHandler.GetGraph)
			r.Get("/graph/nodes", graphHandler.GetNodes)
			r.Get("/graph/edges", graphHandler.GetEdges)

			// Graph Query endpoints (Neo4j-style)
			graphQueryHandler := handlers.NewGraphQueryHandler(s.graphStore)
			r.Get("/graph/nodes/{id}", graphQueryHandler.GetNode)
			r.Get("/graph/path", graphQueryHandler.GetPath)
			r.Get("/graph/impact/{id}", graphQueryHandler.GetImpactRadius)
			r.Get("/graph/dependencies/{id}", graphQueryHandler.GetDependencies)
			r.Get("/graph/dependents/{id}", graphQueryHandler.GetDependents)
			r.Get("/graph/critical", graphQueryHandler.GetCriticalNodes)
			r.Get("/graph/neighbors/{id}", graphQueryHandler.GetNeighbors)
			r.Get("/graph/relationships/{id}", graphQueryHandler.GetRelationships)
			r.Get("/graph/stats", graphQueryHandler.GetGraphStats)
			r.Post("/graph/match", graphQueryHandler.MatchPattern)

			// State endpoints (Phase 2)
			stateHandler := handlers.NewStateHandler(s.stateManager)
			r.Get("/state", stateHandler.GetState)
			r.Get("/state/resources", stateHandler.GetResources)
			r.Get("/state/resource/{id}", stateHandler.GetResource)

			// Events endpoints (Phase 2)
			eventsHandler := handlers.NewEventsHandler(s.graphStore)
			r.Get("/events", eventsHandler.GetEvents)
			r.Get("/events/{id}", eventsHandler.GetEvent)
			r.Patch("/events/{id}", eventsHandler.UpdateEventStatus)

			// Drifts endpoints (Phase 2)
			driftsHandler := handlers.NewDriftsHandler(s.graphStore)
			r.Get("/drifts", driftsHandler.GetDrifts)
			r.Get("/drifts/{id}", driftsHandler.GetDrift)

			// Stats endpoints (Phase 2)
			statsHandler := handlers.NewStatsHandler(s.graphStore)
			r.Get("/stats", statsHandler.GetStats)

			// Analytics endpoints (v0.6.0 - Dashboard)
			analyticsHandler := handlers.NewAnalyticsHandler(s.graphStore)
			r.Get("/analytics/summary", analyticsHandler.GetSummary)
			r.Get("/analytics/timeline", analyticsHandler.GetTimeline)
			r.Get("/analytics/breakdown", analyticsHandler.GetBreakdown)

			// Config endpoints (v0.6.0 - Dashboard)
			configHandler := handlers.NewConfigHandler(s.cfg)
			r.Get("/config", configHandler.GetConfig)
			r.Post("/config/webhooks/test", configHandler.TestWebhook)

			// Discovery endpoints (AWS resource discovery and drift detection)
			discoveryHandler := handlers.NewDiscoveryHandler(s.stateManager)
			r.Get("/discovery/scan", discoveryHandler.DiscoverAWSResources)
			r.Get("/discovery/drift", discoveryHandler.DetectDrift)
			r.Get("/discovery/drift/summary", discoveryHandler.GetDriftSummary)
		})
	})

	// WebSocket endpoint (Phase 3)
	r.Get("/ws", s.wsHandler.HandleWebSocket)

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
	log.Infof("WebSocket URL: ws://localhost%s/ws", addr)
	log.Infof("SSE Stream URL: http://localhost%s/api/v1/stream", addr)

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

// convertAPIKeys converts config API key entries to middleware API key entries.
func convertAPIKeys(entries []config.APIKeyEntry) []apimiddleware.APIKeyEntry {
	result := make([]apimiddleware.APIKeyEntry, len(entries))
	for i, e := range entries {
		result[i] = apimiddleware.APIKeyEntry{
			Name:      e.Name,
			Key:       e.Key,
			Scopes:    e.Scopes,
			CreatedAt: e.CreatedAt,
		}
	}
	return result
}
