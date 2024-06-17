package main

import (
	"os"
	"os/signal"
	"syscall"
)

func main() {

	// gracefully shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

}
