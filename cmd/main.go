package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	segsrv "github.com/ryanreadbooks/folium/internal/segment/server"
)

func main() {
	segsrv.InitHttp()

	// gracefully shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	sig := <-quit
	log.Printf("folium got signal: %v\n", sig.String())
}
