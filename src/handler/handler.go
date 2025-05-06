package handler

import (
	"bufio"
	"encoding/json"
	"errors"
	"io"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path"
	"strconv"
	"strings"
	"syscall"

	"github.com/vTCP-Foundation/vtcpd-cli/src/conf"
	"github.com/vTCP-Foundation/vtcpd-cli/src/logger"
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

type NodesHandler struct {
	// Stores node instances
	node *Node
}

func InitNodesHandler() (*NodesHandler, error) {
	nodesHandler := &NodesHandler{
		node: NewNode(),
	}
	return nodesHandler, nil
}

func (handler *NodesHandler) RestoreNode() error {
	ioDirPath := conf.Params.Handler.NodeDirPath

	_, err := os.Stat(ioDirPath)
	if err != nil {
		return wrap("Can't restore node, there is no node folder ", err)
	}

	err = handler.ensureNodeConfigurationIsPresent()
	if err != nil {
		return wrap("Can't restore node, there is no config file", err)
	}

	handler.node = NewNode()

	if _, err := handler.node.Start(); err != nil {
		return wrap("Can't start node ", err)
	}

	return nil
}

func (handler *NodesHandler) StartNodeForCommunication() error {
	_, err := os.Stat(conf.Params.Handler.NodeDirPath)
	if err != nil {
		return wrap("Can't check node, there is no node folder ", err)
	}

	nodePID, err := getProcessPID(path.Join(conf.Params.Handler.NodeDirPath, "process.pid"))
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

	handler.node = NewNode()

	if _, _, err := handler.node.StartCommunication(); err != nil {
		return wrap("Can't start node ", err)
	}
	return nil
}

func (handler *NodesHandler) RestoreNodeWithCommunication() error {
	ioDirPath := conf.Params.Handler.NodeDirPath

	_, err := os.Stat(ioDirPath)
	if err != nil {
		return wrap("Can't restore node, there is no node folder ", err)
	}

	err = handler.ensureNodeConfigurationIsPresent()
	if err != nil {
		return wrap("Can't restore node, there is no config file", err)
	}

	handler.node = NewNode()

	process, err := handler.node.Start()
	if err != nil {
		return wrap("Can't start node ", err)
	}

	commandsControlEventsChanel, resultsControlEventsChanel, err := handler.node.StartCommunication()
	if err != nil {
		return wrap("Can't start communication with node ", err)
	}

	go handler.node.BeginMonitorInternalProcessCrashes(
		process,
		commandsControlEventsChanel,
		resultsControlEventsChanel)

	return nil
}

func (handler *NodesHandler) StopNodeCommunication() error {
	// Nodes map changes must be sequential,
	// otherwise - there will be a non-zero probability of memory corruption.
	err := handler.node.StopCommunication()
	if err != nil {
		return wrap("can't stop node communication", err)
	}
	return nil
}

func (handler *NodesHandler) CheckNodeRunning() (bool, error) {
	_, err := os.Stat(conf.Params.Handler.NodeDirPath)
	if err != nil {
		return false, wrap("Can't check node, there is no node folder ", err)
	}

	nodePID, err := getProcessPID(path.Join(conf.Params.Handler.NodeDirPath, "process.pid"))
	if err != nil {
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
func (handler *NodesHandler) ensureNodeConfigurationIsPresent() error {
	// No automatic node configuration should be done.
	// Original node config must be preserved.
	// Only checking if configuration is present.
	if _, err := os.Stat(path.Join(conf.Params.Handler.NodeDirPath, "conf.json")); os.IsNotExist(err) {
		return wrap("Node doesn't exists.", err)
	}

	return nil
}

func (handler *NodesHandler) IfNodeWaitForResult() bool {
	return len(handler.node.results) > 0
}

func (handler *NodesHandler) StopNode() error {

	nodePID, err := getProcessPID(path.Join(conf.Params.Handler.NodeDirPath, "process.pid"))
	if err != nil {
		return wrap("Can't read node PID", err)
	}

	process, err := os.FindProcess(int(nodePID))
	if err != nil {
		return wrap("There is no node process with PID "+strconv.Itoa(int(nodePID)), err)
	}

	handler.node.shouldNotBeRestarted = true

	err = process.Kill()
	if err != nil {
		return wrap("Can't kill node process", err)
	}
	return nil
}

func (handler *NodesHandler) StartEventsMonitoring() error {

	// Starting child process.
	process := exec.Command(conf.Params.Service.EventsMonitorExecutableFullPath)
	process.Dir = path.Join(conf.Params.Handler.NodeDirPath, "..")

	err := process.Start()
	if err != nil {
		return wrap("Can't start events monitor process", err)
	}

	pidFile, err := os.OpenFile("events-monitor.pid", os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0777)
	if err != nil {
		process.Process.Kill()
		return wrap("Can't open PID file for writing.", err)
	}
	defer pidFile.Close()

	writer := bufio.NewWriter(pidFile)
	_, err = writer.WriteString(strconv.Itoa(process.Process.Pid))
	if err != nil {
		process.Process.Kill()
		return wrap("Can't write events-monitor PID", err)
	}
	writer.Flush()

	process.Process.Release()

	// It seems, that child process started well.
	// Now the process descriptor must be transferred to the top, for further control.
	return nil
}

func (handler *NodesHandler) StopEventsMonitoring() error {
	eventsMonitorPID, err := getProcessPID("events-monitor.pid")
	if err != nil {
		return wrap("Can't read events-monitor PID", err)
	}

	process, err := os.FindProcess(int(eventsMonitorPID))
	if err != nil {
		return wrap("There is no process with PID "+strconv.Itoa(int(eventsMonitorPID)), err)
	}

	err = process.Kill()
	if err != nil {
		return wrap("Can't kill events-monitor process", err)
	}
	return nil
}

func (handler *NodesHandler) ClearEventsMonitoringPID() {
	pidFile, _ := os.OpenFile("events-monitor.pid", os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0777)
	pidFile.Close()
}

func (handler *NodesHandler) CheckEventsMonitoringRunning() bool {
	eventsMonitorPID, err := getProcessPID("events-monitor.pid")
	if err != nil {
		return false
	}

	process, err := os.FindProcess(int(eventsMonitorPID))
	if err != nil {
		return false
	}
	err = process.Signal(syscall.SIGCHLD)
	return err == nil
}

func getProcessPID(pidFileName string) (int, error) {
	pidFile, err := os.OpenFile(pidFileName, os.O_RDONLY, 0600)
	if err != nil {
		return -1, wrap("Can't open PID file for reading.", err)
	}
	defer pidFile.Close()

	reader := bufio.NewReader(pidFile)
	line, isLinePresent, err := reader.ReadLine()
	if err != nil {
		return -1, wrap("Can't read events-monitor PID", err)
	}
	if isLinePresent {
		return -1, errors.New("events-monitor PID is empty")
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

func writeHTTPResponse(w http.ResponseWriter, statusCode int, data interface{}) {
	w.WriteHeader(statusCode)
	writeJSONResponse(data, w)
}

func writeJSONResponse(data interface{}, w http.ResponseWriter) {
	type Response struct {
		Data interface{} `json:"data"`
	}

	response := Response{Data: data}
	js, err := json.Marshal(response)
	if err != nil {
		logger.Error("Can't marshall data. Details are: " + err.Error())
		writeServerError("JSON forming error", w)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(js)
}

func writeServerError(message string, w http.ResponseWriter) {
	w.WriteHeader(SERVER_ERROR)
	w.Header().Set("Content-Type", "application/json")

	content := make(map[string]string)
	content["error"] = message

	js, _ := json.Marshal(content)
	w.Write(js)
}

func preprocessRequest(r *http.Request) (string, error) {
	url := ""
	if r.Method == "GET" {
		url = r.Method + ": " + r.URL.String()
	} else {
		bodyBytes, _ := io.ReadAll(r.Body)
		url = r.Method + ": " + r.URL.String() + "{ " + string(bodyBytes) + "}"
	}
	logger.Info(url)
	requesterIP := getRealAddr(r)
	logger.Info("Requester IP: " + requesterIP)
	if len(conf.Params.Security.AllowableIPs) > 0 {
		ipIsAllow := false
		for _, allowableIP := range conf.Params.Security.AllowableIPs {
			if allowableIP == requesterIP {
				ipIsAllow = true
				break
			}
		}
		if !ipIsAllow {
			return url, errors.New("IP " + requesterIP + " is not allow")
		}
	}
	apiKey := r.Header.Get("api-key")
	if conf.Params.Security.ApiKey != "" {
		if apiKey != conf.Params.Security.ApiKey {
			return url, errors.New("Invalid api-key " + apiKey)
		}
	}
	return url, nil
}

func getRealAddr(r *http.Request) string {
	remoteIP := ""
	// the default is the originating ip. but we try to find better options because this is almost
	// never the right IP
	if parts := strings.Split(r.RemoteAddr, ":"); len(parts) == 2 {
		remoteIP = parts[0]
	}
	// If we have a forwarded-for header, take the address from there
	if xff := strings.Trim(r.Header.Get("X-Forwarded-For"), ","); len(xff) > 0 {
		addrs := strings.Split(xff, ",")
		lastFwd := addrs[len(addrs)-1]
		if ip := net.ParseIP(lastFwd); ip != nil {
			remoteIP = ip.String()
		}
		// parse X-Real-Ip header
	} else if xri := r.Header.Get("X-Real-Ip"); len(xri) > 0 {
		if ip := net.ParseIP(xri); ip != nil {
			remoteIP = ip.String()
		}
	}

	return remoteIP
}

// Shortcut method for the errors wrapping.
func wrap(message string, err error) error {
	return errors.New(message + " -> " + err.Error())
}
