package main

import (
	"conf"
	"fmt"
	"gopkg.in/alecthomas/kingpin.v2"
	"handler"
	"logger"
	"net/http"
	"os"
	"server"
	"time"
)

var (
	command                   = kingpin.Arg("command", "Command name.").Required().String()
	commandType               = kingpin.Arg("type", "Command type.").Default("").String()
	addresses                 = kingpin.Flag("address", "Contractor address").Default("").Strings()
	contractorID              = kingpin.Flag("contractorID", "Contractor ID").Default("").String()
	channelIDOnContractorSide = kingpin.Flag("channel-id-on-contractor-side",
		"Channel ID on contractor side").Default("").String()
	amount         = kingpin.Flag("amount", "Amount").Default("").String()
	offset         = kingpin.Flag("offset", "Offset of list of requested data.").Default("").String()
	count          = kingpin.Flag("count", "Count requested data.").Default("").String()
	equivalent     = kingpin.Flag("eq", "Equivalent.").Default("").String()
	historyFrom    = kingpin.Flag("history-from", "Lower value of history date.").Default("").String()
	historyTo      = kingpin.Flag("history-to", "Higher value of history date.").Default("").String()
	amountFrom     = kingpin.Flag("amount-from", "Lower value of history amount.").Default("").String()
	amountTo       = kingpin.Flag("amount-to", "Higher value of history amount.").Default("").String()
	cryptoKey      = kingpin.Flag("crypto-key", "Channel crypto key.").Default("").String()
	payload        = kingpin.Flag("payload", "Payload for payment transaction.").Default("").String()
	auditNumber    = kingpin.Flag("audit-number", "Number for audit.").Default("").String()
	incomingAmount = kingpin.Flag("incoming-amount", "Incoming trust amount.").Default("").String()
	outgoingAmount = kingpin.Flag("outgoing-amount", "Outgoing trust amount.").Default("").String()
	balance        = kingpin.Flag("balance", "Trust line balance.").Default("").String()
)

func main() {
	err := conf.LoadSettings()
	if err != nil {
		println("ERROR: Settings can't be loaded." + err.Error())
		os.Exit(1)
	}

	err = logger.Init()
	if err != nil {
		logger.Error("Can't init logger.")
		os.Exit(-1)
	}

	nodesHandler, err := handler.InitNodesHandler()
	if err != nil {
		logger.Error("Can't initialise node handler. Details: " + err.Error())
		os.Exit(-1)
	}

	kingpin.Version("0.0.1")
	kingpin.Parse()

	handler.CommandType = *commandType
	handler.Addresses = *addresses
	handler.ContractorID = *contractorID
	handler.ChannelIDOnContractorSide = *channelIDOnContractorSide
	handler.Amount = *amount
	handler.Offset = *offset
	handler.Count = *count
	handler.Equivalent = *equivalent
	handler.HistoryFrom = *historyFrom
	handler.HistoryTo = *historyTo
	handler.AmountFrom = *amountFrom
	handler.AmountTo = *amountTo
	handler.CryptoKey = *cryptoKey
	handler.Payload = *payload
	handler.AuditNumber = *auditNumber
	handler.IncomingAmount = *incomingAmount
	handler.OutgoingAmount = *outgoingAmount
	handler.Balance = *balance

	if *command == "start" {
		isNodeRunning, err := nodesHandler.CheckNodeRunning()
		if err != nil {
			logger.Error("Can't check if node is running. Details: " + err.Error())
		}
		if isNodeRunning {
			logger.Error("Node already running")
			fmt.Println("Node already running")
			os.Exit(0)
		}
		isEventsMonitorRunning := nodesHandler.CheckEventsMonitoringRunning()
		if isEventsMonitorRunning {
			logger.Info("Events-monitor running. Try stop it")
			err = nodesHandler.StopEventsMonitoring()
			if err != nil {
				logger.Error("Can't stop events-monitor. Details: " + err.Error())
				// todo : need correct reaction
			}
		}
		if conf.Params.Service.SendEvents || conf.Params.Service.SendLogs {
			err := nodesHandler.StartEventsMonitoring()
			if err != nil {
				logger.Error("Can't start events-monitor. Details: " + err.Error())
				// todo : need correct reaction
			}
			logger.Info("events-monitor started")
		} else {
			nodesHandler.ClearEventsMonitoringPID()
		}
		err = nodesHandler.RestoreNode()
		if err != nil {
			logger.Error("Node can't be restored. Error details: " + err.Error())
			fmt.Println("Can't start. " + err.Error())
		} else {
			fmt.Println("Started")
		}
		os.Exit(0)
	} else if *command == "stop" {
		err = nodesHandler.StopNode()
		if err != nil {
			logger.Error("Can't stop node " + err.Error())
			fmt.Println("Can't stop node " + err.Error())
		} else {
			logger.Info("Node stopped")
			fmt.Println("Stopped")
		}
		err = nodesHandler.StopEventsMonitoring()
		if err != nil {
			logger.Error("Can't stop events-monitor. Details: " + err.Error())
		} else {
			logger.Info("Events-monitor stopped")
		}
		nodesHandler.ClearEventsMonitoringPID()
		os.Exit(0)

	} else if *command == "http" {
		err = nodesHandler.StartNodeForCommunication()
		if err != nil {
			logger.Error("Node is not running. Details: " + err.Error())
			fmt.Println("Node is not running. Details: " + err.Error())
			os.Exit(1)
		}
		server.InitNodesHandlerServer(nodesHandler)
		err = http.ListenAndServe(conf.Params.Handler.HTTPInterface(), nil)
		if err != nil {
			logger.Error("Can't start listener. Error details: " + err.Error())
			os.Exit(-1)
		}

	} else if *command == "start-http" {
		err = nodesHandler.RestoreNodeWithCommunication()
		if err != nil {
			logger.Error("Can't start. Details: " + err.Error())
			fmt.Println("Can't start. Details: " + err.Error())
			os.Exit(1)
		}
		server.InitNodesHandlerServer(nodesHandler)
		err = http.ListenAndServe(conf.Params.Handler.HTTPInterface(), nil)
		if err != nil {
			logger.Error("Can't start listener. Error details: " + err.Error())
			os.Exit(-1)
		}

	} else if *command == "channels" {
		err = nodesHandler.StartNodeForCommunication()
		if err != nil {
			logger.Error("Node is not running. Details: " + err.Error())
			fmt.Println("Node is not running. Details: " + err.Error())
			os.Exit(1)
		}
		nodesHandler.Channels()

	} else if *command == "trust-lines" {
		err = nodesHandler.StartNodeForCommunication()
		if err != nil {
			logger.Error("Node is not running. Details: " + err.Error())
			fmt.Println("Node is not running. Details: " + err.Error())
			os.Exit(1)
		}
		nodesHandler.TrustLines()

	} else if *command == "max-flow" {
		err = nodesHandler.StartNodeForCommunication()
		if err != nil {
			logger.Error("Node is not running. Details: " + err.Error())
			fmt.Println("Node is not running. Details: " + err.Error())
			os.Exit(1)
		}
		nodesHandler.MaxFlow()

	} else if *command == "payment" {
		err = nodesHandler.StartNodeForCommunication()
		if err != nil {
			logger.Error("Node is not running. Details: " + err.Error())
			fmt.Println("Node is not running. Details: " + err.Error())
			os.Exit(1)
		}
		nodesHandler.Payment()

	} else if *command == "history" {
		err = nodesHandler.StartNodeForCommunication()
		if err != nil {
			logger.Error("Node is not running. Details: " + err.Error())
			fmt.Println("Node is not running. Details: " + err.Error())
			os.Exit(1)
		}
		nodesHandler.History()

	} else {
		logger.Error("Invalid command " + *command)
		fmt.Println("Invalid command")
		os.Exit(1)
	}

	logger.Info("Handler started")

	if *command != "http" {
		for {
			time.Sleep(time.Millisecond * 10)
			if nodesHandler.IfNodeWaitForResult() {
				continue
			}
			logger.Info("There are no node results")
			err = nodesHandler.StopNodeCommunication()
			if err != nil {
				logger.Error("Can't stop handler")
			}
			break
		}
	}
}
