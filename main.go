package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/danthegoodman1/GoAPITemplate/gologger"
	"github.com/danthegoodman1/GoAPITemplate/http_server"
	"github.com/danthegoodman1/GoAPITemplate/observability"
	"github.com/danthegoodman1/GoAPITemplate/resources"
	"github.com/danthegoodman1/GoAPITemplate/utils"
	"github.com/danthegoodman1/GoAPITemplate/workload"
	"github.com/joho/godotenv"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

var logger = gologger.NewLogger()

func main() {
	if _, err := os.Stat(".env"); err == nil {
		err = godotenv.Load()
		if err != nil {
			logger.Error().Err(err).Msg("error loading .env file, exiting")
			os.Exit(1)
		}
	}

	go func() {
		prometheusReporter := observability.NewPrometheusReporter()
		err := observability.StartInternalHTTPServer(":8042", prometheusReporter)
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error().Err(err).Msg("internal server couldn't start")
			os.Exit(1)
		}
	}()

	httpServer := http_server.StartHTTPServer()

	// Figure out what process to run
	if utils.IsWorker {
		go startWorker(httpServer)
	} else if utils.IsScheduler {
		go startScheduler(httpServer)
	} else {
		logger.Fatal().Msgf("unknown role '%s'", utils.Role)
	}

	// Listen for shutdown signals
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c
	logger.Warn().Msg("received shutdown signal!")

	// For AWS ALB needing some time to de-register pod
	// Convert the time to seconds
	sleepTime := utils.GetEnvOrDefaultInt("SHUTDOWN_SLEEP_SEC", 0)
	logger.Info().Msg(fmt.Sprintf("sleeping for %ds before exiting", sleepTime))

	time.Sleep(time.Second * time.Duration(sleepTime))
	logger.Info().Msg(fmt.Sprintf("slept for %ds, exiting", sleepTime))

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	if err := httpServer.Shutdown(ctx); err != nil {
		logger.Error().Err(err).Msg("failed to shutdown HTTP server")
	} else {
		logger.Info().Msg("successfully shutdown HTTP server")
	}
}

func startWorker(s *http_server.HTTPServer) {
	logger.Debug().Msgf("starting compute WORKER '%s'", utils.Hostname)

	resources.InitResourceManager()
	err := workload.Init()
	if err != nil {
		logger.Fatal().Err(err).Msg("error initializing workloads")
	}
	http_server.RegisterWorkerHandlers(s)
}

func startScheduler(s *http_server.HTTPServer) {
	logger.Debug().Msgf("starting compute SCHEDULER '%s'", utils.Hostname)
	http_server.RegisterSchedulerHandlers(s)
}
