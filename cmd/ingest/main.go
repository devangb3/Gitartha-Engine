package main

import (
	"context"
	"flag"
	"log"
	"time"

	"github.com/devang/Gitartha-Engine/internal/config"
	"github.com/devang/Gitartha-Engine/internal/db"
	"github.com/devang/Gitartha-Engine/internal/ingest"
)

func main() {
	csvPath := flag.String("csv", "bg.csv", "path to Bhagavad Gita CSV file")
	timeout := flag.Duration("timeout", 2*time.Minute, "ingestion timeout")
	flag.Parse()

	cfg := config.Load()

	dbConn, err := db.New(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("database connection error: %v", err)
	}
	defer dbConn.Close()

	loader := ingest.New(dbConn)

	ctx, cancel := context.WithTimeout(context.Background(), *timeout)
	defer cancel()

	if err := loader.LoadCSV(ctx, *csvPath); err != nil {
		log.Fatalf("ingestion failed: %v", err)
	}

	log.Println("ingestion completed successfully")
}
