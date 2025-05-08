package main

import (
	"fmt"
	"os"

	"github.com/alecthomas/kingpin/v2"
	"github.com/vTCP-Foundation/vtcpd-cli/internal/cmd_handler"
	"github.com/vTCP-Foundation/vtcpd-cli/internal/conf"
	"github.com/vTCP-Foundation/vtcpd-cli/internal/handler"
	"github.com/vTCP-Foundation/vtcpd-cli/internal/logger"
)

var (
	command                   = kingpin.Arg("command", "Command name.").Required().String()
	commandType               = kingpin.Arg("type", "Command type.").Default("").String()
	addresses                 = kingpin.Flag("address", "Contractor address").Default("").Strings()
	contractorID              = kingpin.Flag("contractorID", "Contractor ID").Default("").String()
	channelIDOnContractorSide = kingpin.Flag("channel-id-on-contractor-side", "Channel ID on contractor side").Default("").String()
	amount                    = kingpin.Flag("amount", "Amount").Default("").String()
	offset                    = kingpin.Flag("offset", "Offset of list of requested data.").Default("").String()
	count                     = kingpin.Flag("count", "Count requested data.").Default("").String()
	equivalent                = kingpin.Flag("eq", "Equivalent.").Default("").String()
	historyFrom               = kingpin.Flag("history-from", "Lower value of history date.").Default("").String()
	historyTo                 = kingpin.Flag("history-to", "Higher value of history date.").Default("").String()
	amountFrom                = kingpin.Flag("amount-from", "Lower value of history amount.").Default("").String()
	amountTo                  = kingpin.Flag("amount-to", "Higher value of history amount.").Default("").String()
	cryptoKey                 = kingpin.Flag("crypto-key", "Channel crypto key.").Default("").String()
	payload                   = kingpin.Flag("payload", "Payload for payment transaction.").Default("").String()
	auditNumber               = kingpin.Flag("audit-number", "Number for audit.").Default("").String()
	maxNegativeBalance        = kingpin.Flag("max-negative-balance", "Max negative balance.").Default("").String()
	maxPositiveBalance        = kingpin.Flag("max-positive-balance", "Max positive balance.").Default("").String()
	balance                   = kingpin.Flag("balance", "Settlement line balance.").Default("").String()
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
	handler.MaxNegativeBalance = *maxNegativeBalance
	handler.MaxPositiveBalance = *maxPositiveBalance
	handler.Balance = *balance

	cmdHandler, err := cmd_handler.NewCommandHandlerTesting()
	if err != nil {
		logger.Error("Can't initialise node handler. Details: " + err.Error())
		os.Exit(-1)
	}

	err = cmdHandler.HandleCommand(*command)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	logger.Info("Handler started")

	if *command != "http" {
		cmdHandler.WaitForNodeResults()
	}

}
