package routes

import (
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/vTCP-Foundation/vtcpd-cli/internal/common"
	"github.com/vTCP-Foundation/vtcpd-cli/internal/handler"
	"github.com/vTCP-Foundation/vtcpd-cli/internal/logger"
)

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
	if _, exists := common.DecimalsMap[equivalentFrom]; !exists {
		logger.Error("Bad request: unknown equivalent scale for equivalent_from: " + equivalentFrom)
		w.WriteHeader(BAD_REQUEST)
		return
	}
	if _, exists := common.DecimalsMap[equivalentTo]; !exists {
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

	if hasRealRate {
		// Process real_rate mode
		value, shift, err := common.ParseAndValidateRealRate(realRate, equivalentFrom, equivalentTo)
		if err != nil {
			logger.Error("Bad request: " + err.Error() + ": " + url)
			w.WriteHeader(BAD_REQUEST)
			return
		}
		valueStr = value
		shiftStr = strconv.Itoa(int(shift))
	}

	// Build command arguments
	args := []string{"SET:RATE", equivalentFrom, equivalentTo, valueStr, shiftStr}
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
	realRate := common.ComputeRealRateString(result.Tokens[2], int16(shift), equivalentFrom, equivalentTo)

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

