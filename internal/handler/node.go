package handler

import (
	"bufio"
	"errors"
	"io"
	"os"
	"os/exec"
	"path"
	"time"

	"github.com/google/uuid"
	"github.com/vTCP-Foundation/vtcpd-cli/internal/conf"
	"github.com/vTCP-Foundation/vtcpd-cli/internal/logger"
)

// This internal type is used for controlling internal node's goroutines behaviour.
type goroutineControlEvent struct {
	MustBeStopped bool
}

// Represents GEO engine node in the handler.
// Handles reading and writing of the fifo files.
type Node struct {
	// Channel of the commands, that must be transferred to the engine.
	commands chan *Command
	// Map of results of the commands, that was executed already.
	// Each result is mapped to it's command by the UUID.
	// Map contains channels, from which http requests handlers should be waiting for the results.
	results map[uuid.UUID]chan *Result

	commandsGoroutineControlChannel chan *goroutineControlEvent
	resultsGoroutineControlChannel  chan *goroutineControlEvent
	eventsGoroutineControlChannel   chan *goroutineControlEvent
	shouldNotBeRestarted            bool
}

func NewNode() *Node {

	return &Node{
		commands:                        make(chan *Command),
		results:                         make(map[uuid.UUID]chan *Result),
		shouldNotBeRestarted:            false,
		commandsGoroutineControlChannel: nil,
		resultsGoroutineControlChannel:  nil,
		eventsGoroutineControlChannel:   nil,
	}
}

func (node *Node) Start() (*exec.Cmd, error) {
	// Attempt to initially start the node.
	// In case if this attempt fails - the error must be returned imminently.

	// Starting child process.
	process := exec.Command(conf.Params.VTCPDPath)
	process.Dir = path.Join(conf.Params.WorkDir)

	err := process.Start()
	if err != nil {
		return nil, wrap("Can't start child process", err)
	}

	node.logInfo("Started")
	return process, nil
}

func (node *Node) StartCommunication() (chan *goroutineControlEvent, chan *goroutineControlEvent, error) {
	// Node instance must wait some time for the child process to start listening for commands.
	// this timeout specifies how long it would wait.
	CHILD_PROCESS_SPAWN_TIMEOUT_SECONDS := 2

	commandsControlEventsChanel := make(chan *goroutineControlEvent, 1)
	commandsGoroutineErrorsChanel := make(chan error, 1)
	go node.beginTransferCommands(
		path.Join(conf.Params.WorkDir),
		CHILD_PROCESS_SPAWN_TIMEOUT_SECONDS,
		commandsControlEventsChanel,
		commandsGoroutineErrorsChanel)

	resultsControlEventsChanel := make(chan *goroutineControlEvent, 1)
	resultsGoroutinesErrorsChanel := make(chan error, 1)
	go node.beginReceiveResults(
		path.Join(conf.Params.WorkDir),
		CHILD_PROCESS_SPAWN_TIMEOUT_SECONDS,
		resultsControlEventsChanel,
		resultsGoroutinesErrorsChanel)

	// Now node handler must wait for the success response from the internal goroutines.
	// In case of no response, or response wasn't received in specified timeout -
	// error would be reported, and internal goroutines would be finished.
	CHILD_PROCESS_MAX_SPAWN_TIMEOUT_SECONDS := CHILD_PROCESS_SPAWN_TIMEOUT_SECONDS
	select {
	case commandsError := <-commandsGoroutineErrorsChanel:
		{
			// Commands transferring goroutine reported the error.
			// It is assumed, that it would exit without external signal,
			// but the results receiving goroutine must be forced to stop.
			resultsControlEventsChanel <- &goroutineControlEvent{MustBeStopped: true}
			return nil, nil, wrap("Can't start commands transferring to the child process", commandsError)
		}

	case responsesError := <-resultsGoroutinesErrorsChanel:
		{
			// Results receiving goroutine reported the error.
			// It is assumed, that it would exit without external signal,
			// but the commands transferring goroutine must be forced to stop.
			commandsControlEventsChanel <- &goroutineControlEvent{MustBeStopped: true}
			return nil, nil, wrap("Can't start results receiving from the child process", responsesError)
		}

	case <-time.After(time.Second * time.Duration(CHILD_PROCESS_MAX_SPAWN_TIMEOUT_SECONDS)):
		{
			// There are no errors from the internal goroutines was reported for specified period of time.
			// It is assumed, that both of goroutines was started well and the child process is executed normally.
			break
		}
	}

	// It seems, that child process started well.
	// Now the process descriptor must be transferred to the top, for further control.

	node.commandsGoroutineControlChannel = commandsControlEventsChanel
	node.resultsGoroutineControlChannel = resultsControlEventsChanel

	node.logInfo("Communication started")
	return commandsControlEventsChanel, resultsControlEventsChanel, nil
}

func (node *Node) StopCommunication() error {
	// Stopping communication with the node.
	if node.commandsGoroutineControlChannel != nil {
		node.commandsGoroutineControlChannel <- &goroutineControlEvent{MustBeStopped: true}
	}
	if node.resultsGoroutineControlChannel != nil {
		node.resultsGoroutineControlChannel <- &goroutineControlEvent{MustBeStopped: true}
	}

	return nil
}

func openFifoFile(commandsFIFOPath string, flag int, perm os.FileMode, node *Node) (*os.File, error) {
	var fifo *os.File
	var counter int8 = 1
	var err error
	for {
		fifo, err = os.OpenFile(commandsFIFOPath, flag, perm)
		if err != nil {
			counter++
			if counter == 5 {
				node.logError("Max tries count expired. Report error and exit")
				return fifo, err
			}
			node.logError("Can't open " + commandsFIFOPath + " for writing. Details: " + err.Error())
			node.logError("Wait 3s before repeat")
			time.Sleep(time.Second * 3)
			continue
		}
		break
	}
	return fifo, err
}

func (node *Node) beginTransferCommands(
	nodeWorkingDirPath string,
	initialStartupDelaySeconds int,
	controlEvents chan *goroutineControlEvent,
	errorsChannel chan error) {

	// Give process some time to open commands.fifo for reading
	time.Sleep(time.Second * time.Duration(initialStartupDelaySeconds))

	commandsFIFOPath := path.Join(nodeWorkingDirPath, "fifo", "commands.fifo")
	fifo, err := openFifoFile(commandsFIFOPath, os.O_WRONLY, 0777, node)
	if err != nil {
		errorsChannel <- wrap("Can't open "+commandsFIFOPath+" for writing", err)
		node.logError("Can't open " + commandsFIFOPath + " for writing. Details: " + err.Error())
		return
	}

	writer := bufio.NewWriter(fifo)
	for {
		select {
		case command := <-node.commands:
			{
				_, err := writer.Write(command.ToBytes())
				if err != nil {
					node.logError("Can't transfer command to the node, command details: " + string(command.ToBytes()))
					node.results[command.UUID] <- &Result{Error: err}

				} else {
					writer.Flush()
				}
			}

		case event := <-controlEvents:
			{
				if event.MustBeStopped {
					node.logInfo("Commands writing goroutine was finished by the external signal.")

					fifo.Close()
					return
				}
			}
		}
	}
}

// Blocks reading
func (node *Node) beginReceiveResults(
	nodeWorkingDirPath string,
	initialStartupDelaySeconds int,
	controlEvents chan *goroutineControlEvent,
	errorsChannel chan error) {

	// Give process some time to open results.fifo for writing.
	time.Sleep(time.Second * time.Duration(initialStartupDelaySeconds))

	resultsFIFOPath := path.Join(nodeWorkingDirPath, "fifo", "results.fifo")
	fifo, err := openFifoFile(resultsFIFOPath, os.O_RDONLY, 0600, node)
	if err != nil {
		errorsChannel <- wrap("Can't open "+resultsFIFOPath+" file for reading", err)
		node.logError("Can't open " + resultsFIFOPath + " for reading. Details: " + err.Error())
		return
	}

	reader := bufio.NewReader(fifo)
	for {
		// In case if this goroutine receives shutdown event -
		// process it and stop reading results.
		if len(controlEvents) > 0 {
			event := <-controlEvents
			if event.MustBeStopped {
				node.logDebug("Results receiving goroutine was finished by the external signal.")
				fifo.Close()
				return
			}
		}

		// Results are divided by "\n", so it is possible to read entire line from file.
		line, err := reader.ReadBytes('\n')
		if err != nil {
			if err == io.EOF {
				time.Sleep(time.Millisecond * 5)
				continue
			}

			node.logError("Error occurred on result reading. Details are: " + err.Error())
			continue
		}

		// In some cases, redundant '\n' is returned as a result.
		// This erroneous results must be ignored.
		if len(line) == 1 && line[0] == '\n' {
			continue
		}

		// Sometimes, empty lines are produced by the reader.
		// It may be result of invalid write by the engine, or by the internal logic of GO's ReadBytes.
		// In all cases, lines that are shorter than uuid hex length must be dropped.
		UUID_HEX_LENGTH := 36
		if len(line) < UUID_HEX_LENGTH {
			node.logError("To short result occurred. Details are: \"" + string(line) + "\". Dropped")
			continue
		}

		result := ResultFromRawInput(line)
		if result.Error != nil {
			node.logError("Invalid result occurred. Details are: \"" + string(line) + "\". Dropped")
			continue
		}

		// Results received well.
		node.logDebug("Received result: " + string(line))

		// Access to the map is safe:
		// there is only read access used here.
		// (the channel for the result is already present)
		//
		// There is no need for the mutex here.
		channel, isPresent := node.results[result.UUID]
		if isPresent {
			// Transferring result for further processing.
			channel <- result

			node.logInfo("OK: Channel " + result.UUID.String() + " found.")

		} else {
			node.logError("No channel found for the result " + result.UUID.String() + ". Details are: \"" + string(line) + "\". Dropped")
			reader.Reset(fifo)
		}
	}
}

func (node *Node) BeginMonitorInternalProcessCrashes(
	process *exec.Cmd,
	commandsGoroutineControlEvents chan *goroutineControlEvent,
	resultsGoroutineControlEvents chan *goroutineControlEvent) {

	MAX_NODE_REINITIALIZATION_ATTEMPTS := 10
	MIN_TIME_INTERVAL_BETWEEN_CRASHES := time.Second * 30

	currentNodeInitializationAttempt := 0
	lastReinitialisationAttemptTimestamp := time.Now()

	var err error = nil
	for {
		if err != nil {
			node.logError("Finished unexpectedly with the error: " + err.Error())

			if node.shouldNotBeRestarted {
				logger.Info("Node was prevented from restarting. It seems that stop method was called.")
				return
			}

			if time.Since(lastReinitialisationAttemptTimestamp) > MIN_TIME_INTERVAL_BETWEEN_CRASHES {
				// Last node crash was far too in the past.
				// Crashes counter must be reinitialised.
				currentNodeInitializationAttempt = 0
			}

			if currentNodeInitializationAttempt > MAX_NODE_REINITIALIZATION_ATTEMPTS {
				node.logError("Has been crashed to much times and would not be restored any more.")
				node.commandsGoroutineControlChannel = nil
				node.resultsGoroutineControlChannel = nil
				return
			}

			currentNodeInitializationAttempt += 1
			lastReinitialisationAttemptTimestamp = time.Now()

			// Previous process communication goroutines must be closed,
			// before several new would be started.
			commandsGoroutineControlEvents <- &goroutineControlEvent{MustBeStopped: true}
			resultsGoroutineControlEvents <- &goroutineControlEvent{MustBeStopped: true}

			// Attempt to restore the node
			process, err = node.Start()
			if err == nil {
				commandsGoroutineControlEvents, resultsGoroutineControlEvents, err = node.StartCommunication()
				if err == nil {
					node.logInfo("Restarted")
				} else {
					node.logError("Can't restart node communication")
				}
			} else {
				node.logError("Can't restart node")
			}
		}

		err = process.Wait()
	}
}

// Sends command to the engine.
func (node *Node) SendCommand(command *Command) error {
	// WARN: order is significant.
	// Channel for the result must be created before sending command to the execution.
	node.results[command.UUID] = make(chan *Result, 1)

	node.logInfo("Command sent: " + string(command.ToBytes()))

	select {
	case node.commands <- command:
		return nil
	case <-time.After(time.Second * 10):
		return errors.New("can't add command to node commands channel")
	}

}

// Wait for command result from the engine.
func (node *Node) WaitCommand(command *Command) {
	// WARN: order is significant.
	// Channel for the result must be created before sending command to the execution.
	node.results[command.UUID] = make(chan *Result, 1)

	node.logInfo("Command wait: " + string(command.ToBytes()))
}

func (node *Node) GetResult(command *Command, timeoutSeconds uint16) (*Result, error) {
	channel, isPresent := node.results[command.UUID]
	if !isPresent {
		return nil, errors.New("no results channel is present for this UUID")
	}

	select {
	case result := <-channel:
		// This method would be called from the http requests handler,
		// each running it's own goroutine.
		// There is a write access here, so the mutex must be used.
		delete(node.results, command.UUID)

		return result, nil

	case <-time.After(time.Second * time.Duration(timeoutSeconds)):
		// This method would be called from the http requests handler,
		// each running it's own goroutine.
		// There is a write access here, so the mutex must be used.
		delete(node.results, command.UUID)

		return nil, errors.New("timeout fired up")
	}

}

func (node *Node) logError(message string) {
	logger.Error(node.logHeader() + message)
}

func (node *Node) logInfo(message string) {
	logger.Info(node.logHeader() + message)
}

func (node *Node) logDebug(message string) {
	logger.Debug(node.logHeader() + message)
}

func (node *Node) logHeader() string {
	return "[Node]: "
}
