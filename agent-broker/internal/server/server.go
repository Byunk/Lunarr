package server

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

// Server wraps http.Server with graceful shutdown and request logging.
type Server struct {
	// httpServer is the underlying HTTP server.
	httpServer *http.Server
	// logger is the structured logger for server events.
	logger *slog.Logger
	// opts holds the server configuration.
	opts Options
}

// Options configures the Server.
type Options struct {
	// Port is the HTTP server port.
	Port int
	// Logger is the structured logger for request logging.
	Logger *slog.Logger
	// ReadTimeout is the max duration for reading the entire request.
	ReadTimeout time.Duration
	// WriteTimeout is the max duration before timing out writes.
	WriteTimeout time.Duration
	// IdleTimeout is the max time to wait for the next request.
	IdleTimeout time.Duration
	// ShutdownTimeout is the max duration for graceful shutdown.
	ShutdownTimeout time.Duration
}

// DefaultOptions returns Options with sensible defaults.
func DefaultOptions() Options {
	return Options{
		Port:            8080,
		Logger:          slog.Default(),
		ReadTimeout:     15 * time.Second,
		WriteTimeout:    15 * time.Second,
		IdleTimeout:     60 * time.Second,
		ShutdownTimeout: 30 * time.Second,
	}
}

// Option is a functional option for configuring Server.
type Option func(*Options)

// WithPort sets the server port.
func WithPort(port int) Option {
	return func(o *Options) {
		o.Port = port
	}
}

// WithLogger sets the structured logger.
func WithLogger(logger *slog.Logger) Option {
	return func(o *Options) {
		o.Logger = logger
	}
}

// WithReadTimeout sets the read timeout.
func WithReadTimeout(d time.Duration) Option {
	return func(o *Options) {
		o.ReadTimeout = d
	}
}

// WithWriteTimeout sets the write timeout.
func WithWriteTimeout(d time.Duration) Option {
	return func(o *Options) {
		o.WriteTimeout = d
	}
}

// WithIdleTimeout sets the idle timeout.
func WithIdleTimeout(d time.Duration) Option {
	return func(o *Options) {
		o.IdleTimeout = d
	}
}

// WithShutdownTimeout sets the graceful shutdown timeout.
func WithShutdownTimeout(d time.Duration) Option {
	return func(o *Options) {
		o.ShutdownTimeout = d
	}
}

// New creates a Server with the given handler and options.
func New(handler http.Handler, opts ...Option) *Server {
	options := DefaultOptions()
	for _, opt := range opts {
		opt(&options)
	}

	return &Server{
		httpServer: &http.Server{
			Addr:         fmt.Sprintf(":%d", options.Port),
			Handler:      loggingMiddleware(options.Logger)(handler),
			ReadTimeout:  options.ReadTimeout,
			WriteTimeout: options.WriteTimeout,
			IdleTimeout:  options.IdleTimeout,
		},
		logger: options.Logger,
		opts:   options,
	}
}

// Run starts the server and blocks until shutdown signal or context cancellation.
func (s *Server) Run(ctx context.Context) error {
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, syscall.SIGINT, syscall.SIGTERM)

	serverErr := make(chan error, 1)

	go func() {
		s.logger.Info("starting server", "addr", s.httpServer.Addr)
		if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			serverErr <- err
		}
	}()

	select {
	case err := <-serverErr:
		return fmt.Errorf("server error: %w", err)
	case sig := <-shutdown:
		s.logger.Info("shutdown signal received", "signal", sig.String())
	case <-ctx.Done():
		s.logger.Info("context cancelled")
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), s.opts.ShutdownTimeout)
	defer cancel()

	s.logger.Info("shutting down server")
	if err := s.httpServer.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("shutdown error: %w", err)
	}

	s.logger.Info("server stopped")
	return nil
}

func loggingMiddleware(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			wrapped := &responseWriter{ResponseWriter: w, status: http.StatusOK}
			next.ServeHTTP(wrapped, r)

			logger.Info("request",
				"method", r.Method,
				"path", r.URL.Path,
				"status", wrapped.status,
				"duration_ms", time.Since(start).Milliseconds(),
				"remote_addr", r.RemoteAddr,
			)
		})
	}
}

type responseWriter struct {
	http.ResponseWriter
	// status is the HTTP status code written to the response.
	status int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.status = code
	rw.ResponseWriter.WriteHeader(code)
}
