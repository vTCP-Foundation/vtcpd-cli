package routes

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/vTCP-Foundation/vtcpd-cli/internal/common"
	"github.com/vTCP-Foundation/vtcpd-cli/internal/handler"
	"github.com/vTCP-Foundation/vtcpd-cli/internal/logger"
)

func (router *RoutesHandler) BatchMaxFullyTransaction(w http.ResponseWriter, r *http.Request) {
	url, err := preprocessRequest(r)
	if err != nil {
		logger.Error("Bad request: invalid security parameters: " + err.Error())
		w.WriteHeader(BAD_REQUEST)
		return
	}

	equivalent, isParamPresent := mux.Vars(r)["equivalent"]
	if !isParamPresent {
		logger.Error("Bad request: missing equivalent parameter: " + url)
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

	// Command generation
	contractorAddresses = append([]string{strconv.Itoa(len(contractorAddresses) / 2)}, contractorAddresses...)
	contractorAddresses = append([]string{"GET:contractors/transactions/max/fully"}, contractorAddresses...)
	contractorAddresses = append(contractorAddresses, []string{equivalent}...)
	command := handler.NewCommand(contractorAddresses...)

	err = router.nodeHandler.Node.SendCommand(command)
	if err != nil {
		logger.Error("Can't send command: " + string(command.ToBytes()) + " to node. Details: " + err.Error())
		writeHTTPResponse(w, COMMAND_TRANSFERRING_ERROR, common.MaxFlowResponse{})
		return
	}

	// Command processing.
	// This command may execute relatively slow.
	// Timeout is set to little bit greater value to be able to handle this.
	result, err := router.nodeHandler.Node.GetResult(command, common.MAX_FLOW_FULLY_TIMEOUT)
	if err != nil {
		logger.Error("Node is inaccessible during processing command: " +
			string(command.ToBytes()) + ". Details: " + err.Error())
		writeHTTPResponse(w, NODE_IS_INACCESSIBLE, common.MaxFlowResponse{})
		return
	}

	if result.Code != OK && result.Code != ENGINE_NO_EQUIVALENT {
		logger.Error("Node return wrong command result: " + strconv.Itoa(result.Code) +
			" on command: " + string(command.ToBytes()))
		writeHTTPResponse(w, result.Code, common.MaxFlowResponse{})
		return
	}
	if result.Code == ENGINE_NO_EQUIVALENT {
		logger.Info("Node hasn't equivalent for command: " + string(command.ToBytes()))
		writeHTTPResponse(w, result.Code, common.MaxFlowResponse{})
		return
	}

	if len(result.Tokens) == 0 {
		logger.Error("Node return invalid result tokens size on command: " + string(command.ToBytes()))
		writeHTTPResponse(w, ENGINE_UNEXPECTED_ERROR, common.MaxFlowResponse{})
		return
	}

	contractorsCount, err := strconv.Atoi(result.Tokens[0])
	if err != nil {
		logger.Error("Node return invalid token on command: " + string(command.ToBytes()) +
			". Details: " + err.Error())
		writeHTTPResponse(w, ENGINE_UNEXPECTED_ERROR, common.MaxFlowResponse{})
		return
	}

	if contractorsCount == 0 {
		writeHTTPResponse(w, OK, common.MaxFlowResponse{Count: 0})
		return
	}

	response := common.MaxFlowResponse{Count: contractorsCount}
	for i := range contractorsCount {
		response.Records = append(response.Records, common.MaxFlowRecord{
			ContractorAddressType: result.Tokens[i*3+1],
			ContractorAddress:     result.Tokens[i*3+2],
			MaxAmount:             result.Tokens[i*3+3],
		})
	}
	writeHTTPResponse(w, OK, response)
}

func (router *RoutesHandler) CreateTransaction(w http.ResponseWriter, r *http.Request) {
	url, err := preprocessRequest(r)
	if err != nil {
		logger.Error("Bad request: invalid security parameters: " + err.Error())
		w.WriteHeader(BAD_REQUEST)
		return
	}

	equivalent, isParamPresent := mux.Vars(r)["equivalent"]
	if !isParamPresent {
		logger.Error("Bad request: missing equivalent parameter: " + url)
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

	amount := r.FormValue("amount")
	if !common.ValidateSettlementLineAmount(amount) {
		logger.Error("Bad request: invalid amount parameter: " + url)
		w.WriteHeader(BAD_REQUEST)
		return
	}

	payload := r.FormValue("payload")

	transactionUUIDStr := r.FormValue("transaction_uuid")
	var transactionUUID uuid.UUID
	if transactionUUIDStr != "" {
		transactionUUID, err = uuid.Parse(transactionUUIDStr)
		if err != nil {
			logger.Error("Bad request: invalid transaction_uuid parameter: " + url)
			w.WriteHeader(BAD_REQUEST)
			return
		}
	}

	// Command processing.
	// This command may execute relatively slow.
	contractorAddresses = append([]string{strconv.Itoa(len(contractorAddresses) / 2)}, contractorAddresses...)
	contractorAddresses = append([]string{"CREATE:contractors/transactions"}, contractorAddresses...)
	contractorAddresses = append(contractorAddresses, []string{amount, equivalent}...)
	if payload != "" {
		contractorAddresses = append(contractorAddresses, []string{payload}...)
	}

	var command *handler.Command
	if transactionUUIDStr == "" {
		command = handler.NewCommand(contractorAddresses...)
	} else {
		command = handler.NewCommandWithUUID(transactionUUID, contractorAddresses...)
	}

	err = router.nodeHandler.Node.SendCommand(command)
	if err != nil {
		logger.Error("Can't send command: " + string(command.ToBytes()) + " to node. Details: " + err.Error())
		writeHTTPResponse(w, COMMAND_TRANSFERRING_ERROR, common.PaymentResponse{})
		return
	}

	result, err := router.nodeHandler.Node.GetResult(command, common.PAYMENT_OPERATION_TIMEOUT)
	if err != nil {
		logger.Error("Node is inaccessible during processing command: " +
			string(command.ToBytes()) + ". Details: " + err.Error())
		writeHTTPResponse(w, NODE_IS_INACCESSIBLE, common.PaymentResponse{})
		return
	}

	if result.Code != CREATED && result.Code != ENGINE_NO_EQUIVALENT {
		logger.Error("Node return wrong command result: " + strconv.Itoa(result.Code) +
			" on command: " + string(command.ToBytes()))
		writeHTTPResponse(w, result.Code, common.PaymentResponse{})
		return
	}
	if result.Code == ENGINE_NO_EQUIVALENT {
		logger.Info("Node hasn't equivalent for command: " + string(command.ToBytes()))
		writeHTTPResponse(w, result.Code, common.PaymentResponse{})
		return
	}

	if len(result.Tokens) == 0 {
		logger.Error("Node return invalid result tokens size on command: " + string(command.ToBytes()))
		writeHTTPResponse(w, ENGINE_UNEXPECTED_ERROR, common.PaymentResponse{})
		return
	}

	writeHTTPResponse(w, OK, common.PaymentResponse{TransactionUUID: result.Tokens[0]})
}

func (router *RoutesHandler) GetTransactionByCommandUUID(w http.ResponseWriter, r *http.Request) {
	url, err := preprocessRequest(r)
	if err != nil {
		logger.Error("Bad request: invalid security parameters: " + err.Error())
		w.WriteHeader(BAD_REQUEST)
		return
	}

	requestedCommandUUID := mux.Vars(r)["command_uuid"]
	if !common.ValidateUUID(requestedCommandUUID) {
		logger.Error("Bad request: invalid command_uuid parameter: " + url)
		w.WriteHeader(BAD_REQUEST)
		return
	}

	command := handler.NewCommand("GET:transaction/command-uuid", requestedCommandUUID)

	err = router.nodeHandler.Node.SendCommand(command)
	if err != nil {
		logger.Error("Can't send command: " + string(command.ToBytes()) + ". Details: " + err.Error())
		writeHTTPResponse(w, COMMAND_TRANSFERRING_ERROR, common.GetTransactionByCommandUUIDResponse{})
		return
	}

	result, err := router.nodeHandler.Node.GetResult(command, common.COMMAND_UUID_TIMEOUT)
	if err != nil {
		logger.Error("Node is inaccessible during processing command: " +
			string(command.ToBytes()) + ". Details: " + err.Error())
		writeHTTPResponse(w, NODE_IS_INACCESSIBLE, common.GetTransactionByCommandUUIDResponse{})
		return
	}

	if result.Code != OK {
		logger.Error("Node return wrong command result: " + strconv.Itoa(result.Code) +
			" on command: " + string(command.ToBytes()))
		writeHTTPResponse(w, result.Code, common.GetTransactionByCommandUUIDResponse{})
		return
	}

	if len(result.Tokens) == 0 {
		logger.Error("Node return invalid result tokens size on command: " + string(command.ToBytes()))
		writeHTTPResponse(w, ENGINE_UNEXPECTED_ERROR, common.GetTransactionByCommandUUIDResponse{})
		return
	}

	count, err := strconv.Atoi(result.Tokens[0])
	if err != nil {
		logger.Error("Node return invalid token on command: " + string(command.ToBytes()) + ". Details: " + err.Error())
		writeHTTPResponse(w, ENGINE_UNEXPECTED_ERROR, common.GetTransactionByCommandUUIDResponse{})
		return
	}

	if count == 0 {
		writeHTTPResponse(w, OK, common.GetTransactionByCommandUUIDResponse{Count: 0})
		return
	}

	if count == 1 {
		writeHTTPResponse(w, OK, common.GetTransactionByCommandUUIDResponse{
			Count:           1,
			TransactionUUID: result.Tokens[1]})
		return
	}

	logger.Error("Node return invalid token `count` on command: " + string(command.ToBytes()))
	writeHTTPResponse(w, ENGINE_UNEXPECTED_ERROR, common.GetTransactionByCommandUUIDResponse{})
}
