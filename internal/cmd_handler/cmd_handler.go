package cmd_handler

import (
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/vTCP-Foundation/vtcpd-cli/internal/conf"
	"github.com/vTCP-Foundation/vtcpd-cli/internal/handler"
	"github.com/vTCP-Foundation/vtcpd-cli/internal/logger"
	"github.com/vTCP-Foundation/vtcpd-cli/internal/routes"
	"github.com/vTCP-Foundation/vtcpd-cli/internal/server"
)

type CommandHandler struct {
	nodeHandler *handler.NodeHandler
}

func NewCommandHandler() (*CommandHandler, error) {
	nodeHandler, err := handler.InitNodeHandler()
	if err != nil {
		return nil, err
	}
	return &CommandHandler{
		nodeHandler: nodeHandler,
	}, nil
}

func (h *CommandHandler) HandleCommand(command string) error {
	switch command {
	case "start":
		return h.nodeHandler.HandleStart()
	case "stop":
		return h.nodeHandler.HandleStop()
	case "http":
		return h.HandleHTTP()
	case "start-http":
		return h.HandleStartHTTP()
	case "channels":
		return h.nodeHandler.HandleChannels()
	case "settlement-lines":
		return h.nodeHandler.HandleSettlementLines()
	case "max-flow":
		return h.nodeHandler.HandleMaxFlow()
	case "payment":
		return h.nodeHandler.HandlePayment()
	case "history":
		return h.nodeHandler.HandleHistory()
	case "remove-outdated-crypto":
		return h.nodeHandler.HandleRemoveOutdatedCrypto()
	default:
		logger.Error("Invalid command " + command)
		fmt.Println("Invalid command")
		os.Exit(1)
		return nil
	}
}

func (h *CommandHandler) WaitForNodeResults() {
	for {
		time.Sleep(time.Millisecond * 10)
		if h.nodeHandler.IfNodeWaitForResult() {
			continue
		}
		logger.Info("There are no node results")
		err := h.nodeHandler.StopNodeCommunication()
		if err != nil {
			logger.Error("Can't stop handler")
		}
		break
	}
}

func (h *CommandHandler) HandleHTTP() error {
	err := h.nodeHandler.StartNodeForCommunication()
	if err != nil {
		logger.Error("Node is not running. Details: " + err.Error())
		fmt.Println("Node is not running. Details: " + err.Error())
		return err
	}
	routesHandler := routes.NewRoutesHandler(h.nodeHandler)
	router := server.InitNodeHandlerServer(routesHandler)
	return http.ListenAndServe(conf.Params.HTTP.HTTPInterface(), router)
}

func (h *CommandHandler) HandleStartHTTP() error {
	isNodeRunning, err := h.nodeHandler.CheckNodeRunning()
	if err != nil {
		logger.Info(err.Error())
		if !os.IsNotExist(err) {
			logger.Error("Can't check if node is running. Details: " + err.Error())
			return err
		}
		logger.Info("No PID file found. Node is not running.")
	}
	if isNodeRunning {
		logger.Error("Node already running")
		fmt.Println("Node already running")
		os.Exit(0)
	}
	err = h.nodeHandler.RestoreNodeWithCommunication()
	if err != nil {
		logger.Error("Can't start. Details: " + err.Error())
		fmt.Println("Can't start. Details: " + err.Error())
		return err
	}
	routesHandler := routes.NewRoutesHandler(h.nodeHandler)
	router := server.InitNodeHandlerServer(routesHandler)
	return http.ListenAndServe(conf.Params.HTTP.HTTPInterface(), router)
}
