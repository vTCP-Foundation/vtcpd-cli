package routes

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
	"github.com/vTCP-Foundation/vtcpd-cli/internal/handler"
	"github.com/vTCP-Foundation/vtcpd-cli/internal/logger"
)

// --- Global structs for history ---

type SettlementLineHistoryRecord struct {
	TransactionUUID           string `json:"transaction_uuid"`
	UnixTimestampMicroseconds string `json:"unix_timestamp_microseconds"`
	Contractor                string `json:"contractor"`
	OperationDirection        string `json:"operation_direction"`
	Amount                    string `json:"amount"`
}

type PaymentHistoryRecord struct {
	TransactionUUID           string `json:"transaction_uuid"`
	UnixTimestampMicroseconds string `json:"unix_timestamp_microseconds"`
	Contractor                string `json:"contractor"`
	OperationDirection        string `json:"operation_direction"`
	Amount                    string `json:"amount"`
	BalanceAfterOperation     string `json:"balance_after_operation"`
	Payload                   string `json:"payload"`
}

type PaymentAllEquivalentsHistoryRecord struct {
	Equivalent                string `json:"equivalent"`
	TransactionUUID           string `json:"transaction_uuid"`
	UnixTimestampMicroseconds string `json:"unix_timestamp_microseconds"`
	Contractor                string `json:"contractor"`
	OperationDirection        string `json:"operation_direction"`
	Amount                    string `json:"amount"`
	BalanceAfterOperation     string `json:"balance_after_operation"`
	Payload                   string `json:"payload"`
}

type ContractorOperationHistoryRecord struct {
	RecordType                string `json:"record_type"` // "payment" or "trustline"
	TransactionUUID           string `json:"transaction_uuid"`
	UnixTimestampMicroseconds string `json:"unix_timestamp_microseconds"`
	OperationDirection        string `json:"operation_direction"`
	Amount                    string `json:"amount"`
	BalanceAfterOperation     string `json:"balance_after_operation"` // Can be "0" для trustline
	Payload                   string `json:"payload"`                 // Can be "" для trustline
}

type AdditionalPaymentHistoryRecord struct {
	TransactionUUID           string `json:"transaction_uuid"`
	UnixTimestampMicroseconds string `json:"unix_timestamp_microseconds"`
	OperationDirection        string `json:"operation_direction"`
	Amount                    string `json:"amount"`
}

// --- Global API responses for history ---

type SettlementLineHistoryResponse struct {
	Count   int                           `json:"count"`
	Records []SettlementLineHistoryRecord `json:"records"`
}

type PaymentHistoryResponse struct {
	Count   int                    `json:"count"`
	Records []PaymentHistoryRecord `json:"records"`
}

type PaymentAllEquivalentsHistoryResponse struct {
	Count   int                                  `json:"count"`
	Records []PaymentAllEquivalentsHistoryRecord `json:"records"`
}

type ContractorOperationsHistoryResponse struct {
	Count   int                                `json:"count"`
	Records []ContractorOperationHistoryRecord `json:"records"`
}

type AdditionalPaymentHistoryResponse struct {
	Count   int                              `json:"count"`
	Records []AdditionalPaymentHistoryRecord `json:"records"`
}

var (
	HISTORY_RESULT_TIMEOUT uint16 = 20 // seconds
)

func (router *RoutesHandler) SettlementLinesHistory(w http.ResponseWriter, r *http.Request) {
	url, err := preprocessRequest(r)
	if err != nil {
		logger.Error("Bad request: invalid security parameters: " + err.Error())
		w.WriteHeader(BAD_REQUEST)
		return
	}

	offset, isParamPresent := mux.Vars(r)["offset"]
	if !isParamPresent || !handler.ValidateInt(offset) {
		logger.Error("Bad request: invalid offset parameter: " + url)
		w.WriteHeader(BAD_REQUEST)
		return
	}

	count, isParamPresent := mux.Vars(r)["count"]
	if !isParamPresent || !handler.ValidateInt(count) {
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

	command := handler.NewCommand(
		"GET:history/trust-lines", offset, count, dateFromUnixTimestamp, dateToUnixTimestamp, equivalent)

	err = router.nodeHandler.Node.SendCommand(command)
	if err != nil {
		logger.Error("Can't send command: " + string(command.ToBytes()) + " to node. Details: " + err.Error())
		writeHTTPResponse(w, COMMAND_TRANSFERRING_ERROR, SettlementLineHistoryResponse{})
		return
	}

	result, err := router.nodeHandler.Node.GetResult(command, HISTORY_RESULT_TIMEOUT)
	if err != nil {
		logger.Error("Node is inaccessible during processing command " +
			string(command.ToBytes()) + ". Details: " + err.Error())
		writeHTTPResponse(w, NODE_IS_INACCESSIBLE, SettlementLineHistoryResponse{})
		return
	}

	if result.Code != OK && result.Code != ENGINE_NO_EQUIVALENT {
		logger.Error("Node return wrong command result: " + strconv.Itoa(result.Code) +
			" on command: " + string(command.ToBytes()))
		writeHTTPResponse(w, result.Code, SettlementLineHistoryResponse{})
		return
	}
	if result.Code == ENGINE_NO_EQUIVALENT {
		logger.Info("Node hasn't equivalent for command: " + string(command.ToBytes()))
		writeHTTPResponse(w, result.Code, SettlementLineHistoryResponse{})
		return
	}

	if len(result.Tokens) == 0 {
		logger.Error("Node return invalid result tokens size on command: " +
			string(command.ToBytes()))
		writeHTTPResponse(w, ENGINE_UNEXPECTED_ERROR, SettlementLineHistoryResponse{})
		return
	}

	recordsCount, err := strconv.Atoi(result.Tokens[0])
	if err != nil {
		logger.Error("Node return invalid token on command: " + string(command.ToBytes()) +
			". Details: " + err.Error())
		writeHTTPResponse(w, ENGINE_UNEXPECTED_ERROR, SettlementLineHistoryResponse{})
		return
	}

	if recordsCount == 0 {
		writeHTTPResponse(w, OK, SettlementLineHistoryResponse{Count: recordsCount})
		return
	}

	response := SettlementLineHistoryResponse{Count: recordsCount}
	for i := range recordsCount {
		response.Records = append(response.Records, SettlementLineHistoryRecord{
			TransactionUUID:           result.Tokens[i*5+1],
			UnixTimestampMicroseconds: result.Tokens[i*5+2],
			Contractor:                result.Tokens[i*5+3],
			OperationDirection:        result.Tokens[i*5+4],
			Amount:                    result.Tokens[i*5+5],
		})
	}
	writeHTTPResponse(w, OK, response)
}

func (router *RoutesHandler) PaymentsHistory(w http.ResponseWriter, r *http.Request) {
	url, err := preprocessRequest(r)
	if err != nil {
		logger.Error("Bad request: invalid security parameters: " + err.Error())
		w.WriteHeader(BAD_REQUEST)
		return
	}

	offset, isParamPresent := mux.Vars(r)["offset"]
	if !isParamPresent || !handler.ValidateInt(offset) {
		logger.Error("Bad request: invalid offset parameter: " + url)
		w.WriteHeader(BAD_REQUEST)
		return
	}

	count, isParamPresent := mux.Vars(r)["count"]
	if !isParamPresent || !handler.ValidateInt(count) {
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

	// operationUUID
	operationUUID := r.URL.Query().Get("operation_uuid")
	if operationUUID == "" {
		operationUUID = "null"
	}

	equivalent, isParamPresent := mux.Vars(r)["equivalent"]
	if !isParamPresent {
		logger.Error("Bad request: missing equivalent parameter: " + url)
		w.WriteHeader(BAD_REQUEST)
		return
	}

	command := handler.NewCommand(
		"GET:history/payments", offset, count, dateFromUnixTimestamp, dateToUnixTimestamp,
		amountFromUnixTimestamp, amountToUnixTimestamp, commandUUID, operationUUID, equivalent)

	err = router.nodeHandler.Node.SendCommand(command)
	if err != nil {
		logger.Error("Can't send command: " + string(command.ToBytes()) + " to node. Details: " + err.Error())
		writeHTTPResponse(w, COMMAND_TRANSFERRING_ERROR, PaymentHistoryResponse{})
		return
	}

	result, err := router.nodeHandler.Node.GetResult(command, HISTORY_RESULT_TIMEOUT)
	if err != nil {
		logger.Error("Node is inaccessible during processing command: " + string(command.ToBytes()) +
			" . Details: " + err.Error())
		writeHTTPResponse(w, NODE_IS_INACCESSIBLE, PaymentHistoryResponse{})
		return
	}

	if result.Code != OK && result.Code != ENGINE_NO_EQUIVALENT {
		logger.Error("Node return wrong command result: " +
			strconv.Itoa(result.Code) + " on command " + string(command.ToBytes()))
		writeHTTPResponse(w, result.Code, PaymentHistoryResponse{})
		return
	}
	if result.Code == ENGINE_NO_EQUIVALENT {
		logger.Info("Node hasn't equivalent for command: " + string(command.ToBytes()))
		writeHTTPResponse(w, result.Code, PaymentHistoryResponse{})
		return
	}

	if len(result.Tokens) == 0 {
		logger.Error("Node return invalid result tokens size on command: " + string(command.ToBytes()))
		writeHTTPResponse(w, ENGINE_UNEXPECTED_ERROR, PaymentHistoryResponse{})
		return
	}

	recordsCount, err := strconv.Atoi(result.Tokens[0])
	if err != nil {
		logger.Error("Node return invalid token on command: " + string(command.ToBytes()) + ". Details: " + err.Error())
		writeHTTPResponse(w, ENGINE_UNEXPECTED_ERROR, PaymentHistoryResponse{})
		return
	}

	if recordsCount == 0 {
		writeHTTPResponse(w, OK, PaymentHistoryResponse{Count: recordsCount})
		return
	}

	response := PaymentHistoryResponse{Count: recordsCount}
	for i := range recordsCount {
		response.Records = append(response.Records, PaymentHistoryRecord{
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

func (router *RoutesHandler) PaymentsHistoryAllEquivalents(w http.ResponseWriter, r *http.Request) {
	url, err := preprocessRequest(r)
	if err != nil {
		logger.Error("Bad request: invalid security parameters: " + err.Error())
		w.WriteHeader(BAD_REQUEST)
		return
	}

	offset, isParamPresent := mux.Vars(r)["offset"]
	if !isParamPresent || !handler.ValidateInt(offset) {
		logger.Error("Bad request: invalid offset parameter: " + url)
		w.WriteHeader(BAD_REQUEST)
		return
	}

	count, isParamPresent := mux.Vars(r)["count"]
	if !isParamPresent || !handler.ValidateInt(count) {
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

	command := handler.NewCommand(
		"GET:history/payments/all", offset, count, dateFromUnixTimestamp, dateToUnixTimestamp,
		amountFromUnixTimestamp, amountToUnixTimestamp, commandUUID)

	err = router.nodeHandler.Node.SendCommand(command)
	if err != nil {
		logger.Error("Can't send command: " + string(command.ToBytes()) + " to node. Details: " + err.Error())
		writeHTTPResponse(w, COMMAND_TRANSFERRING_ERROR, PaymentAllEquivalentsHistoryResponse{})
		return
	}

	result, err := router.nodeHandler.Node.GetResult(command, HISTORY_RESULT_TIMEOUT)
	if err != nil {
		logger.Error("Node is inaccessible during processing command: " + string(command.ToBytes()) +
			" . Details: " + err.Error())
		writeHTTPResponse(w, NODE_IS_INACCESSIBLE, PaymentAllEquivalentsHistoryResponse{})
		return
	}

	if result.Code != OK {
		logger.Error("Node return wrong command result: " +
			strconv.Itoa(result.Code) + " on command " + string(command.ToBytes()))
		writeHTTPResponse(w, result.Code, PaymentAllEquivalentsHistoryResponse{})
		return
	}

	if len(result.Tokens) == 0 {
		logger.Error("Node return invalid result tokens size on command: " + string(command.ToBytes()))
		writeHTTPResponse(w, ENGINE_UNEXPECTED_ERROR, PaymentAllEquivalentsHistoryResponse{})
		return
	}

	recordsCount, err := strconv.Atoi(result.Tokens[0])
	if err != nil {
		logger.Error("Node return invalid token on command: " + string(command.ToBytes()) + ". Details: " + err.Error())
		writeHTTPResponse(w, ENGINE_UNEXPECTED_ERROR, PaymentAllEquivalentsHistoryResponse{})
		return
	}

	if recordsCount == 0 {
		writeHTTPResponse(w, OK, PaymentAllEquivalentsHistoryResponse{Count: recordsCount})
		return
	}

	response := PaymentAllEquivalentsHistoryResponse{Count: recordsCount}
	for i := range recordsCount {
		response.Records = append(response.Records, PaymentAllEquivalentsHistoryRecord{
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
	writeHTTPResponse(w, OK, response)
}

func (router *RoutesHandler) HistoryWithContractor(w http.ResponseWriter, r *http.Request) {
	url, err := preprocessRequest(r)
	if err != nil {
		logger.Error("Bad request: invalid security parameters: " + err.Error())
		w.WriteHeader(BAD_REQUEST)
		return
	}

	offset, isParamPresent := mux.Vars(r)["offset"]
	if !isParamPresent || !handler.ValidateInt(offset) {
		logger.Error("Bad request: invalid offset parameter: " + url)
		w.WriteHeader(BAD_REQUEST)
		return
	}

	count, isParamPresent := mux.Vars(r)["count"]
	if !isParamPresent || !handler.ValidateInt(count) {
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
	command := handler.NewCommand(contractorAddresses...)

	err = router.nodeHandler.Node.SendCommand(command)
	if err != nil {
		logger.Error("Can't send command: " + string(command.ToBytes()) + " to node. Details: " + err.Error())
		writeHTTPResponse(w, COMMAND_TRANSFERRING_ERROR, ContractorOperationsHistoryResponse{})
		return
	}

	result, err := router.nodeHandler.Node.GetResult(command, HISTORY_RESULT_TIMEOUT)
	if err != nil {
		// Remote node is inaccessible
		logger.Error("Node is inaccessible during processing command: " +
			string(command.ToBytes()) + ". Details: " + err.Error())
		writeHTTPResponse(w, NODE_IS_INACCESSIBLE, ContractorOperationsHistoryResponse{})
		return
	}

	if result.Code != OK && result.Code != ENGINE_NO_EQUIVALENT {
		logger.Error("Node return wrong command result: " + strconv.Itoa(result.Code) +
			" on command: " + string(command.ToBytes()))
		writeHTTPResponse(w, result.Code, ContractorOperationsHistoryResponse{})
		return
	}
	if result.Code == ENGINE_NO_EQUIVALENT {
		logger.Info("Node hasn't equivalent for command: " + string(command.ToBytes()))
		writeHTTPResponse(w, result.Code, ContractorOperationsHistoryResponse{})
		return
	}

	if len(result.Tokens) == 0 {
		logger.Error("Node return invalid result tokens size on command: " + string(command.ToBytes()))
		writeHTTPResponse(w, ENGINE_UNEXPECTED_ERROR, ContractorOperationsHistoryResponse{})
		return
	}

	recordsCount, err := strconv.Atoi(result.Tokens[0])
	if err != nil {
		logger.Error("Node return invalid token on command" + string(command.ToBytes()) +
			". Details: " + err.Error())
		writeHTTPResponse(w, ENGINE_UNEXPECTED_ERROR, ContractorOperationsHistoryResponse{})
		return
	}

	if recordsCount == 0 {
		writeHTTPResponse(w, OK, ContractorOperationsHistoryResponse{Count: recordsCount})
		return
	}

	response := ContractorOperationsHistoryResponse{Count: recordsCount}
	tokenIdx := 1
	for range recordsCount {
		if result.Tokens[tokenIdx] == "payment" {
			response.Records = append(response.Records, ContractorOperationHistoryRecord{
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
			response.Records = append(response.Records, ContractorOperationHistoryRecord{
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

func (router *RoutesHandler) PaymentsAdditionalHistory(w http.ResponseWriter, r *http.Request) {
	url, err := preprocessRequest(r)
	if err != nil {
		logger.Error("Bad request: invalid security parameters: " + err.Error())
		w.WriteHeader(BAD_REQUEST)
		return
	}

	offset, isParamPresent := mux.Vars(r)["offset"]
	if !isParamPresent || !handler.ValidateInt(offset) {
		logger.Error("Bad request: invalid offset parameter: " + url)
		w.WriteHeader(BAD_REQUEST)
		return
	}

	count, isParamPresent := mux.Vars(r)["count"]
	if !isParamPresent || !handler.ValidateInt(count) {
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

	command := handler.NewCommand(
		"GET:history/payments/additional", offset, count, dateFromUnixTimestamp, dateToUnixTimestamp,
		amountFromUnixTimestamp, amountToUnixTimestamp, equivalent)

	err = router.nodeHandler.Node.SendCommand(command)
	if err != nil {
		logger.Error("Can't send command: " + string(command.ToBytes()) + " to node. Details: " + err.Error())
		writeHTTPResponse(w, COMMAND_TRANSFERRING_ERROR, AdditionalPaymentHistoryResponse{})
		return
	}

	result, err := router.nodeHandler.Node.GetResult(command, HISTORY_RESULT_TIMEOUT)
	if err != nil {
		// Remote node is inaccessible
		logger.Error("Node is inaccessible during processing command: " + string(command.ToBytes()) +
			". Details: " + err.Error())
		writeHTTPResponse(w, NODE_IS_INACCESSIBLE, AdditionalPaymentHistoryResponse{})
		return
	}

	if result.Code != OK && result.Code != ENGINE_NO_EQUIVALENT {
		logger.Error("Node return wrong command result: " + strconv.Itoa(result.Code) +
			" on command: " + string(command.ToBytes()))
		writeHTTPResponse(w, result.Code, AdditionalPaymentHistoryResponse{})
		return
	}
	if result.Code == ENGINE_NO_EQUIVALENT {
		logger.Info("Node hasn't equivalent for command: " + string(command.ToBytes()))
		writeHTTPResponse(w, result.Code, AdditionalPaymentHistoryResponse{})
		return
	}

	if len(result.Tokens) == 0 {
		logger.Error("Node return invalid result tokens size on command: " + string(command.ToBytes()))
		writeHTTPResponse(w, ENGINE_UNEXPECTED_ERROR, AdditionalPaymentHistoryResponse{})
		return
	}

	recordsCount, err := strconv.Atoi(result.Tokens[0])
	if err != nil {
		logger.Error("Node return invalid token on command " + string(command.ToBytes()) +
			". Details: " + err.Error())
		writeHTTPResponse(w, ENGINE_UNEXPECTED_ERROR, AdditionalPaymentHistoryResponse{})
		return
	}

	if recordsCount == 0 {
		writeHTTPResponse(w, OK, AdditionalPaymentHistoryResponse{Count: recordsCount})
		return
	}

	response := AdditionalPaymentHistoryResponse{Count: recordsCount}
	tokenIdx := 1
	for range recordsCount {
		response.Records = append(response.Records, AdditionalPaymentHistoryRecord{
			TransactionUUID:           result.Tokens[tokenIdx],
			UnixTimestampMicroseconds: result.Tokens[tokenIdx+1],
			OperationDirection:        result.Tokens[tokenIdx+2],
			Amount:                    result.Tokens[tokenIdx+3],
		})
		tokenIdx += 4
	}
	writeHTTPResponse(w, OK, response)
}
