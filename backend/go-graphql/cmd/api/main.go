package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/99designs/gqlgen/graphql"
	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/handler/extension"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/faizp/zenlist/backend/go-graphql/graph"
	"github.com/faizp/zenlist/backend/go-graphql/internal/config"
	"github.com/faizp/zenlist/backend/go-graphql/internal/db"
	"github.com/faizp/zenlist/backend/go-graphql/internal/db/repo"
	"github.com/faizp/zenlist/backend/go-graphql/internal/graphql/middleware"
	platformlogger "github.com/faizp/zenlist/backend/go-graphql/internal/platform/logger"
	"github.com/faizp/zenlist/backend/go-graphql/internal/service"
	"github.com/joho/godotenv"
	"github.com/vektah/gqlparser/v2/gqlerror"
)

func main() {
	_ = godotenv.Load()

	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "config error: %v\n", err)
		os.Exit(1)
	}

	log := platformlogger.New(cfg.AppEnv)
	ctx := context.Background()

	pool, err := db.NewPool(ctx, db.PoolConfig{
		URL:               cfg.DatabaseURL,
		MaxConns:          cfg.DBMaxConns,
		MinConns:          cfg.DBMinConns,
		HealthCheckPeriod: cfg.DBHealthCheckEvery,
	})
	if err != nil {
		log.Error("db_connect_failed", "error", err)
		os.Exit(1)
	}
	defer pool.Close()

	store := repo.New(pool)
	svc := service.New(store, cfg)
	if err := svc.Bootstrap(ctx); err != nil {
		log.Error("bootstrap_failed", "error", err)
		os.Exit(1)
	}

	resolver := &graph.Resolver{Service: svc}
	srv := handler.New(graph.NewExecutableSchema(graph.Config{Resolvers: resolver}))
	srv.Use(extension.FixedComplexityLimit(200))
	srv.SetErrorPresenter(func(ctx context.Context, err error) *gqlerror.Error {
		presented := graphql.DefaultErrorPresenter(ctx, err)
		if presented.Extensions == nil {
			presented.Extensions = make(map[string]interface{})
		}
		if _, ok := presented.Extensions["code"]; !ok {
			presented.Extensions["code"] = "INTERNAL"
		}
		if reqID := middleware.RequestIDFromContext(ctx); reqID != "" {
			presented.Extensions["request_id"] = reqID
		}
		return presented
	})

	mux := http.NewServeMux()
	mux.Handle("/", playground.Handler("ZenList GraphQL", "/query"))
	mux.Handle("/query", chain(
		srv,
		middleware.Timeout(cfg.RequestTimeout),
		middleware.RequestID,
		middleware.Logging(log),
	))
	mux.Handle("/healthz", healthHandler(pool, log))

	httpServer := &http.Server{
		Addr:              ":" + strings.TrimSpace(cfg.HTTPPort),
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      cfg.RequestTimeout + 2*time.Second,
		IdleTimeout:       60 * time.Second,
	}

	go func() {
		log.Info("server_starting", "addr", httpServer.Addr)
		if err := httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Error("server_failed", "error", err)
			os.Exit(1)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGTERM, syscall.SIGINT)
	<-stop

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		log.Error("server_shutdown_failed", "error", err)
	}
	log.Info("server_stopped")
}

func chain(h http.Handler, middlewares ...func(http.Handler) http.Handler) http.Handler {
	for i := len(middlewares) - 1; i >= 0; i-- {
		h = middlewares[i](h)
	}
	return h
}

func healthHandler(pool interface{ Ping(context.Context) error }, logger *platformlogger.Logger) http.Handler {
	type response struct {
		Status string `json:"status"`
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
		defer cancel()

		if err := pool.Ping(ctx); err != nil {
			logger.Error("health_check_failed", "error", err)
			w.WriteHeader(http.StatusServiceUnavailable)
			_ = json.NewEncoder(w).Encode(response{Status: "unhealthy"})
			return
		}

		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(response{Status: "ok"})
	})
}
