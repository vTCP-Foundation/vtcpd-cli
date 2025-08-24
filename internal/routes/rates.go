package routes

import (
	"fmt"
	"math"
	"net/http"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
	"github.com/vTCP-Foundation/vtcpd-cli/internal/common"
	"github.com/vTCP-Foundation/vtcpd-cli/internal/handler"
	"github.com/vTCP-Foundation/vtcpd-cli/internal/logger"
)

const (
	DECIMALS_101  = 2
	DECIMALS_1001 = 8
	DECIMALS_1002 = 8
	DECIMALS_2002 = 6
)

var decimalsMap = map[string]int{
	"101":  DECIMALS_101,
	"1001": DECIMALS_1001,
	"1002": DECIMALS_1002,
	"2002": DECIMALS_2002,
}

func (router *RoutesHandler) SetRate(w http.ResponseWriter, r *http.Request) {
	logger.Info("SetRate controller")
	url, err := preprocessRequest(r)
	if err != nil {
		logger.Error("Bad request: invalid security parameters: " + err.Error())
		w.WriteHeader(BAD_REQUEST)
		return
	}

	equivalentFrom := mux.Vars(r)["equivalent_from"]
	equivalentTo := mux.Vars(r)["equivalent_to"]

	// Validate equivalents exist in decimals map
	if _, exists := decimalsMap[equivalentFrom]; !exists {
		logger.Error("Bad request: unknown equivalent scale for equivalent_from: " + equivalentFrom)
		w.WriteHeader(BAD_REQUEST)
		return
	}
	if _, exists := decimalsMap[equivalentTo]; !exists {
		logger.Error("Bad request: unknown equivalent scale for equivalent_to: " + equivalentTo)
		w.WriteHeader(BAD_REQUEST)
		return
	}

	// Get query parameters
	realRate := r.FormValue("real_rate")
	valueStr := r.FormValue("value")
	shiftStr := r.FormValue("shift")
	minExchangeAmount := r.FormValue("min_exchange_amount")
	maxExchangeAmount := r.FormValue("max_exchange_amount")

	// Validate mode exclusivity
	hasRealRate := realRate != ""
	hasValueShift := valueStr != "" && shiftStr != ""

	if hasRealRate && hasValueShift {
		logger.Error("Bad request: real_rate and (value, shift) are mutually exclusive: " + url)
		w.WriteHeader(BAD_REQUEST)
		return
	}

	if !hasRealRate && !hasValueShift {
		logger.Error("Bad request: either real_rate or (value, shift) is required: " + url)
		w.WriteHeader(BAD_REQUEST)
		return
	}

	var finalValue string
	var finalShift int16

	if hasRealRate {
		// Process real_rate mode
		value, shift, err := parseAndValidateRealRate(realRate, equivalentFrom, equivalentTo)
		if err != nil {
			logger.Error("Bad request: " + err.Error() + ": " + url)
			w.WriteHeader(BAD_REQUEST)
			return
		}
		finalValue = value
		finalShift = shift
	} else {
		// Process native mode
		shift64, err := strconv.ParseInt(shiftStr, 10, 16)
		if err != nil {
			logger.Error("Bad request: shift is out of int16 range: " + url)
			w.WriteHeader(BAD_REQUEST)
			return
		}

		// Apply scale adjustment
		decimalsFrom := decimalsMap[equivalentFrom]
		decimalsTo := decimalsMap[equivalentTo]
		adjustedShift := shift64 - int64(decimalsFrom-decimalsTo)

		if adjustedShift < math.MinInt16 || adjustedShift > math.MaxInt16 {
			logger.Error("Bad request: shift is out of int16 range after scale adjustment: " + url)
			w.WriteHeader(BAD_REQUEST)
			return
		}

		finalValue = valueStr
		finalShift = int16(adjustedShift)
	}

	// Build command arguments
	args := []string{"SET:RATE", equivalentFrom, equivalentTo, finalValue, strconv.Itoa(int(finalShift))}
	if minExchangeAmount != "" {
		args = append(args, minExchangeAmount)
	} else {
		args = append(args, "")
	}
	if maxExchangeAmount != "" {
		args = append(args, maxExchangeAmount)
	} else {
		args = append(args, "")
	}

	command := handler.NewCommand(args...)

	err = router.nodeHandler.Node.SendCommand(command)
	if err != nil {
		logger.Error("Can't send command: " + string(command.ToBytes()) + " to node. Details: " + err.Error())
		writeHTTPResponse(w, COMMAND_TRANSFERRING_ERROR, common.RatesActionResponse{})
		return
	}

	result, err := router.nodeHandler.Node.GetResult(command, common.SETTLEMENT_LINE_RESULT_TIMEOUT)
	if err != nil {
		logger.Error("Node is inaccessible during processing command: " +
			string(command.ToBytes()) + ". Details: " + err.Error())
		writeHTTPResponse(w, NODE_IS_INACCESSIBLE, common.RatesActionResponse{})
		return
	}

	if result.Code != OK {
		logger.Error("Node return wrong command result: " + strconv.Itoa(result.Code) +
			" on command: " + string(command.ToBytes()))
	}

	writeHTTPResponse(w, result.Code, common.RatesActionResponse{})
}

func (router *RoutesHandler) GetRate(w http.ResponseWriter, r *http.Request) {
	logger.Info("GetRate controller")
	_, err := preprocessRequest(r)
	if err != nil {
		logger.Error("Bad request: invalid security parameters: " + err.Error())
		w.WriteHeader(BAD_REQUEST)
		return
	}

	equivalentFrom := mux.Vars(r)["equivalent_from"]
	equivalentTo := mux.Vars(r)["equivalent_to"]

	command := handler.NewCommand("GET:RATE", equivalentFrom, equivalentTo)

	err = router.nodeHandler.Node.SendCommand(command)
	if err != nil {
		logger.Error("Can't send command: " + string(command.ToBytes()) + " to node. Details: " + err.Error())
		writeHTTPResponse(w, COMMAND_TRANSFERRING_ERROR, common.RateGetResponse{})
		return
	}

	result, err := router.nodeHandler.Node.GetResult(command, common.SETTLEMENT_LINE_RESULT_TIMEOUT)
	if err != nil {
		logger.Error("Node is inaccessible during processing command: " +
			string(command.ToBytes()) + ". Details: " + err.Error())
		writeHTTPResponse(w, NODE_IS_INACCESSIBLE, common.RateGetResponse{})
		return
	}

	if result.Code != OK && result.Code != NODE_NOT_FOUND {
		logger.Error("Node return wrong command result: " + strconv.Itoa(result.Code) +
			" on command: " + string(command.ToBytes()))
		writeHTTPResponse(w, result.Code, common.RateGetResponse{})
		return
	}

	if result.Code == NODE_NOT_FOUND {
		logger.Info("Node hasn't rate for command: " + string(command.ToBytes()))
		writeHTTPResponse(w, result.Code, common.RateGetResponse{})
		return
	}

	if len(result.Tokens) < 7 {
		logger.Error("Node return invalid result tokens size on command: " + string(command.ToBytes()))
		writeHTTPResponse(w, ENGINE_UNEXPECTED_ERROR, common.RateGetResponse{})
		return
	}

	// Parse shift from token
	shift, err := strconv.ParseInt(result.Tokens[3], 10, 16)
	if err != nil {
		logger.Error("Node return invalid shift token on command: " + string(command.ToBytes()) + ". Details: " + err.Error())
		writeHTTPResponse(w, ENGINE_UNEXPECTED_ERROR, common.RateGetResponse{})
		return
	}

	// Compute real rate for display
	realRate := computeRealRateString(result.Tokens[2], int16(shift), equivalentFrom, equivalentTo)

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
	writeHTTPResponse(w, OK, response)
}

func (router *RoutesHandler) ListRates(w http.ResponseWriter, r *http.Request) {
	logger.Info("ListRates controller")
	_, err := preprocessRequest(r)
	if err != nil {
		logger.Error("Bad request: invalid security parameters: " + err.Error())
		w.WriteHeader(BAD_REQUEST)
		return
	}

	command := handler.NewCommand("LIST:RATES")

	err = router.nodeHandler.Node.SendCommand(command)
	if err != nil {
		logger.Error("Can't send command: " + string(command.ToBytes()) + " to node. Details: " + err.Error())
		writeHTTPResponse(w, COMMAND_TRANSFERRING_ERROR, common.RatesListResponse{})
		return
	}

	result, err := router.nodeHandler.Node.GetResult(command, common.SETTLEMENT_LINE_RESULT_TIMEOUT)
	if err != nil {
		logger.Error("Node is inaccessible during processing command: " +
			string(command.ToBytes()) + ". Details: " + err.Error())
		writeHTTPResponse(w, NODE_IS_INACCESSIBLE, common.RatesListResponse{})
		return
	}

	if result.Code != OK {
		logger.Error("Node return wrong command result: " + strconv.Itoa(result.Code) +
			" on command: " + string(command.ToBytes()))
		writeHTTPResponse(w, result.Code, common.RatesListResponse{})
		return
	}

	if len(result.Tokens) == 0 {
		logger.Error("Node return invalid result tokens size on command: " + string(command.ToBytes()))
		writeHTTPResponse(w, ENGINE_UNEXPECTED_ERROR, common.RatesListResponse{})
		return
	}

	ratesCount, err := strconv.Atoi(result.Tokens[0])
	if err != nil {
		logger.Error("Node return invalid token on command: " + string(command.ToBytes()) +
			". Details: " + err.Error())
		writeHTTPResponse(w, ENGINE_UNEXPECTED_ERROR, common.RatesListResponse{})
		return
	}

	if ratesCount == 0 {
		writeHTTPResponse(w, OK, common.RatesListResponse{Count: ratesCount})
		return
	}

	response := common.RatesListResponse{Count: ratesCount}
	response.Rates = make([]common.RateItem, 0, ratesCount)

	for i := range ratesCount {
		// Parse shift from token
		shift, err := strconv.ParseInt(result.Tokens[i*7+4], 10, 16)
		if err != nil {
			logger.Error("Node return invalid shift token on command: " + string(command.ToBytes()) + ". Details: " + err.Error())
			writeHTTPResponse(w, ENGINE_UNEXPECTED_ERROR, common.RatesListResponse{})
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
	writeHTTPResponse(w, OK, response)
}

func (router *RoutesHandler) DeleteRate(w http.ResponseWriter, r *http.Request) {
	logger.Info("DeleteRate controller")
	_, err := preprocessRequest(r)
	if err != nil {
		logger.Error("Bad request: invalid security parameters: " + err.Error())
		w.WriteHeader(BAD_REQUEST)
		return
	}

	equivalentFrom := mux.Vars(r)["equivalent_from"]
	equivalentTo := mux.Vars(r)["equivalent_to"]

	command := handler.NewCommand("DEL:RATE", equivalentFrom, equivalentTo)

	err = router.nodeHandler.Node.SendCommand(command)
	if err != nil {
		logger.Error("Can't send command: " + string(command.ToBytes()) + " to node. Details: " + err.Error())
		writeHTTPResponse(w, COMMAND_TRANSFERRING_ERROR, common.RatesActionResponse{})
		return
	}

	result, err := router.nodeHandler.Node.GetResult(command, common.SETTLEMENT_LINE_RESULT_TIMEOUT)
	if err != nil {
		logger.Error("Node is inaccessible during processing command: " +
			string(command.ToBytes()) + ". Details: " + err.Error())
		writeHTTPResponse(w, NODE_IS_INACCESSIBLE, common.RatesActionResponse{})
		return
	}

	if result.Code != OK {
		logger.Error("Node return wrong command result: " + strconv.Itoa(result.Code) +
			" on command: " + string(command.ToBytes()))
	}

	writeHTTPResponse(w, result.Code, common.RatesActionResponse{})
}

func (router *RoutesHandler) ClearRates(w http.ResponseWriter, r *http.Request) {
	logger.Info("ClearRates controller")
	_, err := preprocessRequest(r)
	if err != nil {
		logger.Error("Bad request: invalid security parameters: " + err.Error())
		w.WriteHeader(BAD_REQUEST)
		return
	}

	command := handler.NewCommand("CLEAR:RATES")

	err = router.nodeHandler.Node.SendCommand(command)
	if err != nil {
		logger.Error("Can't send command: " + string(command.ToBytes()) + " to node. Details: " + err.Error())
		writeHTTPResponse(w, COMMAND_TRANSFERRING_ERROR, common.RatesActionResponse{})
		return
	}

	result, err := router.nodeHandler.Node.GetResult(command, common.SETTLEMENT_LINE_RESULT_TIMEOUT)
	if err != nil {
		logger.Error("Node is inaccessible during processing command: " +
			string(command.ToBytes()) + ". Details: " + err.Error())
		writeHTTPResponse(w, NODE_IS_INACCESSIBLE, common.RatesActionResponse{})
		return
	}

	if result.Code != OK {
		logger.Error("Node return wrong command result: " + strconv.Itoa(result.Code) +
			" on command: " + string(command.ToBytes()))
	}

	writeHTTPResponse(w, result.Code, common.RatesActionResponse{})
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
	decimalsFrom := decimalsMap[equivalentFrom]
	decimalsTo := decimalsMap[equivalentTo]
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
	decimalsFrom := decimalsMap[equivalentFrom]
	decimalsTo := decimalsMap[equivalentTo]
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
