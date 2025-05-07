package cmd_handler

import (
	"fmt"
	"net/http"
	"os"

	"github.com/vTCP-Foundation/vtcpd-cli/internal/conf"
	"github.com/vTCP-Foundation/vtcpd-cli/internal/handler"
	"github.com/vTCP-Foundation/vtcpd-cli/internal/logger"
	"github.com/vTCP-Foundation/vtcpd-cli/internal/routes"
	"github.com/vTCP-Foundation/vtcpd-cli/internal/server"
)

type CommandHandlerTes struct {
	nodeHandler *handler.NodeHandler
}

func NewCommandHandlerTes() (*CommandHandlerTes, error) {
	nodeHandler, err := handler.InitNodeHandler()
	if err != nil {
		return nil, err
	}
	handler.CommandType = CommandType
	handler.Addresses = Addresses
	handler.ContractorID = ContractorID
	handler.ChannelIDOnContractorSide = ChannelIDOnContractorSide
	handler.Amount = Amount
	handler.Offset = Offset
	handler.Count = Count
	handler.Equivalent = Equivalent
	handler.HistoryFrom = HistoryFrom
	handler.HistoryTo = HistoryTo
	handler.AmountFrom = AmountFrom
	handler.AmountTo = AmountTo
	handler.CryptoKey = CryptoKey
	handler.Payload = Payload
	handler.AuditNumber = AuditNumber
	handler.MaxNegativeBalance = MaxNegativeBalance
	handler.MaxPositiveBalance = MaxPositiveBalance
	handler.Balance = Balance
	return &CommandHandlerTes{
		nodeHandler: nodeHandler,
	}, nil
}

func (h *CommandHandlerTes) HandleCommand(command string) error {
	switch command {
	case "start":
		return h.nodeHandler.HandleStart()
	case "stop":
		return h.nodeHandler.HandleStop()
	case "http":
		return h.HandleTestHTTP()
	case "start-http":
		return h.HandleTestStartHTTP()
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

func (h *CommandHandlerTes) HandleTestHTTP() error {
	err := h.nodeHandler.StartNodeForCommunication()
	if err != nil {
		logger.Error("Node is not running. Details: " + err.Error())
		fmt.Println("Node is not running. Details: " + err.Error())
		return err
	}
	// todo : add new server
	routesHandler := routes.NewRoutesHandler(h.nodeHandler)
	router := server.InitNodeHandlerServer(routesHandler)
	return http.ListenAndServe(conf.Params.Handler.HTTPInterface(), router)
}

func (h *CommandHandlerTes) HandleTestStartHTTP() error {
	isNodeRunning, err := h.nodeHandler.CheckNodeRunning()
	if err != nil {
		logger.Error("Can't check if node is running. Details: " + err.Error())
		return err
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
	// todo : add new server
	routesHandler := routes.NewRoutesHandler(h.nodeHandler)
	router := server.InitNodeHandlerServer(routesHandler)
	return http.ListenAndServe(conf.Params.Handler.HTTPInterface(), router)
}
