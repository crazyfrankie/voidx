package main

import (
	"context"
	"log"
	"net/http"
	"path/filepath"
	"syscall"
	"time"

	"github.com/joho/godotenv"
	"github.com/oklog/run"

	"github.com/crazyfrankie/voidx/conf"
	"github.com/crazyfrankie/voidx/ioc"
)

func main() {
	g := &run.Group{}

	prefix := "conf"
	envFile := filepath.Join(prefix, ".env")

	err := godotenv.Load(envFile)
	if err != nil {
		panic(err)
	}

	application := ioc.InitApplication()

	srv := &http.Server{
		Addr:    conf.GetConf().Server,
		Handler: application.Server,
	}

	g.Add(func() error {
		log.Printf("Server is running at http://localhost%s\n", conf.GetConf().Server)
		return srv.ListenAndServe()
	}, func(err error) {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := srv.Shutdown(ctx); err != nil {
			log.Printf("failed to shutdown main server: %v", err)
		}
	})

	consumerManager := application.Consumer
	ctx, cancel := context.WithCancel(context.Background())
	if err := consumerManager.Start(ctx); err != nil {
		cancel()
		panic(err)
	}

	g.Add(run.SignalHandler(context.Background(), syscall.SIGINT, syscall.SIGTERM))

	if err := g.Run(); err != nil {
		log.Printf("program interrupted, err:%s", err)
	}
}
