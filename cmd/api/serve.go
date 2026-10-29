package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"time"
)

// serve sobe o servidor HTTP e bloqueia até ctx ser cancelado (SIGTERM no
// Kubernetes) ou o servidor falhar. No cancelamento, para de aceitar novas
// conexões e espera as requisições em andamento por até shutdownTimeout.
func serve(ctx context.Context, srv *http.Server, shutdownTimeout time.Duration) error {
	errCh := make(chan error, 1)
	go func() {
		errCh <- srv.ListenAndServe()
	}()

	select {
	case err := <-errCh:
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}
		return err
	case <-ctx.Done():
	}

	log.Printf("shutting down server (timeout %s)\n", shutdownTimeout)
	shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	return srv.Shutdown(shutdownCtx)
}
