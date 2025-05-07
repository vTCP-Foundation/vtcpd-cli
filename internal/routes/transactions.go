package routes

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/vTCP-Foundation/vtcpd-cli/internal/handler"
	"github.com/vTCP-Foundation/vtcpd-cli/internal/logger"
)

// --- Global structs for payments ---

type MaxFlowRecord struct {
	ContractorAddressType string `json:"address_type"`
	ContractorAddress     string `json:"contractor_address"`
	MaxAmount             string `json:"max_amount"`
}

// --- Global API responses for payments ---

type MaxFlowResponse struct {
	Count   int             `json:"count"`
	Records []MaxFlowRecord `json:"records"`
}

type MaxFlowPartialResponse struct {
	State   int             `json:"state"`
	Count   int             `json:"count"`
	Records []MaxFlowRecord `json:"records"`
}

type PaymentResponse struct {
	TransactionUUID string `json:"transaction_uuid"`
}

type GetTransactionByCommandUUIDResponse struct {
	Count           int    `json:"count"`
	TransactionUUID string `json:"transaction_uuid"`
}

var (
	PAYMENT_OPERATION_TIMEOUT uint16 = 60
	MAX_FLOW_FIRST_TIMEOUT    uint16 = 30
	MAX_FLOW_FULLY_TIMEOUT    uint16 = 60
	COMMAND_UUID_TIMEOUT      uint16 = 20
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
		writeHTTPResponse(w, COMMAND_TRANSFERRING_ERROR, MaxFlowResponse{})
		return
	}

	// Command processing.
	// This command may execute relatively slow.
	// Timeout is set to little bit greater value to be able to handle this.
	result, err := router.nodeHandler.Node.GetResult(command, MAX_FLOW_FULLY_TIMEOUT)
	if err != nil {
		logger.Error("Node is inaccessible during processing command: " +
			string(command.ToBytes()) + ". Details: " + err.Error())
		writeHTTPResponse(w, NODE_IS_INACCESSIBLE, MaxFlowResponse{})
		return
	}

	if result.Code != OK && result.Code != ENGINE_NO_EQUIVALENT {
		logger.Error("Node return wrong command result: " + strconv.Itoa(result.Code) +
			" on command: " + string(command.ToBytes()))
		writeHTTPResponse(w, result.Code, MaxFlowResponse{})
		return
	}
	if result.Code == ENGINE_NO_EQUIVALENT {
		logger.Info("Node hasn't equivalent for command: " + string(command.ToBytes()))
		writeHTTPResponse(w, result.Code, MaxFlowResponse{})
		return
	}

	if len(result.Tokens) == 0 {
		logger.Error("Node return invalid result tokens size on command: " + string(command.ToBytes()))
		writeHTTPResponse(w, ENGINE_UNEXPECTED_ERROR, MaxFlowResponse{})
		return
	}

	contractorsCount, err := strconv.Atoi(result.Tokens[0])
	if err != nil {
		logger.Error("Node return invalid token on command: " + string(command.ToBytes()) +
			". Details: " + err.Error())
		writeHTTPResponse(w, ENGINE_UNEXPECTED_ERROR, MaxFlowResponse{})
		return
	}

	if contractorsCount == 0 {
		writeHTTPResponse(w, OK, MaxFlowResponse{Count: 0})
		return
	}

	response := MaxFlowResponse{Count: contractorsCount}
	for i := range contractorsCount {
		response.Records = append(response.Records, MaxFlowRecord{
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
	if !handler.ValidateSettlementLineAmount(amount) {
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
		writeHTTPResponse(w, COMMAND_TRANSFERRING_ERROR, PaymentResponse{})
		return
	}

	result, err := router.nodeHandler.Node.GetResult(command, PAYMENT_OPERATION_TIMEOUT)
	if err != nil {
		logger.Error("Node is inaccessible during processing command: " +
			string(command.ToBytes()) + ". Details: " + err.Error())
		writeHTTPResponse(w, NODE_IS_INACCESSIBLE, PaymentResponse{})
		return
	}

	if result.Code != CREATED && result.Code != ENGINE_NO_EQUIVALENT {
		logger.Error("Node return wrong command result: " + strconv.Itoa(result.Code) +
			" on command: " + string(command.ToBytes()))
		writeHTTPResponse(w, result.Code, PaymentResponse{})
		return
	}
	if result.Code == ENGINE_NO_EQUIVALENT {
		logger.Info("Node hasn't equivalent for command: " + string(command.ToBytes()))
		writeHTTPResponse(w, result.Code, PaymentResponse{})
		return
	}

	if len(result.Tokens) == 0 {
		logger.Error("Node return invalid result tokens size on command: " + string(command.ToBytes()))
		writeHTTPResponse(w, ENGINE_UNEXPECTED_ERROR, PaymentResponse{})
		return
	}

	writeHTTPResponse(w, OK, PaymentResponse{TransactionUUID: result.Tokens[0]})
}

func (router *RoutesHandler) GetTransactionByCommandUUID(w http.ResponseWriter, r *http.Request) {
	url, err := preprocessRequest(r)
	if err != nil {
		logger.Error("Bad request: invalid security parameters: " + err.Error())
		w.WriteHeader(BAD_REQUEST)
		return
	}

	requestedCommandUUID := mux.Vars(r)["command_uuid"]
	if !handler.ValidateUUID(requestedCommandUUID) {
		logger.Error("Bad request: invalid command_uuid parameter: " + url)
		w.WriteHeader(BAD_REQUEST)
		return
	}

	command := handler.NewCommand("GET:transaction/command-uuid", requestedCommandUUID)

	err = router.nodeHandler.Node.SendCommand(command)
	if err != nil {
		logger.Error("Can't send command: " + string(command.ToBytes()) + ". Details: " + err.Error())
		writeHTTPResponse(w, COMMAND_TRANSFERRING_ERROR, GetTransactionByCommandUUIDResponse{})
		return
	}

	result, err := router.nodeHandler.Node.GetResult(command, COMMAND_UUID_TIMEOUT)
	if err != nil {
		logger.Error("Node is inaccessible during processing command: " +
			string(command.ToBytes()) + ". Details: " + err.Error())
		writeHTTPResponse(w, NODE_IS_INACCESSIBLE, GetTransactionByCommandUUIDResponse{})
		return
	}

	if result.Code != OK {
		logger.Error("Node return wrong command result: " + strconv.Itoa(result.Code) +
			" on command: " + string(command.ToBytes()))
		writeHTTPResponse(w, result.Code, GetTransactionByCommandUUIDResponse{})
		return
	}

	if len(result.Tokens) == 0 {
		logger.Error("Node return invalid result tokens size on command: " + string(command.ToBytes()))
		writeHTTPResponse(w, ENGINE_UNEXPECTED_ERROR, GetTransactionByCommandUUIDResponse{})
		return
	}

	count, err := strconv.Atoi(result.Tokens[0])
	if err != nil {
		logger.Error("Node return invalid token on command: " + string(command.ToBytes()) + ". Details: " + err.Error())
		writeHTTPResponse(w, ENGINE_UNEXPECTED_ERROR, GetTransactionByCommandUUIDResponse{})
		return
	}

	if count == 0 {
		writeHTTPResponse(w, OK, GetTransactionByCommandUUIDResponse{Count: 0})
		return
	}

	if count == 1 {
		writeHTTPResponse(w, OK, GetTransactionByCommandUUIDResponse{
			Count:           1,
			TransactionUUID: result.Tokens[1]})
		return
	}

	logger.Error("Node return invalid token `count` on command: " + string(command.ToBytes()))
	writeHTTPResponse(w, ENGINE_UNEXPECTED_ERROR, GetTransactionByCommandUUIDResponse{})
}
