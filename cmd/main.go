package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"

	segsrv "github.com/ryanreadbooks/folium/internal/segment/server"
)

var (
	httpPort int
	grpcPort int
)

func init() {
	flag.IntVar(&httpPort, "httpPort", 9527, "the http server port")
	flag.IntVar(&grpcPort, "grpcPort", 9528, "the grpc server port")
}

func ServeSegment() {
	segsrv.InitHttp(httpPort)
	segsrv.InitGrpc(grpcPort)
}

func main() {
	flag.Parse()

	ServeSegment()

	// gracefully shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	sig := <-quit
	log.Printf("folium got a signal: %v\n", sig.String())

	segsrv.CloseServer()
}
