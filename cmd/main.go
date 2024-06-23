package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	segsrv "github.com/ryanreadbooks/folium/internal/segment/server"
)

func ServeSegment() {
	segsrv.InitHttp()
	segsrv.InitGrpc()
}

func main() {
	ServeSegment()

	// gracefully shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	sig := <-quit
	log.Printf("folium got a signal: %v\n", sig.String())

	segsrv.CloseServer()
}
