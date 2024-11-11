package main

import (
	"log/slog"
	"net/http"
	"os"

	hs "github.com/swaggo/http-swagger"
	_ "github.com/sysarmor/guard/server/docs/swag/docs"
)

// @title Swagger Guard API
// @version 2.0
// @description This is api document for Guard
// @termsOfService https://github.com/sysarmor/guard

// @schemes http
// @host localhost:8081
// @BasePath /
func main() {
	port := os.Getenv("SWAG_PORT")
	if port == "" {
		port = ":8081"
	}

	http.Handle("/", hs.WrapHandler)

	slog.Info("docs server start", "port", port)
	err := http.ListenAndServe(port, nil)
	if err != nil {
		panic(err)
	}
}
