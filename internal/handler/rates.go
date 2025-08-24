package handler

import (
	"fmt"
	"math"
	"strconv"
	"strings"

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
	value, shift, err := parseAndValidateRealRate(RealRate, EquivalentFrom, EquivalentTo)
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

	// Parse shift
	shift64, err := strconv.ParseInt(Shift, 10, 16)
	if err != nil {
		logger.Error("Bad request: shift is out of int16 range")
		fmt.Println("Bad request: shift is out of int16 range")
		return
	}

	// Apply scale adjustment
	decimalsFrom := common.DecimalsMap[EquivalentFrom]
	decimalsTo := common.DecimalsMap[EquivalentTo]
	adjustedShift := shift64 - int64(decimalsFrom-decimalsTo)

	if adjustedShift < math.MinInt16 || adjustedShift > math.MaxInt16 {
		logger.Error("Bad request: shift is out of int16 range after scale adjustment")
		fmt.Println("Bad request: shift is out of int16 range after scale adjustment")
		return
	}

	// Build command arguments
	args := []string{"SET:RATE", EquivalentFrom, EquivalentTo, Value, strconv.Itoa(int(adjustedShift))}
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
	realRate := computeRealRateString(result.Tokens[2], int16(shift), result.Tokens[0], result.Tokens[1])

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
		realRate := computeRealRateString(value, int16(shift), equivalentFrom, equivalentTo)

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

// parseAndValidateRealRate converts a real decimal rate to native (value, shift) format
func parseAndValidateRealRate(realRate string, equivalentFrom string, equivalentTo string) (string, int16, error) {
	// Check for maximum 12 fractional digits
	parts := strings.Split(realRate, ".")
	if len(parts) == 2 && len(parts[1]) > 12 {
		return "", 0, fmt.Errorf("real_rate has more than 12 fractional digits")
	}

	// Parse the decimal value
	value, shift, err := normalizeDecimal(realRate)
	if err != nil {
		return "", 0, fmt.Errorf("invalid real_rate format: %v", err)
	}

	// Apply scale difference
	decimalsFrom := common.DecimalsMap[equivalentFrom]
	decimalsTo := common.DecimalsMap[equivalentTo]
	adjustedShift := int64(shift) - int64(decimalsFrom-decimalsTo)

	// Validate int16 range
	if adjustedShift < math.MinInt16 || adjustedShift > math.MaxInt16 {
		return "", 0, fmt.Errorf("shift is out of int16 range after scale adjustment")
	}

	return value, int16(adjustedShift), nil
}

// normalizeDecimal converts decimal string to (value, shift) format
func normalizeDecimal(decimal string) (string, int16, error) {
	// Remove leading/trailing whitespace
	decimal = strings.TrimSpace(decimal)

	// Handle empty string
	if decimal == "" {
		return "0", 0, nil
	}

	// Handle negative sign
	negative := false
	if strings.HasPrefix(decimal, "-") {
		negative = true
		decimal = decimal[1:]
	} else if strings.HasPrefix(decimal, "+") {
		decimal = decimal[1:]
	}

	// Split by decimal point
	parts := strings.Split(decimal, ".")
	if len(parts) > 2 {
		return "", 0, fmt.Errorf("invalid decimal format")
	}

	var integerPart, fractionalPart string
	if len(parts) == 1 {
		integerPart = parts[0]
		fractionalPart = ""
	} else {
		integerPart = parts[0]
		fractionalPart = parts[1]
	}

	// Handle empty integer part
	if integerPart == "" {
		integerPart = "0"
	}

	// Remove leading zeros from integer part
	integerPart = strings.TrimLeft(integerPart, "0")
	if integerPart == "" {
		integerPart = "0"
	}

	// Remove trailing zeros from fractional part
	fractionalPart = strings.TrimRight(fractionalPart, "0")

	// Build the value
	value := integerPart + fractionalPart
	shift := len(fractionalPart)

	// Remove leading zeros from final value
	value = strings.TrimLeft(value, "0")
	if value == "" {
		value = "0"
		shift = 0
	}

	// Add negative sign if needed
	if negative && value != "0" {
		value = "-" + value
	}

	return value, int16(shift), nil
}

// computeRealRateString computes the real decimal rate from native (value, shift) format
func computeRealRateString(value string, shift int16, equivalentFrom string, equivalentTo string) string {
	// Reverse scale adjustment
	decimalsFrom := common.DecimalsMap[equivalentFrom]
	decimalsTo := common.DecimalsMap[equivalentTo]
	originalShift := int(shift) + (decimalsFrom - decimalsTo)

	return applyShiftToValue(value, originalShift)
}

// applyShiftToValue applies decimal shift to a value string
func applyShiftToValue(value string, shift int) string {
	if value == "0" {
		return "0"
	}

	negative := false
	if strings.HasPrefix(value, "-") {
		negative = true
		value = value[1:]
	}

	// If shift is 0, return as is
	if shift == 0 {
		if negative {
			return "-" + value
		}
		return value
	}

	// If shift is positive, we need to place decimal point within the number
	if shift > 0 {
		if len(value) <= shift {
			// Need to add leading zeros
			zerosNeeded := shift - len(value) + 1
			result := "0." + strings.Repeat("0", zerosNeeded-1) + value
			if negative {
				return "-" + result
			}
			return result
		} else {
			// Place decimal point within the number
			integerPart := value[:len(value)-shift]
			fractionalPart := value[len(value)-shift:]
			result := integerPart + "." + fractionalPart
			if negative {
				return "-" + result
			}
			return result
		}
	} else {
		// shift is negative, add zeros to the right
		result := value + strings.Repeat("0", -shift)
		if negative {
			return "-" + result
		}
		return result
	}
}