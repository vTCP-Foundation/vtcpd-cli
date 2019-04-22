package handler

import (
	"fmt"
	"github.com/gorilla/mux"
	"logger"
	"net/http"
	"strconv"
	"strings"
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

func (handler *NodesHandler) trustLinesHistory() {

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
		TransactionUUID           string `json:"transaction_uuid"`
		UnixTimestampMicroseconds string `json:"unix_timestamp_microseconds"`
		Contractor                string `json:"contractor"`
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

func (handler *NodesHandler) TrustLinesHistory(w http.ResponseWriter, r *http.Request) {
	url := logRequest(r)

	offset, isParamPresent := mux.Vars(r)["offset"]
	if !isParamPresent || !ValidateInt(offset) {
		logger.Error("Bad request: invalid offset parameter: " + url)
		w.WriteHeader(BAD_REQUEST)
		return
	}

	count, isParamPresent := mux.Vars(r)["count"]
	if !isParamPresent || !ValidateInt(count) {
		logger.Error("Bad request: invalid count parameter: " + url)
		w.WriteHeader(BAD_REQUEST)
		return
	}

	dateFromUnixTimestamp := r.URL.Query().Get("date_from")
	if dateFromUnixTimestamp == "" {
		dateFromUnixTimestamp = "null"
	}

	dateToUnixTimestamp := r.URL.Query().Get("date_to")
	if dateToUnixTimestamp == "" {
		dateToUnixTimestamp = "null"
	}

	equivalent, isParamPresent := mux.Vars(r)["equivalent"]
	if !isParamPresent {
		logger.Error("Bad request: missing equivalent parameter: " + url)
		w.WriteHeader(BAD_REQUEST)
		return
	}

	command := NewCommand(
		"GET:history/trust-lines", offset, count, dateFromUnixTimestamp, dateToUnixTimestamp, equivalent)

	type Record struct {
		TransactionUUID           string `json:"transaction_uuid"`
		UnixTimestampMicroseconds string `json:"unix_timestamp_microseconds"`
		Contractor                string `json:"contractor"`
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
		writeHTTPResponse(w, COMMAND_TRANSFERRING_ERROR, Response{})
		return
	}

	result, err := handler.node.GetResult(command, HISTORY_RESULT_TIMEOUT)
	if err != nil {
		logger.Error("Node is inaccessible during processing command " +
			string(command.ToBytes()) + ". Details: " + err.Error())
		writeHTTPResponse(w, NODE_IS_INACCESSIBLE, Response{})
		return
	}

	if result.Code != OK && result.Code != ENGINE_NO_EQUIVALENT {
		logger.Error("Node return wrong command result: " + strconv.Itoa(result.Code) +
			" on command: " + string(command.ToBytes()))
		writeHTTPResponse(w, result.Code, Response{})
		return
	}
	if result.Code == ENGINE_NO_EQUIVALENT {
		logger.Info("Node hasn't equivalent for command: " + string(command.ToBytes()))
		writeHTTPResponse(w, result.Code, Response{})
		return
	}

	if len(result.Tokens) == 0 {
		logger.Error("Node return invalid result tokens size on command: " +
			string(command.ToBytes()))
		writeHTTPResponse(w, ENGINE_UNEXPECTED_ERROR, Response{})
		return
	}

	recordsCount, err := strconv.Atoi(result.Tokens[0])
	if err != nil {
		logger.Error("Node return invalid token on command: " + string(command.ToBytes()) +
			". Details: " + err.Error())
		writeHTTPResponse(w, ENGINE_UNEXPECTED_ERROR, Response{})
		return
	}

	if recordsCount == 0 {
		writeHTTPResponse(w, OK, Response{Count: recordsCount})
		return
	}

	response := Response{Count: recordsCount}
	for i := 0; i < recordsCount; i++ {
		response.Records = append(response.Records, Record{
			TransactionUUID:           result.Tokens[i*5+1],
			UnixTimestampMicroseconds: result.Tokens[i*5+2],
			Contractor:                result.Tokens[i*5+3],
			OperationDirection:        result.Tokens[i*5+4],
			Amount:                    result.Tokens[i*5+5],
		})
	}
	writeHTTPResponse(w, OK, response)
}

func (handler *NodesHandler) paymentsHistory() {

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
		TransactionUUID           string `json:"transaction_uuid"`
		UnixTimestampMicroseconds string `json:"unix_timestamp_microseconds"`
		Contractor                string `json:"contractor"`
		OperationDirection        string `json:"operation_direction"`
		Amount                    string `json:"amount"`
		BalanceAfterOperation     string `json:"balance_after_operation"`
		Payload                   string `json:"payload"`
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

func (handler *NodesHandler) PaymentsHistory(w http.ResponseWriter, r *http.Request) {
	url := logRequest(r)

	offset, isParamPresent := mux.Vars(r)["offset"]
	if !isParamPresent || !ValidateInt(offset) {
		logger.Error("Bad request: invalid offset parameter: " + url)
		w.WriteHeader(BAD_REQUEST)
		return
	}

	count, isParamPresent := mux.Vars(r)["count"]
	if !isParamPresent || !ValidateInt(count) {
		logger.Error("Bad request: invalid count parameter: " + url)
		w.WriteHeader(BAD_REQUEST)
		return
	}

	dateFromUnixTimestamp := r.URL.Query().Get("date_from")
	if dateFromUnixTimestamp == "" {
		dateFromUnixTimestamp = "null"
	}

	dateToUnixTimestamp := r.URL.Query().Get("date_to")
	if dateToUnixTimestamp == "" {
		dateToUnixTimestamp = "null"
	}

	// Amount from
	amountFromUnixTimestamp := r.URL.Query().Get("amount_from")
	if amountFromUnixTimestamp == "" {
		amountFromUnixTimestamp = "null"
	}

	// Amount to
	amountToUnixTimestamp := r.URL.Query().Get("amount_to")
	if amountToUnixTimestamp == "" {
		amountToUnixTimestamp = "null"
	}

	// commandUUID
	commandUUID := r.URL.Query().Get("command_uuid")
	if commandUUID == "" {
		commandUUID = "null"
	}

	equivalent, isParamPresent := mux.Vars(r)["equivalent"]
	if !isParamPresent {
		logger.Error("Bad request: missing equivalent parameter: " + url)
		w.WriteHeader(BAD_REQUEST)
		return
	}

	command := NewCommand(
		"GET:history/payments", offset, count, dateFromUnixTimestamp, dateToUnixTimestamp,
		amountFromUnixTimestamp, amountToUnixTimestamp, commandUUID, equivalent)

	type Record struct {
		TransactionUUID           string `json:"transaction_uuid"`
		UnixTimestampMicroseconds string `json:"unix_timestamp_microseconds"`
		Contractor                string `json:"contractor"`
		OperationDirection        string `json:"operation_direction"`
		Amount                    string `json:"amount"`
		BalanceAfterOperation     string `json:"balance_after_operation"`
		Payload                   string `json:"payload"`
	}

	type Response struct {
		Count   int      `json:"count"`
		Records []Record `json:"records"`
	}

	err := handler.node.SendCommand(command)
	if err != nil {
		logger.Error("Can't send command: " + string(command.ToBytes()) + " to node. Details: " + err.Error())
		writeHTTPResponse(w, COMMAND_TRANSFERRING_ERROR, Response{})
		return
	}

	result, err := handler.node.GetResult(command, HISTORY_RESULT_TIMEOUT)
	if err != nil {
		logger.Error("Node is inaccessible during processing command: " + string(command.ToBytes()) +
			" . Details: " + err.Error())
		writeHTTPResponse(w, NODE_IS_INACCESSIBLE, Response{})
		return
	}

	if result.Code != OK && result.Code != ENGINE_NO_EQUIVALENT {
		logger.Error("Node return wrong command result: " +
			strconv.Itoa(result.Code) + " on command " + string(command.ToBytes()))
		writeHTTPResponse(w, result.Code, Response{})
		return
	}
	if result.Code == ENGINE_NO_EQUIVALENT {
		logger.Info("Node hasn't equivalent for command: " + string(command.ToBytes()))
		writeHTTPResponse(w, result.Code, Response{})
		return
	}

	if len(result.Tokens) == 0 {
		logger.Error("Node return invalid result tokens size on command: " + string(command.ToBytes()))
		writeHTTPResponse(w, ENGINE_UNEXPECTED_ERROR, Response{})
		return
	}

	recordsCount, err := strconv.Atoi(result.Tokens[0])
	if err != nil {
		logger.Error("Node return invalid token on command: " + string(command.ToBytes()) + ". Details: " + err.Error())
		writeHTTPResponse(w, ENGINE_UNEXPECTED_ERROR, Response{})
		return
	}

	if recordsCount == 0 {
		writeHTTPResponse(w, OK, Response{Count: recordsCount})
		return
	}

	response := Response{Count: recordsCount}
	for i := 0; i < recordsCount; i++ {
		response.Records = append(response.Records, Record{
			TransactionUUID:           result.Tokens[i*7+1],
			UnixTimestampMicroseconds: result.Tokens[i*7+2],
			Contractor:                result.Tokens[i*7+3],
			OperationDirection:        result.Tokens[i*7+4],
			Amount:                    result.Tokens[i*7+5],
			BalanceAfterOperation:     result.Tokens[i*7+6],
			Payload:                   result.Tokens[i*7+7],
		})
	}
	writeHTTPResponse(w, OK, response)
}

func (handler *NodesHandler) contractorOperationsHistory() {
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
	for idx := 0; idx < len(Addresses); idx++ {
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
		Payload                   string `json:"payload"`
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
				Payload:                   result.Tokens[tokenIdx+6],
			})
			tokenIdx += 7
		} else if result.Tokens[tokenIdx] == "trustline" {
			response.Records = append(response.Records, Record{
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

func (handler *NodesHandler) HistoryWithContractor(w http.ResponseWriter, r *http.Request) {
	url := logRequest(r)

	offset, isParamPresent := mux.Vars(r)["offset"]
	if !isParamPresent || !ValidateInt(offset) {
		logger.Error("Bad request: invalid offset parameter: " + url)
		w.WriteHeader(BAD_REQUEST)
		return
	}

	count, isParamPresent := mux.Vars(r)["count"]
	if !isParamPresent || !ValidateInt(count) {
		logger.Error("Bad request: invalid count parameter: " + url)
		w.WriteHeader(BAD_REQUEST)
		return
	}

	contractorAddresses := []string{}
	for key, values := range r.URL.Query() {
		if key != "contractor_address" {
			continue
		}
		for _, value := range values {
			typeAndAddress := strings.Split(value, "-")
			contractorAddresses = append(contractorAddresses, typeAndAddress[0])
			contractorAddresses = append(contractorAddresses, typeAndAddress[1])
		}
		break
	}
	if len(contractorAddresses) == 0 {
		logger.Error("Bad request: there are no contractor_addresses parameters: " + url)
		w.WriteHeader(BAD_REQUEST)
		return
	}

	equivalent, isParamPresent := mux.Vars(r)["equivalent"]
	if !isParamPresent {
		logger.Error("Bad request: missing equivalent parameter: " + url)
		w.WriteHeader(BAD_REQUEST)
		return
	}

	// Command generation
	contractorAddresses = append([]string{offset, count, strconv.Itoa(len(contractorAddresses) / 2)}, contractorAddresses...)
	contractorAddresses = append([]string{"GET:history/contractor"}, contractorAddresses...)
	contractorAddresses = append(contractorAddresses, []string{equivalent}...)
	command := NewCommand(contractorAddresses...)

	type Record struct {
		RecordType                string `json:"record_type"`
		TransactionUUID           string `json:"transaction_uuid"`
		UnixTimestampMicroseconds string `json:"unix_timestamp_microseconds"`
		OperationDirection        string `json:"operation_direction"`
		Amount                    string `json:"amount"`
		BalanceAfterOperation     string `json:"balance_after_operation"`
		Payload                   string `json:"payload"`
	}

	type Response struct {
		Count   int      `json:"count"`
		Records []Record `json:"records"`
	}

	err := handler.node.SendCommand(command)
	if err != nil {
		logger.Error("Can't send command: " + string(command.ToBytes()) + " to node. Details: " + err.Error())
		writeHTTPResponse(w, COMMAND_TRANSFERRING_ERROR, Response{})
		return
	}

	result, err := handler.node.GetResult(command, HISTORY_RESULT_TIMEOUT)
	if err != nil {
		// Remote node is inaccessible
		logger.Error("Node is inaccessible during processing command: " +
			string(command.ToBytes()) + ". Details: " + err.Error())
		writeHTTPResponse(w, NODE_IS_INACCESSIBLE, Response{})
		return
	}

	if result.Code != OK && result.Code != ENGINE_NO_EQUIVALENT {
		logger.Error("Node return wrong command result: " + strconv.Itoa(result.Code) +
			" on command: " + string(command.ToBytes()))
		writeHTTPResponse(w, result.Code, Response{})
		return
	}
	if result.Code == ENGINE_NO_EQUIVALENT {
		logger.Info("Node hasn't equivalent for command: " + string(command.ToBytes()))
		writeHTTPResponse(w, result.Code, Response{})
		return
	}

	if len(result.Tokens) == 0 {
		logger.Error("Node return invalid result tokens size on command: " + string(command.ToBytes()))
		writeHTTPResponse(w, ENGINE_UNEXPECTED_ERROR, Response{})
		return
	}

	recordsCount, err := strconv.Atoi(result.Tokens[0])
	if err != nil {
		logger.Error("Node return invalid token on command" + string(command.ToBytes()) +
			". Details: " + err.Error())
		writeHTTPResponse(w, ENGINE_UNEXPECTED_ERROR, Response{})
		return
	}

	if recordsCount == 0 {
		writeHTTPResponse(w, OK, Response{Count: recordsCount})
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
				Payload:                   result.Tokens[tokenIdx+6],
			})
			tokenIdx += 7
		} else if result.Tokens[tokenIdx] == "trustline" {
			response.Records = append(response.Records, Record{
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
	writeHTTPResponse(w, OK, response)
}

func (handler *NodesHandler) additionalHistory() {

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

func (handler *NodesHandler) PaymentsAdditionalHistory(w http.ResponseWriter, r *http.Request) {
	url := logRequest(r)

	offset, isParamPresent := mux.Vars(r)["offset"]
	if !isParamPresent || !ValidateInt(offset) {
		logger.Error("Bad request: invalid offset parameter: " + url)
		w.WriteHeader(BAD_REQUEST)
		return
	}

	count, isParamPresent := mux.Vars(r)["count"]
	if !isParamPresent || !ValidateInt(count) {
		logger.Error("Bad request: invalid count parameter: " + url)
		w.WriteHeader(BAD_REQUEST)
		return
	}

	dateFromUnixTimestamp := r.URL.Query().Get("date_from")
	if dateFromUnixTimestamp == "" {
		dateFromUnixTimestamp = "null"
	}

	dateToUnixTimestamp := r.URL.Query().Get("date_to")
	if dateToUnixTimestamp == "" {
		dateToUnixTimestamp = "null"
	}

	// Amount from
	amountFromUnixTimestamp := r.URL.Query().Get("amount_from")
	if amountFromUnixTimestamp == "" {
		amountFromUnixTimestamp = "null"
	}

	// Amount to
	amountToUnixTimestamp := r.URL.Query().Get("amount_to")
	if amountToUnixTimestamp == "" {
		amountToUnixTimestamp = "null"
	}

	equivalent, isParamPresent := mux.Vars(r)["equivalent"]
	if !isParamPresent {
		logger.Error("Bad request: missing equivalent parameter: " + url)
		w.WriteHeader(BAD_REQUEST)
		return
	}

	command := NewCommand(
		"GET:history/payments/additional", offset, count, dateFromUnixTimestamp, dateToUnixTimestamp,
		amountFromUnixTimestamp, amountToUnixTimestamp, equivalent)

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
		writeHTTPResponse(w, COMMAND_TRANSFERRING_ERROR, Response{})
		return
	}

	result, err := handler.node.GetResult(command, HISTORY_RESULT_TIMEOUT)
	if err != nil {
		// Remote node is inaccessible
		logger.Error("Node is inaccessible during processing command: " + string(command.ToBytes()) +
			". Details: " + err.Error())
		writeHTTPResponse(w, NODE_IS_INACCESSIBLE, Response{})
		return
	}

	if result.Code != OK && result.Code != ENGINE_NO_EQUIVALENT {
		logger.Error("Node return wrong command result: " + strconv.Itoa(result.Code) +
			" on command: " + string(command.ToBytes()))
		writeHTTPResponse(w, result.Code, Response{})
		return
	}
	if result.Code == ENGINE_NO_EQUIVALENT {
		logger.Info("Node hasn't equivalent for command: " + string(command.ToBytes()))
		writeHTTPResponse(w, result.Code, Response{})
		return
	}

	if len(result.Tokens) == 0 {
		logger.Error("Node return invalid result tokens size on command: " + string(command.ToBytes()))
		writeHTTPResponse(w, ENGINE_UNEXPECTED_ERROR, Response{})
		return
	}

	recordsCount, err := strconv.Atoi(result.Tokens[0])
	if err != nil {
		logger.Error("Node return invalid token on command " + string(command.ToBytes()) +
			". Details: " + err.Error())
		writeHTTPResponse(w, ENGINE_UNEXPECTED_ERROR, Response{})
		return
	}

	if recordsCount == 0 {
		writeHTTPResponse(w, OK, Response{Count: recordsCount})
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
	writeHTTPResponse(w, OK, response)
}
