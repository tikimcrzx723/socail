package main

import (
	"log"

	"github.com/tikimcrzx723/social/internal/db"
	"github.com/tikimcrzx723/social/internal/env"
	"github.com/tikimcrzx723/social/internal/store"
)

func main() {
	addr := env.GetString("DB_ADDR", "postgres://admin:adminpassword@localhost/social?sslmode=disable")
	conn, err := db.New(addr, 3, 3, "15m")
	if err != nil {
		log.Fatal(err)
		return
	}
	defer conn.Close()

	store := store.NewStorage(conn)
	db.Seed(store, conn)
}
