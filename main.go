package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/rs/zerolog/log"
	"github.com/vTCP-Foundation/vtcpd-cli/pkg/conf"
	"github.com/vTCP-Foundation/vtcpd-cli/pkg/logger"
	"github.com/vTCP-Foundation/vtcpd-cli/pkg/vtcpd"
)

const (
	ExitCodeSuccess     = 0
	ExitCodeConfigError = 1
)

func main() {
	err := conf.LoadConfiguration()
	if err != nil {
		fmt.Println("Failed to load configuration: ", err)
		os.Exit(ExitCodeConfigError)
	}

	logger.Init(conf.LogLevel())

	wg := &sync.WaitGroup{}
	node := vtcpd.NewNode()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go node.BlockingBeginReceiveResults(ctx, wg)
	go node.BlockingBeginCommandsTransfer(ctx, wg)

	// Setup SIGTERM and SIGINT signal handling
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGTERM, syscall.SIGINT)

	sig := <-sigChan
	log.Info().Msgf("Received signal %v, shutting down...", sig)
	cancel() // Cancel context to initiate shutdown
	wg.Wait()

	log.Debug().Msg("All goroutines finished, exiting")
	os.Exit(ExitCodeSuccess)
}
