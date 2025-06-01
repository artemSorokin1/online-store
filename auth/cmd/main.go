package main

import (
	"auth/internal/config"
	"auth/internal/routes"
	"auth/pkg/server"
	"log"
)

func main() {
	cfg := config.MustLoad()

	router, db, err := routes.SetupApp(cfg)
	if err != nil {
		log.Fatal("cant setup app", err)
	}
	defer db.DB.Close()

	server.MustRun(router, cfg)

}
