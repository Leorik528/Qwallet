package main

import (
	"fmt"
	"qwalletrestapi/httpserver"
	"qwalletrestapi/internal/config"
	"qwalletrestapi/internal/storage/postges"
)

func main() {

	cfg := config.MustLoad()

	fmt.Println(cfg)

	database, err := postges.NewPostgresStore(cfg.DSN)

	if err != nil {
		fmt.Println("Проблемы с подключением к БД ", err)
	} else {
		fmt.Println()
	}

	httpHandlers := httpserver.NewHTTPHandlers(database)
	httpServer := httpserver.NewHTTPServer(httpHandlers)

	if err := httpServer.StartServer(); err != nil {
		fmt.Println("failed to start http server:", err)
	}
}
