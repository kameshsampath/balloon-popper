package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/kameshsampath/balloon-popper-server/web"
)

func main() {
	server := web.NewServer()

	// Graceful shutdown
	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
		<-sigChan
		if err := server.Stop(); err != nil {
			log.Printf("Error stopping server: %v", err)
		}
		os.Exit(0)
	}()

	if err := server.Start(); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
