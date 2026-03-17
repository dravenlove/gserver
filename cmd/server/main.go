package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"

	"openclaw-go/internal/gateway"
	"openclaw-go/internal/handlers"
	"openclaw-go/internal/router"
	"openclaw-go/internal/storage/mysql"
)

func main() {
	addr := flag.String("addr", ":9000", "gateway listen address")
	mysqlDSN := flag.String("mysql-dsn", "gserver:gserver@tcp(127.0.0.1:3306)/gserver?charset=utf8mb4&parseTime=true&loc=Local", "mysql dsn")
	flag.Parse()

	logger := log.New(os.Stdout, "[game-server] ", log.LstdFlags|log.Lmicroseconds)

	store, err := mysql.NewPlayerStore(*mysqlDSN)
	if err != nil {
		logger.Fatalf("init mysql store failed: %v", err)
	}
	defer func() {
		if cerr := store.Close(); cerr != nil {
			logger.Printf("close mysql store failed: %v", cerr)
		}
	}()

	r := router.New()
	handlers.RegisterPing(r, logger)
	handlers.RegisterEcho(r, logger)
	handlers.RegisterPlayer(r, logger, store)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	srv := gateway.NewServer(*addr, r, logger)
	if err := srv.Start(ctx); err != nil {
		logger.Fatalf("server exited with error: %v", err)
	}
}
