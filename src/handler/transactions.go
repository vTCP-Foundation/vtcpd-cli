package handler

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/vTCP-Foundation/vtcpd-cli/src/logger"
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

func (handler *NodesHandler) MaxFlow() {

	if CommandType == "" {
		handler.maxFlowFully()

	} else if CommandType == "partly" {
		handler.maxFlowPartly()

	} else {
		logger.Error("Invalid max-flow command " + CommandType)
		fmt.Println("Invalid max-flow command")
		return
	}
}

func (handler *NodesHandler) maxFlowFully() {
	if len(Addresses) == 0 {
		logger.Error("Bad request: there are no contractor addresses parameters in max-flow request")
		fmt.Println("Bad request: there are no contractor addresses parameters")
		return
	}

	if !ValidateInt(Equivalent) {
		logger.Error("Bad request: invalid equivalent parameter in max-flow request")
		fmt.Println("Bad request: invalid equivalent parameter")
		return
	}

	var addresses []string
	for idx := 0; idx < len(Addresses); idx++ {
		addressType, address := ValidateAddress(Addresses[idx])
		if addressType == "" {
			logger.Error("Bad request: invalid address parameter in max-flow request")
			fmt.Println("Bad request: invalid address parameter")
			return
		}
		addresses = append(addresses, addressType, address)
	}

	addresses = append([]string{strconv.Itoa(len(Addresses))}, addresses...)
	addresses = append([]string{"GET:contractors/transactions/max/fully"}, addresses...)
	addresses = append(addresses, []string{Equivalent}...)
	command := NewCommand(addresses...)

	go handler.maxFlowGetResult(command)
}

func (handler *NodesHandler) maxFlowGetResult(command *Command) {
	err := handler.node.SendCommand(command)
	if err != nil {
		logger.Error("Can't send command: " + string(command.ToBytes()) + " to node. Details: " + err.Error())
		resultJSON := buildJSONResponse(COMMAND_TRANSFERRING_ERROR, MaxFlowResponse{})
		fmt.Println(string(resultJSON))
		return
	}

	result, err := handler.node.GetResult(command, MAX_FLOW_FULLY_TIMEOUT)
	if err != nil {
		logger.Error("Node is inaccessible during processing command: " +
			string(command.ToBytes()) + ". Details: " + err.Error())
		resultJSON := buildJSONResponse(NODE_IS_INACCESSIBLE, MaxFlowResponse{})
		fmt.Println(string(resultJSON))
		return
	}

	if result.Code != OK && result.Code != ENGINE_NO_EQUIVALENT {
		logger.Error("Node return wrong command result: " + strconv.Itoa(result.Code) +
			" on command: " + string(command.ToBytes()))
		resultJSON := buildJSONResponse(result.Code, MaxFlowResponse{})
		fmt.Println(string(resultJSON))
		return
	}
	if result.Code == ENGINE_NO_EQUIVALENT {
		logger.Info("Node hasn't equivalent for command: " + string(command.ToBytes()))
		resultJSON := buildJSONResponse(result.Code, MaxFlowResponse{})
		fmt.Println(string(resultJSON))
		return
	}

	if len(result.Tokens) == 0 {
		logger.Error("Node return invalid result tokens size on command: " + string(command.ToBytes()))
		resultJSON := buildJSONResponse(ENGINE_UNEXPECTED_ERROR, MaxFlowResponse{})
		fmt.Println(string(resultJSON))
		return
	}

	contractorsCount, err := strconv.Atoi(result.Tokens[0])
	if err != nil {
		logger.Error("Node return invalid token on command: " + string(command.ToBytes()) +
			". Details: " + err.Error())
		resultJSON := buildJSONResponse(ENGINE_UNEXPECTED_ERROR, MaxFlowResponse{})
		fmt.Println(string(resultJSON))
		return
	}

	if contractorsCount == 0 {
		resultJSON := buildJSONResponse(OK, MaxFlowResponse{Count: 0})
		fmt.Println(string(resultJSON))
		return
	}

	response := MaxFlowResponse{Count: contractorsCount}
	for i := 0; i < contractorsCount; i++ {
		response.Records = append(response.Records, MaxFlowRecord{
			ContractorAddressType: result.Tokens[i*3+1],
			ContractorAddress:     result.Tokens[i*3+2],
			MaxAmount:             result.Tokens[i*3+3],
		})
	}
	resultJSON := buildJSONResponse(OK, response)
	fmt.Println(string(resultJSON))
}

func (handler *NodesHandler) BatchMaxFullyTransaction(w http.ResponseWriter, r *http.Request) {
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
	command := NewCommand(contractorAddresses...)

	err = handler.node.SendCommand(command)
	if err != nil {
		logger.Error("Can't send command: " + string(command.ToBytes()) + " to node. Details: " + err.Error())
		writeHTTPResponse(w, COMMAND_TRANSFERRING_ERROR, MaxFlowResponse{})
		return
	}

	// Command processing.
	// This command may execute relatively slow.
	// Timeout is set to little bit greater value to be able to handle this.
	result, err := handler.node.GetResult(command, MAX_FLOW_FULLY_TIMEOUT)
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
	for i := 0; i < contractorsCount; i++ {
		response.Records = append(response.Records, MaxFlowRecord{
			ContractorAddressType: result.Tokens[i*3+1],
			ContractorAddress:     result.Tokens[i*3+2],
			MaxAmount:             result.Tokens[i*3+3],
		})
	}
	writeHTTPResponse(w, OK, response)
}

func (handler *NodesHandler) maxFlowPartly() {
	if len(Addresses) == 0 {
		logger.Error("Bad request: there are no contractor addresses parameters in max-flow partly request")
		fmt.Println("Bad request: there are no contractor addresses parameters")
		return
	}

	if !ValidateInt(Equivalent) {
		logger.Error("Bad request: invalid equivalent parameter in max-flow partly request")
		fmt.Println("Bad request: invalid equivalent parameter")
		return
	}

	var addresses []string
	for idx := 0; idx < len(Addresses); idx++ {
		addressType, address := ValidateAddress(Addresses[idx])
		if addressType == "" {
			logger.Error("Bad request: invalid address parameter in max-flow partly request")
			fmt.Println("Bad request: invalid address parameter")
			return
		}
		addresses = append(addresses, addressType, address)
	}

	addresses = append([]string{strconv.Itoa(len(Addresses))}, addresses...)
	addresses = append([]string{"GET:contractors/transactions/max"}, addresses...)
	addresses = append(addresses, []string{Equivalent}...)
	command := NewCommand(addresses...)

	go handler.maxFlowPartlyGetResult(command)
}

func (handler *NodesHandler) maxFlowPartlyGetResult(command *Command) {
	err := handler.node.SendCommand(command)
	if err != nil {
		logger.Error("Can't send command: " + string(command.ToBytes()) + " to node. Details: " + err.Error())
		resultJSON := buildJSONResponse(COMMAND_TRANSFERRING_ERROR, MaxFlowPartialResponse{})
		fmt.Println(string(resultJSON))
		return
	}

	result, err := handler.node.GetResult(command, MAX_FLOW_FIRST_TIMEOUT)
	if err != nil {
		logger.Error("Node is inaccessible during processing command: " +
			string(command.ToBytes()) + ". Details: " + err.Error())
		resultJSON := buildJSONResponse(NODE_IS_INACCESSIBLE, MaxFlowPartialResponse{})
		fmt.Println(string(resultJSON))
		return
	}

	if result.Code != OK && result.Code != ENGINE_NO_EQUIVALENT {
		logger.Error("Node return wrong command result: " + strconv.Itoa(result.Code) +
			" on command: " + string(command.ToBytes()))
		resultJSON := buildJSONResponse(result.Code, MaxFlowPartialResponse{})
		fmt.Println(string(resultJSON))
		return
	}
	if result.Code == ENGINE_NO_EQUIVALENT {
		logger.Info("Node hasn't equivalent for command: " + string(command.ToBytes()))
		resultJSON := buildJSONResponse(result.Code, MaxFlowPartialResponse{})
		fmt.Println(string(resultJSON))
		return
	}

	if len(result.Tokens) == 0 {
		logger.Error("Node return invalid result tokens size on command: " + string(command.ToBytes()))
		resultJSON := buildJSONResponse(ENGINE_UNEXPECTED_ERROR, MaxFlowPartialResponse{})
		fmt.Println(string(resultJSON))
		return
	}

	stateResult, err := strconv.Atoi(result.Tokens[0])
	if err != nil {
		logger.Error("Node return invalid stateResult on command: " + string(command.ToBytes()) +
			". Details: " + err.Error())
		resultJSON := buildJSONResponse(ENGINE_UNEXPECTED_ERROR, MaxFlowPartialResponse{})
		fmt.Println(string(resultJSON))
		return
	}

	contractorsCount, err := strconv.Atoi(result.Tokens[1])
	if err != nil {
		logger.Error("Node return invalid contractorsCount on command: " + string(command.ToBytes()) +
			". Details: " + err.Error())
		resultJSON := buildJSONResponse(ENGINE_UNEXPECTED_ERROR, MaxFlowPartialResponse{})
		fmt.Println(string(resultJSON))
		return
	}

	if contractorsCount == 0 {
		resultJSON := buildJSONResponse(OK, MaxFlowPartialResponse{State: stateResult, Count: 0})
		fmt.Println(string(resultJSON))
		return
	}

	response := MaxFlowPartialResponse{
		State: stateResult,
		Count: contractorsCount}
	for i := 0; i < contractorsCount; i++ {
		response.Records = append(response.Records, MaxFlowRecord{
			ContractorAddressType: result.Tokens[i*3+2],
			ContractorAddress:     result.Tokens[i*3+3],
			MaxAmount:             result.Tokens[i*3+4],
		})
	}
	resultJSON := buildJSONResponse(OK, response)
	fmt.Println(string(resultJSON))

	// if max flows are not final : wait for final results
	if stateResult != 10 {
		go handler.maxFlowPartlyStepTwoGetResult(command)
	}
}

func (handler *NodesHandler) maxFlowPartlyStepTwoGetResult(command *Command) {
	handler.node.WaitCommand(command)

	// Command processing.
	// This command may execute relatively slow.
	// Timeout is set to little bit greater value to be able to handle this.
	result, err := handler.node.GetResult(command, MAX_FLOW_FULLY_TIMEOUT)
	if err != nil {
		logger.Error("Node is inaccessible during processing command 2: " + string(command.ToBytes()) +
			". Details: " + err.Error())
		resultJSON := buildJSONResponse(NODE_IS_INACCESSIBLE, MaxFlowPartialResponse{})
		fmt.Println(string(resultJSON))
		return
	}

	if result.Code != OK {
		logger.Error("Node return wrong command result: " + strconv.Itoa(result.Code) +
			" on command 2: " + string(command.ToBytes()))
		resultJSON := buildJSONResponse(result.Code, MaxFlowPartialResponse{})
		fmt.Println(string(resultJSON))
		return
	}

	if len(result.Tokens) == 0 {
		logger.Error("Node return invalid result tokens size on command 2: " + string(command.ToBytes()))
		resultJSON := buildJSONResponse(ENGINE_UNEXPECTED_ERROR, MaxFlowPartialResponse{})
		fmt.Println(string(resultJSON))
		return
	}

	stateResult, err := strconv.Atoi(result.Tokens[0])
	if err != nil {
		logger.Error("Node return invalid stateResult on command 2: " + string(command.ToBytes()) + ". Details: " + err.Error())
		resultJSON := buildJSONResponse(ENGINE_UNEXPECTED_ERROR, MaxFlowPartialResponse{})
		fmt.Println(string(resultJSON))
		return
	}

	contractorsCount, err := strconv.Atoi(result.Tokens[1])
	if err != nil {
		logger.Error("Node return invalid contractorsCount on command 2: " + string(command.ToBytes()) + ". Details: " + err.Error())
		resultJSON := buildJSONResponse(ENGINE_UNEXPECTED_ERROR, MaxFlowPartialResponse{})
		fmt.Println(string(resultJSON))
		return
	}

	if contractorsCount == 0 {
		resultJSON := buildJSONResponse(OK, MaxFlowPartialResponse{State: stateResult, Count: 0})
		fmt.Println(string(resultJSON))
		return
	}

	response := MaxFlowPartialResponse{
		State: stateResult,
		Count: contractorsCount}
	for i := 0; i < contractorsCount; i++ {
		response.Records = append(response.Records, MaxFlowRecord{
			ContractorAddressType: result.Tokens[i*3+2],
			ContractorAddress:     result.Tokens[i*3+3],
			MaxAmount:             result.Tokens[i*3+4],
		})
	}
	resultJSON := buildJSONResponse(OK, response)
	fmt.Println(string(resultJSON))

	// if max flows are not final : wait for final results
	if stateResult != 10 {
		go handler.maxFlowPartlyStepTwoGetResult(command)
	}
}

func (handler *NodesHandler) Payment() {
	if len(Addresses) == 0 {
		logger.Error("Bad request: there are no contractor addresses parameters in payment request")
		fmt.Println("Bad request: there are no contractor addresses parameters")
		return
	}

	if !ValidateSettlementLineAmount(Amount) {
		logger.Error("Bad request: invalid amount parameter in payment request")
		fmt.Println("Bad request: invalid amount parameter")
		return
	}

	if !ValidateInt(Equivalent) {
		logger.Error("Bad request: invalid equivalent parameter in payment request")
		fmt.Println("Bad request: invalid equivalent parameter")
		return
	}

	var addresses []string
	for idx := 0; idx < len(Addresses); idx++ {
		addressType, address := ValidateAddress(Addresses[idx])
		if addressType == "" {
			logger.Error("Bad request: invalid address parameter in payment request")
			fmt.Println("Bad request: invalid address parameter")
			return
		}
		addresses = append(addresses, addressType, address)
	}

	addresses = append([]string{strconv.Itoa(len(Addresses))}, addresses...)
	addresses = append([]string{"CREATE:contractors/transactions"}, addresses...)
	addresses = append(addresses, []string{Amount, Equivalent}...)
	if Payload != "" {
		addresses = append(addresses, []string{Payload}...)
	}
	command := NewCommand(addresses...)

	go handler.paymentResult(command)
}

func (handler *NodesHandler) paymentResult(command *Command) {
	err := handler.node.SendCommand(command)
	if err != nil {
		logger.Error("Can't send command: " + string(command.ToBytes()) + " to node. Details: " + err.Error())
		resultJSON := buildJSONResponse(COMMAND_TRANSFERRING_ERROR, PaymentResponse{})
		fmt.Println(string(resultJSON))
		return
	}

	result, err := handler.node.GetResult(command, PAYMENT_OPERATION_TIMEOUT)
	if err != nil {
		logger.Error("Node is inaccessible during processing command: " +
			string(command.ToBytes()) + ". Details: " + err.Error())
		resultJSON := buildJSONResponse(NODE_IS_INACCESSIBLE, PaymentResponse{})
		fmt.Println(string(resultJSON))
		return
	}

	if result.Code != CREATED && result.Code != ENGINE_NO_EQUIVALENT {
		logger.Error("Node return wrong command result: " + strconv.Itoa(result.Code) +
			" on command: " + string(command.ToBytes()))
		resultJSON := buildJSONResponse(result.Code, PaymentResponse{})
		fmt.Println(string(resultJSON))
		return
	}
	if result.Code == ENGINE_NO_EQUIVALENT {
		logger.Info("Node hasn't equivalent for command: " + string(command.ToBytes()))
		resultJSON := buildJSONResponse(result.Code, PaymentResponse{})
		fmt.Println(string(resultJSON))
		return
	}

	if len(result.Tokens) == 0 {
		logger.Error("Node return invalid result tokens size on command: " + string(command.ToBytes()))
		resultJSON := buildJSONResponse(ENGINE_UNEXPECTED_ERROR, PaymentResponse{})
		fmt.Println(string(resultJSON))
		return
	}

	resultJSON := buildJSONResponse(OK, PaymentResponse{TransactionUUID: result.Tokens[0]})
	fmt.Println(string(resultJSON))
}

func (handler *NodesHandler) CreateTransaction(w http.ResponseWriter, r *http.Request) {
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
	if !ValidateSettlementLineAmount(amount) {
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

	var command *Command
	if transactionUUIDStr == "" {
		command = NewCommand(contractorAddresses...)
	} else {
		command = NewCommandWithUUID(transactionUUID, contractorAddresses...)
	}

	err = handler.node.SendCommand(command)
	if err != nil {
		logger.Error("Can't send command: " + string(command.ToBytes()) + " to node. Details: " + err.Error())
		writeHTTPResponse(w, COMMAND_TRANSFERRING_ERROR, PaymentResponse{})
		return
	}

	result, err := handler.node.GetResult(command, PAYMENT_OPERATION_TIMEOUT)
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

func (handler *NodesHandler) GetTransactionByCommandUUID(w http.ResponseWriter, r *http.Request) {
	url, err := preprocessRequest(r)
	if err != nil {
		logger.Error("Bad request: invalid security parameters: " + err.Error())
		w.WriteHeader(BAD_REQUEST)
		return
	}

	requestedCommandUUID := mux.Vars(r)["command_uuid"]
	if !validateUUID(requestedCommandUUID) {
		logger.Error("Bad request: invalid command_uuid parameter: " + url)
		w.WriteHeader(BAD_REQUEST)
		return
	}

	command := NewCommand("GET:transaction/command-uuid", requestedCommandUUID)

	err = handler.node.SendCommand(command)
	if err != nil {
		logger.Error("Can't send command: " + string(command.ToBytes()) + ". Details: " + err.Error())
		writeHTTPResponse(w, COMMAND_TRANSFERRING_ERROR, GetTransactionByCommandUUIDResponse{})
		return
	}

	result, err := handler.node.GetResult(command, COMMAND_UUID_TIMEOUT)
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
