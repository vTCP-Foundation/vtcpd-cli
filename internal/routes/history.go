package routes

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
	"github.com/vTCP-Foundation/vtcpd-cli/internal/common"
	"github.com/vTCP-Foundation/vtcpd-cli/internal/handler"
	"github.com/vTCP-Foundation/vtcpd-cli/internal/logger"
)

func (router *RoutesHandler) SettlementLinesHistory(w http.ResponseWriter, r *http.Request) {
	url, err := preprocessRequest(r)
	if err != nil {
		logger.Error("Bad request: invalid security parameters: " + err.Error())
		w.WriteHeader(BAD_REQUEST)
		return
	}

	offset, isParamPresent := mux.Vars(r)["offset"]
	if !isParamPresent || !common.ValidateInt(offset) {
		logger.Error("Bad request: invalid offset parameter: " + url)
		w.WriteHeader(BAD_REQUEST)
		return
	}

	count, isParamPresent := mux.Vars(r)["count"]
	if !isParamPresent || !common.ValidateInt(count) {
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
		writeHTTPResponse(w, COMMAND_TRANSFERRING_ERROR, common.SettlementLineHistoryResponse{})
		return
	}

	result, err := router.nodeHandler.Node.GetResult(command, common.HISTORY_RESULT_TIMEOUT)
	if err != nil {
		logger.Error("Node is inaccessible during processing command " +
			string(command.ToBytes()) + ". Details: " + err.Error())
		writeHTTPResponse(w, NODE_IS_INACCESSIBLE, common.SettlementLineHistoryResponse{})
		return
	}

	if result.Code != OK && result.Code != ENGINE_NO_EQUIVALENT {
		logger.Error("Node return wrong command result: " + strconv.Itoa(result.Code) +
			" on command: " + string(command.ToBytes()))
		writeHTTPResponse(w, result.Code, common.SettlementLineHistoryResponse{})
		return
	}
	if result.Code == ENGINE_NO_EQUIVALENT {
		logger.Info("Node hasn't equivalent for command: " + string(command.ToBytes()))
		writeHTTPResponse(w, result.Code, common.SettlementLineHistoryResponse{})
		return
	}

	if len(result.Tokens) == 0 {
		logger.Error("Node return invalid result tokens size on command: " +
			string(command.ToBytes()))
		writeHTTPResponse(w, ENGINE_UNEXPECTED_ERROR, common.SettlementLineHistoryResponse{})
		return
	}

	recordsCount, err := strconv.Atoi(result.Tokens[0])
	if err != nil {
		logger.Error("Node return invalid token on command: " + string(command.ToBytes()) +
			". Details: " + err.Error())
		writeHTTPResponse(w, ENGINE_UNEXPECTED_ERROR, common.SettlementLineHistoryResponse{})
		return
	}

	if recordsCount == 0 {
		writeHTTPResponse(w, OK, common.SettlementLineHistoryResponse{Count: recordsCount})
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
	if !isParamPresent || !common.ValidateInt(offset) {
		logger.Error("Bad request: invalid offset parameter: " + url)
		w.WriteHeader(BAD_REQUEST)
		return
	}

	count, isParamPresent := mux.Vars(r)["count"]
	if !isParamPresent || !common.ValidateInt(count) {
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
		writeHTTPResponse(w, COMMAND_TRANSFERRING_ERROR, common.PaymentHistoryResponse{})
		return
	}

	result, err := router.nodeHandler.Node.GetResult(command, common.HISTORY_RESULT_TIMEOUT)
	if err != nil {
		logger.Error("Node is inaccessible during processing command: " + string(command.ToBytes()) +
			" . Details: " + err.Error())
		writeHTTPResponse(w, NODE_IS_INACCESSIBLE, common.PaymentHistoryResponse{})
		return
	}

	if result.Code != OK && result.Code != ENGINE_NO_EQUIVALENT {
		logger.Error("Node return wrong command result: " +
			strconv.Itoa(result.Code) + " on command " + string(command.ToBytes()))
		writeHTTPResponse(w, result.Code, common.PaymentHistoryResponse{})
		return
	}
	if result.Code == ENGINE_NO_EQUIVALENT {
		logger.Info("Node hasn't equivalent for command: " + string(command.ToBytes()))
		writeHTTPResponse(w, result.Code, common.PaymentHistoryResponse{})
		return
	}

	if len(result.Tokens) == 0 {
		logger.Error("Node return invalid result tokens size on command: " + string(command.ToBytes()))
		writeHTTPResponse(w, ENGINE_UNEXPECTED_ERROR, common.PaymentHistoryResponse{})
		return
	}

	recordsCount, err := strconv.Atoi(result.Tokens[0])
	if err != nil {
		logger.Error("Node return invalid token on command: " + string(command.ToBytes()) + ". Details: " + err.Error())
		writeHTTPResponse(w, ENGINE_UNEXPECTED_ERROR, common.PaymentHistoryResponse{})
		return
	}

	if recordsCount == 0 {
		writeHTTPResponse(w, OK, common.PaymentHistoryResponse{Count: recordsCount})
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
	if !isParamPresent || !common.ValidateInt(offset) {
		logger.Error("Bad request: invalid offset parameter: " + url)
		w.WriteHeader(BAD_REQUEST)
		return
	}

	count, isParamPresent := mux.Vars(r)["count"]
	if !isParamPresent || !common.ValidateInt(count) {
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
		writeHTTPResponse(w, COMMAND_TRANSFERRING_ERROR, common.PaymentAllEquivalentsHistoryResponse{})
		return
	}

	result, err := router.nodeHandler.Node.GetResult(command, common.HISTORY_RESULT_TIMEOUT)
	if err != nil {
		logger.Error("Node is inaccessible during processing command: " + string(command.ToBytes()) +
			" . Details: " + err.Error())
		writeHTTPResponse(w, NODE_IS_INACCESSIBLE, common.PaymentAllEquivalentsHistoryResponse{})
		return
	}

	if result.Code != OK {
		logger.Error("Node return wrong command result: " +
			strconv.Itoa(result.Code) + " on command " + string(command.ToBytes()))
		writeHTTPResponse(w, result.Code, common.PaymentAllEquivalentsHistoryResponse{})
		return
	}

	if len(result.Tokens) == 0 {
		logger.Error("Node return invalid result tokens size on command: " + string(command.ToBytes()))
		writeHTTPResponse(w, ENGINE_UNEXPECTED_ERROR, common.PaymentAllEquivalentsHistoryResponse{})
		return
	}

	recordsCount, err := strconv.Atoi(result.Tokens[0])
	if err != nil {
		logger.Error("Node return invalid token on command: " + string(command.ToBytes()) + ". Details: " + err.Error())
		writeHTTPResponse(w, ENGINE_UNEXPECTED_ERROR, common.PaymentAllEquivalentsHistoryResponse{})
		return
	}

	if recordsCount == 0 {
		writeHTTPResponse(w, OK, common.PaymentAllEquivalentsHistoryResponse{Count: recordsCount})
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
	if !isParamPresent || !common.ValidateInt(offset) {
		logger.Error("Bad request: invalid offset parameter: " + url)
		w.WriteHeader(BAD_REQUEST)
		return
	}

	count, isParamPresent := mux.Vars(r)["count"]
	if !isParamPresent || !common.ValidateInt(count) {
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
		writeHTTPResponse(w, COMMAND_TRANSFERRING_ERROR, common.ContractorOperationsHistoryResponse{})
		return
	}

	result, err := router.nodeHandler.Node.GetResult(command, common.HISTORY_RESULT_TIMEOUT)
	if err != nil {
		// Remote node is inaccessible
		logger.Error("Node is inaccessible during processing command: " +
			string(command.ToBytes()) + ". Details: " + err.Error())
		writeHTTPResponse(w, NODE_IS_INACCESSIBLE, common.ContractorOperationsHistoryResponse{})
		return
	}

	if result.Code != OK && result.Code != ENGINE_NO_EQUIVALENT {
		logger.Error("Node return wrong command result: " + strconv.Itoa(result.Code) +
			" on command: " + string(command.ToBytes()))
		writeHTTPResponse(w, result.Code, common.ContractorOperationsHistoryResponse{})
		return
	}
	if result.Code == ENGINE_NO_EQUIVALENT {
		logger.Info("Node hasn't equivalent for command: " + string(command.ToBytes()))
		writeHTTPResponse(w, result.Code, common.ContractorOperationsHistoryResponse{})
		return
	}

	if len(result.Tokens) == 0 {
		logger.Error("Node return invalid result tokens size on command: " + string(command.ToBytes()))
		writeHTTPResponse(w, ENGINE_UNEXPECTED_ERROR, common.ContractorOperationsHistoryResponse{})
		return
	}

	recordsCount, err := strconv.Atoi(result.Tokens[0])
	if err != nil {
		logger.Error("Node return invalid token on command" + string(command.ToBytes()) +
			". Details: " + err.Error())
		writeHTTPResponse(w, ENGINE_UNEXPECTED_ERROR, common.ContractorOperationsHistoryResponse{})
		return
	}

	if recordsCount == 0 {
		writeHTTPResponse(w, OK, common.ContractorOperationsHistoryResponse{Count: recordsCount})
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
	if !isParamPresent || !common.ValidateInt(offset) {
		logger.Error("Bad request: invalid offset parameter: " + url)
		w.WriteHeader(BAD_REQUEST)
		return
	}

	count, isParamPresent := mux.Vars(r)["count"]
	if !isParamPresent || !common.ValidateInt(count) {
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
		writeHTTPResponse(w, COMMAND_TRANSFERRING_ERROR, common.AdditionalPaymentHistoryResponse{})
		return
	}

	result, err := router.nodeHandler.Node.GetResult(command, common.HISTORY_RESULT_TIMEOUT)
	if err != nil {
		// Remote node is inaccessible
		logger.Error("Node is inaccessible during processing command: " + string(command.ToBytes()) +
			". Details: " + err.Error())
		writeHTTPResponse(w, NODE_IS_INACCESSIBLE, common.AdditionalPaymentHistoryResponse{})
		return
	}

	if result.Code != OK && result.Code != ENGINE_NO_EQUIVALENT {
		logger.Error("Node return wrong command result: " + strconv.Itoa(result.Code) +
			" on command: " + string(command.ToBytes()))
		writeHTTPResponse(w, result.Code, common.AdditionalPaymentHistoryResponse{})
		return
	}
	if result.Code == ENGINE_NO_EQUIVALENT {
		logger.Info("Node hasn't equivalent for command: " + string(command.ToBytes()))
		writeHTTPResponse(w, result.Code, common.AdditionalPaymentHistoryResponse{})
		return
	}

	if len(result.Tokens) == 0 {
		logger.Error("Node return invalid result tokens size on command: " + string(command.ToBytes()))
		writeHTTPResponse(w, ENGINE_UNEXPECTED_ERROR, common.AdditionalPaymentHistoryResponse{})
		return
	}

	recordsCount, err := strconv.Atoi(result.Tokens[0])
	if err != nil {
		logger.Error("Node return invalid token on command " + string(command.ToBytes()) +
			". Details: " + err.Error())
		writeHTTPResponse(w, ENGINE_UNEXPECTED_ERROR, common.AdditionalPaymentHistoryResponse{})
		return
	}

	if recordsCount == 0 {
		writeHTTPResponse(w, OK, common.AdditionalPaymentHistoryResponse{Count: recordsCount})
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
	writeHTTPResponse(w, OK, response)
}
