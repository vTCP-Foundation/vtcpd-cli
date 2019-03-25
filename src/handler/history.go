package handler

import (
	"logger"
	"strconv"
	"fmt"
)

var (
	HISTORY_RESULT_TIMEOUT uint16 = 20 // seconds
)

func (handler *NodesHandler) History() {

	if CommandType == "trust-lines" {
		handler.trustLinesHistory()

	} else if CommandType == "payments" {
		handler.paymentsHistory()

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

func (handler *NodesHandler)trustLinesHistory() {

	if !ValidateInt(Offset) {
		logger.Error("Bad request: invalid offset parameter in history trust-lines request")
		fmt.Println("Bad request: invalid offset parameter")
		return
	}

	if !ValidateInt(Count) {
		logger.Error("Bad request: invalid count parameter in history trust-lines request")
		fmt.Println("Bad request: invalid count parameter")
		return
	}

	if !ValidateInt(Equivalent) {
		logger.Error("Bad request: invalid equivalent parameter in history trust-lines request")
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

	go handler.trustLinesHistoryResult(command)
}

func (handler *NodesHandler) trustLinesHistoryResult(command *Command) {
	type Record struct {
		TransactionUUID				string	`json:"transaction_uuid"`
		UnixTimestampMicroseconds	string	`json:"unix_timestamp_microseconds"`
		Contractor					string	`json:"contractor"`
		OperationDirection			string	`json:"operation_direction"`
		Amount						string	`json:"amount"`
	}

	type Response struct {
		Count   int      `json:"count"`
		Records []Record `json:"records"`
	}

	err := handler.node.SendCommand(command)
	if err != nil {
		logger.Error("Can't send command: " + string(command.ToBytes()) + " to node. Details: " + err.Error())
		resultJSON := buildJSONResponse(COMMAND_TRANSFERRING_ERROR, Response{})
		fmt.Println(string(resultJSON))
		return
	}

	result, err := handler.node.GetResult(command, HISTORY_RESULT_TIMEOUT)
	if err != nil {
		logger.Error("Node is inaccessible during processing command " +
			string(command.ToBytes()) + ". Details: " + err.Error())
		resultJSON := buildJSONResponse(NODE_IS_INACCESSIBLE, Response{})
		fmt.Println(string(resultJSON))
		return
	}

	if result.Code != OK && result.Code != ENGINE_NO_EQUIVALENT {
		logger.Error("Node return wrong command result: " + strconv.Itoa(result.Code) +
			" on command: " + string(command.ToBytes()))
		resultJSON := buildJSONResponse(result.Code, Response{})
		fmt.Println(string(resultJSON))
		return
	}
	if result.Code == ENGINE_NO_EQUIVALENT {
		logger.Info("Node hasn't equivalent for command: " + string(command.ToBytes()))
		resultJSON := buildJSONResponse(result.Code, Response{})
		fmt.Println(string(resultJSON))
		return
	}

	if len(result.Tokens) == 0 {
		logger.Error("Node return invalid result tokens size on command: " +
			string(command.ToBytes()))
		resultJSON := buildJSONResponse(ENGINE_UNEXPECTED_ERROR, Response{})
		fmt.Println(string(resultJSON))
		return
	}

	recordsCount, err := strconv.Atoi(result.Tokens[0])
	if err != nil {
		logger.Error("Node return invalid token on command: " + string(command.ToBytes()) +
			". Details: " + err.Error())
		resultJSON := buildJSONResponse(ENGINE_UNEXPECTED_ERROR, Response{})
		fmt.Println(string(resultJSON))
		return
	}

	if recordsCount == 0 {
		resultJSON := buildJSONResponse(OK, Response{Count: recordsCount})
		fmt.Println(string(resultJSON))
		return
	}

	response := Response{Count: recordsCount}
	for i := 0; i < recordsCount; i++ {
		response.Records = append(response.Records, Record{
			TransactionUUID:			result.Tokens[i*5+1],
			UnixTimestampMicroseconds:	result.Tokens[i*5+2],
			Contractor:					result.Tokens[i*5+3],
			OperationDirection:			result.Tokens[i*5+4],
			Amount:						result.Tokens[i*5+5],
		})
	}
	resultJSON := buildJSONResponse(OK, response)
	fmt.Println(string(resultJSON))
}

func (handler *NodesHandler)paymentsHistory() {

	if !ValidateInt(Offset) {
		logger.Error("Bad request: invalid offset parameter in history payments request")
		fmt.Println("Bad request: invalid offset parameter")
		return
	}

	if !ValidateInt(Count) {
		logger.Error("Bad request: invalid count parameter in history payments request")
		fmt.Println("Bad request: invalid count parameter")
		return
	}

	if !ValidateInt(Equivalent) {
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
		if !ValidateTrustLineAmount(amountFromUnixTimestamp) {
			logger.Error("Bad request: invalid amount-from parameter in history payments request")
			fmt.Println("Bad request: invalid amount-from parameter")
			return
		}
	}

	amountToUnixTimestamp := AmountTo
	if amountToUnixTimestamp == "" {
		amountToUnixTimestamp = "null"
	} else {
		if !ValidateTrustLineAmount(amountToUnixTimestamp) {
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

func (handler *NodesHandler) paymentsHistoryResult(command *Command) {
	type Record struct {
		TransactionUUID				string 	`json:"transaction_uuid"`
		UnixTimestampMicroseconds	string 	`json:"unix_timestamp_microseconds"`
		Contractor					string 	`json:"contractor"`
		OperationDirection			string 	`json:"operation_direction"`
		Amount						string 	`json:"amount"`
		BalanceAfterOperation		string 	`json:"balance_after_operation"`
	}

	type Response struct {
		Count   int      `json:"count"`
		Records []Record `json:"records"`
	}

	err := handler.node.SendCommand(command)
	if err != nil {
		logger.Error("Can't send command: " + string(command.ToBytes()) + " to node. Details: " + err.Error())
		resultJSON := buildJSONResponse(COMMAND_TRANSFERRING_ERROR, Response{})
		fmt.Println(string(resultJSON))
		return
	}

	result, err := handler.node.GetResult(command, HISTORY_RESULT_TIMEOUT)
	if err != nil {
		logger.Error("Node is inaccessible during processing command: " + string(command.ToBytes()) +
			" . Details: " + err.Error())
		resultJSON := buildJSONResponse(NODE_IS_INACCESSIBLE, Response{})
		fmt.Println(string(resultJSON))
		return
	}

	if result.Code != OK && result.Code != ENGINE_NO_EQUIVALENT {
		logger.Error("Node return wrong command result: " +
			strconv.Itoa(result.Code) + " on command " + string(command.ToBytes()))
		resultJSON := buildJSONResponse(result.Code, Response{})
		fmt.Println(string(resultJSON))
		return
	}
	if result.Code == ENGINE_NO_EQUIVALENT {
		logger.Info("Node hasn't equivalent for command: " + string(command.ToBytes()))
		resultJSON := buildJSONResponse(result.Code, Response{})
		fmt.Println(string(resultJSON))
		return
	}

	if len(result.Tokens) == 0 {
		logger.Error("Node return invalid result tokens size on command: " + string(command.ToBytes()))
		resultJSON := buildJSONResponse(ENGINE_UNEXPECTED_ERROR, Response{})
		fmt.Println(string(resultJSON))
		return
	}

	recordsCount, err := strconv.Atoi(result.Tokens[0])
	if err != nil {
		logger.Error("Node return invalid token on command: " + string(command.ToBytes()) + ". Details: " + err.Error())
		resultJSON := buildJSONResponse(ENGINE_UNEXPECTED_ERROR, Response{})
		fmt.Println(string(resultJSON))
		return
	}

	if recordsCount == 0 {
		resultJSON := buildJSONResponse(OK, Response{Count: recordsCount})
		fmt.Println(string(resultJSON))
		return
	}

	response := Response{Count: recordsCount}
	for i := 0; i < recordsCount; i++ {
		response.Records = append(response.Records, Record{
			TransactionUUID:			result.Tokens[i*6+1],
			UnixTimestampMicroseconds:	result.Tokens[i*6+2],
			Contractor:					result.Tokens[i*6+3],
			OperationDirection:			result.Tokens[i*6+4],
			Amount:						result.Tokens[i*6+5],
			BalanceAfterOperation:		result.Tokens[i*6+6],
		})
	}
	resultJSON := buildJSONResponse(OK, response)
	fmt.Println(string(resultJSON))
}

func (handler *NodesHandler)contractorOperationsHistory() {
	if !ValidateInt(Offset) {
		logger.Error("Bad request: invalid offset parameter in history with-contractor request")
		fmt.Println("Bad request: invalid offset parameter")
		return
	}

	if !ValidateInt(Count) {
		logger.Error("Bad request: invalid count parameter in history with-contractor request")
		fmt.Println("Bad request: invalid count parameter")
		return
	}

	if len(Addresses) == 0 {
		logger.Error("Bad request: there are no contractor addresses parameters in history with-contractor request")
		fmt.Println("Bad request: there are no contractor addresses parameters")
		return
	}

	if !ValidateInt(Equivalent) {
		logger.Error("Bad request: invalid equivalent parameter in history with-contractor request")
		fmt.Println("Bad request: invalid equivalent parameter")
		return
	}

	var addresses []string
	for idx:= 0; idx < len(Addresses); idx++ {
		addressType, address := ValidateAddress(Addresses[idx])
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

func (handler *NodesHandler) contractorOperationsHistoryResult(command *Command) {
	type Record struct {
		RecordType                string `json:"record_type"`
		TransactionUUID           string `json:"transaction_uuid"`
		UnixTimestampMicroseconds string `json:"unix_timestamp_microseconds"`
		OperationDirection        string `json:"operation_direction"`
		Amount                    string `json:"amount"`
		BalanceAfterOperation     string `json:"balance_after_operation"`
	}

	type Response struct {
		Count   int      `json:"count"`
		Records []Record `json:"records"`
	}

	err := handler.node.SendCommand(command)
	if err != nil {
		logger.Error("Can't send command: " + string(command.ToBytes()) + " to node. Details: " + err.Error())
		resultJSON := buildJSONResponse(COMMAND_TRANSFERRING_ERROR, Response{})
		fmt.Println(string(resultJSON))
		return
	}

	result, err := handler.node.GetResult(command, HISTORY_RESULT_TIMEOUT)
	if err != nil {
		// Remote node is inaccessible
		logger.Error("Node is inaccessible during processing command: " +
			string(command.ToBytes()) + ". Details: " + err.Error())
		resultJSON := buildJSONResponse(NODE_IS_INACCESSIBLE, Response{})
		fmt.Println(string(resultJSON))
		return
	}

	if result.Code != OK && result.Code != ENGINE_NO_EQUIVALENT {
		logger.Error("Node return wrong command result: " + strconv.Itoa(result.Code) +
			" on command: " + string(command.ToBytes()))
		resultJSON := buildJSONResponse(result.Code, Response{})
		fmt.Println(string(resultJSON))
		return
	}
	if result.Code == ENGINE_NO_EQUIVALENT {
		logger.Info("Node hasn't equivalent for command: " + string(command.ToBytes()))
		resultJSON := buildJSONResponse(result.Code, Response{})
		fmt.Println(string(resultJSON))
		return
	}

	if len(result.Tokens) == 0 {
		logger.Error("Node return invalid result tokens size on command: " + string(command.ToBytes()))
		resultJSON := buildJSONResponse(ENGINE_UNEXPECTED_ERROR, Response{})
		fmt.Println(string(resultJSON))
		return
	}

	recordsCount, err := strconv.Atoi(result.Tokens[0])
	if err != nil {
		logger.Error("Node return invalid token on command" + string(command.ToBytes()) +
			". Details: " + err.Error())
		resultJSON := buildJSONResponse(ENGINE_UNEXPECTED_ERROR, Response{})
		fmt.Println(string(resultJSON))
		return
	}

	if recordsCount == 0 {
		resultJSON := buildJSONResponse(OK, Response{Count: recordsCount})
		fmt.Println(string(resultJSON))
		return
	}

	response := Response{Count: recordsCount}
	tokenIdx := 1
	for i := 0; i < recordsCount; i++ {
		if result.Tokens[tokenIdx] == "payment" {
			response.Records = append(response.Records, Record{
				RecordType:                result.Tokens[tokenIdx],
				TransactionUUID:           result.Tokens[tokenIdx+1],
				UnixTimestampMicroseconds: result.Tokens[tokenIdx+2],
				OperationDirection:        result.Tokens[tokenIdx+3],
				Amount:                    result.Tokens[tokenIdx+4],
				BalanceAfterOperation:     result.Tokens[tokenIdx+5],
			})
			tokenIdx += 6
		} else if result.Tokens[tokenIdx] == "trustline" {
			response.Records = append(response.Records, Record{
				RecordType:                result.Tokens[tokenIdx],
				TransactionUUID:           result.Tokens[tokenIdx+1],
				UnixTimestampMicroseconds: result.Tokens[tokenIdx+2],
				OperationDirection:        result.Tokens[tokenIdx+3],
				Amount:                    result.Tokens[tokenIdx+4],
				BalanceAfterOperation:     "0",
			})
			tokenIdx += 5
		}
	}
	resultJSON := buildJSONResponse(OK, response)
	fmt.Println(string(resultJSON))
}

func (handler *NodesHandler)additionalHistory() {

	if !ValidateInt(Offset) {
		logger.Error("Bad request: invalid offset parameter in history additional request")
		fmt.Println("Bad request: invalid offset parameter")
		return
	}

	if !ValidateInt(Count) {
		logger.Error("Bad request: invalid count parameter in history additional request")
		fmt.Println("Bad request: invalid count parameter")
		return
	}

	if !ValidateInt(Equivalent) {
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
		if !ValidateTrustLineAmount(amountFromUnixTimestamp) {
			logger.Error("Bad request: invalid amount-from parameter in history additional request")
			fmt.Println("Bad request: invalid amount-from parameter")
			return
		}
	}

	amountToUnixTimestamp := AmountTo
	if amountToUnixTimestamp == "" {
		amountToUnixTimestamp = "null"
	} else {
		if !ValidateTrustLineAmount(amountToUnixTimestamp) {
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

func (handler *NodesHandler) additionalHistoryResult(command *Command) {
	type Record struct {
		TransactionUUID           string `json:"transaction_uuid"`
		UnixTimestampMicroseconds string `json:"unix_timestamp_microseconds"`
		OperationDirection        string `json:"operation_direction"`
		Amount                    string `json:"amount"`
	}

	type Response struct {
		Count   int      `json:"count"`
		Records []Record `json:"records"`
	}

	err := handler.node.SendCommand(command)
	if err != nil {
		logger.Error("Can't send command: " + string(command.ToBytes()) + " to node. Details: " + err.Error())
		resultJSON := buildJSONResponse(COMMAND_TRANSFERRING_ERROR, Response{})
		fmt.Println(string(resultJSON))
		return
	}

	result, err := handler.node.GetResult(command, HISTORY_RESULT_TIMEOUT)

	if err != nil {
		// Remote node is inaccessible
		logger.Error("Node is inaccessible during processing command: " + string(command.ToBytes()) +
			". Details: " + err.Error())
		resultJSON := buildJSONResponse(NODE_IS_INACCESSIBLE, Response{})
		fmt.Println(string(resultJSON))
		return
	}

	if result.Code != OK && result.Code != ENGINE_NO_EQUIVALENT {
		logger.Error("Node return wrong command result: " + strconv.Itoa(result.Code) +
			" on command: " + string(command.ToBytes()))
		resultJSON := buildJSONResponse(result.Code, Response{})
		fmt.Println(string(resultJSON))
		return
	}
	if result.Code == ENGINE_NO_EQUIVALENT {
		logger.Info("Node hasn't equivalent for command: " + string(command.ToBytes()))
		resultJSON := buildJSONResponse(result.Code, Response{})
		fmt.Println(string(resultJSON))
		return
	}

	if len(result.Tokens) == 0 {
		logger.Error("Node return invalid result tokens size on command: " + string(command.ToBytes()))
		resultJSON := buildJSONResponse(ENGINE_UNEXPECTED_ERROR, Response{})
		fmt.Println(string(resultJSON))
		return
	}

	recordsCount, err := strconv.Atoi(result.Tokens[0])
	if err != nil {
		logger.Error("Node return invalid token on command " + string(command.ToBytes()) +
			". Details: " + err.Error())
		resultJSON := buildJSONResponse(ENGINE_UNEXPECTED_ERROR, Response{})
		fmt.Println(string(resultJSON))
		return
	}

	if recordsCount == 0 {
		resultJSON := buildJSONResponse(OK, Response{Count: recordsCount})
		fmt.Println(string(resultJSON))
		return
	}

	response := Response{Count: recordsCount}
	tokenIdx := 1
	for i := 0; i < recordsCount; i++ {
		response.Records = append(response.Records, Record{
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