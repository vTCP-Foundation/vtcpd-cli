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

func (router *RoutesHandler) BatchMaxExchangeTransaction(w http.ResponseWriter, r *http.Request) {
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

	// Parse exchange equivalents
	exchangeEquivalents := []string{}
	for key, values := range r.URL.Query() {
		if key != "exchange_equivalent" {
			continue
		}
		// Expecting plain integer equivalents as strings
		exchangeEquivalents = append(exchangeEquivalents, values...)
		break
	}
	if len(exchangeEquivalents) == 0 {
		logger.Error("Bad request: there are no exchange_equivalents parameters: " + url)
		w.WriteHeader(BAD_REQUEST)
		return
	}

	// Command generation
	contractorAddresses = append([]string{strconv.Itoa(len(contractorAddresses) / 2)}, contractorAddresses...)
	contractorAddresses = append([]string{"GET:contractors/transactions/max/exchange"}, contractorAddresses...)
	// Insert exchange equivalents block after addresses and before target equivalent
	contractorAddresses = append(contractorAddresses, []string{equivalent}...)
	contractorAddresses = append(contractorAddresses, exchangeEquivalents...)

	command := handler.NewCommand(contractorAddresses...)

	err = router.nodeHandler.Node.SendCommand(command)
	if err != nil {
		logger.Error("Can't send command: " + string(command.ToBytes()) + " to node. Details: " + err.Error())
		writeHTTPResponse(w, COMMAND_TRANSFERRING_ERROR, common.MaxFlowResponse{})
		return
	}

	// Command processing.
	// This command may execute relatively slow.
	// Timeout is set to a slightly greater value to be able to handle this.
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
	for i := 0; i < contractorsCount; i++ {
		response.Records = append(response.Records, common.MaxFlowRecord{
			ContractorAddressType: result.Tokens[i*3+1],
			ContractorAddress:     result.Tokens[i*3+2],
			MaxAmount:             result.Tokens[i*3+3],
		})
	}
	writeHTTPResponse(w, OK, response)
}

func (router *RoutesHandler) EstimatePayment(w http.ResponseWriter, r *http.Request) {
	url, err := preprocessRequest(r)
	if err != nil {
		logger.Error("Bad request: invalid security parameters: " + err.Error())
		w.WriteHeader(BAD_REQUEST)
		return
	}

	vars := mux.Vars(r)
	senderEquivalent, isPresent := vars["sender_equivalent"]
	if !isPresent {
		logger.Error("Bad request: missing sender_equivalent parameter: " + url)
		w.WriteHeader(BAD_REQUEST)
		return
	}
	if !common.ValidateInt(senderEquivalent) {
		logger.Error("Bad request: invalid sender_equivalent parameter: " + url)
		w.WriteHeader(BAD_REQUEST)
		return
	}

	receiverEquivalent, isPresent := vars["receiver_equivalent"]
	if !isPresent {
		logger.Error("Bad request: missing receiver_equivalent parameter: " + url)
		w.WriteHeader(BAD_REQUEST)
		return
	}
	if !common.ValidateInt(receiverEquivalent) {
		logger.Error("Bad request: invalid receiver_equivalent parameter: " + url)
		w.WriteHeader(BAD_REQUEST)
		return
	}

	contractorAddress := r.URL.Query().Get("contractor_address")
	if contractorAddress == "" {
		logger.Error("Bad request: missing contractor_address parameter: " + url)
		w.WriteHeader(BAD_REQUEST)
		return
	}

	typeAndAddress := strings.SplitN(contractorAddress, "-", 2)
	if len(typeAndAddress) != 2 || typeAndAddress[0] == "" || typeAndAddress[1] == "" {
		logger.Error("Bad request: invalid contractor_address parameter: " + url)
		w.WriteHeader(BAD_REQUEST)
		return
	}
	if !common.ValidateInt(typeAndAddress[0]) {
		logger.Error("Bad request: invalid contractor_address type parameter: " + url)
		w.WriteHeader(BAD_REQUEST)
		return
	}

	receiveAmount := r.URL.Query().Get("receive_amount")
	if !common.ValidateSettlementLineAmount(receiveAmount) {
		logger.Error("Bad request: invalid receive_amount parameter: " + url)
		w.WriteHeader(BAD_REQUEST)
		return
	}

	command := handler.NewCommand(
		"GET:contractors/transactions/estimate/payment",
		typeAndAddress[0],
		typeAndAddress[1],
		receiveAmount,
		receiverEquivalent,
		senderEquivalent,
	)

	if err = router.nodeHandler.Node.SendCommand(command); err != nil {
		logger.Error("Can't send command: " + string(command.ToBytes()) + " to node. Details: " + err.Error())
		writeHTTPResponse(w, COMMAND_TRANSFERRING_ERROR, common.EstimatePaymentResponse{})
		return
	}

	result, err := router.nodeHandler.Node.GetResult(command, common.PAYMENT_OPERATION_TIMEOUT)
	if err != nil {
		logger.Error("Node is inaccessible during processing command: " + string(command.ToBytes()) + ". Details: " + err.Error())
		writeHTTPResponse(w, NODE_IS_INACCESSIBLE, common.EstimatePaymentResponse{})
		return
	}

	if result.Code != OK {
		logger.Error("Node return wrong command result: " + strconv.Itoa(result.Code) +
			" on command: " + string(command.ToBytes()))
		writeHTTPResponse(w, result.Code, common.EstimateReceiveResponse{})
		return
	}

	commandStr := string(command.ToBytes())

	if len(result.Tokens) == 0 {
		logger.Error("Node returned invalid result tokens size on command: " + commandStr)
		writeHTTPResponse(w, ENGINE_UNEXPECTED_ERROR, common.EstimatePaymentResponse{})
		return
	}
	writeHTTPResponse(w, OK, common.EstimatePaymentResponse{EstimatedPaymentAmount: result.Tokens[0]})
}

func (router *RoutesHandler) EstimateReceive(w http.ResponseWriter, r *http.Request) {
	url, err := preprocessRequest(r)
	if err != nil {
		logger.Error("Bad request: invalid security parameters: " + err.Error())
		w.WriteHeader(BAD_REQUEST)
		return
	}

	vars := mux.Vars(r)
	senderEquivalent, isPresent := vars["sender_equivalent"]
	if !isPresent {
		logger.Error("Bad request: missing sender_equivalent parameter: " + url)
		w.WriteHeader(BAD_REQUEST)
		return
	}
	if !common.ValidateInt(senderEquivalent) {
		logger.Error("Bad request: invalid sender_equivalent parameter: " + url)
		w.WriteHeader(BAD_REQUEST)
		return
	}

	receiverEquivalent, isPresent := vars["receiver_equivalent"]
	if !isPresent {
		logger.Error("Bad request: missing receiver_equivalent parameter: " + url)
		w.WriteHeader(BAD_REQUEST)
		return
	}
	if !common.ValidateInt(receiverEquivalent) {
		logger.Error("Bad request: invalid receiver_equivalent parameter: " + url)
		w.WriteHeader(BAD_REQUEST)
		return
	}

	contractorAddress := r.URL.Query().Get("contractor_address")
	if contractorAddress == "" {
		logger.Error("Bad request: missing contractor_address parameter: " + url)
		w.WriteHeader(BAD_REQUEST)
		return
	}

	typeAndAddress := strings.SplitN(contractorAddress, "-", 2)
	if len(typeAndAddress) != 2 || typeAndAddress[0] == "" || typeAndAddress[1] == "" {
		logger.Error("Bad request: invalid contractor_address parameter: " + url)
		w.WriteHeader(BAD_REQUEST)
		return
	}
	if !common.ValidateInt(typeAndAddress[0]) {
		logger.Error("Bad request: invalid contractor_address type parameter: " + url)
		w.WriteHeader(BAD_REQUEST)
		return
	}

	paymentAmount := r.URL.Query().Get("payment_amount")
	if !common.ValidateSettlementLineAmount(paymentAmount) {
		logger.Error("Bad request: invalid payment_amount parameter: " + url)
		w.WriteHeader(BAD_REQUEST)
		return
	}

	command := handler.NewCommand(
		"GET:contractors/transactions/estimate/receive",
		typeAndAddress[0],
		typeAndAddress[1],
		paymentAmount,
		senderEquivalent,
		receiverEquivalent,
	)

	if err = router.nodeHandler.Node.SendCommand(command); err != nil {
		logger.Error("Can't send command: " + string(command.ToBytes()) + " to node. Details: " + err.Error())
		writeHTTPResponse(w, COMMAND_TRANSFERRING_ERROR, common.EstimateReceiveResponse{})
		return
	}

	result, err := router.nodeHandler.Node.GetResult(command, common.PAYMENT_OPERATION_TIMEOUT)
	if err != nil {
		logger.Error("Node is inaccessible during processing command: " + string(command.ToBytes()) + ". Details: " + err.Error())
		writeHTTPResponse(w, NODE_IS_INACCESSIBLE, common.EstimateReceiveResponse{})
		return
	}

	if result.Code != OK {
		logger.Error("Node return wrong command result: " + strconv.Itoa(result.Code) +
			" on command: " + string(command.ToBytes()))
		writeHTTPResponse(w, result.Code, common.EstimateReceiveResponse{})
		return
	}

	commandStr := string(command.ToBytes())

	if len(result.Tokens) == 0 {
		logger.Error("Node returned invalid result tokens size on command: " + commandStr)
		writeHTTPResponse(w, ENGINE_UNEXPECTED_ERROR, common.EstimateReceiveResponse{})
		return
	}
	writeHTTPResponse(w, OK, common.EstimateReceiveResponse{EstimatedReceiveAmount: result.Tokens[0]})

}

func (router *RoutesHandler) CreateExchangeTransaction(w http.ResponseWriter, r *http.Request) {
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
	if !common.ValidateInt(equivalent) {
		logger.Error("Bad request: invalid equivalent parameter: " + url)
		w.WriteHeader(BAD_REQUEST)
		return
	}

	query := r.URL.Query()
	addressValues, hasAddresses := query["contractor_address"]
	if !hasAddresses || len(addressValues) == 0 {
		logger.Error("Bad request: there are no contractor_addresses parameters: " + url)
		w.WriteHeader(BAD_REQUEST)
		return
	}

	contractorAddresses := make([]string, 0, len(addressValues)*2)
	for _, value := range addressValues {
		typeAndAddress := strings.SplitN(value, "-", 2)
		if len(typeAndAddress) != 2 || typeAndAddress[0] == "" || typeAndAddress[1] == "" {
			logger.Error("Bad request: invalid contractor_address parameter: " + url)
			w.WriteHeader(BAD_REQUEST)
			return
		}
		if !common.ValidateInt(typeAndAddress[0]) {
			logger.Error("Bad request: invalid contractor_address type parameter: " + url)
			w.WriteHeader(BAD_REQUEST)
			return
		}
		contractorAddresses = append(contractorAddresses, typeAndAddress[0], typeAndAddress[1])
	}

	amount := r.FormValue("amount")
	if !common.ValidateSettlementLineAmount(amount) {
		logger.Error("Bad request: invalid amount parameter: " + url)
		w.WriteHeader(BAD_REQUEST)
		return
	}

	exchangeEquivalents, hasExchange := query["exchange_equivalent"]
	if !hasExchange || len(exchangeEquivalents) == 0 {
		logger.Error("Bad request: there are no exchange_equivalents parameters: " + url)
		w.WriteHeader(BAD_REQUEST)
		return
	}
	if len(exchangeEquivalents) > 5 {
		logger.Error("Bad request: too many exchange_equivalents parameters: " + url)
		w.WriteHeader(BAD_REQUEST)
		return
	}
	for _, exchangeEquivalent := range exchangeEquivalents {
		if !common.ValidateInt(exchangeEquivalent) {
			logger.Error("Bad request: invalid exchange_equivalent parameter: " + url)
			w.WriteHeader(BAD_REQUEST)
			return
		}
	}

	// Get and validate max_allowable_payment_amount (optional)
	maxAllowableAmount := "0"
	maxAllowablePaymentAmountStr := r.FormValue("max_allowable_payment_amount")
	if maxAllowablePaymentAmountStr != "" {
		if !common.ValidateSettlementLineAmount(maxAllowablePaymentAmountStr) {
			logger.Error("Bad request: invalid max_allowable_payment_amount parameter: " + url)
			w.WriteHeader(BAD_REQUEST)
			return
		}
		maxAllowableAmount = maxAllowablePaymentAmountStr
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

	commandParts := append([]string{strconv.Itoa(len(contractorAddresses) / 2)}, contractorAddresses...)
	commandParts = append([]string{"CREATE:contractors/transactions/exchange"}, commandParts...)
	commandParts = append(commandParts, amount, equivalent)
	commandParts = append(commandParts, exchangeEquivalents...)
	commandParts = append(commandParts, maxAllowableAmount)
	if payload != "" {
		commandParts = append(commandParts, payload)
	}

	var command *handler.Command
	if transactionUUIDStr == "" {
		command = handler.NewCommand(commandParts...)
	} else {
		command = handler.NewCommandWithUUID(transactionUUID, commandParts...)
	}

	if err = router.nodeHandler.Node.SendCommand(command); err != nil {
		logger.Error("Can't send command: " + string(command.ToBytes()) + " to node. Details: " + err.Error())
		writeHTTPResponse(w, COMMAND_TRANSFERRING_ERROR, common.PaymentResponse{})
		return
	}

	result, err := router.nodeHandler.Node.GetResult(command, common.PAYMENT_OPERATION_TIMEOUT)
	if err != nil {
		logger.Error("Node is inaccessible during processing command: " + string(command.ToBytes()) + ". Details: " + err.Error())
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
