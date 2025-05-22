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

func (router *RoutesHandler) InitSettlementLine(w http.ResponseWriter, r *http.Request) {
	logger.Info("InitSettlementLine controller")
	url, err := preprocessRequest(r)
	if err != nil {
		logger.Error("Bad request: invalid security parameters: " + err.Error())
		w.WriteHeader(BAD_REQUEST)
		return
	}

	contractorID := mux.Vars(r)["contractor_id"]
	if !common.ValidateInt(contractorID) {
		logger.Error("Bad request: invalid contractor_id parameter: " + url)
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
	command := handler.NewCommand("INIT:contractors/trust-line", contractorID, equivalent)

	err = router.nodeHandler.Node.SendCommand(command)
	if err != nil {
		logger.Error("Can't send command: " + string(command.ToBytes()) + " to node. Details: " + err.Error())
		writeHTTPResponse(w, COMMAND_TRANSFERRING_ERROR, common.ActionResponse{})
		return
	}

	result, err := router.nodeHandler.Node.GetResult(command, common.SETTLEMENT_LINE_RESULT_TIMEOUT)
	if err != nil {
		logger.Error("Node is inaccessible during processing command: " +
			string(command.ToBytes()) + ". Details: " + err.Error())
		writeHTTPResponse(w, NODE_IS_INACCESSIBLE, common.ActionResponse{})
		return
	}

	if result.Code != OK {
		logger.Error("Node return wrong command result: " + strconv.Itoa(result.Code) +
			" on command: " + string(command.ToBytes()))
	}

	writeHTTPResponse(w, result.Code, common.ActionResponse{})
}

func (router *RoutesHandler) SetMaxPositiveBalance(w http.ResponseWriter, r *http.Request) {
	url, err := preprocessRequest(r)
	if err != nil {
		logger.Error("Bad request: invalid security parameters: " + err.Error())
		w.WriteHeader(BAD_REQUEST)
		return
	}

	contractorID := mux.Vars(r)["contractor_id"]
	if !common.ValidateInt(contractorID) {
		logger.Error("Bad request: invalid contractor_id parameter: " + url)
		w.WriteHeader(BAD_REQUEST)
		return
	}

	equivalent := mux.Vars(r)["equivalent"]
	if !common.ValidateInt(equivalent) {
		logger.Error("Bad request: invalid equivalent parameter: " + url)
		w.WriteHeader(BAD_REQUEST)
		return
	}

	amount := r.FormValue("amount")
	if !common.ValidateSettlementLineAmount(amount) {
		logger.Error("Bad request: invalid amount parameter: " + url)
		w.WriteHeader(BAD_REQUEST)
		return
	}

	command := handler.NewCommand(
		"SET:contractors/trust-lines", contractorID, amount, equivalent)

	err = router.nodeHandler.Node.SendCommand(command)
	if err != nil {
		logger.Error("Can't send command: " + string(command.ToBytes()) + " to node. Details: " + err.Error())
		writeHTTPResponse(w, COMMAND_TRANSFERRING_ERROR, common.ActionResponse{})
		return
	}

	result, err := router.nodeHandler.Node.GetResult(command, common.SETTLEMENT_LINE_RESULT_TIMEOUT)
	if err != nil {
		logger.Error("Node is inaccessible during processing command: " +
			string(command.ToBytes()) + ". Details: " + err.Error())
		writeHTTPResponse(w, NODE_IS_INACCESSIBLE, common.ActionResponse{})
		return
	}

	if result.Code != OK {
		logger.Error("Node return wrong command result: " + strconv.Itoa(result.Code) +
			" on command: " + string(command.ToBytes()))
	}

	writeHTTPResponse(w, result.Code, common.ActionResponse{})
}

func (router *RoutesHandler) ZeroOutMaxNegativeBalance(w http.ResponseWriter, r *http.Request) {
	url, err := preprocessRequest(r)
	if err != nil {
		logger.Error("Bad request: invalid security parameters: " + err.Error())
		w.WriteHeader(BAD_REQUEST)
		return
	}

	contractorID := mux.Vars(r)["contractor_id"]
	if !common.ValidateInt(contractorID) {
		logger.Error("Bad request: invalid contractor_id parameter: " + url)
		w.WriteHeader(BAD_REQUEST)
		return
	}

	equivalent := mux.Vars(r)["equivalent"]
	if !common.ValidateInt(equivalent) {
		logger.Error("Bad request: invalid equivalent parameter: " + url)
		w.WriteHeader(BAD_REQUEST)
		return
	}

	command := handler.NewCommand(
		"DELETE:contractors/incoming-trust-line", contractorID, equivalent)

	err = router.nodeHandler.Node.SendCommand(command)
	if err != nil {
		logger.Error("Can't send command: " + string(command.ToBytes()) + " to node. Details: " + err.Error())
		writeHTTPResponse(w, COMMAND_TRANSFERRING_ERROR, common.ActionResponse{})
		return
	}

	result, err := router.nodeHandler.Node.GetResult(command, common.SETTLEMENT_LINE_RESULT_TIMEOUT)
	if err != nil {
		logger.Error("Node is inaccessible during processing command: " +
			string(command.ToBytes()) + ". Details: " + err.Error())
		writeHTTPResponse(w, NODE_IS_INACCESSIBLE, common.ActionResponse{})
		return
	}

	if result.Code != OK {
		logger.Error("Node return wrong command result: " + strconv.Itoa(result.Code) +
			" on command: " + string(command.ToBytes()))
	}

	writeHTTPResponse(w, result.Code, common.ActionResponse{})
}

func (router *RoutesHandler) PublicKeysSharing(w http.ResponseWriter, r *http.Request) {
	url, err := preprocessRequest(r)
	if err != nil {
		logger.Error("Bad request: invalid security parameters: " + err.Error())
		w.WriteHeader(BAD_REQUEST)
		return
	}

	contractorID := mux.Vars(r)["contractor_id"]
	if !common.ValidateInt(contractorID) {
		logger.Error("Bad request: invalid contractor_id parameter: " + url)
		w.WriteHeader(BAD_REQUEST)
		return
	}

	equivalent := mux.Vars(r)["equivalent"]
	if !common.ValidateInt(equivalent) {
		logger.Error("Bad request: invalid equivalent parameter: " + url)
		w.WriteHeader(BAD_REQUEST)
		return
	}

	command := handler.NewCommand(
		"SET:contractors/trust-line-keys", contractorID, equivalent)

	err = router.nodeHandler.Node.SendCommand(command)
	if err != nil {
		logger.Error("Can't send command: " + string(command.ToBytes()) + " to node. Details: " + err.Error())
		writeHTTPResponse(w, COMMAND_TRANSFERRING_ERROR, common.ActionResponse{})
		return
	}

	result, err := router.nodeHandler.Node.GetResult(command, common.SETTLEMENT_LINE_RESULT_TIMEOUT)
	if err != nil {
		logger.Error("Node is inaccessible during processing command: " +
			string(command.ToBytes()) + ". Details: " + err.Error())
		writeHTTPResponse(w, NODE_IS_INACCESSIBLE, common.ActionResponse{})
		return
	}

	if result.Code != OK {
		logger.Error("Node return wrong command result: " + strconv.Itoa(result.Code) +
			" on command: " + string(command.ToBytes()))
	}

	writeHTTPResponse(w, result.Code, common.ActionResponse{})
}

func (router *RoutesHandler) RemoveSettlementLine(w http.ResponseWriter, r *http.Request) {
	url, err := preprocessRequest(r)
	if err != nil {
		logger.Error("Bad request: invalid security parameters: " + err.Error())
		w.WriteHeader(BAD_REQUEST)
		return
	}

	contractorID := mux.Vars(r)["contractor_id"]
	if !common.ValidateInt(contractorID) {
		logger.Error("Bad request: invalid contractor_id parameter: " + url)
		w.WriteHeader(BAD_REQUEST)
		return
	}

	equivalent := mux.Vars(r)["equivalent"]
	if !common.ValidateInt(equivalent) {
		logger.Error("Bad request: invalid equivalent parameter: " + url)
		w.WriteHeader(BAD_REQUEST)
		return
	}

	command := handler.NewCommand(
		"DELETE:contractors/trust-line", contractorID, equivalent)

	err = router.nodeHandler.Node.SendCommand(command)
	if err != nil {
		logger.Error("Can't send command: " + string(command.ToBytes()) + " to node. Details: " + err.Error())
		writeHTTPResponse(w, COMMAND_TRANSFERRING_ERROR, common.ActionResponse{})
		return
	}

	result, err := router.nodeHandler.Node.GetResult(command, common.SETTLEMENT_LINE_RESULT_TIMEOUT)
	if err != nil {
		logger.Error("Node is inaccessible during processing command: " +
			string(command.ToBytes()) + ". Details: " + err.Error())
		writeHTTPResponse(w, NODE_IS_INACCESSIBLE, common.ActionResponse{})
		return
	}

	if result.Code != OK {
		logger.Error("Node return wrong command result: " + strconv.Itoa(result.Code) +
			" on command: " + string(command.ToBytes()))
	}

	writeHTTPResponse(w, result.Code, common.ActionResponse{})
}

func (router *RoutesHandler) ResetSettlementLine(w http.ResponseWriter, r *http.Request) {
	url, err := preprocessRequest(r)
	if err != nil {
		logger.Error("Bad request: invalid security parameters: " + err.Error())
		w.WriteHeader(BAD_REQUEST)
		return
	}

	contractorID := mux.Vars(r)["contractor_id"]
	if !common.ValidateInt(contractorID) {
		logger.Error("Bad request: invalid contractor_id parameter: " + url)
		w.WriteHeader(BAD_REQUEST)
		return
	}

	equivalent := mux.Vars(r)["equivalent"]
	if !common.ValidateInt(equivalent) {
		logger.Error("Bad request: invalid equivalent parameter: " + url)
		w.WriteHeader(BAD_REQUEST)
		return
	}

	auditNumber := r.FormValue("audit_number")
	if !common.ValidateInt(auditNumber) {
		logger.Error("Bad request: invalid audit_number parameter: " + url)
		w.WriteHeader(BAD_REQUEST)
		return
	}

	maxNegativeBalance := r.FormValue("max_negative_balance")
	if !common.ValidateSettlementLineAmount(maxNegativeBalance) {
		logger.Error("Bad request: invalid max_negative_balance parameter: " + url)
		w.WriteHeader(BAD_REQUEST)
		return
	}

	maxPositiveBalance := r.FormValue("max_positive_balance")
	if !common.ValidateSettlementLineAmount(maxPositiveBalance) {
		logger.Error("Bad request: invalid max_positive_balance parameter: " + url)
		w.WriteHeader(BAD_REQUEST)
		return
	}

	balance := r.FormValue("balance")
	if balance == "" {
		logger.Error("Bad request: invalid balance parameter: " + url)
		w.WriteHeader(BAD_REQUEST)
		return
	}

	command := handler.NewCommand(
		"SET:contractors/trust-lines/reset", contractorID, auditNumber,
		maxNegativeBalance, maxPositiveBalance, balance, equivalent)

	err = router.nodeHandler.Node.SendCommand(command)
	if err != nil {
		logger.Error("Can't send command: " + string(command.ToBytes()) + " to node. Details: " + err.Error())
		writeHTTPResponse(w, COMMAND_TRANSFERRING_ERROR, common.ActionResponse{})
		return
	}

	result, err := router.nodeHandler.Node.GetResult(command, common.SETTLEMENT_LINE_RESULT_TIMEOUT)
	if err != nil {
		logger.Error("Node is inaccessible during processing command: " +
			string(command.ToBytes()) + ". Details: " + err.Error())
		writeHTTPResponse(w, NODE_IS_INACCESSIBLE, common.ActionResponse{})
		return
	}

	if result.Code != OK {
		logger.Error("Node return wrong command result: " + strconv.Itoa(result.Code) +
			" on command: " + string(command.ToBytes()))
	}

	writeHTTPResponse(w, result.Code, common.ActionResponse{})
}

func (router *RoutesHandler) ListSettlementLines(w http.ResponseWriter, r *http.Request) {
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

	command := handler.NewCommand("GET:contractors/trust-lines", common.DEFAULT_SETTLEMENT_LINES_OFFSET, common.DFEAULT_SETTLEMENT_LINES_COUNT, equivalent)

	err = router.nodeHandler.Node.SendCommand(command)
	if err != nil {
		logger.Error("Can't send command: " + string(command.ToBytes()) + " to node. Details: " + err.Error())
		writeHTTPResponse(w, COMMAND_TRANSFERRING_ERROR, common.SettlementLineListResponse{})
		return
	}

	result, err := router.nodeHandler.Node.GetResult(command, common.SETTLEMENT_LINE_RESULT_TIMEOUT)
	if err != nil {
		logger.Error("Node is inaccessible during processing command: " +
			string(command.ToBytes()) + ". Details: " + err.Error())
		writeHTTPResponse(w, NODE_IS_INACCESSIBLE, common.SettlementLineListResponse{})
		return
	}

	if result.Code != OK && result.Code != ENGINE_NO_EQUIVALENT {
		logger.Error("Node return wrong command result: " + strconv.Itoa(result.Code) +
			" on command: " + string(command.ToBytes()))
		writeHTTPResponse(w, result.Code, common.SettlementLineListResponse{})
		return
	}
	if result.Code == ENGINE_NO_EQUIVALENT {
		logger.Info("Node hasn't equivalent for command: " + string(command.ToBytes()))
		writeHTTPResponse(w, result.Code, common.SettlementLineListResponse{})
		return
	}

	if len(result.Tokens) == 0 {
		logger.Error("Node return invalid result tokens size on command: " + string(command.ToBytes()))
		writeHTTPResponse(w, ENGINE_UNEXPECTED_ERROR, common.SettlementLineListResponse{})
		return
	}

	contractorsCount, err := strconv.Atoi(result.Tokens[0])
	if err != nil {
		logger.Error("Node return invalid token on command: " + string(command.ToBytes()) +
			". Details: " + err.Error())
		writeHTTPResponse(w, ENGINE_UNEXPECTED_ERROR, common.SettlementLineListResponse{})
		return
	}

	if contractorsCount == 0 {
		writeHTTPResponse(w, OK, common.SettlementLineListResponse{Count: contractorsCount})
		return
	}

	response := common.SettlementLineListResponse{Count: contractorsCount}
	response.SettlementLines = make([]common.SettlementLineListItem, 0, contractorsCount)

	for i := range contractorsCount {
		response.SettlementLines = append(response.SettlementLines, common.SettlementLineListItem{
			ID:                    result.Tokens[i*8+1],
			Contractor:            result.Tokens[i*8+2],
			State:                 result.Tokens[i*8+3],
			OwnKeysPresent:        result.Tokens[i*8+4],
			ContractorKeysPresent: result.Tokens[i*8+5],
			MaxNegativeBalance:    result.Tokens[i*8+6],
			MaxPositiveBalance:    result.Tokens[i*8+7],
			Balance:               result.Tokens[i*8+8],
		})
	}
	writeHTTPResponse(w, OK, response)
}

func (router *RoutesHandler) ListSettlementLinesPortions(w http.ResponseWriter, r *http.Request) {
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

	equivalent, isParamPresent := mux.Vars(r)["equivalent"]
	if !isParamPresent {
		logger.Error("Bad request: missing equivalent parameter: " + url)
		w.WriteHeader(BAD_REQUEST)
		return
	}

	command := handler.NewCommand("GET:contractors/trust-lines", offset, count, equivalent)

	err = router.nodeHandler.Node.SendCommand(command)
	if err != nil {
		logger.Error("Can't send command: " + string(command.ToBytes()) + " to node. Details: " + err.Error())
		writeHTTPResponse(w, COMMAND_TRANSFERRING_ERROR, common.SettlementLineListResponse{})
		return
	}

	result, err := router.nodeHandler.Node.GetResult(command, common.SETTLEMENT_LINE_RESULT_TIMEOUT)
	if err != nil {
		logger.Error("Node is inaccessible during processing command: " +
			string(command.ToBytes()) + ". Details: " + err.Error())
		writeHTTPResponse(w, NODE_IS_INACCESSIBLE, common.SettlementLineListResponse{})
		return
	}

	if result.Code != OK && result.Code != ENGINE_NO_EQUIVALENT {
		logger.Error("Node return wrong command result: " + strconv.Itoa(result.Code) +
			" on command: " + string(command.ToBytes()))
		writeHTTPResponse(w, result.Code, common.SettlementLineListResponse{})
		return
	}
	if result.Code == ENGINE_NO_EQUIVALENT {
		logger.Info("Node hasn't equivalent for command: " + string(command.ToBytes()))
		writeHTTPResponse(w, result.Code, common.SettlementLineListResponse{})
		return
	}

	if len(result.Tokens) == 0 {
		logger.Error("Node return invalid result tokens size on command: " + string(command.ToBytes()))
		writeHTTPResponse(w, ENGINE_UNEXPECTED_ERROR, common.SettlementLineListResponse{})
		return
	}

	contractorsCount, err := strconv.Atoi(result.Tokens[0])
	if err != nil {
		logger.Error("Node return invalid token on command: " +
			string(command.ToBytes()) + ". Details: " + err.Error())
		writeHTTPResponse(w, ENGINE_UNEXPECTED_ERROR, common.SettlementLineListResponse{})
		return
	}

	if contractorsCount == 0 {
		writeHTTPResponse(w, OK, common.SettlementLineListResponse{Count: contractorsCount})
		return
	}

	response := common.SettlementLineListResponse{Count: contractorsCount}
	response.SettlementLines = make([]common.SettlementLineListItem, 0, contractorsCount)

	for i := range contractorsCount {
		response.SettlementLines = append(response.SettlementLines, common.SettlementLineListItem{
			ID:                    result.Tokens[i*8+1],
			Contractor:            result.Tokens[i*8+2],
			State:                 result.Tokens[i*8+3],
			OwnKeysPresent:        result.Tokens[i*8+4],
			ContractorKeysPresent: result.Tokens[i*8+5],
			MaxNegativeBalance:    result.Tokens[i*8+6],
			MaxPositiveBalance:    result.Tokens[i*8+7],
			Balance:               result.Tokens[i*8+8],
		})
	}
	writeHTTPResponse(w, OK, response)
}

func (router *RoutesHandler) ListSettlementLinesAllEquivalents(w http.ResponseWriter, r *http.Request) {
	_, err := preprocessRequest(r)
	if err != nil {
		logger.Error("Bad request: invalid security parameters: " + err.Error())
		w.WriteHeader(BAD_REQUEST)
		return
	}

	command := handler.NewCommand("GET:contractors/trust-lines-all", common.DEFAULT_SETTLEMENT_LINES_OFFSET, common.DFEAULT_SETTLEMENT_LINES_COUNT)

	err = router.nodeHandler.Node.SendCommand(command)
	if err != nil {
		logger.Error("Can't send command: " + string(command.ToBytes()) + " to node. Details: " + err.Error())
		writeHTTPResponse(w, COMMAND_TRANSFERRING_ERROR, common.AllEquivalentsResponse{})
		return
	}

	result, err := router.nodeHandler.Node.GetResult(command, common.SETTLEMENT_LINE_RESULT_TIMEOUT)
	if err != nil {
		logger.Error("Node is inaccessible during processing command: " +
			string(command.ToBytes()) + ". Details: " + err.Error())
		writeHTTPResponse(w, NODE_IS_INACCESSIBLE, common.AllEquivalentsResponse{})
		return
	}

	if result.Code != OK {
		logger.Error("Node return wrong command result: " + strconv.Itoa(result.Code) +
			" on command: " + string(command.ToBytes()))
		writeHTTPResponse(w, result.Code, common.AllEquivalentsResponse{})
		return
	}

	if len(result.Tokens) == 0 {
		logger.Error("Node return invalid result tokens size on command: " + string(command.ToBytes()))
		writeHTTPResponse(w, ENGINE_UNEXPECTED_ERROR, common.AllEquivalentsResponse{})
		return
	}

	equivalentsCount, err := strconv.Atoi(result.Tokens[0])
	if err != nil {
		logger.Error("Node return invalid token on command: " +
			string(command.ToBytes()) + ". Details: " + err.Error())
		writeHTTPResponse(w, ENGINE_UNEXPECTED_ERROR, common.AllEquivalentsResponse{})
		return
	}

	if equivalentsCount == 0 {
		writeHTTPResponse(w, OK, common.AllEquivalentsResponse{Count: equivalentsCount})
		return
	}

	tokenIdx := 1

	response := common.AllEquivalentsResponse{Count: equivalentsCount}
	for range equivalentsCount {
		contractorsCount, err := strconv.Atoi(result.Tokens[tokenIdx+1])
		if err != nil {
			logger.Error("Node return invalid token on command: " +
				string(command.ToBytes()) + ". Details: " + err.Error())
			writeHTTPResponse(w, ENGINE_UNEXPECTED_ERROR, common.AllEquivalentsResponse{})
			return
		}
		responseSettlementLines := common.EquivalentStatistics{Eq: result.Tokens[tokenIdx], Count: contractorsCount}
		tokenIdx = tokenIdx + 2
		for i := range contractorsCount {
			responseSettlementLines.SettlementLines = append(responseSettlementLines.SettlementLines, common.SettlementLineListItem{
				ID:                    result.Tokens[tokenIdx+i*8],
				Contractor:            result.Tokens[tokenIdx+i*8+1],
				State:                 result.Tokens[tokenIdx+i*8+2],
				OwnKeysPresent:        result.Tokens[tokenIdx+i*8+3],
				ContractorKeysPresent: result.Tokens[tokenIdx+i*8+4],
				MaxNegativeBalance:    result.Tokens[tokenIdx+i*8+5],
				MaxPositiveBalance:    result.Tokens[tokenIdx+i*8+6],
				Balance:               result.Tokens[tokenIdx+i*8+7],
			})
		}
		tokenIdx = tokenIdx + responseSettlementLines.Count*8
		response.Equivalents = append(response.Equivalents, responseSettlementLines)
	}
	writeHTTPResponse(w, OK, response)
}

func (router *RoutesHandler) ListContractors(w http.ResponseWriter, r *http.Request) {
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

	command := handler.NewCommand("GET:contractors", equivalent)

	err = router.nodeHandler.Node.SendCommand(command)
	if err != nil {
		logger.Error("Can't send command: " + string(command.ToBytes()) + " to node. Details: " + err.Error())
		writeHTTPResponse(w, COMMAND_TRANSFERRING_ERROR, common.ContractorsListResponse{})
		return
	}

	result, err := router.nodeHandler.Node.GetResult(command, common.CONTRACTORS_RESULT_TIMEOUT)
	if err != nil {
		logger.Error("Node is inaccessible during processing command: " +
			string(command.ToBytes()) + ". Details: " + err.Error())
		writeHTTPResponse(w, NODE_IS_INACCESSIBLE, common.ContractorsListResponse{})
		return
	}

	if result.Code != OK && result.Code != ENGINE_NO_EQUIVALENT {
		logger.Error("Node return wrong command result: " + strconv.Itoa(result.Code) +
			" on command: " + string(command.ToBytes()))
		writeHTTPResponse(w, result.Code, common.ContractorsListResponse{})
		return
	}
	if result.Code == ENGINE_NO_EQUIVALENT {
		logger.Info("Node hasn't equivalent for command: " + string(command.ToBytes()))
		writeHTTPResponse(w, result.Code, common.ContractorsListResponse{})
		return
	}

	if len(result.Tokens) == 0 {
		logger.Error("Node return invalid result tokens size on command: " + string(command.ToBytes()))
		writeHTTPResponse(w, ENGINE_UNEXPECTED_ERROR, common.ContractorsListResponse{})
		return
	}

	contractorsCount, err := strconv.Atoi(result.Tokens[0])
	if err != nil {
		logger.Error("Node return invalid token on command: " + string(command.ToBytes()) +
			". Details: " + err.Error())
		writeHTTPResponse(w, ENGINE_UNEXPECTED_ERROR, common.ContractorsListResponse{})
		return
	}

	if contractorsCount == 0 {
		writeHTTPResponse(w, OK, common.ContractorsListResponse{Count: contractorsCount})
		return
	}

	response := common.ContractorsListResponse{Count: contractorsCount}
	for i := range contractorsCount {
		response.Contractors = append(response.Contractors, common.ContractorInfo{
			ContractorID:        result.Tokens[i*2+1],
			ContractorAddresses: result.Tokens[i*2+2],
		})
	}

	writeHTTPResponse(w, OK, response)
}

func (router *RoutesHandler) GetSettlementLineByID(w http.ResponseWriter, r *http.Request) {
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

	contractorID := r.FormValue("contractor_id")
	if !common.ValidateInt(contractorID) {
		logger.Error("Bad request: invalid contractorID parameter: " + url)
		w.WriteHeader(BAD_REQUEST)
		return
	}

	command := handler.NewCommand(
		"GET:contractors/trust-lines/one/id", contractorID, equivalent)

	err = router.nodeHandler.Node.SendCommand(command)
	if err != nil {
		logger.Error("Can't send command: " + string(command.ToBytes()) + " to node. Details: " + err.Error())
		writeHTTPResponse(w, COMMAND_TRANSFERRING_ERROR, common.SettlementLineDetailResponse{})
		return
	}

	result, err := router.nodeHandler.Node.GetResult(command, common.SETTLEMENT_LINE_RESULT_TIMEOUT)
	if err != nil {
		logger.Error("Node is inaccessible during processing command: " +
			string(command.ToBytes()) + ". Details: " + err.Error())
		writeHTTPResponse(w, NODE_IS_INACCESSIBLE, common.SettlementLineDetailResponse{})
		return
	}

	if result.Code != OK && result.Code != ENGINE_NO_EQUIVALENT && result.Code != NODE_NOT_FOUND {
		logger.Error("Node return wrong command result: " + strconv.Itoa(result.Code) +
			" on command: " + string(command.ToBytes()))
		writeHTTPResponse(w, result.Code, common.SettlementLineDetailResponse{})
		return
	}
	if result.Code == ENGINE_NO_EQUIVALENT {
		logger.Info("Node hasn't equivalent for command: " + string(command.ToBytes()))
		writeHTTPResponse(w, result.Code, common.SettlementLineDetailResponse{})
		return
	}
	if result.Code == NODE_NOT_FOUND {
		logger.Info("Node hasn't TL with requested contractor for command " + string(command.ToBytes()))
		writeHTTPResponse(w, result.Code, common.SettlementLineDetailResponse{})
		return
	}

	if len(result.Tokens) < 4 {
		logger.Error("Node return invalid result tokens size on command: " + string(command.ToBytes()))
		writeHTTPResponse(w, ENGINE_UNEXPECTED_ERROR, common.SettlementLineDetailResponse{})
		return
	}

	response := common.SettlementLineDetailResponse{SettlementLine: common.SettlementLineDetail{
		ID:                    result.Tokens[0],
		State:                 result.Tokens[1],
		OwnKeysPresent:        result.Tokens[2],
		ContractorKeysPresent: result.Tokens[3],
		AuditNumber:           result.Tokens[4],
		MaxNegativeBalance:    result.Tokens[5],
		MaxPositiveBalance:    result.Tokens[6],
		Balance:               result.Tokens[7],
	}}
	writeHTTPResponse(w, OK, response)
}

func (router *RoutesHandler) GetSettlementLineByAddress(w http.ResponseWriter, r *http.Request) {
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

	contractorAddresses = append([]string{strconv.Itoa(len(contractorAddresses) / 2)}, contractorAddresses...)
	contractorAddresses = append([]string{"GET:contractors/trust-lines/one/address"}, contractorAddresses...)
	contractorAddresses = append(contractorAddresses, []string{equivalent}...)
	command := handler.NewCommand(contractorAddresses...)

	err = router.nodeHandler.Node.SendCommand(command)
	if err != nil {
		logger.Error("Can't send command: " + string(command.ToBytes()) + " to node. Details: " + err.Error())
		writeHTTPResponse(w, COMMAND_TRANSFERRING_ERROR, common.SettlementLineDetailResponse{})
		return
	}

	result, err := router.nodeHandler.Node.GetResult(command, common.SETTLEMENT_LINE_RESULT_TIMEOUT)
	if err != nil {
		logger.Error("Node is inaccessible during processing command: " +
			string(command.ToBytes()) + ". Details: " + err.Error())
		writeHTTPResponse(w, NODE_IS_INACCESSIBLE, common.SettlementLineDetailResponse{})
		return
	}

	if result.Code != OK && result.Code != ENGINE_NO_EQUIVALENT && result.Code != NODE_NOT_FOUND {
		logger.Error("Node return wrong command result: " + strconv.Itoa(result.Code) +
			" on command: " + string(command.ToBytes()))
		writeHTTPResponse(w, result.Code, common.SettlementLineDetailResponse{})
		return
	}
	if result.Code == ENGINE_NO_EQUIVALENT {
		logger.Info("Node hasn't equivalent for command: " + string(command.ToBytes()))
		writeHTTPResponse(w, result.Code, common.SettlementLineDetailResponse{})
		return
	}
	if result.Code == NODE_NOT_FOUND {
		logger.Info("Node hasn't TL with requested contractor for command " + string(command.ToBytes()))
		writeHTTPResponse(w, result.Code, common.SettlementLineDetailResponse{})
		return
	}

	if len(result.Tokens) < 8 {
		logger.Error("Node return invalid result tokens size on command: " + string(command.ToBytes()))
		writeHTTPResponse(w, ENGINE_UNEXPECTED_ERROR, common.SettlementLineDetailResponse{})
		return
	}

	response := common.SettlementLineDetailResponse{SettlementLine: common.SettlementLineDetail{
		ID:                    result.Tokens[0],
		State:                 result.Tokens[1],
		OwnKeysPresent:        result.Tokens[2],
		ContractorKeysPresent: result.Tokens[3],
		AuditNumber:           result.Tokens[4],
		MaxNegativeBalance:    result.Tokens[5],
		MaxPositiveBalance:    result.Tokens[6],
		Balance:               result.Tokens[7],
	}}
	writeHTTPResponse(w, OK, response)
}

func (router *RoutesHandler) ListEquivalents(w http.ResponseWriter, r *http.Request) {
	_, err := preprocessRequest(r)
	if err != nil {
		logger.Error("Bad request: invalid security parameters: " + err.Error())
		w.WriteHeader(BAD_REQUEST)
		return
	}

	command := handler.NewCommand("GET:equivalents")

	err = router.nodeHandler.Node.SendCommand(command)
	if err != nil {
		logger.Error("Can't send command: " + string(command.ToBytes()) + " to node. Details: " + err.Error())
		writeHTTPResponse(w, COMMAND_TRANSFERRING_ERROR, common.EquivalentsListResponse{})
		return
	}

	result, err := router.nodeHandler.Node.GetResult(command, common.SETTLEMENT_LINE_RESULT_TIMEOUT)
	if err != nil {
		logger.Error("Node is inaccessible during processing command: " +
			string(command.ToBytes()) + ". Details: " + err.Error())
		writeHTTPResponse(w, NODE_IS_INACCESSIBLE, common.EquivalentsListResponse{})
		return
	}

	if result.Code != OK {
		logger.Error("Node return wrong command result: " + strconv.Itoa(result.Code) +
			" on command: " + string(command.ToBytes()))
		writeHTTPResponse(w, result.Code, common.EquivalentsListResponse{})
		return
	}

	if len(result.Tokens) == 0 {
		logger.Error("Node return invalid result tokens size on command: " + string(command.ToBytes()))
		writeHTTPResponse(w, ENGINE_UNEXPECTED_ERROR, common.EquivalentsListResponse{})
		return
	}

	equivalentsCount, err := strconv.Atoi(result.Tokens[0])
	if err != nil {
		logger.Error("Node return invalid token on command: " +
			string(command.ToBytes()) + ". Details: " + err.Error())
		writeHTTPResponse(w, ENGINE_UNEXPECTED_ERROR, common.EquivalentsListResponse{})
		return
	}

	if equivalentsCount == 0 {
		writeHTTPResponse(w, OK, common.EquivalentsListResponse{Count: equivalentsCount})
		return
	}

	response := common.EquivalentsListResponse{Count: equivalentsCount}
	for i := range equivalentsCount {
		response.Equivalents = append(response.Equivalents, result.Tokens[i+1])
	}
	writeHTTPResponse(w, OK, response)
}

func (router *RoutesHandler) TotalBalance(w http.ResponseWriter, r *http.Request) {
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

	command := handler.NewCommand("GET:stats/balance/total", equivalent)

	err = router.nodeHandler.Node.SendCommand(command)
	if err != nil {
		logger.Error("Can't send command: " + string(command.ToBytes()) + " to node. Details: " + err.Error())
		writeHTTPResponse(w, COMMAND_TRANSFERRING_ERROR, common.TotalBalanceResponse{})
		return
	}

	result, err := router.nodeHandler.Node.GetResult(command, common.STATS_RESULT_TIMEOUT)
	if err != nil {
		logger.Error("Node is inaccessible during processing command: " + string(command.ToBytes()) +
			". Details: " + err.Error())
		writeHTTPResponse(w, NODE_IS_INACCESSIBLE, common.TotalBalanceResponse{})
		return
	}

	if result.Code != OK && result.Code != ENGINE_NO_EQUIVALENT {
		logger.Error("Node return wrong command result: " + strconv.Itoa(result.Code) +
			" on command: " + string(command.ToBytes()))
		writeHTTPResponse(w, result.Code, common.TotalBalanceResponse{})
		return
	}
	if result.Code == ENGINE_NO_EQUIVALENT {
		logger.Info("Node hasn't equivalent for command: " + string(command.ToBytes()))
		writeHTTPResponse(w, result.Code, common.TotalBalanceResponse{})
		return
	}

	if len(result.Tokens) < 4 {
		logger.Error("Node return invalid result tokens size on command: " + string(command.ToBytes()))
		writeHTTPResponse(w, ENGINE_UNEXPECTED_ERROR, common.TotalBalanceResponse{})
		return
	}

	response := common.TotalBalanceResponse{
		TotalMaxNegativeBalance: result.Tokens[0],
		TotalNegativeBalance:    result.Tokens[1],
		TotalMaxPositiveBalance: result.Tokens[2],
		TotalPositiveBalance:    result.Tokens[3],
	}
	writeHTTPResponse(w, OK, response)
}
