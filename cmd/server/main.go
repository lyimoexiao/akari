package main

import (
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/lyimoexiao/akari/pkg/version"
)

func main() {
	app, cleanup, err := Initialize(os.Getenv("CONFIG_PATH"))
	if err != nil {
		panic(err)
	}
	defer cleanup()

	shutdownTimeout := time.Duration(app.Config.Server.ShutdownTimeout) * time.Second

	srv := &http.Server{
		Addr:    ":" + app.Config.Server.Port,
		Handler: app.Engine,
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	app.Logger.Infow("server starting",
		"port", app.Config.Server.Port,
		"mode", app.Config.Server.Mode,
		"version", version.Short(),
	)

	go func() {
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			app.Logger.Fatalw("failed to start server", "error", err)
		}
	}()

	sig := <-quit
	app.Logger.Infow("received signal, shutting down", "signal", sig, "timeout", shutdownTimeout)

	ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		app.Logger.Errorw("forced shutdown", "error", err)
	}

	app.Logger.Info("server stopped")
}
