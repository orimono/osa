package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/orimono/osa/internal/api"
	"github.com/orimono/osa/internal/config"
)

func main() {
	cfg := config.MustLoad()

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	nc, err := nats.Connect(cfg.NatsURL)
	if err != nil {
		slog.Error("failed to connect to nats", "err", err)
		os.Exit(1)
	}
	defer nc.Drain()

	timeout := time.Duration(cfg.RequestTimeout)
	if timeout == 0 {
		timeout = 5 * time.Second
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/api/nodes", api.NodesHandler(nc, timeout))
	mux.HandleFunc("/api/history", api.HistoryHandler(nc, timeout))
	mux.HandleFunc("/api/stream", api.StreamHandler(nc))
	mux.HandleFunc("/api/nodes/{nodeId}/executors", api.ExecutorsHandler(nc, timeout))
	mux.HandleFunc("POST /api/nodes/{nodeId}/executors", api.RegisterExecutorHandler(nc, timeout))

	srv := &http.Server{Addr: cfg.Addr, Handler: mux}

	go func() {
		<-ctx.Done()
		srv.Shutdown(context.Background())
	}()

	slog.Info("osa listening", "addr", cfg.Addr)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		slog.Error("server error", "err", err)
		os.Exit(1)
	}
}
