package server

import (
	"context"
	"errors"
	"log/slog"
	"marketplace/internal/config"
	healthhandler "marketplace/internal/http/handlers/health"
	postshandler "marketplace/internal/http/handlers/posts"
	sessionhandler "marketplace/internal/http/handlers/session"
	userhandler "marketplace/internal/http/handlers/user"
	"marketplace/internal/http/middleware"
	"marketplace/internal/models"
	utils "marketplace/internal/utils/http_errors"
	"net/http"
	"time"

	"github.com/gorilla/mux"
)

func StartServer(
	ctx context.Context,
	cfg *config.HTTPServer,
	log *slog.Logger,
	authService AuthService,
	postService PostService,
) error {
	r := mux.NewRouter()

	r.Use(middleware.Logger(log))
	r.Use(middleware.AuthOptional(log, authService))

	setupRoutes(r, log, authService, postService)

	srv := &http.Server{
		Addr:         cfg.Address,
		ReadTimeout:  cfg.Timeout,
		WriteTimeout: cfg.Timeout,
		IdleTimeout:  cfg.IdleTimeout,
		Handler:      r,
	}

	errChan := make(chan error, 1)

	go func() {
		log.Info("server started", slog.String("address", cfg.Address))
		if err := srv.ListenAndServe(); err != nil {
			if errors.Is(err, http.ErrServerClosed) {
				log.Info("server closed gracefully")
			} else {
				log.Error("could not start server:", "error", err)
				errChan <- err
			}
		}
	}()
	select {
	case <-ctx.Done():
		log.Info("shutting down server...")

		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if err := srv.Shutdown(shutdownCtx); err != nil {
			log.Error("error shutting down server", "error", err)
			return err
		}
		log.Info("server exited gracefully")
		return nil
	case err := <-errChan:
		return err
	}

}

func setupRoutes(r *mux.Router, log *slog.Logger, auth AuthService, post PostService) {

	// POST user
	r.HandleFunc("/api/register", func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		userhandler.Add(ctx, log, w, r, auth)
	}).Methods(http.MethodPost)

	// POST session
	r.HandleFunc("/api/auth", func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		sessionhandler.Add(ctx, log, w, r, auth)
	}).Methods(http.MethodPost)

	// DELETE session
	r.HandleFunc("/api/auth", func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		sessionhandler.Delete(ctx, log, w, r, auth)
	}).Methods(http.MethodDelete)

	// GET posts
	r.HandleFunc("/api/posts", func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		postshandler.Get(ctx, log, w, r, post)
	}).Methods(http.MethodGet)

	// HEAD posts
	r.HandleFunc("/api/posts", func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		postshandler.Head(ctx, log, w, r, post)
	}).Methods(http.MethodHead)

	// GET health
	r.HandleFunc("/api/health", func(w http.ResponseWriter, r *http.Request) {
		healthhandler.Get(w, r)
	}).Methods(http.MethodGet)

	requiredAuth := r.NewRoute().Subrouter()
	requiredAuth.Use(middleware.AuthRequired(log, auth))

	// POST posts
	requiredAuth.HandleFunc("/api/posts", func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		postshandler.Add(ctx, log, w, r, post)
	}).Methods(http.MethodPost)

	// Not allowed
	r.MethodNotAllowedHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		utils.WriteJSONError(w, http.StatusMethodNotAllowed, models.ErrMethodNotAllowed.Error())
	})
}
