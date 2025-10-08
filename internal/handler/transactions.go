package handler

import (
	"fmt"
	"strconv"

	"github.com/google/uuid"
	"github.com/vTCP-Foundation/vtcpd-cli/internal/common"
	"github.com/vTCP-Foundation/vtcpd-cli/internal/logger"
)

func (handler *NodeHandler) MaxFlow() {

	if CommandType == "fully" {
		handler.maxFlowFully()

	} else if CommandType == "partly" {
		handler.maxFlowPartly()

	} else if CommandType == "exchange" {
		handler.maxFlowExchange()

	} else {
		logger.Error("Invalid max-flow command " + CommandType)
		fmt.Println("Invalid max-flow command")
		return
	}
}

func (handler *NodeHandler) maxFlowFully() {
	if len(Addresses) == 0 {
		logger.Error("Bad request: there are no contractor addresses parameters in max-flow request")
		fmt.Println("Bad request: there are no contractor addresses parameters")
		return
	}

	if !common.ValidateInt(Equivalent) {
		logger.Error("Bad request: invalid equivalent parameter in max-flow request")
		fmt.Println("Bad request: invalid equivalent parameter")
		return
	}

	var addresses []string
	for idx := range len(Addresses) {
		addressType, address := common.ValidateAddress(Addresses[idx])
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

func (handler *NodeHandler) maxFlowExchange() {
	if len(Addresses) == 0 {
		logger.Error("Bad request: there are no contractor addresses parameters in max-flow exchange request")
		fmt.Println("Bad request: there are no contractor addresses parameters")
		return
	}

	if !common.ValidateInt(Equivalent) {
		logger.Error("Bad request: invalid equivalent parameter in max-flow exchange request")
		fmt.Println("Bad request: invalid equivalent parameter")
		return
	}

	if len(ExchangeEquivalents) == 0 {
		logger.Error("Bad request: there are no exchange_equivalents parameters in max-flow exchange request")
		fmt.Println("Bad request: there are no exchange_equivalents parameters")
		return
	}

	var addresses []string
	for idx := range len(Addresses) {
		addressType, address := common.ValidateAddress(Addresses[idx])
		if addressType == "" {
			logger.Error("Bad request: invalid address parameter in max-flow exchange request")
			fmt.Println("Bad request: invalid address parameter")
			return
		}
		addresses = append(addresses, addressType, address)
	}

	// Command generation: prepend count, prepend command, then append target equivalent and exchange equivalents
	addresses = append([]string{strconv.Itoa(len(Addresses))}, addresses...)
	addresses = append([]string{"GET:contractors/transactions/max/exchange"}, addresses...)
	addresses = append(addresses, []string{Equivalent}...)
	addresses = append(addresses, ExchangeEquivalents...)

	command := NewCommand(addresses...)

	go handler.maxFlowGetResult(command)
}

func (handler *NodeHandler) maxFlowGetResult(command *Command) {
	err := handler.Node.SendCommand(command)
	if err != nil {
		logger.Error("Can't send command: " + string(command.ToBytes()) + " to node. Details: " + err.Error())
		resultJSON := buildJSONResponse(COMMAND_TRANSFERRING_ERROR, common.MaxFlowResponse{})
		fmt.Println(string(resultJSON))
		return
	}

	result, err := handler.Node.GetResult(command, common.MAX_FLOW_FULLY_TIMEOUT)
	if err != nil {
		logger.Error("Node is inaccessible during processing command: " +
			string(command.ToBytes()) + ". Details: " + err.Error())
		resultJSON := buildJSONResponse(NODE_IS_INACCESSIBLE, common.MaxFlowResponse{})
		fmt.Println(string(resultJSON))
		return
	}

	if result.Code != OK && result.Code != ENGINE_NO_EQUIVALENT {
		logger.Error("Node return wrong command result: " + strconv.Itoa(result.Code) +
			" on command: " + string(command.ToBytes()))
		resultJSON := buildJSONResponse(result.Code, common.MaxFlowResponse{})
		fmt.Println(string(resultJSON))
		return
	}
	if result.Code == ENGINE_NO_EQUIVALENT {
		logger.Info("Node hasn't equivalent for command: " + string(command.ToBytes()))
		resultJSON := buildJSONResponse(result.Code, common.MaxFlowResponse{})
		fmt.Println(string(resultJSON))
		return
	}

	if len(result.Tokens) == 0 {
		logger.Error("Node return invalid result tokens size on command: " + string(command.ToBytes()))
		resultJSON := buildJSONResponse(ENGINE_UNEXPECTED_ERROR, common.MaxFlowResponse{})
		fmt.Println(string(resultJSON))
		return
	}

	contractorsCount, err := strconv.Atoi(result.Tokens[0])
	if err != nil {
		logger.Error("Node return invalid token on command: " + string(command.ToBytes()) +
			". Details: " + err.Error())
		resultJSON := buildJSONResponse(ENGINE_UNEXPECTED_ERROR, common.MaxFlowResponse{})
		fmt.Println(string(resultJSON))
		return
	}

	if contractorsCount == 0 {
		resultJSON := buildJSONResponse(OK, common.MaxFlowResponse{Count: 0})
		fmt.Println(string(resultJSON))
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
	resultJSON := buildJSONResponse(OK, response)
	fmt.Println(string(resultJSON))
}

func (handler *NodeHandler) maxFlowPartly() {
	if len(Addresses) == 0 {
		logger.Error("Bad request: there are no contractor addresses parameters in max-flow partly request")
		fmt.Println("Bad request: there are no contractor addresses parameters")
		return
	}

	if !common.ValidateInt(Equivalent) {
		logger.Error("Bad request: invalid equivalent parameter in max-flow partly request")
		fmt.Println("Bad request: invalid equivalent parameter")
		return
	}

	var addresses []string
	for idx := range len(Addresses) {
		addressType, address := common.ValidateAddress(Addresses[idx])
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

func (handler *NodeHandler) maxFlowPartlyGetResult(command *Command) {
	err := handler.Node.SendCommand(command)
	if err != nil {
		logger.Error("Can't send command: " + string(command.ToBytes()) + " to node. Details: " + err.Error())
		resultJSON := buildJSONResponse(COMMAND_TRANSFERRING_ERROR, common.MaxFlowPartialResponse{})
		fmt.Println(string(resultJSON))
		return
	}

	result, err := handler.Node.GetResult(command, common.MAX_FLOW_FIRST_TIMEOUT)
	if err != nil {
		logger.Error("Node is inaccessible during processing command: " +
			string(command.ToBytes()) + ". Details: " + err.Error())
		resultJSON := buildJSONResponse(NODE_IS_INACCESSIBLE, common.MaxFlowPartialResponse{})
		fmt.Println(string(resultJSON))
		return
	}

	if result.Code != OK && result.Code != ENGINE_NO_EQUIVALENT {
		logger.Error("Node return wrong command result: " + strconv.Itoa(result.Code) +
			" on command: " + string(command.ToBytes()))
		resultJSON := buildJSONResponse(result.Code, common.MaxFlowPartialResponse{})
		fmt.Println(string(resultJSON))
		return
	}
	if result.Code == ENGINE_NO_EQUIVALENT {
		logger.Info("Node hasn't equivalent for command: " + string(command.ToBytes()))
		resultJSON := buildJSONResponse(result.Code, common.MaxFlowPartialResponse{})
		fmt.Println(string(resultJSON))
		return
	}

	if len(result.Tokens) == 0 {
		logger.Error("Node return invalid result tokens size on command: " + string(command.ToBytes()))
		resultJSON := buildJSONResponse(ENGINE_UNEXPECTED_ERROR, common.MaxFlowPartialResponse{})
		fmt.Println(string(resultJSON))
		return
	}

	stateResult, err := strconv.Atoi(result.Tokens[0])
	if err != nil {
		logger.Error("Node return invalid stateResult on command: " + string(command.ToBytes()) +
			". Details: " + err.Error())
		resultJSON := buildJSONResponse(ENGINE_UNEXPECTED_ERROR, common.MaxFlowPartialResponse{})
		fmt.Println(string(resultJSON))
		return
	}

	contractorsCount, err := strconv.Atoi(result.Tokens[1])
	if err != nil {
		logger.Error("Node return invalid contractorsCount on command: " + string(command.ToBytes()) +
			". Details: " + err.Error())
		resultJSON := buildJSONResponse(ENGINE_UNEXPECTED_ERROR, common.MaxFlowPartialResponse{})
		fmt.Println(string(resultJSON))
		return
	}

	if contractorsCount == 0 {
		resultJSON := buildJSONResponse(OK, common.MaxFlowPartialResponse{State: stateResult, Count: 0})
		fmt.Println(string(resultJSON))
		return
	}

	response := common.MaxFlowPartialResponse{
		State: stateResult,
		Count: contractorsCount}
	for i := range contractorsCount {
		response.Records = append(response.Records, common.MaxFlowRecord{
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

func (handler *NodeHandler) maxFlowPartlyStepTwoGetResult(command *Command) {
	handler.Node.WaitCommand(command)

	// Command processing.
	// This command may execute relatively slow.
	// Timeout is set to little bit greater value to be able to handle this.
	result, err := handler.Node.GetResult(command, common.MAX_FLOW_FULLY_TIMEOUT)
	if err != nil {
		logger.Error("Node is inaccessible during processing command 2: " + string(command.ToBytes()) +
			". Details: " + err.Error())
		resultJSON := buildJSONResponse(NODE_IS_INACCESSIBLE, common.MaxFlowPartialResponse{})
		fmt.Println(string(resultJSON))
		return
	}

	if result.Code != OK {
		logger.Error("Node return wrong command result: " + strconv.Itoa(result.Code) +
			" on command 2: " + string(command.ToBytes()))
		resultJSON := buildJSONResponse(result.Code, common.MaxFlowPartialResponse{})
		fmt.Println(string(resultJSON))
		return
	}

	if len(result.Tokens) == 0 {
		logger.Error("Node return invalid result tokens size on command 2: " + string(command.ToBytes()))
		resultJSON := buildJSONResponse(ENGINE_UNEXPECTED_ERROR, common.MaxFlowPartialResponse{})
		fmt.Println(string(resultJSON))
		return
	}

	stateResult, err := strconv.Atoi(result.Tokens[0])
	if err != nil {
		logger.Error("Node return invalid stateResult on command 2: " + string(command.ToBytes()) + ". Details: " + err.Error())
		resultJSON := buildJSONResponse(ENGINE_UNEXPECTED_ERROR, common.MaxFlowPartialResponse{})
		fmt.Println(string(resultJSON))
		return
	}

	contractorsCount, err := strconv.Atoi(result.Tokens[1])
	if err != nil {
		logger.Error("Node return invalid contractorsCount on command 2: " + string(command.ToBytes()) + ". Details: " + err.Error())
		resultJSON := buildJSONResponse(ENGINE_UNEXPECTED_ERROR, common.MaxFlowPartialResponse{})
		fmt.Println(string(resultJSON))
		return
	}

	if contractorsCount == 0 {
		resultJSON := buildJSONResponse(OK, common.MaxFlowPartialResponse{State: stateResult, Count: 0})
		fmt.Println(string(resultJSON))
		return
	}

	response := common.MaxFlowPartialResponse{
		State: stateResult,
		Count: contractorsCount}
	for i := range contractorsCount {
		response.Records = append(response.Records, common.MaxFlowRecord{
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

func (handler *NodeHandler) PaymentEstimate() {
	hasReceiveAmount := ReceiveAmount != ""
	hasPaymentAmount := PaymentAmount != ""

	if hasReceiveAmount && hasPaymentAmount {
		logger.Error("Bad request: receive-amount and payment-amount parameters can't be used together in payment estimate request")
		fmt.Println("Bad request: specify only receive-amount or payment-amount")
		return
	}

	if !hasReceiveAmount && !hasPaymentAmount {
		logger.Error("Bad request: missing receive-amount or payment-amount parameter in payment estimate request")
		fmt.Println("Bad request: missing receive-amount or payment-amount parameter")
		return
	}

	if hasReceiveAmount {
		handler.estimatePayment()
		return
	}

	handler.estimateReceive()
}

func (handler *NodeHandler) estimatePayment() {
	if len(Addresses) == 0 {
		logger.Error("Bad request: missing address parameter in payment estimate request")
		fmt.Println("Bad request: missing address parameter")
		return
	}

	addressType, address := common.ValidateAddress(Addresses[0])
	if addressType == "" || address == "" {
		logger.Error("Bad request: invalid address parameter in payment estimate request")
		fmt.Println("Bad request: invalid address parameter")
		return
	}

	if !common.ValidateSettlementLineAmount(ReceiveAmount) {
		logger.Error("Bad request: invalid receive-amount parameter in payment estimate request")
		fmt.Println("Bad request: invalid receive-amount parameter")
		return
	}

	if SenderEquivalent == "" || ReceiverEquivalent == "" {
		logger.Error("Bad request: missing sender-eq or receiver-eq parameter in payment estimate request")
		fmt.Println("Bad request: missing sender-eq or receiver-eq parameter")
		return
	}

	if !common.ValidateInt(SenderEquivalent) || !common.ValidateInt(ReceiverEquivalent) {
		logger.Error("Bad request: invalid sender-eq or receiver-eq parameter in payment estimate request")
		fmt.Println("Bad request: invalid sender-eq or receiver-eq parameter")
		return
	}

	command := NewCommand(
		"GET:contractors/transactions/estimate/payment",
		addressType,
		address,
		ReceiveAmount,
		ReceiverEquivalent,
		SenderEquivalent,
	)

	go handler.estimatePaymentGetResult(command)
}

func (handler *NodeHandler) estimatePaymentGetResult(command *Command) {
	err := handler.Node.SendCommand(command)
	if err != nil {
		logger.Error("Can't send command: " + string(command.ToBytes()) + " to node. Details: " + err.Error())
		resultJSON := buildJSONResponse(COMMAND_TRANSFERRING_ERROR, common.EstimatePaymentResponse{})
		fmt.Println(string(resultJSON))
		return
	}

	result, err := handler.Node.GetResult(command, common.PAYMENT_OPERATION_TIMEOUT)
	if err != nil {
		logger.Error("Node is inaccessible during processing command: " +
			string(command.ToBytes()) + ". Details: " + err.Error())
		resultJSON := buildJSONResponse(NODE_IS_INACCESSIBLE, common.EstimatePaymentResponse{})
		fmt.Println(string(resultJSON))
		return
	}

	commandStr := string(command.ToBytes())
	switch result.Code {
	case OK:
		if len(result.Tokens) == 0 {
			logger.Error("Node return invalid result tokens size on command: " + commandStr)
			resultJSON := buildJSONResponse(ENGINE_UNEXPECTED_ERROR, common.EstimatePaymentResponse{})
			fmt.Println(string(resultJSON))
			return
		}
		resultJSON := buildJSONResponse(OK, common.EstimatePaymentResponse{EstimatedPaymentAmount: result.Tokens[0]})
		fmt.Println(string(resultJSON))
		return
	case 401:
		logger.Error("Node return unexpected error code 401 on command: " + commandStr)
	case 412:
		logger.Info("Node return insufficient paths (412) on command: " + commandStr)
	case 462:
		logger.Info("Node return no cached paths (462) on command: " + commandStr)
	default:
		logger.Error("Node return wrong command result: " + strconv.Itoa(result.Code) +
			" on command: " + commandStr)
	}
	resultJSON := buildJSONResponse(result.Code, common.EstimatePaymentResponse{})
	fmt.Println(string(resultJSON))
}

func (handler *NodeHandler) estimateReceive() {
	if len(Addresses) == 0 {
		logger.Error("Bad request: missing address parameter in receive estimate request")
		fmt.Println("Bad request: missing address parameter")
		return
	}

	addressType, address := common.ValidateAddress(Addresses[0])
	if addressType == "" || address == "" {
		logger.Error("Bad request: invalid address parameter in receive estimate request")
		fmt.Println("Bad request: invalid address parameter")
		return
	}

	if !common.ValidateSettlementLineAmount(PaymentAmount) {
		logger.Error("Bad request: invalid payment-amount parameter in receive estimate request")
		fmt.Println("Bad request: invalid payment-amount parameter")
		return
	}

	if SenderEquivalent == "" || ReceiverEquivalent == "" {
		logger.Error("Bad request: missing sender-eq or receiver-eq parameter in receive estimate request")
		fmt.Println("Bad request: missing sender-eq or receiver-eq parameter")
		return
	}

	if !common.ValidateInt(SenderEquivalent) || !common.ValidateInt(ReceiverEquivalent) {
		logger.Error("Bad request: invalid sender-eq or receiver-eq parameter in receive estimate request")
		fmt.Println("Bad request: invalid sender-eq or receiver-eq parameter")
		return
	}

	command := NewCommand(
		"GET:contractors/transactions/estimate/receive",
		addressType,
		address,
		PaymentAmount,
		SenderEquivalent,
		ReceiverEquivalent,
	)

	go handler.estimateReceiveGetResult(command)
}

func (handler *NodeHandler) estimateReceiveGetResult(command *Command) {
	err := handler.Node.SendCommand(command)
	if err != nil {
		logger.Error("Can't send command: " + string(command.ToBytes()) + " to node. Details: " + err.Error())
		resultJSON := buildJSONResponse(COMMAND_TRANSFERRING_ERROR, common.EstimateReceiveResponse{})
		fmt.Println(string(resultJSON))
		return
	}

	result, err := handler.Node.GetResult(command, common.PAYMENT_OPERATION_TIMEOUT)
	if err != nil {
		logger.Error("Node is inaccessible during processing command: " +
			string(command.ToBytes()) + ". Details: " + err.Error())
		resultJSON := buildJSONResponse(NODE_IS_INACCESSIBLE, common.EstimateReceiveResponse{})
		fmt.Println(string(resultJSON))
		return
	}

	commandStr := string(command.ToBytes())
	switch result.Code {
	case OK:
		if len(result.Tokens) == 0 {
			logger.Error("Node return invalid result tokens size on command: " + commandStr)
			resultJSON := buildJSONResponse(ENGINE_UNEXPECTED_ERROR, common.EstimateReceiveResponse{})
			fmt.Println(string(resultJSON))
			return
		}
		resultJSON := buildJSONResponse(OK, common.EstimateReceiveResponse{EstimatedReceiveAmount: result.Tokens[0]})
		fmt.Println(string(resultJSON))
		return
	case 401:
		logger.Error("Node return unexpected error code 401 on command: " + commandStr)
	case 462:
		logger.Info("Node return no cached paths (462) on command: " + commandStr)
	default:
		logger.Error("Node return wrong command result: " + strconv.Itoa(result.Code) +
			" on command: " + commandStr)
	}
	resultJSON := buildJSONResponse(result.Code, common.EstimateReceiveResponse{})
	fmt.Println(string(resultJSON))
}

func (handler *NodeHandler) PaymentExchange() {
	if len(Addresses) == 0 {
		logger.Error("Bad request: there are no contractor addresses parameters in payment exchange request")
		fmt.Println("Bad request: there are no contractor addresses parameters")
		return
	}

	if !common.ValidateSettlementLineAmount(Amount) {
		logger.Error("Bad request: invalid amount parameter in payment exchange request")
		fmt.Println("Bad request: invalid amount parameter")
		return
	}

	if ReceiverEquivalent == "" {
		logger.Error("Bad request: missing receiver-eq parameter in payment exchange request")
		fmt.Println("Bad request: missing receiver-eq parameter")
		return
	}
	if !common.ValidateInt(ReceiverEquivalent) {
		logger.Error("Bad request: invalid receiver-eq parameter in payment exchange request")
		fmt.Println("Bad request: invalid receiver-eq parameter")
		return
	}

	if len(ExchangeEquivalents) == 0 {
		logger.Error("Bad request: there are no exchange equivalents parameters in payment exchange request")
		fmt.Println("Bad request: there are no exchange equivalents parameters")
		return
	}
	if len(ExchangeEquivalents) > 5 {
		logger.Error("Bad request: too many exchange equivalents parameters in payment exchange request")
		fmt.Println("Bad request: too many exchange equivalents parameters")
		return
	}
	for _, exchangeEquivalent := range ExchangeEquivalents {
		if !common.ValidateInt(exchangeEquivalent) {
			logger.Error("Bad request: invalid exchange equivalent parameter in payment exchange request")
			fmt.Println("Bad request: invalid exchange equivalent parameter")
			return
		}
	}

	var addresses []string
	for idx := range len(Addresses) {
		addressType, address := common.ValidateAddress(Addresses[idx])
		if addressType == "" {
			logger.Error("Bad request: invalid address parameter in payment exchange request")
			fmt.Println("Bad request: invalid address parameter")
			return
		}
		addresses = append(addresses, addressType, address)
	}

	commandParts := append([]string{strconv.Itoa(len(Addresses))}, addresses...)
	commandParts = append([]string{"CREATE:contractors/transactions/exchange"}, commandParts...)
	commandParts = append(commandParts, Amount, ReceiverEquivalent)
	commandParts = append(commandParts, ExchangeEquivalents...)
	if Payload != "" {
		commandParts = append(commandParts, Payload)
	}

	var command *Command
	if TransactionUUID != "" {
		parsedUUID, err := uuid.Parse(TransactionUUID)
		if err != nil {
			logger.Error("Bad request: invalid transaction-uuid parameter in payment exchange request")
			fmt.Println("Bad request: invalid transaction-uuid parameter")
			return
		}
		command = NewCommandWithUUID(parsedUUID, commandParts...)
	} else {
		command = NewCommand(commandParts...)
	}

	go handler.paymentResult(command)
}

func (handler *NodeHandler) Payment() {
	if len(Addresses) == 0 {
		logger.Error("Bad request: there are no contractor addresses parameters in payment request")
		fmt.Println("Bad request: there are no contractor addresses parameters")
		return
	}

	if !common.ValidateSettlementLineAmount(Amount) {
		logger.Error("Bad request: invalid amount parameter in payment request")
		fmt.Println("Bad request: invalid amount parameter")
		return
	}

	if !common.ValidateInt(Equivalent) {
		logger.Error("Bad request: invalid equivalent parameter in payment request")
		fmt.Println("Bad request: invalid equivalent parameter")
		return
	}

	var addresses []string
	for idx := range len(Addresses) {
		addressType, address := common.ValidateAddress(Addresses[idx])
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

func (handler *NodeHandler) paymentResult(command *Command) {
	err := handler.Node.SendCommand(command)
	if err != nil {
		logger.Error("Can't send command: " + string(command.ToBytes()) + " to node. Details: " + err.Error())
		resultJSON := buildJSONResponse(COMMAND_TRANSFERRING_ERROR, common.PaymentResponse{})
		fmt.Println(string(resultJSON))
		return
	}

	result, err := handler.Node.GetResult(command, common.PAYMENT_OPERATION_TIMEOUT)
	if err != nil {
		logger.Error("Node is inaccessible during processing command: " +
			string(command.ToBytes()) + ". Details: " + err.Error())
		resultJSON := buildJSONResponse(NODE_IS_INACCESSIBLE, common.PaymentResponse{})
		fmt.Println(string(resultJSON))
		return
	}

	if result.Code != CREATED && result.Code != ENGINE_NO_EQUIVALENT {
		logger.Error("Node return wrong command result: " + strconv.Itoa(result.Code) +
			" on command: " + string(command.ToBytes()))
		resultJSON := buildJSONResponse(result.Code, common.PaymentResponse{})
		fmt.Println(string(resultJSON))
		return
	}
	if result.Code == ENGINE_NO_EQUIVALENT {
		logger.Info("Node hasn't equivalent for command: " + string(command.ToBytes()))
		resultJSON := buildJSONResponse(result.Code, common.PaymentResponse{})
		fmt.Println(string(resultJSON))
		return
	}

	if len(result.Tokens) == 0 {
		logger.Error("Node return invalid result tokens size on command: " + string(command.ToBytes()))
		resultJSON := buildJSONResponse(ENGINE_UNEXPECTED_ERROR, common.PaymentResponse{})
		fmt.Println(string(resultJSON))
		return
	}

	resultJSON := buildJSONResponse(OK, common.PaymentResponse{TransactionUUID: result.Tokens[0]})
	fmt.Println(string(resultJSON))
}
