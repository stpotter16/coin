package main

import (
	"context"
	"flag"
	"log"
	"os"

	"github.com/stpotter16/coin/internal/auth"
	"github.com/stpotter16/coin/internal/store/db"
	"github.com/stpotter16/coin/internal/store/sqlite"
)

func main() {
	username := flag.String("username", "", "Username for the new user")
	password := flag.String("password", "", "Password for the new user")
	flag.Parse()

	if *username == "" || *password == "" {
		log.Fatal("Usage: go run cmd/tools/update_user.go -username <username> -password <password>")
	}

	dbPath := os.Getenv("COIN_DB_PATH")
	if dbPath == "" {
		log.Fatal("COIN_DB_PATH environment variable not set")
	}

	database, err := db.New(dbPath)
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}

	store, err := sqlite.New(database)
	if err != nil {
		log.Fatalf("Failed to initialise store: %v", err)
	}

	hash, err := auth.HashPassword(*password)
	if err != nil {
		log.Fatalf("Failed to hash password: %v", err)
	}

	if err := store.UpdateUser(context.Background(), *username, hash); err != nil {
		log.Fatalf("Failed to update user: %v", err)
	}

	log.Printf("User %q updated successfully", *username)
}
