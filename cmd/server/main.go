package main

import (
	"context"
	"encoding/base64"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/stpotter16/coin/internal/handlers"
	"github.com/stpotter16/coin/internal/handlers/authentication"
	"github.com/stpotter16/coin/internal/handlers/sessions"
	"github.com/stpotter16/coin/internal/plaidclient"
	"github.com/stpotter16/coin/internal/store/db"
	"github.com/stpotter16/coin/internal/store/sqlite"
	"github.com/stpotter16/coin/internal/sync"
)

func run(
	ctx context.Context,
	args []string,
	getenv func(string) string,
	stdin io.Reader,
	stdout, stderr io.Writer,
) error {
	ctx, cancel := signal.NotifyContext(ctx, os.Interrupt)
	defer cancel()

	log.SetFlags(log.LstdFlags | log.Lshortfile)

	dbPath := getenv("COIN_DB_PATH")
	if dbPath == "" {
		return errors.New("database environment variable not set")
	}
	flag.Parse()

	encryptionKey, err := loadEncryptionKey(getenv)
	if err != nil {
		return err
	}

	log.Printf("Opening database in %v", dbPath)
	database, err := db.New(dbPath)
	if err != nil {
		return err
	}

	store, err := sqlite.New(database)
	if err != nil {
		return err
	}

	sessionManager, err := sessions.New(database, getenv)
	if err != nil {
		return err
	}

	authenticator := authentication.New(store)

	plaidClient, err := plaidclient.New(getenv)
	if err != nil {
		return err
	}

	syncer := sync.New(store, plaidClient, encryptionKey)

	go startSyncPoller(ctx, syncer)

	handler := handlers.NewServer(store, sessionManager, authenticator, plaidClient, syncer, encryptionKey)
	port := getenv("PORT")
	if port == "" {
		port = "8080"
	}
	server := &http.Server{
		Addr:    ":" + port,
		Handler: handler,
	}

	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("Server error: %v\n", err)
		}
	}()

	<-ctx.Done()
	log.Println("Received termination signal. Shutting down")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("server shutdown error: %w", err)
	}
	return nil
}

func loadEncryptionKey(getenv func(string) string) ([]byte, error) {
	encoded := getenv("COIN_ENCRYPTION_KEY")
	if encoded == "" {
		return nil, errors.New("COIN_ENCRYPTION_KEY environment variable not set")
	}
	key, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return nil, fmt.Errorf("COIN_ENCRYPTION_KEY is not valid base64: %w", err)
	}
	if len(key) != 32 {
		return nil, fmt.Errorf("COIN_ENCRYPTION_KEY must be 32 bytes, got %d", len(key))
	}
	return key, nil
}

func startSyncPoller(ctx context.Context, syncer sync.Syncer) {
	ticker := time.NewTicker(time.Hour)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			log.Println("sync: starting scheduled sync")
			syncer.SyncAll(ctx)
		case <-ctx.Done():
			return
		}
	}
}

func main() {
	ctx := context.Background()
	if err := run(
		ctx,
		os.Args,
		os.Getenv,
		nil,
		os.Stdout,
		os.Stderr,
	); err != nil {
		log.Fatalf("%s", err)
	}
}
