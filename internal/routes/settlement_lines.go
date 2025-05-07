package routes

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
	"github.com/vTCP-Foundation/vtcpd-cli/internal/handler"
	"github.com/vTCP-Foundation/vtcpd-cli/internal/logger"
)

var (
	SETTLEMENT_LINE_RESULT_TIMEOUT  uint16 = 20 // seconds
	CONTRACTORS_RESULT_TIMEOUT      uint16 = 20
	STATS_RESULT_TIMEOUT            uint16 = 20 // seconds
	DEFAULT_SETTLEMENT_LINES_OFFSET        = "0"
	DFEAULT_SETTLEMENT_LINES_COUNT         = "10000"
)

// --- Global structs ---

type SettlementLineListItem struct {
	ID                    string `json:"contractor_id"`
	Contractor            string `json:"contractor"`
	State                 string `json:"state"`
	OwnKeysPresent        string `json:"own_keys_present"`
	ContractorKeysPresent string `json:"contractor_keys_present"`
	MaxNegativeBalance    string `json:"max_negative_balance"`
	MaxPositiveBalance    string `json:"max_positive_balance"`
	Balance               string `json:"balance"`
}

type SettlementLineDetail struct {
	ID                    string `json:"id"` // todo : contractor_id
	State                 string `json:"state"`
	OwnKeysPresent        string `json:"own_keys_present"`
	ContractorKeysPresent string `json:"contractor_keys_present"`
	AuditNumber           string `json:"audit_number"`
	MaxNegativeBalance    string `json:"max_negative_balance"`
	MaxPositiveBalance    string `json:"max_positive_balance"`
	Balance               string `json:"balance"`
}

type EquivalentStatistics struct {
	Eq              string                   `json:"equivalent"`
	Count           int                      `json:"count"`
	SettlementLines []SettlementLineListItem `json:"settlement_lines"`
}

type ContractorInfo struct {
	ContractorID        string `json:"contractor_id"`
	ContractorAddresses string `json:"contractor_addresses"`
}

// --- Global API responses ---

// ActionResponse used for operations that return only status
type ActionResponse struct{}

type SettlementLineListResponse struct {
	Count           int                      `json:"count"`
	SettlementLines []SettlementLineListItem `json:"settlement_lines"`
}

type SettlementLineDetailResponse struct {
	SettlementLine SettlementLineDetail `json:"settlement_line"`
}

type AllEquivalentsResponse struct {
	Count       int                    `json:"count"`
	Equivalents []EquivalentStatistics `json:"equivalents"`
}

type ContractorsListResponse struct {
	Count       int              `json:"count"`
	Contractors []ContractorInfo `json:"contractors"`
}

type EquivalentsListResponse struct {
	Count       int      `json:"count"`
	Equivalents []string `json:"equivalents"`
}

type TotalBalanceResponse struct {
	TotalMaxNegativeBalance string `json:"total_max_negative_balance"`
	TotalNegativeBalance    string `json:"total_negative_balance"`
	TotalMaxPositiveBalance string `json:"total_max_positive_balance"`
	TotalPositiveBalance    string `json:"total_positive_balance"`
}

func (router *RoutesHandler) InitSettlementLine(w http.ResponseWriter, r *http.Request) {
	logger.Info("InitSettlementLine controller")
	url, err := preprocessRequest(r)
	if err != nil {
		logger.Error("Bad request: invalid security parameters: " + err.Error())
		w.WriteHeader(BAD_REQUEST)
		return
	}

	contractorID := mux.Vars(r)["contractor_id"]
	if !handler.ValidateInt(contractorID) {
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
		writeHTTPResponse(w, COMMAND_TRANSFERRING_ERROR, ActionResponse{})
		return
	}

	result, err := router.nodeHandler.Node.GetResult(command, SETTLEMENT_LINE_RESULT_TIMEOUT)
	if err != nil {
		logger.Error("Node is inaccessible during processing command: " +
			string(command.ToBytes()) + ". Details: " + err.Error())
		writeHTTPResponse(w, NODE_IS_INACCESSIBLE, ActionResponse{})
		return
	}

	if result.Code != OK {
		logger.Error("Node return wrong command result: " + strconv.Itoa(result.Code) +
			" on command: " + string(command.ToBytes()))
	}

	writeHTTPResponse(w, result.Code, ActionResponse{})
}

func (router *RoutesHandler) SetMaxPositiveBalance(w http.ResponseWriter, r *http.Request) {
	url, err := preprocessRequest(r)
	if err != nil {
		logger.Error("Bad request: invalid security parameters: " + err.Error())
		w.WriteHeader(BAD_REQUEST)
		return
	}

	contractorID := mux.Vars(r)["contractor_id"]
	if !handler.ValidateInt(contractorID) {
		logger.Error("Bad request: invalid contractor_id parameter: " + url)
		w.WriteHeader(BAD_REQUEST)
		return
	}

	equivalent := mux.Vars(r)["equivalent"]
	if !handler.ValidateInt(equivalent) {
		logger.Error("Bad request: invalid equivalent parameter: " + url)
		w.WriteHeader(BAD_REQUEST)
		return
	}

	amount := r.FormValue("amount")
	if !handler.ValidateSettlementLineAmount(amount) {
		logger.Error("Bad request: invalid amount parameter: " + url)
		w.WriteHeader(BAD_REQUEST)
		return
	}

	command := handler.NewCommand(
		"SET:contractors/trust-lines", contractorID, amount, equivalent)

	err = router.nodeHandler.Node.SendCommand(command)
	if err != nil {
		logger.Error("Can't send command: " + string(command.ToBytes()) + " to node. Details: " + err.Error())
		writeHTTPResponse(w, COMMAND_TRANSFERRING_ERROR, ActionResponse{})
		return
	}

	result, err := router.nodeHandler.Node.GetResult(command, SETTLEMENT_LINE_RESULT_TIMEOUT)
	if err != nil {
		logger.Error("Node is inaccessible during processing command: " +
			string(command.ToBytes()) + ". Details: " + err.Error())
		writeHTTPResponse(w, NODE_IS_INACCESSIBLE, ActionResponse{})
		return
	}

	if result.Code != OK {
		logger.Error("Node return wrong command result: " + strconv.Itoa(result.Code) +
			" on command: " + string(command.ToBytes()))
	}

	writeHTTPResponse(w, result.Code, ActionResponse{})
}

func (router *RoutesHandler) ZeroOutMaxNegativeBalance(w http.ResponseWriter, r *http.Request) {
	url, err := preprocessRequest(r)
	if err != nil {
		logger.Error("Bad request: invalid security parameters: " + err.Error())
		w.WriteHeader(BAD_REQUEST)
		return
	}

	contractorID := mux.Vars(r)["contractor_id"]
	if !handler.ValidateInt(contractorID) {
		logger.Error("Bad request: invalid contractor_id parameter: " + url)
		w.WriteHeader(BAD_REQUEST)
		return
	}

	equivalent := mux.Vars(r)["equivalent"]
	if !handler.ValidateInt(equivalent) {
		logger.Error("Bad request: invalid equivalent parameter: " + url)
		w.WriteHeader(BAD_REQUEST)
		return
	}

	command := handler.NewCommand(
		"DELETE:contractors/incoming-trust-line", contractorID, equivalent)

	err = router.nodeHandler.Node.SendCommand(command)
	if err != nil {
		logger.Error("Can't send command: " + string(command.ToBytes()) + " to node. Details: " + err.Error())
		writeHTTPResponse(w, COMMAND_TRANSFERRING_ERROR, ActionResponse{})
		return
	}

	result, err := router.nodeHandler.Node.GetResult(command, SETTLEMENT_LINE_RESULT_TIMEOUT)
	if err != nil {
		logger.Error("Node is inaccessible during processing command: " +
			string(command.ToBytes()) + ". Details: " + err.Error())
		writeHTTPResponse(w, NODE_IS_INACCESSIBLE, ActionResponse{})
		return
	}

	if result.Code != OK {
		logger.Error("Node return wrong command result: " + strconv.Itoa(result.Code) +
			" on command: " + string(command.ToBytes()))
	}

	writeHTTPResponse(w, result.Code, ActionResponse{})
}

func (router *RoutesHandler) PublicKeysSharing(w http.ResponseWriter, r *http.Request) {
	url, err := preprocessRequest(r)
	if err != nil {
		logger.Error("Bad request: invalid security parameters: " + err.Error())
		w.WriteHeader(BAD_REQUEST)
		return
	}

	contractorID := mux.Vars(r)["contractor_id"]
	if !handler.ValidateInt(contractorID) {
		logger.Error("Bad request: invalid contractor_id parameter: " + url)
		w.WriteHeader(BAD_REQUEST)
		return
	}

	equivalent := mux.Vars(r)["equivalent"]
	if !handler.ValidateInt(equivalent) {
		logger.Error("Bad request: invalid equivalent parameter: " + url)
		w.WriteHeader(BAD_REQUEST)
		return
	}

	command := handler.NewCommand(
		"SET:contractors/trust-line-keys", contractorID, equivalent)

	err = router.nodeHandler.Node.SendCommand(command)
	if err != nil {
		logger.Error("Can't send command: " + string(command.ToBytes()) + " to node. Details: " + err.Error())
		writeHTTPResponse(w, COMMAND_TRANSFERRING_ERROR, ActionResponse{})
		return
	}

	result, err := router.nodeHandler.Node.GetResult(command, SETTLEMENT_LINE_RESULT_TIMEOUT)
	if err != nil {
		logger.Error("Node is inaccessible during processing command: " +
			string(command.ToBytes()) + ". Details: " + err.Error())
		writeHTTPResponse(w, NODE_IS_INACCESSIBLE, ActionResponse{})
		return
	}

	if result.Code != OK {
		logger.Error("Node return wrong command result: " + strconv.Itoa(result.Code) +
			" on command: " + string(command.ToBytes()))
	}

	writeHTTPResponse(w, result.Code, ActionResponse{})
}

func (router *RoutesHandler) RemoveSettlementLine(w http.ResponseWriter, r *http.Request) {
	url, err := preprocessRequest(r)
	if err != nil {
		logger.Error("Bad request: invalid security parameters: " + err.Error())
		w.WriteHeader(BAD_REQUEST)
		return
	}

	contractorID := mux.Vars(r)["contractor_id"]
	if !handler.ValidateInt(contractorID) {
		logger.Error("Bad request: invalid contractor_id parameter: " + url)
		w.WriteHeader(BAD_REQUEST)
		return
	}

	equivalent := mux.Vars(r)["equivalent"]
	if !handler.ValidateInt(equivalent) {
		logger.Error("Bad request: invalid equivalent parameter: " + url)
		w.WriteHeader(BAD_REQUEST)
		return
	}

	command := handler.NewCommand(
		"DELETE:contractors/trust-line", contractorID, equivalent)

	err = router.nodeHandler.Node.SendCommand(command)
	if err != nil {
		logger.Error("Can't send command: " + string(command.ToBytes()) + " to node. Details: " + err.Error())
		writeHTTPResponse(w, COMMAND_TRANSFERRING_ERROR, ActionResponse{})
		return
	}

	result, err := router.nodeHandler.Node.GetResult(command, SETTLEMENT_LINE_RESULT_TIMEOUT)
	if err != nil {
		logger.Error("Node is inaccessible during processing command: " +
			string(command.ToBytes()) + ". Details: " + err.Error())
		writeHTTPResponse(w, NODE_IS_INACCESSIBLE, ActionResponse{})
		return
	}

	if result.Code != OK {
		logger.Error("Node return wrong command result: " + strconv.Itoa(result.Code) +
			" on command: " + string(command.ToBytes()))
	}

	writeHTTPResponse(w, result.Code, ActionResponse{})
}

func (router *RoutesHandler) ResetSettlementLine(w http.ResponseWriter, r *http.Request) {
	url, err := preprocessRequest(r)
	if err != nil {
		logger.Error("Bad request: invalid security parameters: " + err.Error())
		w.WriteHeader(BAD_REQUEST)
		return
	}

	contractorID := mux.Vars(r)["contractor_id"]
	if !handler.ValidateInt(contractorID) {
		logger.Error("Bad request: invalid contractor_id parameter: " + url)
		w.WriteHeader(BAD_REQUEST)
		return
	}

	equivalent := mux.Vars(r)["equivalent"]
	if !handler.ValidateInt(equivalent) {
		logger.Error("Bad request: invalid equivalent parameter: " + url)
		w.WriteHeader(BAD_REQUEST)
		return
	}

	auditNumber := r.FormValue("audit_number")
	if !handler.ValidateInt(auditNumber) {
		logger.Error("Bad request: invalid audit_number parameter: " + url)
		w.WriteHeader(BAD_REQUEST)
		return
	}

	maxNegativeBalance := r.FormValue("max_negative_balance")
	if !handler.ValidateSettlementLineAmount(maxNegativeBalance) {
		logger.Error("Bad request: invalid max_negative_balance parameter: " + url)
		w.WriteHeader(BAD_REQUEST)
		return
	}

	maxPositiveBalance := r.FormValue("max_positive_balance")
	if !handler.ValidateSettlementLineAmount(maxPositiveBalance) {
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
		writeHTTPResponse(w, COMMAND_TRANSFERRING_ERROR, ActionResponse{})
		return
	}

	result, err := router.nodeHandler.Node.GetResult(command, SETTLEMENT_LINE_RESULT_TIMEOUT)
	if err != nil {
		logger.Error("Node is inaccessible during processing command: " +
			string(command.ToBytes()) + ". Details: " + err.Error())
		writeHTTPResponse(w, NODE_IS_INACCESSIBLE, ActionResponse{})
		return
	}

	if result.Code != OK {
		logger.Error("Node return wrong command result: " + strconv.Itoa(result.Code) +
			" on command: " + string(command.ToBytes()))
	}

	writeHTTPResponse(w, result.Code, ActionResponse{})
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

	command := handler.NewCommand("GET:contractors/trust-lines", DEFAULT_SETTLEMENT_LINES_OFFSET, DFEAULT_SETTLEMENT_LINES_COUNT, equivalent)

	err = router.nodeHandler.Node.SendCommand(command)
	if err != nil {
		logger.Error("Can't send command: " + string(command.ToBytes()) + " to node. Details: " + err.Error())
		writeHTTPResponse(w, COMMAND_TRANSFERRING_ERROR, SettlementLineListResponse{})
		return
	}

	result, err := router.nodeHandler.Node.GetResult(command, SETTLEMENT_LINE_RESULT_TIMEOUT)
	if err != nil {
		logger.Error("Node is inaccessible during processing command: " +
			string(command.ToBytes()) + ". Details: " + err.Error())
		writeHTTPResponse(w, NODE_IS_INACCESSIBLE, SettlementLineListResponse{})
		return
	}

	if result.Code != OK && result.Code != ENGINE_NO_EQUIVALENT {
		logger.Error("Node return wrong command result: " + strconv.Itoa(result.Code) +
			" on command: " + string(command.ToBytes()))
		writeHTTPResponse(w, result.Code, SettlementLineListResponse{})
		return
	}
	if result.Code == ENGINE_NO_EQUIVALENT {
		logger.Info("Node hasn't equivalent for command: " + string(command.ToBytes()))
		writeHTTPResponse(w, result.Code, SettlementLineListResponse{})
		return
	}

	if len(result.Tokens) == 0 {
		logger.Error("Node return invalid result tokens size on command: " + string(command.ToBytes()))
		writeHTTPResponse(w, ENGINE_UNEXPECTED_ERROR, SettlementLineListResponse{})
		return
	}

	contractorsCount, err := strconv.Atoi(result.Tokens[0])
	if err != nil {
		logger.Error("Node return invalid token on command: " + string(command.ToBytes()) +
			". Details: " + err.Error())
		writeHTTPResponse(w, ENGINE_UNEXPECTED_ERROR, SettlementLineListResponse{})
		return
	}

	if contractorsCount == 0 {
		writeHTTPResponse(w, OK, SettlementLineListResponse{Count: contractorsCount})
		return
	}

	response := SettlementLineListResponse{Count: contractorsCount}
	response.SettlementLines = make([]SettlementLineListItem, 0, contractorsCount)

	for i := range contractorsCount {
		response.SettlementLines = append(response.SettlementLines, SettlementLineListItem{
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
	if !isParamPresent || !handler.ValidateInt(offset) {
		logger.Error("Bad request: invalid offset parameter: " + url)
		w.WriteHeader(BAD_REQUEST)
		return
	}

	count, isParamPresent := mux.Vars(r)["count"]
	if !isParamPresent || !handler.ValidateInt(count) {
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
		writeHTTPResponse(w, COMMAND_TRANSFERRING_ERROR, SettlementLineListResponse{})
		return
	}

	result, err := router.nodeHandler.Node.GetResult(command, SETTLEMENT_LINE_RESULT_TIMEOUT)
	if err != nil {
		logger.Error("Node is inaccessible during processing command: " +
			string(command.ToBytes()) + ". Details: " + err.Error())
		writeHTTPResponse(w, NODE_IS_INACCESSIBLE, SettlementLineListResponse{})
		return
	}

	if result.Code != OK && result.Code != ENGINE_NO_EQUIVALENT {
		logger.Error("Node return wrong command result: " + strconv.Itoa(result.Code) +
			" on command: " + string(command.ToBytes()))
		writeHTTPResponse(w, result.Code, SettlementLineListResponse{})
		return
	}
	if result.Code == ENGINE_NO_EQUIVALENT {
		logger.Info("Node hasn't equivalent for command: " + string(command.ToBytes()))
		writeHTTPResponse(w, result.Code, SettlementLineListResponse{})
		return
	}

	if len(result.Tokens) == 0 {
		logger.Error("Node return invalid result tokens size on command: " + string(command.ToBytes()))
		writeHTTPResponse(w, ENGINE_UNEXPECTED_ERROR, SettlementLineListResponse{})
		return
	}

	contractorsCount, err := strconv.Atoi(result.Tokens[0])
	if err != nil {
		logger.Error("Node return invalid token on command: " +
			string(command.ToBytes()) + ". Details: " + err.Error())
		writeHTTPResponse(w, ENGINE_UNEXPECTED_ERROR, SettlementLineListResponse{})
		return
	}

	if contractorsCount == 0 {
		writeHTTPResponse(w, OK, SettlementLineListResponse{Count: contractorsCount})
		return
	}

	response := SettlementLineListResponse{Count: contractorsCount}
	response.SettlementLines = make([]SettlementLineListItem, 0, contractorsCount)

	for i := range contractorsCount {
		response.SettlementLines = append(response.SettlementLines, SettlementLineListItem{
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

	command := handler.NewCommand("GET:contractors/trust-lines-all", DEFAULT_SETTLEMENT_LINES_OFFSET, DFEAULT_SETTLEMENT_LINES_COUNT)

	err = router.nodeHandler.Node.SendCommand(command)
	if err != nil {
		logger.Error("Can't send command: " + string(command.ToBytes()) + " to node. Details: " + err.Error())
		writeHTTPResponse(w, COMMAND_TRANSFERRING_ERROR, AllEquivalentsResponse{})
		return
	}

	result, err := router.nodeHandler.Node.GetResult(command, SETTLEMENT_LINE_RESULT_TIMEOUT)
	if err != nil {
		logger.Error("Node is inaccessible during processing command: " +
			string(command.ToBytes()) + ". Details: " + err.Error())
		writeHTTPResponse(w, NODE_IS_INACCESSIBLE, AllEquivalentsResponse{})
		return
	}

	if result.Code != OK {
		logger.Error("Node return wrong command result: " + strconv.Itoa(result.Code) +
			" on command: " + string(command.ToBytes()))
		writeHTTPResponse(w, result.Code, AllEquivalentsResponse{})
		return
	}

	if len(result.Tokens) == 0 {
		logger.Error("Node return invalid result tokens size on command: " + string(command.ToBytes()))
		writeHTTPResponse(w, ENGINE_UNEXPECTED_ERROR, AllEquivalentsResponse{})
		return
	}

	equivalentsCount, err := strconv.Atoi(result.Tokens[0])
	if err != nil {
		logger.Error("Node return invalid token on command: " +
			string(command.ToBytes()) + ". Details: " + err.Error())
		writeHTTPResponse(w, ENGINE_UNEXPECTED_ERROR, AllEquivalentsResponse{})
		return
	}

	if equivalentsCount == 0 {
		writeHTTPResponse(w, OK, AllEquivalentsResponse{Count: equivalentsCount})
		return
	}

	tokenIdx := 1

	response := AllEquivalentsResponse{Count: equivalentsCount}
	for range equivalentsCount {
		contractorsCount, err := strconv.Atoi(result.Tokens[tokenIdx+1])
		if err != nil {
			logger.Error("Node return invalid token on command: " +
				string(command.ToBytes()) + ". Details: " + err.Error())
			writeHTTPResponse(w, ENGINE_UNEXPECTED_ERROR, AllEquivalentsResponse{})
			return
		}
		responseSettlementLines := EquivalentStatistics{Eq: result.Tokens[tokenIdx], Count: contractorsCount}
		tokenIdx = tokenIdx + 2
		for i := range contractorsCount {
			responseSettlementLines.SettlementLines = append(responseSettlementLines.SettlementLines, SettlementLineListItem{
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
		writeHTTPResponse(w, COMMAND_TRANSFERRING_ERROR, ContractorsListResponse{})
		return
	}

	result, err := router.nodeHandler.Node.GetResult(command, CONTRACTORS_RESULT_TIMEOUT)
	if err != nil {
		logger.Error("Node is inaccessible during processing command: " +
			string(command.ToBytes()) + ". Details: " + err.Error())
		writeHTTPResponse(w, NODE_IS_INACCESSIBLE, ContractorsListResponse{})
		return
	}

	if result.Code != OK && result.Code != ENGINE_NO_EQUIVALENT {
		logger.Error("Node return wrong command result: " + strconv.Itoa(result.Code) +
			" on command: " + string(command.ToBytes()))
		writeHTTPResponse(w, result.Code, ContractorsListResponse{})
		return
	}
	if result.Code == ENGINE_NO_EQUIVALENT {
		logger.Info("Node hasn't equivalent for command: " + string(command.ToBytes()))
		writeHTTPResponse(w, result.Code, ContractorsListResponse{})
		return
	}

	if len(result.Tokens) == 0 {
		logger.Error("Node return invalid result tokens size on command: " + string(command.ToBytes()))
		writeHTTPResponse(w, ENGINE_UNEXPECTED_ERROR, ContractorsListResponse{})
		return
	}

	contractorsCount, err := strconv.Atoi(result.Tokens[0])
	if err != nil {
		logger.Error("Node return invalid token on command: " + string(command.ToBytes()) +
			". Details: " + err.Error())
		writeHTTPResponse(w, ENGINE_UNEXPECTED_ERROR, ContractorsListResponse{})
		return
	}

	if contractorsCount == 0 {
		writeHTTPResponse(w, OK, ContractorsListResponse{Count: contractorsCount})
		return
	}

	response := ContractorsListResponse{Count: contractorsCount}
	for i := range contractorsCount {
		response.Contractors = append(response.Contractors, ContractorInfo{
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
	if !handler.ValidateInt(contractorID) {
		logger.Error("Bad request: invalid contractorID parameter: " + url)
		w.WriteHeader(BAD_REQUEST)
		return
	}

	command := handler.NewCommand(
		"GET:contractors/trust-lines/one/id", contractorID, equivalent)

	err = router.nodeHandler.Node.SendCommand(command)
	if err != nil {
		logger.Error("Can't send command: " + string(command.ToBytes()) + " to node. Details: " + err.Error())
		writeHTTPResponse(w, COMMAND_TRANSFERRING_ERROR, SettlementLineDetailResponse{})
		return
	}

	result, err := router.nodeHandler.Node.GetResult(command, SETTLEMENT_LINE_RESULT_TIMEOUT)
	if err != nil {
		logger.Error("Node is inaccessible during processing command: " +
			string(command.ToBytes()) + ". Details: " + err.Error())
		writeHTTPResponse(w, NODE_IS_INACCESSIBLE, SettlementLineDetailResponse{})
		return
	}

	if result.Code != OK && result.Code != ENGINE_NO_EQUIVALENT && result.Code != NODE_NOT_FOUND {
		logger.Error("Node return wrong command result: " + strconv.Itoa(result.Code) +
			" on command: " + string(command.ToBytes()))
		writeHTTPResponse(w, result.Code, SettlementLineDetailResponse{})
		return
	}
	if result.Code == ENGINE_NO_EQUIVALENT {
		logger.Info("Node hasn't equivalent for command: " + string(command.ToBytes()))
		writeHTTPResponse(w, result.Code, SettlementLineDetailResponse{})
		return
	}
	if result.Code == NODE_NOT_FOUND {
		logger.Info("Node hasn't TL with requested contractor for command " + string(command.ToBytes()))
		writeHTTPResponse(w, result.Code, SettlementLineDetailResponse{})
		return
	}

	if len(result.Tokens) < 4 {
		logger.Error("Node return invalid result tokens size on command: " + string(command.ToBytes()))
		writeHTTPResponse(w, ENGINE_UNEXPECTED_ERROR, SettlementLineDetailResponse{})
		return
	}

	response := SettlementLineDetailResponse{SettlementLine: SettlementLineDetail{
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
		writeHTTPResponse(w, COMMAND_TRANSFERRING_ERROR, SettlementLineDetailResponse{})
		return
	}

	result, err := router.nodeHandler.Node.GetResult(command, SETTLEMENT_LINE_RESULT_TIMEOUT)
	if err != nil {
		logger.Error("Node is inaccessible during processing command: " +
			string(command.ToBytes()) + ". Details: " + err.Error())
		writeHTTPResponse(w, NODE_IS_INACCESSIBLE, SettlementLineDetailResponse{})
		return
	}

	if result.Code != OK && result.Code != ENGINE_NO_EQUIVALENT && result.Code != NODE_NOT_FOUND {
		logger.Error("Node return wrong command result: " + strconv.Itoa(result.Code) +
			" on command: " + string(command.ToBytes()))
		writeHTTPResponse(w, result.Code, SettlementLineDetailResponse{})
		return
	}
	if result.Code == ENGINE_NO_EQUIVALENT {
		logger.Info("Node hasn't equivalent for command: " + string(command.ToBytes()))
		writeHTTPResponse(w, result.Code, SettlementLineDetailResponse{})
		return
	}
	if result.Code == NODE_NOT_FOUND {
		logger.Info("Node hasn't TL with requested contractor for command " + string(command.ToBytes()))
		writeHTTPResponse(w, result.Code, SettlementLineDetailResponse{})
		return
	}

	if len(result.Tokens) < 4 {
		logger.Error("Node return invalid result tokens size on command: " + string(command.ToBytes()))
		writeHTTPResponse(w, ENGINE_UNEXPECTED_ERROR, SettlementLineDetailResponse{})
		return
	}

	response := SettlementLineDetailResponse{SettlementLine: SettlementLineDetail{
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
		writeHTTPResponse(w, COMMAND_TRANSFERRING_ERROR, EquivalentsListResponse{})
		return
	}

	result, err := router.nodeHandler.Node.GetResult(command, SETTLEMENT_LINE_RESULT_TIMEOUT)
	if err != nil {
		logger.Error("Node is inaccessible during processing command: " +
			string(command.ToBytes()) + ". Details: " + err.Error())
		writeHTTPResponse(w, NODE_IS_INACCESSIBLE, EquivalentsListResponse{})
		return
	}

	if result.Code != OK {
		logger.Error("Node return wrong command result: " + strconv.Itoa(result.Code) +
			" on command: " + string(command.ToBytes()))
		writeHTTPResponse(w, result.Code, EquivalentsListResponse{})
		return
	}

	if len(result.Tokens) == 0 {
		logger.Error("Node return invalid result tokens size on command: " + string(command.ToBytes()))
		writeHTTPResponse(w, ENGINE_UNEXPECTED_ERROR, EquivalentsListResponse{})
		return
	}

	equivalentsCount, err := strconv.Atoi(result.Tokens[0])
	if err != nil {
		logger.Error("Node return invalid token on command: " +
			string(command.ToBytes()) + ". Details: " + err.Error())
		writeHTTPResponse(w, ENGINE_UNEXPECTED_ERROR, EquivalentsListResponse{})
		return
	}

	if equivalentsCount == 0 {
		writeHTTPResponse(w, OK, EquivalentsListResponse{Count: equivalentsCount})
		return
	}

	response := EquivalentsListResponse{Count: equivalentsCount}
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
		writeHTTPResponse(w, COMMAND_TRANSFERRING_ERROR, TotalBalanceResponse{})
		return
	}

	result, err := router.nodeHandler.Node.GetResult(command, STATS_RESULT_TIMEOUT)
	if err != nil {
		logger.Error("Node is inaccessible during processing command: " + string(command.ToBytes()) +
			". Details: " + err.Error())
		writeHTTPResponse(w, NODE_IS_INACCESSIBLE, TotalBalanceResponse{})
		return
	}

	if result.Code != OK && result.Code != ENGINE_NO_EQUIVALENT {
		logger.Error("Node return wrong command result: " + strconv.Itoa(result.Code) +
			" on command: " + string(command.ToBytes()))
		writeHTTPResponse(w, result.Code, TotalBalanceResponse{})
		return
	}
	if result.Code == ENGINE_NO_EQUIVALENT {
		logger.Info("Node hasn't equivalent for command: " + string(command.ToBytes()))
		writeHTTPResponse(w, result.Code, TotalBalanceResponse{})
		return
	}

	if len(result.Tokens) < 4 {
		logger.Error("Node return invalid result tokens size on command: " + string(command.ToBytes()))
		writeHTTPResponse(w, ENGINE_UNEXPECTED_ERROR, TotalBalanceResponse{})
		return
	}

	response := TotalBalanceResponse{
		TotalMaxNegativeBalance: result.Tokens[0],
		TotalNegativeBalance:    result.Tokens[1],
		TotalMaxPositiveBalance: result.Tokens[2],
		TotalPositiveBalance:    result.Tokens[3],
	}
	writeHTTPResponse(w, OK, response)
}
