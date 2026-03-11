package main

import (
	"flag"
	"fmt"
	"log"
	"time"

	"github.com/stpotter16/coin/internal/auth"
)

func main() {
	username := flag.String("username", "", "Username for the new user")
	password := flag.String("password", "", "Password for the new user")
	isAdmin := flag.Bool("admin", false, "Grant admin privileges")
	flag.Parse()

	if *username == "" || *password == "" {
		log.Fatal("Usage: go run tools/create_user.go -username <username> -password <password> [-admin]")
	}

	hash, err := auth.HashPassword(*password)
	if err != nil {
		log.Fatalf("Failed to hash password: %v", err)
	}

	now := time.Now().UTC().Format(time.RFC3339)
	adminVal := 0
	if *isAdmin {
		adminVal = 1
	}

	fmt.Printf(
		"INSERT INTO user (username, password, is_admin, created_time, last_modified_time) VALUES ('%s', '%s', %d, '%s', '%s');\n",
		*username, hash, adminVal, now, now,
	)
}
