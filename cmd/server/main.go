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
)

func main() {
	addr := flag.String("addr", ":9000", "gateway listen address")
	flag.Parse()

	logger := log.New(os.Stdout, "[game-server] ", log.LstdFlags|log.Lmicroseconds)
	r := router.New()
	handlers.RegisterPing(r, logger)
	handlers.RegisterEcho(r, logger)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	srv := gateway.NewServer(*addr, r, logger)
	if err := srv.Start(ctx); err != nil {
		logger.Fatalf("server exited with error: %v", err)
	}
}
