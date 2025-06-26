package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/crazyfrankie/voidx/conf"
	"github.com/crazyfrankie/voidx/ioc"
)

func main() {
	engine := ioc.InitEngine()

	srv := &http.Server{
		Addr:    conf.GetConf().Server,
		Handler: engine,
	}

	log.Printf("Server is running at http://localhost%s", conf.GetConf().Server)
	go func() {
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Printf("failed start server: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	<-quit
	log.Println("Shutting down server...")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), time.Second*5)
	defer shutdownCancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Printf("failed to shutdown main server: %v", err)
	}
	log.Println("Server exited gracefully")
}
