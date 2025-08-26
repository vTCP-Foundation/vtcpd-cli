package handler

import (
	"fmt"
	"strconv"

	"github.com/vTCP-Foundation/vtcpd-cli/internal/common"
	"github.com/vTCP-Foundation/vtcpd-cli/internal/logger"
)

func (handler *NodeHandler) Rates() {
	switch CommandType {
	case "set":
		handler.setRate()
	case "set-native":
		handler.setNativeRate()
	case "get":
		handler.getRate()
	case "list":
		handler.listRates()
	case "del":
		handler.deleteRate()
	case "clear":
		handler.clearRates()
	default:
		logger.Error("Invalid rates command " + CommandType)
		fmt.Println("Invalid rates command")
		return
	}
}

func (handler *NodeHandler) setRate() {
	// Validate required parameters
	if EquivalentFrom == "" {
		logger.Error("Bad request: --from parameter is required for rates set command")
		fmt.Println("Bad request: --from parameter is required")
		return
	}
	if EquivalentTo == "" {
		logger.Error("Bad request: --to parameter is required for rates set command")
		fmt.Println("Bad request: --to parameter is required")
		return
	}

	// Validate equivalents exist in decimals map
	if _, exists := common.DecimalsMap[EquivalentFrom]; !exists {
		logger.Error("Bad request: unknown equivalent scale for equivalent_from: " + EquivalentFrom)
		fmt.Println("Bad request: unknown equivalent scale for equivalent_from")
		return
	}
	if _, exists := common.DecimalsMap[EquivalentTo]; !exists {
		logger.Error("Bad request: unknown equivalent scale for equivalent_to: " + EquivalentTo)
		fmt.Println("Bad request: unknown equivalent scale for equivalent_to")
		return
	}

	// Validate real rate parameter is provided
	if RealRate == "" {
		logger.Error("Bad request: --real parameter is required for rates set command")
		fmt.Println("Bad request: --real parameter is required")
		return
	}

	// Parse and validate real rate
	value, shift, err := common.ParseAndValidateRealRate(RealRate, EquivalentFrom, EquivalentTo)
	if err != nil {
		logger.Error("Bad request: " + err.Error())
		fmt.Println("Bad request: " + err.Error())
		return
	}

	// Build command arguments
	args := []string{"SET:RATE", EquivalentFrom, EquivalentTo, value, strconv.Itoa(int(shift))}
	if MinExchangeAmount != "" {
		args = append(args, MinExchangeAmount)
	} else {
		args = append(args, "")
	}
	if MaxExchangeAmount != "" {
		args = append(args, MaxExchangeAmount)
	} else {
		args = append(args, "")
	}

	command := NewCommand(args...)
	go handler.actionRatesGetResult(command)
}

func (handler *NodeHandler) setNativeRate() {
	// Validate required parameters
	if EquivalentFrom == "" {
		logger.Error("Bad request: --from parameter is required for rates set-native command")
		fmt.Println("Bad request: --from parameter is required")
		return
	}
	if EquivalentTo == "" {
		logger.Error("Bad request: --to parameter is required for rates set-native command")
		fmt.Println("Bad request: --to parameter is required")
		return
	}

	// Validate equivalents exist in decimals map
	if _, exists := common.DecimalsMap[EquivalentFrom]; !exists {
		logger.Error("Bad request: unknown equivalent scale for equivalent_from: " + EquivalentFrom)
		fmt.Println("Bad request: unknown equivalent scale for equivalent_from")
		return
	}
	if _, exists := common.DecimalsMap[EquivalentTo]; !exists {
		logger.Error("Bad request: unknown equivalent scale for equivalent_to: " + EquivalentTo)
		fmt.Println("Bad request: unknown equivalent scale for equivalent_to")
		return
	}

	// Validate required parameters for native mode
	if Value == "" {
		logger.Error("Bad request: --value parameter is required for rates set-native command")
		fmt.Println("Bad request: --value parameter is required")
		return
	}
	if Shift == "" {
		logger.Error("Bad request: --shift parameter is required for rates set-native command")
		fmt.Println("Bad request: --shift parameter is required")
		return
	}

	// Build command arguments
	args := []string{"SET:RATE", EquivalentFrom, EquivalentTo, Value, Shift}
	if MinExchangeAmount != "" {
		args = append(args, MinExchangeAmount)
	} else {
		args = append(args, "")
	}
	if MaxExchangeAmount != "" {
		args = append(args, MaxExchangeAmount)
	} else {
		args = append(args, "")
	}

	command := NewCommand(args...)
	go handler.actionRatesGetResult(command)
}

func (handler *NodeHandler) getRate() {
	// Validate required parameters
	if EquivalentFrom == "" {
		logger.Error("Bad request: --from parameter is required for rates get command")
		fmt.Println("Bad request: --from parameter is required")
		return
	}
	if EquivalentTo == "" {
		logger.Error("Bad request: --to parameter is required for rates get command")
		fmt.Println("Bad request: --to parameter is required")
		return
	}

	command := NewCommand("GET:RATE", EquivalentFrom, EquivalentTo)
	go handler.getRateResult(command)
}

func (handler *NodeHandler) listRates() {
	command := NewCommand("LIST:RATES")
	go handler.listRatesResult(command)
}

func (handler *NodeHandler) deleteRate() {
	// Validate required parameters
	if EquivalentFrom == "" {
		logger.Error("Bad request: --from parameter is required for rates del command")
		fmt.Println("Bad request: --from parameter is required")
		return
	}
	if EquivalentTo == "" {
		logger.Error("Bad request: --to parameter is required for rates del command")
		fmt.Println("Bad request: --to parameter is required")
		return
	}

	command := NewCommand("DEL:RATE", EquivalentFrom, EquivalentTo)
	go handler.actionRatesGetResult(command)
}

func (handler *NodeHandler) clearRates() {
	command := NewCommand("CLEAR:RATES")
	go handler.actionRatesGetResult(command)
}

func (handler *NodeHandler) actionRatesGetResult(command *Command) {
	err := handler.Node.SendCommand(command)
	if err != nil {
		logger.Error("Can't send command: " + string(command.ToBytes()) + " to node. Details: " + err.Error())
		resultJSON := buildJSONResponse(COMMAND_TRANSFERRING_ERROR, common.RatesActionResponse{})
		fmt.Println(string(resultJSON))
		return
	}

	result, err := handler.Node.GetResult(command, common.SETTLEMENT_LINE_RESULT_TIMEOUT)
	if err != nil {
		logger.Error("Node is inaccessible during processing command: " +
			string(command.ToBytes()) + ". Details: " + err.Error())
		resultJSON := buildJSONResponse(NODE_IS_INACCESSIBLE, common.RatesActionResponse{})
		fmt.Println(string(resultJSON))
		return
	}

	if result.Code != OK {
		logger.Error("Node return wrong command result: " + strconv.Itoa(result.Code) +
			" on command: " + string(command.ToBytes()))
	}

	resultJSON := buildJSONResponse(result.Code, common.RatesActionResponse{})
	fmt.Println(string(resultJSON))
}

func (handler *NodeHandler) getRateResult(command *Command) {
	err := handler.Node.SendCommand(command)
	if err != nil {
		logger.Error("Can't send command: " + string(command.ToBytes()) + " to node. Details: " + err.Error())
		resultJSON := buildJSONResponse(COMMAND_TRANSFERRING_ERROR, common.RateGetResponse{})
		fmt.Println(string(resultJSON))
		return
	}

	result, err := handler.Node.GetResult(command, common.SETTLEMENT_LINE_RESULT_TIMEOUT)
	if err != nil {
		logger.Error("Node is inaccessible during processing command: " +
			string(command.ToBytes()) + ". Details: " + err.Error())
		resultJSON := buildJSONResponse(NODE_IS_INACCESSIBLE, common.RateGetResponse{})
		fmt.Println(string(resultJSON))
		return
	}

	if result.Code != OK && result.Code != NODE_NOT_FOUND {
		logger.Error("Node return wrong command result: " + strconv.Itoa(result.Code) +
			" on command: " + string(command.ToBytes()))
		resultJSON := buildJSONResponse(result.Code, common.RateGetResponse{})
		fmt.Println(string(resultJSON))
		return
	}

	if result.Code == NODE_NOT_FOUND {
		logger.Info("Node hasn't rate for command: " + string(command.ToBytes()))
		resultJSON := buildJSONResponse(result.Code, common.RateGetResponse{})
		fmt.Println(string(resultJSON))
		return
	}

	if len(result.Tokens) < 7 {
		logger.Error("Node return invalid result tokens size on command: " + string(command.ToBytes()))
		resultJSON := buildJSONResponse(ENGINE_UNEXPECTED_ERROR, common.RateGetResponse{})
		fmt.Println(string(resultJSON))
		return
	}

	// Parse shift from token
	shift, err := strconv.ParseInt(result.Tokens[3], 10, 16)
	if err != nil {
		logger.Error("Node return invalid shift token on command: " + string(command.ToBytes()) + ". Details: " + err.Error())
		resultJSON := buildJSONResponse(ENGINE_UNEXPECTED_ERROR, common.RateGetResponse{})
		fmt.Println(string(resultJSON))
		return
	}

	// Compute real rate for display
	realRate := common.ComputeRealRateString(result.Tokens[2], int16(shift), result.Tokens[0], result.Tokens[1])

	rate := common.RateItem{
		EquivalentFrom:            result.Tokens[0],
		EquivalentTo:              result.Tokens[1],
		Value:                     result.Tokens[2],
		Shift:                     int16(shift),
		RealRate:                  realRate,
		MinExchangeAmount:         result.Tokens[4],
		MaxExchangeAmount:         result.Tokens[5],
		ExpiresAtUnixMicroseconds: result.Tokens[6],
	}

	response := common.RateGetResponse{Rate: rate}
	resultJSON := buildJSONResponse(OK, response)
	fmt.Println(string(resultJSON))
}

func (handler *NodeHandler) listRatesResult(command *Command) {
	err := handler.Node.SendCommand(command)
	if err != nil {
		logger.Error("Can't send command: " + string(command.ToBytes()) + " to node. Details: " + err.Error())
		resultJSON := buildJSONResponse(COMMAND_TRANSFERRING_ERROR, common.RatesListResponse{})
		fmt.Println(string(resultJSON))
		return
	}

	result, err := handler.Node.GetResult(command, common.SETTLEMENT_LINE_RESULT_TIMEOUT)
	if err != nil {
		logger.Error("Node is inaccessible during processing command: " +
			string(command.ToBytes()) + ". Details: " + err.Error())
		resultJSON := buildJSONResponse(NODE_IS_INACCESSIBLE, common.RatesListResponse{})
		fmt.Println(string(resultJSON))
		return
	}

	if result.Code != OK {
		logger.Error("Node return wrong command result: " + strconv.Itoa(result.Code) +
			" on command: " + string(command.ToBytes()))
		resultJSON := buildJSONResponse(result.Code, common.RatesListResponse{})
		fmt.Println(string(resultJSON))
		return
	}

	if len(result.Tokens) == 0 {
		logger.Error("Node return invalid result tokens size on command: " + string(command.ToBytes()))
		resultJSON := buildJSONResponse(ENGINE_UNEXPECTED_ERROR, common.RatesListResponse{})
		fmt.Println(string(resultJSON))
		return
	}

	ratesCount, err := strconv.Atoi(result.Tokens[0])
	if err != nil {
		logger.Error("Node return invalid token on command: " + string(command.ToBytes()) +
			". Details: " + err.Error())
		resultJSON := buildJSONResponse(ENGINE_UNEXPECTED_ERROR, common.RatesListResponse{})
		fmt.Println(string(resultJSON))
		return
	}

	if ratesCount == 0 {
		resultJSON := buildJSONResponse(OK, common.RatesListResponse{Count: ratesCount})
		fmt.Println(string(resultJSON))
		return
	}

	response := common.RatesListResponse{Count: ratesCount}
	response.Rates = make([]common.RateItem, 0, ratesCount)

	for i := range ratesCount {
		// Parse shift from token
		shift, err := strconv.ParseInt(result.Tokens[i*7+4], 10, 16)
		if err != nil {
			logger.Error("Node return invalid shift token on command: " + string(command.ToBytes()) + ". Details: " + err.Error())
			resultJSON := buildJSONResponse(ENGINE_UNEXPECTED_ERROR, common.RatesListResponse{})
			fmt.Println(string(resultJSON))
			return
		}

		equivalentFrom := result.Tokens[i*7+1]
		equivalentTo := result.Tokens[i*7+2]
		value := result.Tokens[i*7+3]

		// Compute real rate for display
		realRate := common.ComputeRealRateString(value, int16(shift), equivalentFrom, equivalentTo)

		response.Rates = append(response.Rates, common.RateItem{
			EquivalentFrom:            equivalentFrom,
			EquivalentTo:              equivalentTo,
			Value:                     value,
			Shift:                     int16(shift),
			RealRate:                  realRate,
			MinExchangeAmount:         result.Tokens[i*7+5],
			MaxExchangeAmount:         result.Tokens[i*7+6],
			ExpiresAtUnixMicroseconds: result.Tokens[i*7+7],
		})
	}
	resultJSON := buildJSONResponse(OK, response)
	fmt.Println(string(resultJSON))
}

