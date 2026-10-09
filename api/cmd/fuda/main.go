package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"fuda/internal/app"
	"fuda/internal/config"
	"fuda/internal/httpapi"
	"fuda/internal/login"
	"fuda/internal/web"
)

func main() {
	log := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	if err := run(log); err != nil {
		log.Error("fuda stopped", "error", err)
		os.Exit(1)
	}
}

func newLogins(log *slog.Logger, cfg config.Config) ([]httpapi.Login, error) {
	var logins []httpapi.Login
	if cfg.Has(config.SourceGitHub) {
		key, err := login.LoadOrCreateKey(filepath.Join(cfg.CacheDir, "cookie.key"))
		if err != nil {
			return nil, fmt.Errorf("cookie key: %w", err)
		}
		github, err := login.NewWebDevice(log, login.WebDeviceConfig{ClientID: cfg.GitHubClientID, Key: key})
		if err != nil {
			return nil, err
		}
		logins = append(logins, github)
	}
	if cfg.Has(config.SourceAzure) {
		azure, err := login.NewAzure(log, login.Config{
			Tenant: cfg.AzureTenant, ClientID: cfg.AzureClientID, ClientSecret: cfg.AzureClientSecret, CookieSecret: cfg.CookieSecret, BaseURL: cfg.BaseURL,
		})
		if err != nil {
			return nil, err
		}
		logins = append(logins, azure)
	}
	return logins, nil
}

func run(log *slog.Logger) error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if ignored := cfg.IgnoredVars(); len(ignored) > 0 {
		log.Warn("these settings are ignored", "vars", ignored)
	}
	logins, err := newLogins(log, cfg)
	if err != nil {
		return err
	}

	folders, err := app.NewFolders("")
	if err != nil {
		return err
	}
	if cfg.Has(config.SourceLocal) {
		if _, err := folders.Add(cfg.LocalPath); err != nil {
			return fmt.Errorf("FUDA_LOCAL_PATH: %w", err)
		}
	}

	server := &http.Server{
		Addr:              cfg.Addr,
		Handler:           httpapi.NewHandler(log, app.NewBoards(log, cfg, folders), web.Dist(), logins...),
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       30 * time.Second,
		IdleTimeout:       2 * time.Minute,
	}

	errs := make(chan error, 1)
	go func() {
		log.Info("listening", "addr", cfg.Addr, "sources", cfg.Sources)
		errs <- server.ListenAndServe()
	}()

	select {
	case err := <-errs:
		if !errors.Is(err, http.ErrServerClosed) {
			return err
		}
	case <-ctx.Done():
	}

	shutdown, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	return server.Shutdown(shutdown)
}
