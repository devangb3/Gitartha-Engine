package main

import (
	"context"
	"fmt"
	"log"
	nethttp "net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/devang/Gitartha-Engine/internal/config"
	"github.com/devang/Gitartha-Engine/internal/data"
	"github.com/devang/Gitartha-Engine/internal/db"
	apirouter "github.com/devang/Gitartha-Engine/internal/http"
)

func main() {
	cfg := config.Load()

	dbConn, err := db.New(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("database connection error: %v", err)
	}
	defer dbConn.Close()

	store := data.NewStore(dbConn)

	router := apirouter.NewRouter(store)

	srv := &nethttp.Server{
		Addr:    fmt.Sprintf(":%s", cfg.Port),
		Handler: router,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != nethttp.ErrServerClosed {
			log.Fatalf("server error: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("server shutdown error: %v", err)
	}
}
