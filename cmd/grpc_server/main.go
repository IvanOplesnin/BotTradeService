package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/IvanOplesnin/BotTradeService.git/internal/app"
	"google.golang.org/grpc"
)

func main() {
	configPath := flag.String("config", "config/config.yaml", "path to config file")
	flag.Parse()

	a, err := app.InitApp(*configPath)
	if err != nil {
		fmt.Println(err)
		return
	}

	// Запускаем сервер в отдельной горутине, чтобы main мог ждать сигнал.
	runErr := make(chan error, 1)
	go func() {
		runErr <- a.Run()
	}()

	// Ждем SIGINT (Ctrl+C) или SIGTERM (docker/k8s stop).
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	defer signal.Stop(sigCh)

	select {
	case sig := <-sigCh:
		fmt.Printf("received signal: %s\n", sig)

		done := make(chan struct{})
		go func() {
			a.GracefulStop()
			close(done)
		}()

		// Ждем graceful N секунд, потом жестко останавливаем.
		const shutdownTimeout = 5 * time.Second
		select {
		case <-done:
			// graceful ok
		case <-time.After(shutdownTimeout):
			fmt.Println("graceful shutdown timeout, forcing stop")
			a.Stop()
		}

		// Дожидаемся завершения Serve()
		if err := <-runErr; err != nil && err != grpc.ErrServerStopped {
			fmt.Println(err)
		}

	case err := <-runErr:
		if err != nil && err != grpc.ErrServerStopped {
			fmt.Println(err)
		}
	}
}
