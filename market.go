package market

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/elastic/go-elasticsearch/v7"
	"github.com/go-redis/redis/v8"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/plifk/market/config"
	"github.com/plifk/market/internal/api"
	"github.com/plifk/market/internal/frontend"
	"github.com/plifk/market/internal/services"
)

// System of the market.
// Used both for running the HTTP servers and CLI tooling.
type System struct {
	Modules *services.Modules
	core    *services.Core

	httpServer *http.Server
	api        *api.Router
	frontend   *frontend.Router
	notFound   notFoundHandler
}

// Load the system, but don't start up the HTTP server.
// This also does not test or close connections.
func (s *System) Load(filename string) (err error) {
	settings, err := config.ReadFile(filename)
	if err != nil {
		return err
	}
	postgres, err := connectPostgres(context.Background(), settings.SQLDataSourceName)
	if err != nil {
		return fmt.Errorf("cannot establish a connection with a PostgreSQL server: %w", err)
	}
	kv := redis.NewClient(&redis.Options{
		Addr:     settings.RedisAddress,
		Username: settings.RedisUsername,
		Password: settings.RedisPassword,
	})
	elasticsearch, err := elasticsearchClient(settings)
	if err != nil {
		return fmt.Errorf("cannot configure Elasticsearch: %w", err)
	}
	s.core = &services.Core{
		Settings:       settings,
		Postgres:       postgres,
		Redis:          kv,
		Elasticsearch:  elasticsearch,
		CSRFProtection: csrfProtectionMiddleware(s),
	}
	s.Modules, err = services.NewModules(s.core)
	return err
}

func connectPostgres(ctx context.Context, connString string) (*pgxpool.Pool, error) {
	config, err := pgxpool.ParseConfig(connString)
	if err != nil {
		return nil, err
	}
	// Do not try to connect to database during initialization.
	// Otherwise, a failure will mean the market server won't go online before the database.
	config.LazyConnect = true
	return pgxpool.ConnectConfig(ctx, config)
}

// ListenAndServe listens on the TCP network address srv.Addr and then
// calls Serve to handle requests on incoming connections.
//
// It uses net/http.*Server.ListenAndServe and ListenAndServerTLS functions behind the scene.
// This should be stateless and register HTTP handlers.
func (s *System) ListenAndServe(ctx context.Context) (err error) {
	go s.checkSQL(ctx)
	go s.checkRedis(ctx)
	defer s.core.Postgres.Close()
	s.httpHandlers()
	go s.handleShutdown(ctx)
	settings := s.core.Settings
	if settings.HTTPCertFile == "" && settings.HTTPKeyFile == "" {
		log.Printf("listening to HTTP traffic on %q\n", settings.HTTPAddress)
		return s.httpServer.ListenAndServe()
	}
	log.Printf("listening to HTTPS traffic on %q\n", settings.HTTPAddress)
	return s.httpServer.ListenAndServeTLS(settings.HTTPCertFile, settings.HTTPKeyFile)
}

func (s *System) handleShutdown(ctx context.Context) {
	<-ctx.Done()
	fmt.Println()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	// Try to gracefully shutdown requests.
	// If long-lived connections such as WebSockets are used, it is necessary to
	// handle them apart with the RegisterOnShutdown function.
	// If long-lived requests are open, they will make the
	if err := s.httpServer.Shutdown(ctx); err != nil {
		log.Printf("cannot graceful shutdown HTTP server: %v\n", err)
		log.Println("Hanging until long-lived connections are closed or SIGKILL is received.")
		return
	}
	log.Println("graceful shutdown of HTTP server completed with success")
}

func (s *System) checkSQL(ctx context.Context) {
	pg := s.core.Postgres
	if _, err := pg.Exec(ctx, "SELECT 1"); err != nil {
		log.Printf("not connected to PostgreSQL yet: %v\n", err)
		return
	}
	stat := pg.Stat()
	log.Printf("total of PostgreSQL connections: %v\n", stat.TotalConns())
}

func (s *System) checkRedis(ctx context.Context) {
	ctx, cancel := context.WithTimeout(ctx, time.Second)
	defer cancel()
	kv := s.core.Redis
	status := kv.Ping(ctx)
	if err := status.Err(); err != nil {
		log.Printf("redis key-value storage: %v", err)
		return
	}
	log.Println("redis server:", status)
}

// elasticsearchClient instance.
func elasticsearchClient(s config.Settings) (*elasticsearch.Client, error) {
	cfg := elasticsearch.Config{
		Addresses: []string{s.ElasticsearchHost},
		Username:  s.ElasticsearchUsername,
		Password:  s.ElasticsearchPassword,
		APIKey:    s.ElasticsearchAPIKey,
	}
	if s.Debug {
		cfg.EnableMetrics = true
		cfg.EnableDebugLogger = true
		cfg.Transport = httpLogger().RoundTripper(cfg.Transport)
	}
	return elasticsearch.NewClient(cfg)
}

func csrfProtectionMiddleware(h http.Handler) *services.CSRFProtection {
	middleware := services.NewCSRFProtection(h)
	middleware.SetFailureHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, `400 Bad Request: service denied. Possible HTTP request forgery.
CSRF token not matching expected value. Try again.`, http.StatusBadRequest)
	}))
	return middleware
}
