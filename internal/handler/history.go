package handler

import (
	"fmt"
	"strconv"

	"github.com/vTCP-Foundation/vtcpd-cli/internal/common"
	"github.com/vTCP-Foundation/vtcpd-cli/internal/logger"
)

func (handler *NodeHandler) History() {

	if CommandType == "settlement-lines" {
		handler.settlementLinesHistory()

	} else if CommandType == "payments" {
		handler.paymentsHistory()

	} else if CommandType == "payments-all" {
		handler.paymentsHistoryAllEquivalents()

	} else if CommandType == "additional" {
		handler.additionalHistory()

	} else if CommandType == "with-contractor" {
		handler.contractorOperationsHistory()

	} else {
		logger.Error("Invalid history command " + CommandType)
		fmt.Println("Invalid history command")
		return
	}
}

func (handler *NodeHandler) settlementLinesHistory() {

	if !common.ValidateInt(Offset) {
		logger.Error("Bad request: invalid offset parameter in history settlement-lines request")
		fmt.Println("Bad request: invalid offset parameter")
		return
	}

	if !common.ValidateInt(Count) {
		logger.Error("Bad request: invalid count parameter in history settlement-lines request")
		fmt.Println("Bad request: invalid count parameter")
		return
	}

	if !common.ValidateInt(Equivalent) {
		logger.Error("Bad request: invalid equivalent parameter in history settlement-lines request")
		fmt.Println("Bad request: invalid equivalent parameter")
		return
	}

	dateFromUnixTimestamp := HistoryFrom
	if dateFromUnixTimestamp == "" {
		dateFromUnixTimestamp = "null"
	}

	dateToUnixTimestamp := HistoryTo
	if dateToUnixTimestamp == "" {
		dateToUnixTimestamp = "null"
	}

	command := NewCommand(
		"GET:history/trust-lines", Offset, Count, dateFromUnixTimestamp, dateToUnixTimestamp, Equivalent)

	go handler.settlementLinesHistoryResult(command)
}

func (handler *NodeHandler) settlementLinesHistoryResult(command *Command) {
	err := handler.Node.SendCommand(command)
	if err != nil {
		logger.Error("Can't send command: " + string(command.ToBytes()) + " to node. Details: " + err.Error())
		resultJSON := buildJSONResponse(COMMAND_TRANSFERRING_ERROR, common.SettlementLineHistoryResponse{})
		fmt.Println(string(resultJSON))
		return
	}

	result, err := handler.Node.GetResult(command, common.HISTORY_RESULT_TIMEOUT)
	if err != nil {
		logger.Error("Node is inaccessible during processing command " +
			string(command.ToBytes()) + ". Details: " + err.Error())
		resultJSON := buildJSONResponse(NODE_IS_INACCESSIBLE, common.SettlementLineHistoryResponse{})
		fmt.Println(string(resultJSON))
		return
	}

	if result.Code != OK && result.Code != ENGINE_NO_EQUIVALENT {
		logger.Error("Node return wrong command result: " + strconv.Itoa(result.Code) +
			" on command: " + string(command.ToBytes()))
		resultJSON := buildJSONResponse(result.Code, common.SettlementLineHistoryResponse{})
		fmt.Println(string(resultJSON))
		return
	}
	if result.Code == ENGINE_NO_EQUIVALENT {
		logger.Info("Node hasn't equivalent for command: " + string(command.ToBytes()))
		resultJSON := buildJSONResponse(result.Code, common.SettlementLineHistoryResponse{})
		fmt.Println(string(resultJSON))
		return
	}

	if len(result.Tokens) == 0 {
		logger.Error("Node return invalid result tokens size on command: " +
			string(command.ToBytes()))
		resultJSON := buildJSONResponse(ENGINE_UNEXPECTED_ERROR, common.SettlementLineHistoryResponse{})
		fmt.Println(string(resultJSON))
		return
	}

	recordsCount, err := strconv.Atoi(result.Tokens[0])
	if err != nil {
		logger.Error("Node return invalid token on command: " + string(command.ToBytes()) +
			". Details: " + err.Error())
		resultJSON := buildJSONResponse(ENGINE_UNEXPECTED_ERROR, common.SettlementLineHistoryResponse{})
		fmt.Println(string(resultJSON))
		return
	}

	if recordsCount == 0 {
		resultJSON := buildJSONResponse(OK, common.SettlementLineHistoryResponse{Count: recordsCount})
		fmt.Println(string(resultJSON))
		return
	}

	response := common.SettlementLineHistoryResponse{Count: recordsCount}
	for i := range recordsCount {
		response.Records = append(response.Records, common.SettlementLineHistoryRecord{
			TransactionUUID:           result.Tokens[i*5+1],
			UnixTimestampMicroseconds: result.Tokens[i*5+2],
			Contractor:                result.Tokens[i*5+3],
			OperationDirection:        result.Tokens[i*5+4],
			Amount:                    result.Tokens[i*5+5],
		})
	}
	resultJSON := buildJSONResponse(OK, response)
	fmt.Println(string(resultJSON))
}

func (handler *NodeHandler) paymentsHistory() {

	if !common.ValidateInt(Offset) {
		logger.Error("Bad request: invalid offset parameter in history payments request")
		fmt.Println("Bad request: invalid offset parameter")
		return
	}

	if !common.ValidateInt(Count) {
		logger.Error("Bad request: invalid count parameter in history payments request")
		fmt.Println("Bad request: invalid count parameter")
		return
	}

	if !common.ValidateInt(Equivalent) {
		logger.Error("Bad request: invalid equivalent parameter in history payments request")
		fmt.Println("Bad request: invalid equivalent parameter")
		return
	}

	dateFromUnixTimestamp := HistoryFrom
	if dateFromUnixTimestamp == "" {
		dateFromUnixTimestamp = "null"
	}

	dateToUnixTimestamp := HistoryTo
	if dateToUnixTimestamp == "" {
		dateToUnixTimestamp = "null"
	}

	amountFromUnixTimestamp := AmountFrom
	if amountFromUnixTimestamp == "" {
		amountFromUnixTimestamp = "null"
	} else {
		if !common.ValidateSettlementLineAmount(amountFromUnixTimestamp) {
			logger.Error("Bad request: invalid amount-from parameter in history payments request")
			fmt.Println("Bad request: invalid amount-from parameter")
			return
		}
	}

	amountToUnixTimestamp := AmountTo
	if amountToUnixTimestamp == "" {
		amountToUnixTimestamp = "null"
	} else {
		if !common.ValidateSettlementLineAmount(amountToUnixTimestamp) {
			logger.Error("Bad request: invalid amount-to parameter in history payments request")
			fmt.Println("Bad request: invalid amount-to parameter")
			return
		}
	}

	command := NewCommand(
		"GET:history/payments", Offset, Count, dateFromUnixTimestamp, dateToUnixTimestamp,
		amountFromUnixTimestamp, amountToUnixTimestamp, "null", Equivalent)

	go handler.paymentsHistoryResult(command)
}

func (handler *NodeHandler) paymentsHistoryResult(command *Command) {
	err := handler.Node.SendCommand(command)
	if err != nil {
		logger.Error("Can't send command: " + string(command.ToBytes()) + " to node. Details: " + err.Error())
		resultJSON := buildJSONResponse(COMMAND_TRANSFERRING_ERROR, common.PaymentHistoryResponse{})
		fmt.Println(string(resultJSON))
		return
	}

	result, err := handler.Node.GetResult(command, common.HISTORY_RESULT_TIMEOUT)
	if err != nil {
		logger.Error("Node is inaccessible during processing command: " + string(command.ToBytes()) +
			" . Details: " + err.Error())
		resultJSON := buildJSONResponse(NODE_IS_INACCESSIBLE, common.PaymentHistoryResponse{})
		fmt.Println(string(resultJSON))
		return
	}

	if result.Code != OK && result.Code != ENGINE_NO_EQUIVALENT {
		logger.Error("Node return wrong command result: " +
			strconv.Itoa(result.Code) + " on command " + string(command.ToBytes()))
		resultJSON := buildJSONResponse(result.Code, common.PaymentHistoryResponse{})
		fmt.Println(string(resultJSON))
		return
	}
	if result.Code == ENGINE_NO_EQUIVALENT {
		logger.Info("Node hasn't equivalent for command: " + string(command.ToBytes()))
		resultJSON := buildJSONResponse(result.Code, common.PaymentHistoryResponse{})
		fmt.Println(string(resultJSON))
		return
	}

	if len(result.Tokens) == 0 {
		logger.Error("Node return invalid result tokens size on command: " + string(command.ToBytes()))
		resultJSON := buildJSONResponse(ENGINE_UNEXPECTED_ERROR, common.PaymentHistoryResponse{})
		fmt.Println(string(resultJSON))
		return
	}

	recordsCount, err := strconv.Atoi(result.Tokens[0])
	if err != nil {
		logger.Error("Node return invalid token on command: " + string(command.ToBytes()) + ". Details: " + err.Error())
		resultJSON := buildJSONResponse(ENGINE_UNEXPECTED_ERROR, common.PaymentHistoryResponse{})
		fmt.Println(string(resultJSON))
		return
	}

	if recordsCount == 0 {
		resultJSON := buildJSONResponse(OK, common.PaymentHistoryResponse{Count: recordsCount})
		fmt.Println(string(resultJSON))
		return
	}

	response := common.PaymentHistoryResponse{Count: recordsCount}
	for i := range recordsCount {
		response.Records = append(response.Records, common.PaymentHistoryRecord{
			TransactionUUID:           result.Tokens[i*7+1],
			UnixTimestampMicroseconds: result.Tokens[i*7+2],
			Contractor:                result.Tokens[i*7+3],
			OperationDirection:        result.Tokens[i*7+4],
			Amount:                    result.Tokens[i*7+5],
			BalanceAfterOperation:     result.Tokens[i*7+6],
			Payload:                   result.Tokens[i*7+7],
		})
	}
	resultJSON := buildJSONResponse(OK, response)
	fmt.Println(string(resultJSON))
}

func (handler *NodeHandler) paymentsHistoryAllEquivalents() {

	if !common.ValidateInt(Offset) {
		logger.Error("Bad request: invalid offset parameter in history-all payments request")
		fmt.Println("Bad request: invalid offset parameter")
		return
	}

	if !common.ValidateInt(Count) {
		logger.Error("Bad request: invalid count parameter in history-all payments request")
		fmt.Println("Bad request: invalid count parameter")
		return
	}

	dateFromUnixTimestamp := HistoryFrom
	if dateFromUnixTimestamp == "" {
		dateFromUnixTimestamp = "null"
	}

	dateToUnixTimestamp := HistoryTo
	if dateToUnixTimestamp == "" {
		dateToUnixTimestamp = "null"
	}

	amountFromUnixTimestamp := AmountFrom
	if amountFromUnixTimestamp == "" {
		amountFromUnixTimestamp = "null"
	} else {
		if !common.ValidateSettlementLineAmount(amountFromUnixTimestamp) {
			logger.Error("Bad request: invalid amount-from parameter in history-all payments request")
			fmt.Println("Bad request: invalid amount-from parameter")
			return
		}
	}

	amountToUnixTimestamp := AmountTo
	if amountToUnixTimestamp == "" {
		amountToUnixTimestamp = "null"
	} else {
		if !common.ValidateSettlementLineAmount(amountToUnixTimestamp) {
			logger.Error("Bad request: invalid amount-to parameter in history-all payments request")
			fmt.Println("Bad request: invalid amount-to parameter")
			return
		}
	}

	command := NewCommand(
		"GET:history/payments/all", Offset, Count, dateFromUnixTimestamp, dateToUnixTimestamp,
		amountFromUnixTimestamp, amountToUnixTimestamp, "null")

	go handler.paymentsHistoryAllEquivalentsResult(command)
}

func (handler *NodeHandler) paymentsHistoryAllEquivalentsResult(command *Command) {
	err := handler.Node.SendCommand(command)
	if err != nil {
		logger.Error("Can't send command: " + string(command.ToBytes()) + " to node. Details: " + err.Error())
		resultJSON := buildJSONResponse(COMMAND_TRANSFERRING_ERROR, common.PaymentAllEquivalentsHistoryResponse{})
		fmt.Println(string(resultJSON))
		return
	}

	result, err := handler.Node.GetResult(command, common.HISTORY_RESULT_TIMEOUT)
	if err != nil {
		logger.Error("Node is inaccessible during processing command: " + string(command.ToBytes()) +
			" . Details: " + err.Error())
		resultJSON := buildJSONResponse(NODE_IS_INACCESSIBLE, common.PaymentAllEquivalentsHistoryResponse{})
		fmt.Println(string(resultJSON))
		return
	}

	if result.Code != OK {
		logger.Error("Node return wrong command result: " +
			strconv.Itoa(result.Code) + " on command " + string(command.ToBytes()))
		resultJSON := buildJSONResponse(result.Code, common.PaymentAllEquivalentsHistoryResponse{})
		fmt.Println(string(resultJSON))
		return
	}

	if len(result.Tokens) == 0 {
		logger.Error("Node return invalid result tokens size on command: " + string(command.ToBytes()))
		resultJSON := buildJSONResponse(ENGINE_UNEXPECTED_ERROR, common.PaymentAllEquivalentsHistoryResponse{})
		fmt.Println(string(resultJSON))
		return
	}

	recordsCount, err := strconv.Atoi(result.Tokens[0])
	if err != nil {
		logger.Error("Node return invalid token on command: " + string(command.ToBytes()) + ". Details: " + err.Error())
		resultJSON := buildJSONResponse(ENGINE_UNEXPECTED_ERROR, common.PaymentAllEquivalentsHistoryResponse{})
		fmt.Println(string(resultJSON))
		return
	}

	if recordsCount == 0 {
		resultJSON := buildJSONResponse(OK, common.PaymentAllEquivalentsHistoryResponse{Count: recordsCount})
		fmt.Println(string(resultJSON))
		return
	}

	response := common.PaymentAllEquivalentsHistoryResponse{Count: recordsCount}
	for i := range recordsCount {
		response.Records = append(response.Records, common.PaymentAllEquivalentsHistoryRecord{
			Equivalent:                result.Tokens[i*8+1],
			TransactionUUID:           result.Tokens[i*8+2],
			UnixTimestampMicroseconds: result.Tokens[i*8+3],
			Contractor:                result.Tokens[i*8+4],
			OperationDirection:        result.Tokens[i*8+5],
			Amount:                    result.Tokens[i*8+6],
			BalanceAfterOperation:     result.Tokens[i*8+7],
			Payload:                   result.Tokens[i*8+8],
		})
	}
	resultJSON := buildJSONResponse(OK, response)
	fmt.Println(string(resultJSON))
}

func (handler *NodeHandler) contractorOperationsHistory() {
	if !common.ValidateInt(Offset) {
		logger.Error("Bad request: invalid offset parameter in history with-contractor request")
		fmt.Println("Bad request: invalid offset parameter")
		return
	}

	if !common.ValidateInt(Count) {
		logger.Error("Bad request: invalid count parameter in history with-contractor request")
		fmt.Println("Bad request: invalid count parameter")
		return
	}

	if len(Addresses) == 0 {
		logger.Error("Bad request: there are no contractor addresses parameters in history with-contractor request")
		fmt.Println("Bad request: there are no contractor addresses parameters")
		return
	}

	if !common.ValidateInt(Equivalent) {
		logger.Error("Bad request: invalid equivalent parameter in history with-contractor request")
		fmt.Println("Bad request: invalid equivalent parameter")
		return
	}

	var addresses []string
	for idx := range len(Addresses) {
		addressType, address := common.ValidateAddress(Addresses[idx])
		if addressType == "" {
			logger.Error("Bad request: invalid address parameter in history with-contractor request")
			fmt.Println("Bad request: invalid address parameter")
			return
		}
		addresses = append(addresses, addressType, address)
	}

	addresses = append([]string{Offset, Count, strconv.Itoa(len(Addresses))}, addresses...)
	addresses = append([]string{"GET:history/contractor"}, addresses...)
	addresses = append(addresses, []string{Equivalent}...)
	command := NewCommand(addresses...)

	go handler.contractorOperationsHistoryResult(command)
}

func (handler *NodeHandler) contractorOperationsHistoryResult(command *Command) {
	err := handler.Node.SendCommand(command)
	if err != nil {
		logger.Error("Can't send command: " + string(command.ToBytes()) + " to node. Details: " + err.Error())
		resultJSON := buildJSONResponse(COMMAND_TRANSFERRING_ERROR, common.ContractorOperationsHistoryResponse{})
		fmt.Println(string(resultJSON))
		return
	}

	result, err := handler.Node.GetResult(command, common.HISTORY_RESULT_TIMEOUT)
	if err != nil {
		// Remote node is inaccessible
		logger.Error("Node is inaccessible during processing command: " +
			string(command.ToBytes()) + ". Details: " + err.Error())
		resultJSON := buildJSONResponse(NODE_IS_INACCESSIBLE, common.ContractorOperationsHistoryResponse{})
		fmt.Println(string(resultJSON))
		return
	}

	if result.Code != OK && result.Code != ENGINE_NO_EQUIVALENT {
		logger.Error("Node return wrong command result: " + strconv.Itoa(result.Code) +
			" on command: " + string(command.ToBytes()))
		resultJSON := buildJSONResponse(result.Code, common.ContractorOperationsHistoryResponse{})
		fmt.Println(string(resultJSON))
		return
	}
	if result.Code == ENGINE_NO_EQUIVALENT {
		logger.Info("Node hasn't equivalent for command: " + string(command.ToBytes()))
		resultJSON := buildJSONResponse(result.Code, common.ContractorOperationsHistoryResponse{})
		fmt.Println(string(resultJSON))
		return
	}

	if len(result.Tokens) == 0 {
		logger.Error("Node return invalid result tokens size on command: " + string(command.ToBytes()))
		resultJSON := buildJSONResponse(ENGINE_UNEXPECTED_ERROR, common.ContractorOperationsHistoryResponse{})
		fmt.Println(string(resultJSON))
		return
	}

	recordsCount, err := strconv.Atoi(result.Tokens[0])
	if err != nil {
		logger.Error("Node return invalid token on command" + string(command.ToBytes()) +
			". Details: " + err.Error())
		resultJSON := buildJSONResponse(ENGINE_UNEXPECTED_ERROR, common.ContractorOperationsHistoryResponse{})
		fmt.Println(string(resultJSON))
		return
	}

	if recordsCount == 0 {
		resultJSON := buildJSONResponse(OK, common.ContractorOperationsHistoryResponse{Count: recordsCount})
		fmt.Println(string(resultJSON))
		return
	}

	response := common.ContractorOperationsHistoryResponse{Count: recordsCount}
	tokenIdx := 1
	for range recordsCount {
		if result.Tokens[tokenIdx] == "payment" {
			response.Records = append(response.Records, common.ContractorOperationHistoryRecord{
				RecordType:                result.Tokens[tokenIdx],
				TransactionUUID:           result.Tokens[tokenIdx+1],
				UnixTimestampMicroseconds: result.Tokens[tokenIdx+2],
				OperationDirection:        result.Tokens[tokenIdx+3],
				Amount:                    result.Tokens[tokenIdx+4],
				BalanceAfterOperation:     result.Tokens[tokenIdx+5],
				Payload:                   result.Tokens[tokenIdx+6],
			})
			tokenIdx += 7
		} else if result.Tokens[tokenIdx] == "trustline" {
			response.Records = append(response.Records, common.ContractorOperationHistoryRecord{
				RecordType:                result.Tokens[tokenIdx],
				TransactionUUID:           result.Tokens[tokenIdx+1],
				UnixTimestampMicroseconds: result.Tokens[tokenIdx+2],
				OperationDirection:        result.Tokens[tokenIdx+3],
				Amount:                    result.Tokens[tokenIdx+4],
				BalanceAfterOperation:     "0",
				Payload:                   "",
			})
			tokenIdx += 5
		}
	}
	resultJSON := buildJSONResponse(OK, response)
	fmt.Println(string(resultJSON))
}

func (handler *NodeHandler) additionalHistory() {

	if !common.ValidateInt(Offset) {
		logger.Error("Bad request: invalid offset parameter in history additional request")
		fmt.Println("Bad request: invalid offset parameter")
		return
	}

	if !common.ValidateInt(Count) {
		logger.Error("Bad request: invalid count parameter in history additional request")
		fmt.Println("Bad request: invalid count parameter")
		return
	}

	if !common.ValidateInt(Equivalent) {
		logger.Error("Bad request: invalid equivalent parameter in history additional request")
		fmt.Println("Bad request: invalid equivalent parameter")
		return
	}

	dateFromUnixTimestamp := HistoryFrom
	if dateFromUnixTimestamp == "" {
		dateFromUnixTimestamp = "null"
	}

	dateToUnixTimestamp := HistoryTo
	if dateToUnixTimestamp == "" {
		dateToUnixTimestamp = "null"
	}

	amountFromUnixTimestamp := AmountFrom
	if amountFromUnixTimestamp == "" {
		amountFromUnixTimestamp = "null"
	} else {
		if !common.ValidateSettlementLineAmount(amountFromUnixTimestamp) {
			logger.Error("Bad request: invalid amount-from parameter in history additional request")
			fmt.Println("Bad request: invalid amount-from parameter")
			return
		}
	}

	amountToUnixTimestamp := AmountTo
	if amountToUnixTimestamp == "" {
		amountToUnixTimestamp = "null"
	} else {
		if !common.ValidateSettlementLineAmount(amountToUnixTimestamp) {
			logger.Error("Bad request: invalid amount-to parameter in history additional request")
			fmt.Println("Bad request: invalid amount-to parameter")
			return
		}
	}

	command := NewCommand(
		"GET:history/payments/additional", Offset, Count, dateFromUnixTimestamp, dateToUnixTimestamp,
		amountFromUnixTimestamp, amountToUnixTimestamp, Equivalent)

	go handler.additionalHistoryResult(command)
}

func (handler *NodeHandler) additionalHistoryResult(command *Command) {
	err := handler.Node.SendCommand(command)
	if err != nil {
		logger.Error("Can't send command: " + string(command.ToBytes()) + " to node. Details: " + err.Error())
		resultJSON := buildJSONResponse(COMMAND_TRANSFERRING_ERROR, common.AdditionalPaymentHistoryResponse{})
		fmt.Println(string(resultJSON))
		return
	}

	result, err := handler.Node.GetResult(command, common.HISTORY_RESULT_TIMEOUT)

	if err != nil {
		// Remote node is inaccessible
		logger.Error("Node is inaccessible during processing command: " + string(command.ToBytes()) +
			". Details: " + err.Error())
		resultJSON := buildJSONResponse(NODE_IS_INACCESSIBLE, common.AdditionalPaymentHistoryResponse{})
		fmt.Println(string(resultJSON))
		return
	}

	if result.Code != OK && result.Code != ENGINE_NO_EQUIVALENT {
		logger.Error("Node return wrong command result: " + strconv.Itoa(result.Code) +
			" on command: " + string(command.ToBytes()))
		resultJSON := buildJSONResponse(result.Code, common.AdditionalPaymentHistoryResponse{})
		fmt.Println(string(resultJSON))
		return
	}
	if result.Code == ENGINE_NO_EQUIVALENT {
		logger.Info("Node hasn't equivalent for command: " + string(command.ToBytes()))
		resultJSON := buildJSONResponse(result.Code, common.AdditionalPaymentHistoryResponse{})
		fmt.Println(string(resultJSON))
		return
	}

	if len(result.Tokens) == 0 {
		logger.Error("Node return invalid result tokens size on command: " + string(command.ToBytes()))
		resultJSON := buildJSONResponse(ENGINE_UNEXPECTED_ERROR, common.AdditionalPaymentHistoryResponse{})
		fmt.Println(string(resultJSON))
		return
	}

	recordsCount, err := strconv.Atoi(result.Tokens[0])
	if err != nil {
		logger.Error("Node return invalid token on command " + string(command.ToBytes()) +
			". Details: " + err.Error())
		resultJSON := buildJSONResponse(ENGINE_UNEXPECTED_ERROR, common.AdditionalPaymentHistoryResponse{})
		fmt.Println(string(resultJSON))
		return
	}

	if recordsCount == 0 {
		resultJSON := buildJSONResponse(OK, common.AdditionalPaymentHistoryResponse{Count: recordsCount})
		fmt.Println(string(resultJSON))
		return
	}

	response := common.AdditionalPaymentHistoryResponse{Count: recordsCount}
	tokenIdx := 1
	for range recordsCount {
		response.Records = append(response.Records, common.AdditionalPaymentHistoryRecord{
			TransactionUUID:           result.Tokens[tokenIdx],
			UnixTimestampMicroseconds: result.Tokens[tokenIdx+1],
			OperationDirection:        result.Tokens[tokenIdx+2],
			Amount:                    result.Tokens[tokenIdx+3],
		})
		tokenIdx += 4
	}
	resultJSON := buildJSONResponse(OK, response)
	fmt.Println(string(resultJSON))
}
