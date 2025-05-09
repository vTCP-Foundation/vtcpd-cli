package handler

import (
	"bufio"
	"encoding/json"
	"errors"
	"os"
	"path"
	"strconv"
	"syscall"

	"github.com/vTCP-Foundation/vtcpd-cli/internal/conf"
	"github.com/vTCP-Foundation/vtcpd-cli/internal/logger"
)

var (
	// Response status codes
	// for the commands
	OK                         = 200
	CREATED                    = 201
	ACCEPTED                   = 202
	BAD_REQUEST                = 400
	NODE_NOT_FOUND             = 405
	SERVER_ERROR               = 500
	NODE_IS_INACCESSIBLE       = 503
	ENGINE_UNEXPECTED_ERROR    = 504
	COMMAND_TRANSFERRING_ERROR = 505
	ENGINE_NO_EQUIVALENT       = 604
)

var (
	CommandType               = ""
	Addresses                 []string
	ContractorID              = ""
	ChannelIDOnContractorSide = ""
	Amount                    = ""
	Offset                    = ""
	Count                     = ""
	Equivalent                = ""
	HistoryFrom               = ""
	HistoryTo                 = ""
	AmountFrom                = ""
	AmountTo                  = ""
	CryptoKey                 = ""
	Payload                   = ""
	AuditNumber               = ""
	MaxNegativeBalance        = ""
	MaxPositiveBalance        = ""
	Balance                   = ""
)

type NodeHandler struct {
	// Stores node instances
	Node *Node
}

func InitNodeHandler() (*NodeHandler, error) {
	nodeHandler := &NodeHandler{
		Node: NewNode(),
	}
	return nodeHandler, nil
}

func (nh *NodeHandler) RestoreNode() error {
	ioDirPath := conf.Params.WorkDir

	_, err := os.Stat(ioDirPath)
	if err != nil {
		return wrap("Can't restore node, there is no node folder ", err)
	}

	err = nh.ensureNodeConfigurationIsPresent()
	if err != nil {
		return wrap("Can't restore node, there is no config file", err)
	}

	nh.Node = NewNode()

	if _, err := nh.Node.Start(); err != nil {
		return wrap("Can't start node ", err)
	}

	return nil
}

func (nh *NodeHandler) StartNodeForCommunication() error {
	_, err := os.Stat(conf.Params.WorkDir)
	if err != nil {
		return wrap("Can't check node, there is no node folder ", err)
	}

	nodePID, err := getProcessPID(path.Join(conf.Params.WorkDir, "process.pid"))
	if err != nil {
		return wrap("can't read node PID", err)
	}
	process, err := os.FindProcess(int(nodePID))
	if err != nil {
		return errors.New("can't find node process")
	}
	err = process.Signal(syscall.SIGCHLD)
	if err != nil {
		return errors.New("can't find node process")
	}

	nh.Node = NewNode()

	if _, _, err := nh.Node.StartCommunication(); err != nil {
		return wrap("Can't start node ", err)
	}
	return nil
}

func (nh *NodeHandler) RestoreNodeWithCommunication() error {
	ioDirPath := conf.Params.WorkDir

	_, err := os.Stat(ioDirPath)
	if err != nil {
		return wrap("Can't restore node, there is no node folder ", err)
	}

	err = nh.ensureNodeConfigurationIsPresent()
	if err != nil {
		return wrap("Can't restore node, there is no config file", err)
	}

	nh.Node = NewNode()

	process, err := nh.Node.Start()
	if err != nil {
		return wrap("Can't start node ", err)
	}

	commandsControlEventsChanel, resultsControlEventsChanel, err := nh.Node.StartCommunication()
	if err != nil {
		return wrap("Can't start communication with node ", err)
	}

	go nh.Node.BeginMonitorInternalProcessCrashes(
		process,
		commandsControlEventsChanel,
		resultsControlEventsChanel)

	return nil
}

func (nh *NodeHandler) StopNodeCommunication() error {
	err := nh.Node.StopCommunication()
	if err != nil {
		return wrap("can't stop node communication", err)
	}
	return nil
}

func (nh *NodeHandler) CheckNodeRunning() (bool, error) {
	_, err := os.Stat(conf.Params.WorkDir)
	if err != nil {
		return false, wrap("Can't check node, there is no node folder ", err)
	}

	nodePID, err := getProcessPID(path.Join(conf.Params.WorkDir, "process.pid"))
	if err != nil {
		if os.IsNotExist(err) {
			return false, err
		}
		return false, wrap("Can't read node PID", err)
	}
	process, err := os.FindProcess(int(nodePID))
	if err != nil {
		return false, nil
	}
	err = process.Signal(syscall.SIGCHLD)
	return err == nil, nil
}

// Creates configuration file for the node.
func (nh *NodeHandler) ensureNodeConfigurationIsPresent() error {
	// No automatic node configuration should be done.
	// Original node config must be preserved.
	// Only checking if configuration is present.
	if _, err := os.Stat(path.Join(conf.Params.WorkDir, "conf.json")); os.IsNotExist(err) {
		return wrap("Node doesn't exists.", err)
	}

	return nil
}

func (nh *NodeHandler) IfNodeWaitForResult() bool {
	return len(nh.Node.results) > 0
}

func (nh *NodeHandler) StopNode() error {

	nodePID, err := getProcessPID(path.Join(conf.Params.WorkDir, "process.pid"))
	if err != nil {
		return wrap("Can't read node PID", err)
	}

	process, err := os.FindProcess(int(nodePID))
	if err != nil {
		return wrap("There is no node process with PID "+strconv.Itoa(int(nodePID)), err)
	}

	nh.Node.shouldNotBeRestarted = true

	err = process.Kill()
	if err != nil {
		return wrap("Can't kill node process", err)
	}
	return nil
}

func getProcessPID(pidFileName string) (int, error) {
	pidFile, err := os.OpenFile(pidFileName, os.O_RDONLY, 0600)
	if err != nil {
		if os.IsNotExist(err) {
			return -1, err
		}
		return -1, wrap("Can't open PID file for reading.", err)
	}
	defer pidFile.Close()

	reader := bufio.NewReader(pidFile)
	line, isLinePresent, err := reader.ReadLine()
	if err != nil {
		return -1, wrap("Can't read PID", err)
	}
	if isLinePresent {
		return -1, errors.New("PID is empty")
	}
	pid, err := strconv.Atoi(string(line))
	if err != nil {
		return -1, wrap("events-monitor PID is invalid", err)
	}
	return pid, nil
}

func buildJSONResponse(status int, data interface{}) []byte {
	type Response struct {
		Status int         `json:"status"`
		Data   interface{} `json:"data"`
	}
	response := Response{
		Status: status,
		Data:   data}
	js, err := json.Marshal(response)
	if err != nil {
		logger.Error("Can't marshall data. Details are: " + err.Error())
		return nil
	}
	return js
}

// Shortcut method for the errors wrapping.
func wrap(message string, err error) error {
	return errors.New(message + " -> " + err.Error())
}
