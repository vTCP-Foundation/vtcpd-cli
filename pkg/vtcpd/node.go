package vtcpd

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"os"
	"sync"

	"github.com/rs/zerolog/log"
	"github.com/vTCP-Foundation/vtcpd-cli/pkg/conf"

	"time"

	"github.com/google/uuid"
)

var (
	ErrFIFOOpen              = errors.New("can't open fifo")
	ErrCommandSubmission     = errors.New("can't submit command")
	ErrResultChannelNotFound = errors.New("no results channel is present for this UUID")
	ErrTimeout               = errors.New("timeout fired up")
)

type Node struct {
	commands chan *Command

	results        map[uuid.UUID]chan *Result
	resultsRWMutex sync.RWMutex
}

func NewNode() *Node {
	return &Node{
		commands: make(chan *Command),
		results:  make(map[uuid.UUID]chan *Result),
	}
}

func (node *Node) BlockingBeginCommandsTransfer(ctx context.Context, wg *sync.WaitGroup) error {
	log.Debug().Msg("Starting commands transferring")

	wg.Add(1)
	defer wg.Done()

mainLoop:
	for {
		select {
		case <-ctx.Done():
			log.Debug().Msg("Commands transferring goroutine was finished due to context cancellation")
			return nil

		default:
			fifo, err := node.connectToCommandsFifo(ctx)
			if err != nil {
				// If context was cancelled, we don't need to log an error
				if errors.Is(err, context.Canceled) {
					log.Debug().Msg("Commands transferring goroutine was finished due to context cancellation")
					return nil
				} else {
					log.Error().Err(err).Msg("Can't connect to node's commands FIFO")
					return err
				}
			}
			writer := bufio.NewWriter(fifo)

			processError := func(command *Command, err error) {
				log.Err(err).
					Str("commands.uuid", command.UUID.String()).
					Str("command.body", command.Body).
					Msg("Can't transfer command to the node's FIFO")
				node.results[command.UUID] <- &Result{Error: err}

				log.Error().Msg(
					"Commands transferring was finished due to inability to write command to the node's FIFO " +
						"It will be restarted automatically in 3 seconds...")

				fifo.Close()
				time.Sleep(3 * time.Second)
				log.Debug().Msg("Restarting commands transferring goroutine...")
			}
			for {
				select {
				case command := <-node.commands:
					{
						_, err := writer.Write(command.ToBytes())
						if err != nil {
							processError(command, err)
							continue mainLoop
						}

						err = writer.Flush()
						if err != nil {
							processError(command, err)
							continue mainLoop
						}
					}
				}
			}
		}
	}
}

func (node *Node) BlockingBeginReceiveResults(ctx context.Context, wg *sync.WaitGroup) error {
	log.Debug().Msg("Starting results receiving")

	wg.Add(1)
	defer wg.Done()

mainLoop:
	for {
		select {
		case <-ctx.Done():
			log.Debug().Msg("Results receiving goroutine was finished due to context cancellation")
			return nil

		default:
			fifo, err := node.connectToResultsFifo(ctx)
			if err != nil {
				// If context was cancelled, we don't need to log an error
				if errors.Is(err, context.Canceled) {
					log.Debug().Msg("Results receiving goroutine was finished due to context cancellation")
					return nil
				} else {
					log.Error().Err(err).Msg("Can't connect to node's results FIFO")
					return err
				}
			}

			reader := bufio.NewReader(fifo)

			readDone := make(chan struct {
				line []byte
				err  error
			})

			for {
				// Start the read operation in a goroutine
				go func() {
					line, err := reader.ReadBytes('\n')
					readDone <- struct {
						line []byte
						err  error
					}{line, err}
				}()

				// Wait for either context cancellation or read completion
				select {
				case <-ctx.Done():
					close(readDone)
					fifo.Close()
					return ctx.Err()

				case readResult := <-readDone:
					if readResult.err != nil {
						log.Error().Err(readResult.err).Msg("Error reading from results FIFO")
						log.Error().Msg(
							"Results receiving was finished due to read error. " +
								"It will be restarted automatically in 3 seconds...")

						close(readDone)
						fifo.Close()
						time.Sleep(3 * time.Second)
						log.Debug().Msg("Restarting results receiving goroutine...")
						continue mainLoop
					}

					// Skip empty lines
					if len(readResult.line) <= 1 {
						continue
					}

					// Parse result
					result := ResultFromRawInput(readResult.line)
					if result.Error != nil {
						log.Error().
							Str("raw_input", string(readResult.line)).
							Msg("Invalid result received, dropping")
						continue
					}

					// Send result to appropriate channel if it exists
					if resultChan, exists := node.results[result.UUID]; exists {
						resultChan <- result
						log.Debug().
							Str("result.uuid", result.UUID.String()).
							Msg("Result delivered to channel")
					} else {
						log.Error().
							Str("result.uuid", result.UUID.String()).
							Msg("No channel found for result UUID, skipped")
					}
				}
			}
		}
	}
}

func (node *Node) SendCommand(command *Command) error {
	// Create channel and store it atomically
	node.resultsRWMutex.Lock()
	node.results[command.UUID] = make(chan *Result, 1)
	node.resultsRWMutex.Unlock()

	// Create a timer to avoid potential memory leak
	timer := time.NewTimer(10 * time.Second)
	defer timer.Stop()

	// If command submission fails, ensure we clean up the channel
	select {
	case node.commands <- command:
		return nil
	case <-timer.C:
		// Clean up the channel on timeout
		node.resultsRWMutex.Lock()
		delete(node.results, command.UUID)
		node.resultsRWMutex.Unlock()
		return ErrCommandSubmission
	}
}

func (node *Node) GetResult(command *Command, timeoutSeconds uint16) (*Result, error) {
	// Get the channel under read lock
	node.resultsRWMutex.RLock()
	channel, isPresent := node.results[command.UUID]
	node.resultsRWMutex.RUnlock()

	if !isPresent {
		return nil, ErrResultChannelNotFound
	}

	// Ensure cleanup happens exactly once
	defer func() {
		node.resultsRWMutex.Lock()
		delete(node.results, command.UUID)
		node.resultsRWMutex.Unlock()
	}()

	// Create a timer to avoid potential memory leak
	timer := time.NewTimer(time.Duration(timeoutSeconds) * time.Second)
	defer timer.Stop()

	// Wait for result or timeout
	select {
	case result := <-channel:
		return result, nil
	case <-timer.C:
		return nil, ErrTimeout
	}
}

func (node *Node) connectToCommandsFifo(ctx context.Context) (fifo *os.File, err error) {
	commandsFIFOPath := conf.CommandsFifoPath()
	log.Debug().Msgf("Opening commands FIFO for writing: %s", commandsFIFOPath)

	// Create a channel to receive the open result
	type openResult struct {
		file *os.File
		err  error
	}
	done := make(chan openResult)

	// Open the FIFO in a separate goroutine since it might block
	go func() {
		f, err := os.OpenFile(commandsFIFOPath, os.O_WRONLY, 0660)
		done <- openResult{f, err}
	}()

	// Wait for either the context to be cancelled or the open to complete
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case result := <-done:
		if result.err != nil {
			return nil, fmt.Errorf("can't open commands fifo (%s): %w", commandsFIFOPath, ErrFIFOOpen)
		}
		log.Debug().Msg("Successfully connected to commands FIFO")
		return result.file, nil
	}
}

func (node *Node) connectToResultsFifo(ctx context.Context) (fifo *os.File, err error) {
	resultsFIFOPath := conf.ResultsFifoPath()
	log.Debug().Msgf("Opening results FIFO for reading: %s", resultsFIFOPath)

	// Create a channel to receive the open result
	type openResult struct {
		file *os.File
		err  error
	}
	done := make(chan openResult)

	// Open the FIFO in a separate goroutine since it might block
	go func() {
		f, err := os.OpenFile(resultsFIFOPath, os.O_RDONLY, 0660)
		done <- openResult{f, err}
	}()

	// Wait for either the context to be cancelled or the open to complete
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case result := <-done:
		if result.err != nil {
			return nil, fmt.Errorf("can't open results fifo (%s): %w", resultsFIFOPath, ErrFIFOOpen)
		}
		log.Debug().Msg("Successfully connected to results FIFO")
		return result.file, nil
	}
}
