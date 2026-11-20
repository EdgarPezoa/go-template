package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/edgarpezoa/go-template/internal/config"
)

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	configuration, err := config.Load()
	if err != nil {
		logger.Error("Configuration failed", "error", err)
		os.Exit(1)
	}

	bootstrap, err := NewBootstrap(configuration)
	if err != nil {
		logger.Error("Api initialization failed", "error", err)
		os.Exit(1)
	}

	if err := run(logger, bootstrap.Handler, configuration.API.Port); err != nil {
		logger.Error("api server failed", "error", err)
		closeBootstrap(logger, bootstrap)
		os.Exit(1)
	}

	closeBootstrap(logger, bootstrap)
}

func closeBootstrap(logger *slog.Logger, bootstrap *Bootstrap) {
	if err := bootstrap.Close(context.Background()); err != nil {
		logger.Error("api cleanup failed", "error", err)
	}
}

func run(logger *slog.Logger, handler http.Handler, port int) error {
	server := &http.Server{
		Addr:    ":" + fmt.Sprint(port),
		Handler: handler,
	}
	serverErrors := make(chan error, 1)
	go listenAndServe(server, serverErrors)

	shutdownSignalContext, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	select {
	case serverError := <-serverErrors:
		return handleServerError(server, serverError)
	case <-shutdownSignalContext.Done():
		return shutdownServer(server)
	}
}

func listenAndServe(server *http.Server, serverErrors chan<- error) {
	serverErrors <- server.ListenAndServe()
}

func handleServerError(server *http.Server, serverError error) error {
	if !errors.Is(serverError, http.ErrServerClosed) {
		_ = server.Close()
		return serverError
	}
	return nil
}

func shutdownServer(server *http.Server) error {
	shutdownContext, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := server.Shutdown(shutdownContext); err != nil {
		_ = server.Close()
		return err
	}
	return nil
}
