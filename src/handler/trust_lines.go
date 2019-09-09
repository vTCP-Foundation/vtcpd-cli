package handler

import (
	"fmt"
	"github.com/gorilla/mux"
	"logger"
	"net/http"
	"strconv"
	"strings"
)

var (
	TRUST_LINE_RESULT_TIMEOUT  uint16 = 20 // seconds
	CONTRACTORS_RESULT_TIMEOUT uint16 = 20
	STATS_RESULT_TIMEOUT       uint16 = 20 // seconds
	DEFAULT_TRUST_LINES_OFFSET        = "0"
	DFEAULT_TRUST_LINES_COUNT         = "10000"
)

func (handler *NodesHandler) TrustLines() {

	if CommandType == "open" {
		handler.initTrustLine()

	} else if CommandType == "set" {
		handler.setOutgoingTrustLine()

	} else if CommandType == "close-incoming" {
		handler.closeIncomingTrustLine()

	} else if CommandType == "share-keys" {
		handler.shareKeysTrustLine()

	} else if CommandType == "delete" {
		handler.deleteTrustLine()

	} else if CommandType == "get" {
		handler.listTrustLinesPortions()

	} else if CommandType == "get-contractors" {
		handler.listContractors()

	} else if CommandType == "get-by-id" {
		handler.trustLineByID()

	} else if CommandType == "get-by-addresses" {
		handler.trustLineByAddresses()

	} else if CommandType == "equivalents" {
		handler.listEquivalents()

	} else if CommandType == "total-balance" {
		handler.totalBalance()

	} else {
		logger.Error("Invalid trust-line command " + CommandType)
		fmt.Println("Invalid trust-line command")
		return
	}
}

func (handler *NodesHandler) initTrustLine() {
	if !ValidateInt(ContractorID) {
		logger.Error("Bad request: invalid contractorID parameter in open request")
		fmt.Println("Bad request: invalid contractorID parameter")
		return
	}
	if !ValidateInt(Equivalent) {
		logger.Error("Bad request: invalid equivalent parameter in open request")
		fmt.Println("Bad request: invalid equivalent parameter")
		return
	}

	command := NewCommand("INIT:contractors/trust-line", ContractorID, Equivalent)

	go handler.actionTrustLineGetResult(command)
}

func (handler *NodesHandler) setOutgoingTrustLine() {
	if !ValidateInt(ContractorID) {
		logger.Error("Bad request: invalid contractorID parameter in set request")
		fmt.Println("Bad request: invalid contractorID parameter")
		return
	}
	if !ValidateTrustLineAmount(Amount) {
		logger.Error("Bad request: invalid amount parameter in set request")
		fmt.Println("Bad request: invalid amount parameter")
		return
	}
	if !ValidateInt(Equivalent) {
		logger.Error("Bad request: invalid equivalent parameter in set request")
		fmt.Println("Bad request: invalid equivalent parameter")
		return
	}

	command := NewCommand("SET:contractors/trust-lines", ContractorID, Amount, Equivalent)

	go handler.actionTrustLineGetResult(command)
}

func (handler *NodesHandler) closeIncomingTrustLine() {
	if !ValidateInt(ContractorID) {
		logger.Error("Bad request: invalid contractorID parameter in close-incoming request")
		fmt.Println("Bad request: invalid contractorID parameter")
		return
	}
	if !ValidateInt(Equivalent) {
		logger.Error("Bad request: invalid equivalent parameter in close-incoming request")
		fmt.Println("Bad request: invalid equivalent parameter")
		return
	}

	command := NewCommand("DELETE:contractors/incoming-trust-line", ContractorID, Equivalent)

	go handler.actionTrustLineGetResult(command)
}

func (handler *NodesHandler) shareKeysTrustLine() {
	if !ValidateInt(ContractorID) {
		logger.Error("Bad request: invalid contractorID parameter in share-keys request")
		fmt.Println("Bad request: invalid contractorID parameter")
		return
	}
	if !ValidateInt(Equivalent) {
		logger.Error("Bad request: invalid equivalent parameter in share-keys request")
		fmt.Println("Bad request: invalid equivalent parameter")
		return
	}

	command := NewCommand("SET:contractors/trust-line-keys", ContractorID, Equivalent)

	go handler.actionTrustLineGetResult(command)
}

func (handler *NodesHandler) deleteTrustLine() {
	if !ValidateInt(ContractorID) {
		logger.Error("Bad request: invalid contractorID parameter in delete request")
		fmt.Println("Bad request: invalid contractorID parameter")
		return
	}
	if !ValidateInt(Equivalent) {
		logger.Error("Bad request: invalid equivalent parameter in delete request")
		fmt.Println("Bad request: invalid equivalent parameter")
		return
	}

	command := NewCommand("DELETE:contractors/trust-line", ContractorID, Equivalent)

	go handler.actionTrustLineGetResult(command)
}

func (handler *NodesHandler) actionTrustLineGetResult(command *Command) {

	type Response struct{}

	err := handler.node.SendCommand(command)
	if err != nil {
		logger.Error("Can't send command: " + string(command.ToBytes()) + " to node. Details: " + err.Error())
		resultJSON := buildJSONResponse(COMMAND_TRANSFERRING_ERROR, Response{})
		fmt.Println(string(resultJSON))
		return
	}

	result, err := handler.node.GetResult(command, TRUST_LINE_RESULT_TIMEOUT)
	if err != nil {
		logger.Error("Node is inaccessible during processing command: " +
			string(command.ToBytes()) + ". Details: " + err.Error())
		resultJSON := buildJSONResponse(NODE_IS_INACCESSIBLE, Response{})
		fmt.Println(string(resultJSON))
		return
	}

	if result.Code != OK {
		logger.Error("Node return wrong command result: " + strconv.Itoa(result.Code) +
			" on command: " + string(command.ToBytes()))
	}

	resultJSON := buildJSONResponse(result.Code, Response{})
	fmt.Println(string(resultJSON))
}

func (handler *NodesHandler) InitTrustLine(w http.ResponseWriter, r *http.Request) {
	url, err := preprocessRequest(r)
	if err != nil {
		logger.Error("Bad request: invalid security parameters: " + err.Error())
		w.WriteHeader(BAD_REQUEST)
		return
	}

	contractorID := mux.Vars(r)["contractor_id"]
	if !ValidateInt(contractorID) {
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
	command := NewCommand("INIT:contractors/trust-line", contractorID, equivalent)

	type Response struct{}

	err = handler.node.SendCommand(command)
	if err != nil {
		logger.Error("Can't send command: " + string(command.ToBytes()) + " to node. Details: " + err.Error())
		writeHTTPResponse(w, COMMAND_TRANSFERRING_ERROR, Response{})
		return
	}

	result, err := handler.node.GetResult(command, TRUST_LINE_RESULT_TIMEOUT)
	if err != nil {
		logger.Error("Node is inaccessible during processing command: " +
			string(command.ToBytes()) + ". Details: " + err.Error())
		writeHTTPResponse(w, NODE_IS_INACCESSIBLE, Response{})
		return
	}

	if result.Code != OK {
		logger.Error("Node return wrong command result: " + strconv.Itoa(result.Code) +
			" on command: " + string(command.ToBytes()))
	}

	writeHTTPResponse(w, result.Code, Response{})
}

func (handler *NodesHandler) SetTrustLine(w http.ResponseWriter, r *http.Request) {
	url, err := preprocessRequest(r)
	if err != nil {
		logger.Error("Bad request: invalid security parameters: " + err.Error())
		w.WriteHeader(BAD_REQUEST)
		return
	}

	contractorID := mux.Vars(r)["contractor_id"]
	if !ValidateInt(contractorID) {
		logger.Error("Bad request: invalid contractor_id parameter: " + url)
		w.WriteHeader(BAD_REQUEST)
		return
	}

	equivalent := mux.Vars(r)["equivalent"]
	if !ValidateInt(equivalent) {
		logger.Error("Bad request: invalid equivalent parameter: " + url)
		w.WriteHeader(BAD_REQUEST)
		return
	}

	amount := r.FormValue("amount")
	if !ValidateTrustLineAmount(amount) {
		logger.Error("Bad request: invalid amount parameter: " + url)
		w.WriteHeader(BAD_REQUEST)
		return
	}

	command := NewCommand(
		"SET:contractors/trust-lines", contractorID, amount, equivalent)

	type Response struct{}

	err = handler.node.SendCommand(command)
	if err != nil {
		logger.Error("Can't send command: " + string(command.ToBytes()) + " to node. Details: " + err.Error())
		writeHTTPResponse(w, COMMAND_TRANSFERRING_ERROR, Response{})
		return
	}

	result, err := handler.node.GetResult(command, TRUST_LINE_RESULT_TIMEOUT)
	if err != nil {
		logger.Error("Node is inaccessible during processing command: " +
			string(command.ToBytes()) + ". Details: " + err.Error())
		writeHTTPResponse(w, NODE_IS_INACCESSIBLE, Response{})
		return
	}

	if result.Code != OK {
		logger.Error("Node return wrong command result: " + strconv.Itoa(result.Code) +
			" on command: " + string(command.ToBytes()))
	}

	writeHTTPResponse(w, result.Code, Response{})
}

func (handler *NodesHandler) CloseIncomingTrustLine(w http.ResponseWriter, r *http.Request) {
	url, err := preprocessRequest(r)
	if err != nil {
		logger.Error("Bad request: invalid security parameters: " + err.Error())
		w.WriteHeader(BAD_REQUEST)
		return
	}

	contractorID := mux.Vars(r)["contractor_id"]
	if !ValidateInt(contractorID) {
		logger.Error("Bad request: invalid contractor_id parameter: " + url)
		w.WriteHeader(BAD_REQUEST)
		return
	}

	equivalent := mux.Vars(r)["equivalent"]
	if !ValidateInt(equivalent) {
		logger.Error("Bad request: invalid equivalent parameter: " + url)
		w.WriteHeader(BAD_REQUEST)
		return
	}

	command := NewCommand(
		"DELETE:contractors/incoming-trust-line", contractorID, equivalent)

	type Response struct{}

	err = handler.node.SendCommand(command)
	if err != nil {
		logger.Error("Can't send command: " + string(command.ToBytes()) + " to node. Details: " + err.Error())
		writeHTTPResponse(w, COMMAND_TRANSFERRING_ERROR, Response{})
		return
	}

	result, err := handler.node.GetResult(command, TRUST_LINE_RESULT_TIMEOUT)
	if err != nil {
		logger.Error("Node is inaccessible during processing command: " +
			string(command.ToBytes()) + ". Details: " + err.Error())
		writeHTTPResponse(w, NODE_IS_INACCESSIBLE, Response{})
		return
	}

	if result.Code != OK {
		logger.Error("Node return wrong command result: " + strconv.Itoa(result.Code) +
			" on command: " + string(command.ToBytes()))
	}

	writeHTTPResponse(w, result.Code, Response{})
}

func (handler *NodesHandler) PublicKeysSharing(w http.ResponseWriter, r *http.Request) {
	url, err := preprocessRequest(r)
	if err != nil {
		logger.Error("Bad request: invalid security parameters: " + err.Error())
		w.WriteHeader(BAD_REQUEST)
		return
	}

	contractorID := mux.Vars(r)["contractor_id"]
	if !ValidateInt(contractorID) {
		logger.Error("Bad request: invalid contractor_id parameter: " + url)
		w.WriteHeader(BAD_REQUEST)
		return
	}

	equivalent := mux.Vars(r)["equivalent"]
	if !ValidateInt(equivalent) {
		logger.Error("Bad request: invalid equivalent parameter: " + url)
		w.WriteHeader(BAD_REQUEST)
		return
	}

	command := NewCommand(
		"SET:contractors/trust-line-keys", contractorID, equivalent)

	type Response struct{}

	err = handler.node.SendCommand(command)
	if err != nil {
		logger.Error("Can't send command: " + string(command.ToBytes()) + " to node. Details: " + err.Error())
		writeHTTPResponse(w, COMMAND_TRANSFERRING_ERROR, Response{})
		return
	}

	result, err := handler.node.GetResult(command, TRUST_LINE_RESULT_TIMEOUT)
	if err != nil {
		logger.Error("Node is inaccessible during processing command: " +
			string(command.ToBytes()) + ". Details: " + err.Error())
		writeHTTPResponse(w, NODE_IS_INACCESSIBLE, Response{})
		return
	}

	if result.Code != OK {
		logger.Error("Node return wrong command result: " + strconv.Itoa(result.Code) +
			" on command: " + string(command.ToBytes()))
	}

	writeHTTPResponse(w, result.Code, Response{})
}

func (handler *NodesHandler) RemoveTrustLine(w http.ResponseWriter, r *http.Request) {
	url, err := preprocessRequest(r)
	if err != nil {
		logger.Error("Bad request: invalid security parameters: " + err.Error())
		w.WriteHeader(BAD_REQUEST)
		return
	}

	contractorID := mux.Vars(r)["contractor_id"]
	if !ValidateInt(contractorID) {
		logger.Error("Bad request: invalid contractor_id parameter: " + url)
		w.WriteHeader(BAD_REQUEST)
		return
	}

	equivalent := mux.Vars(r)["equivalent"]
	if !ValidateInt(equivalent) {
		logger.Error("Bad request: invalid equivalent parameter: " + url)
		w.WriteHeader(BAD_REQUEST)
		return
	}

	command := NewCommand(
		"DELETE:contractors/trust-line", contractorID, equivalent)

	type Response struct{}

	err = handler.node.SendCommand(command)
	if err != nil {
		logger.Error("Can't send command: " + string(command.ToBytes()) + " to node. Details: " + err.Error())
		writeHTTPResponse(w, COMMAND_TRANSFERRING_ERROR, Response{})
		return
	}

	result, err := handler.node.GetResult(command, TRUST_LINE_RESULT_TIMEOUT)
	if err != nil {
		logger.Error("Node is inaccessible during processing command: " +
			string(command.ToBytes()) + ". Details: " + err.Error())
		writeHTTPResponse(w, NODE_IS_INACCESSIBLE, Response{})
		return
	}

	if result.Code != OK {
		logger.Error("Node return wrong command result: " + strconv.Itoa(result.Code) +
			" on command: " + string(command.ToBytes()))
	}

	writeHTTPResponse(w, result.Code, Response{})
}

func (handler *NodesHandler) listTrustLinesPortions() {
	if Offset == "" {
		Offset = DEFAULT_TRUST_LINES_OFFSET
	} else if !ValidateInt(Offset) {
		logger.Error("Bad request: invalid offset parameter in get request")
		fmt.Println("Bad request: invalid offset parameter")
		return
	}

	if Count == "" {
		Count = DFEAULT_TRUST_LINES_COUNT
	} else if !ValidateInt(Count) {
		logger.Error("Bad request: invalid count parameter in get request")
		fmt.Println("Bad request: invalid count parameter")
		return
	}

	if !ValidateInt(Equivalent) {
		logger.Error("Bad request: invalid equivalent parameter in get request")
		fmt.Println("Bad request: invalid equivalent parameter")
		return
	}

	command := NewCommand("GET:contractors/trust-lines", Offset, Count, Equivalent)

	go handler.listTrustLinesResult(command)
}

func (handler *NodesHandler) listTrustLinesResult(command *Command) {
	type TrustLine struct {
		ID                    string `json:"contractor_id"`
		Contractor            string `json:"contractor"`
		State                 string `json:"state"`
		OwnKeysPresent        string `json:"own_keys_present"`
		ContractorKeysPresent string `json:"contractor_keys_present"`
		IncomingTrustAmount   string `json:"incoming_trust_amount"`
		OutgoingTrustAmount   string `json:"outgoing_trust_amount"`
		Balance               string `json:"balance"`
	}

	type Response struct {
		Count      int         `json:"count"`
		TrustLines []TrustLine `json:"trust_lines"`
	}

	err := handler.node.SendCommand(command)
	if err != nil {
		logger.Error("Can't send command: " + string(command.ToBytes()) + " to node. Details: " + err.Error())
		resultJSON := buildJSONResponse(COMMAND_TRANSFERRING_ERROR, Response{})
		fmt.Println(string(resultJSON))
		return
	}

	result, err := handler.node.GetResult(command, TRUST_LINE_RESULT_TIMEOUT)
	if err != nil {
		logger.Error("Node is inaccessible during processing command: " +
			string(command.ToBytes()) + ". Details: " + err.Error())
		resultJSON := buildJSONResponse(NODE_IS_INACCESSIBLE, Response{})
		fmt.Println(string(resultJSON))
		return
	}

	if result.Code != OK && result.Code != ENGINE_NO_EQUIVALENT {
		logger.Error("Node returned wrong command result: " + strconv.Itoa(result.Code) +
			" on command: " + string(command.ToBytes()))
		resultJSON := buildJSONResponse(result.Code, Response{})
		fmt.Println(string(resultJSON))
		return
	}
	if result.Code == ENGINE_NO_EQUIVALENT {
		logger.Info("Node hasn't equivalent for command: " + string(command.ToBytes()))
		resultJSON := buildJSONResponse(result.Code, Response{})
		fmt.Println(string(resultJSON))
		return
	}

	if len(result.Tokens) == 0 {
		logger.Error("Node returned invalid result tokens size on command: " + string(command.ToBytes()))
		resultJSON := buildJSONResponse(ENGINE_UNEXPECTED_ERROR, Response{})
		fmt.Println(string(resultJSON))
		return
	}

	// Contractors received well
	contractorsCount, err := strconv.Atoi(result.Tokens[0])
	if err != nil {
		logger.Error("Node returned invalid token on command: " +
			string(command.ToBytes()) + ". Details: " + err.Error())
		resultJSON := buildJSONResponse(ENGINE_UNEXPECTED_ERROR, Response{})
		fmt.Println(string(resultJSON))
		return
	}

	if contractorsCount == 0 {
		resultJSON := buildJSONResponse(OK, Response{Count: contractorsCount})
		fmt.Println(string(resultJSON))
		return
	}

	response := Response{Count: contractorsCount}
	for i := 0; i < contractorsCount; i++ {
		response.TrustLines = append(response.TrustLines, TrustLine{
			ID:                    result.Tokens[i*8+1],
			Contractor:            result.Tokens[i*8+2],
			State:                 result.Tokens[i*8+3],
			OwnKeysPresent:        result.Tokens[i*8+4],
			ContractorKeysPresent: result.Tokens[i*8+5],
			IncomingTrustAmount:   result.Tokens[i*8+6],
			OutgoingTrustAmount:   result.Tokens[i*8+7],
			Balance:               result.Tokens[i*8+8],
		})
	}
	resultJSON := buildJSONResponse(OK, response)
	fmt.Println(string(resultJSON))
}

func (handler *NodesHandler) ListTrustLines(w http.ResponseWriter, r *http.Request) {
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

	command := NewCommand("GET:contractors/trust-lines", DEFAULT_TRUST_LINES_OFFSET, DFEAULT_TRUST_LINES_COUNT, equivalent)

	type TrustLine struct {
		ID                    string `json:"contractor_id"`
		Contractor            string `json:"contractor"`
		State                 string `json:"state"`
		OwnKeysPresent        string `json:"own_keys_present"`
		ContractorKeysPresent string `json:"contractor_keys_present"`
		IncomingTrustAmount   string `json:"incoming_trust_amount"`
		OutgoingTrustAmount   string `json:"outgoing_trust_amount"`
		Balance               string `json:"balance"`
	}

	type Response struct {
		Count      int         `json:"count"`
		TrustLines []TrustLine `json:"trust_lines"`
	}

	err = handler.node.SendCommand(command)
	if err != nil {
		logger.Error("Can't send command: " + string(command.ToBytes()) + " to node. Details: " + err.Error())
		writeHTTPResponse(w, COMMAND_TRANSFERRING_ERROR, Response{})
		return
	}

	result, err := handler.node.GetResult(command, TRUST_LINE_RESULT_TIMEOUT)
	if err != nil {
		logger.Error("Node is inaccessible during processing command: " +
			string(command.ToBytes()) + ". Details: " + err.Error())
		writeHTTPResponse(w, NODE_IS_INACCESSIBLE, Response{})
		return
	}

	if result.Code != OK && result.Code != ENGINE_NO_EQUIVALENT {
		logger.Error("Node return wrong command result: " + strconv.Itoa(result.Code) +
			" on command: " + string(command.ToBytes()))
		writeHTTPResponse(w, result.Code, Response{})
		return
	}
	if result.Code == ENGINE_NO_EQUIVALENT {
		logger.Info("Node hasn't equivalent for command: " + string(command.ToBytes()))
		writeHTTPResponse(w, result.Code, Response{})
		return
	}

	if len(result.Tokens) == 0 {
		logger.Error("Node return invalid result tokens size on command: " + string(command.ToBytes()))
		writeHTTPResponse(w, ENGINE_UNEXPECTED_ERROR, Response{})
		return
	}

	// Contractors received well
	contractorsCount, err := strconv.Atoi(result.Tokens[0])
	if err != nil {
		logger.Error("Node return invalid token on command: " +
			string(command.ToBytes()) + ". Details: " + err.Error())
		writeHTTPResponse(w, ENGINE_UNEXPECTED_ERROR, Response{})
		return
	}

	if contractorsCount == 0 {
		writeHTTPResponse(w, OK, Response{Count: contractorsCount})
		return
	}

	response := Response{Count: contractorsCount}
	for i := 0; i < contractorsCount; i++ {
		response.TrustLines = append(response.TrustLines, TrustLine{
			ID:                    result.Tokens[i*8+1],
			Contractor:            result.Tokens[i*8+2],
			State:                 result.Tokens[i*8+3],
			OwnKeysPresent:        result.Tokens[i*8+4],
			ContractorKeysPresent: result.Tokens[i*8+5],
			IncomingTrustAmount:   result.Tokens[i*8+6],
			OutgoingTrustAmount:   result.Tokens[i*8+7],
			Balance:               result.Tokens[i*8+8],
		})
	}
	writeHTTPResponse(w, OK, response)
}

func (handler *NodesHandler) ListTrustLinesPortions(w http.ResponseWriter, r *http.Request) {
	url, err := preprocessRequest(r)
	if err != nil {
		logger.Error("Bad request: invalid security parameters: " + err.Error())
		w.WriteHeader(BAD_REQUEST)
		return
	}

	offset, isParamPresent := mux.Vars(r)["offset"]
	if !isParamPresent || !ValidateInt(offset) {
		logger.Error("Bad request: invalid offset parameter: " + url)
		w.WriteHeader(BAD_REQUEST)
		return
	}

	count, isParamPresent := mux.Vars(r)["count"]
	if !isParamPresent || !ValidateInt(count) {
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

	command := NewCommand("GET:contractors/trust-lines", offset, count, equivalent)

	type TrustLine struct {
		ID                    string `json:"contractor_id"`
		Contractor            string `json:"contractor"`
		State                 string `json:"state"`
		OwnKeysPresent        string `json:"own_keys_present"`
		ContractorKeysPresent string `json:"contractor_keys_present"`
		IncomingTrustAmount   string `json:"incoming_trust_amount"`
		OutgoingTrustAmount   string `json:"outgoing_trust_amount"`
		Balance               string `json:"balance"`
	}

	type Response struct {
		Count      int         `json:"count"`
		TrustLines []TrustLine `json:"trust_lines"`
	}

	err = handler.node.SendCommand(command)
	if err != nil {
		logger.Error("Can't send command: " + string(command.ToBytes()) + " to node. Details: " + err.Error())
		writeHTTPResponse(w, COMMAND_TRANSFERRING_ERROR, Response{})
		return
	}

	result, err := handler.node.GetResult(command, TRUST_LINE_RESULT_TIMEOUT)
	if err != nil {
		logger.Error("Node is inaccessible during processing command: " +
			string(command.ToBytes()) + ". Details: " + err.Error())
		writeHTTPResponse(w, NODE_IS_INACCESSIBLE, Response{})
		return
	}

	if result.Code != OK && result.Code != ENGINE_NO_EQUIVALENT {
		logger.Error("Node return wrong command result: " + strconv.Itoa(result.Code) +
			" on command: " + string(command.ToBytes()))
		writeHTTPResponse(w, result.Code, Response{})
		return
	}
	if result.Code == ENGINE_NO_EQUIVALENT {
		logger.Info("Node hasn't equivalent for command: " + string(command.ToBytes()))
		writeHTTPResponse(w, result.Code, Response{})
		return
	}

	if len(result.Tokens) == 0 {
		logger.Error("Node return invalid result tokens size on command: " + string(command.ToBytes()))
		writeHTTPResponse(w, ENGINE_UNEXPECTED_ERROR, Response{})
		return
	}

	// Contractors received well
	contractorsCount, err := strconv.Atoi(result.Tokens[0])
	if err != nil {
		logger.Error("Node return invalid token on command: " +
			string(command.ToBytes()) + ". Details: " + err.Error())
		writeHTTPResponse(w, ENGINE_UNEXPECTED_ERROR, Response{})
		return
	}

	if contractorsCount == 0 {
		writeHTTPResponse(w, OK, Response{Count: contractorsCount})
		return
	}

	response := Response{Count: contractorsCount}
	for i := 0; i < contractorsCount; i++ {
		response.TrustLines = append(response.TrustLines, TrustLine{
			ID:                    result.Tokens[i*8+1],
			Contractor:            result.Tokens[i*8+2],
			State:                 result.Tokens[i*8+3],
			OwnKeysPresent:        result.Tokens[i*8+4],
			ContractorKeysPresent: result.Tokens[i*8+5],
			IncomingTrustAmount:   result.Tokens[i*8+6],
			OutgoingTrustAmount:   result.Tokens[i*8+7],
			Balance:               result.Tokens[i*8+8],
		})
	}
	writeHTTPResponse(w, OK, response)
}

func (handler *NodesHandler) listContractors() {

	if !ValidateInt(Equivalent) {
		logger.Error("Bad request: invalid equivalent parameter in get-contractors request")
		fmt.Println("Bad request: invalid equivalent parameter")
		return
	}

	command := NewCommand("GET:contractors", Equivalent)

	go handler.listContractorsResult(command)
}

func (handler *NodesHandler) listContractorsResult(command *Command) {
	type Record struct {
		ContractorID        string `json:"contractor_id"`
		ContractorAddresses string `json:"contractor_addresses"`
	}

	type Response struct {
		Count       int      `json:"count"`
		Contractors []Record `json:"contractors"`
	}

	err := handler.node.SendCommand(command)
	if err != nil {
		logger.Error("Can't send command: " + string(command.ToBytes()) + " to node. Details: " + err.Error())
		resultJSON := buildJSONResponse(COMMAND_TRANSFERRING_ERROR, Response{})
		fmt.Println(string(resultJSON))
		return
	}

	result, err := handler.node.GetResult(command, CONTRACTORS_RESULT_TIMEOUT)
	if err != nil {
		logger.Error("Node is inaccessible during processing command: " +
			string(command.ToBytes()) + ". Details: " + err.Error())
		resultJSON := buildJSONResponse(NODE_IS_INACCESSIBLE, Response{})
		fmt.Println(string(resultJSON))
		return
	}

	if result.Code != OK && result.Code != ENGINE_NO_EQUIVALENT {
		logger.Error("Node return wrong command result: " + strconv.Itoa(result.Code) +
			" on command: " + string(command.ToBytes()))
		resultJSON := buildJSONResponse(result.Code, Response{})
		fmt.Println(string(resultJSON))
		return
	}
	if result.Code == ENGINE_NO_EQUIVALENT {
		logger.Info("Node hasn't equivalent for command: " + string(command.ToBytes()))
		resultJSON := buildJSONResponse(result.Code, Response{})
		fmt.Println(string(resultJSON))
		return
	}

	if len(result.Tokens) == 0 {
		logger.Error("Node return invalid result tokens size on command: " + string(command.ToBytes()))
		resultJSON := buildJSONResponse(ENGINE_UNEXPECTED_ERROR, Response{})
		fmt.Println(string(resultJSON))
		return
	}

	contractorsCount, err := strconv.Atoi(result.Tokens[0])
	if err != nil {
		logger.Error("Node return invalid token on command: " + string(command.ToBytes()) +
			". Details: " + err.Error())
		resultJSON := buildJSONResponse(ENGINE_UNEXPECTED_ERROR, Response{})
		fmt.Println(string(resultJSON))
		return
	}

	if contractorsCount == 0 {
		resultJSON := buildJSONResponse(OK, Response{Count: contractorsCount})
		fmt.Println(string(resultJSON))
		return
	}

	response := Response{Count: contractorsCount}
	for i := 0; i < contractorsCount; i++ {
		response.Contractors = append(response.Contractors, Record{
			ContractorID:        result.Tokens[i*2+1],
			ContractorAddresses: result.Tokens[i*2+2],
		})
	}

	resultJSON := buildJSONResponse(OK, response)
	fmt.Println(string(resultJSON))
}

func (handler *NodesHandler) ListContractors(w http.ResponseWriter, r *http.Request) {
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

	command := NewCommand("GET:contractors", equivalent)

	type Record struct {
		ContractorID        string `json:"contractor_id"`
		ContractorAddresses string `json:"contractor_addresses"`
	}

	type Response struct {
		Count       int      `json:"count"`
		Contractors []Record `json:"contractors"`
	}

	err = handler.node.SendCommand(command)
	if err != nil {
		logger.Error("Can't send command: " + string(command.ToBytes()) + " to node. Details: " + err.Error())
		writeHTTPResponse(w, COMMAND_TRANSFERRING_ERROR, Response{})
		return
	}

	result, err := handler.node.GetResult(command, CONTRACTORS_RESULT_TIMEOUT)
	if err != nil {
		logger.Error("Node is inaccessible during processing command: " +
			string(command.ToBytes()) + ". Details: " + err.Error())
		writeHTTPResponse(w, NODE_IS_INACCESSIBLE, Response{})
		return
	}

	if result.Code != OK && result.Code != ENGINE_NO_EQUIVALENT {
		logger.Error("Node return wrong command result: " + strconv.Itoa(result.Code) +
			" on command: " + string(command.ToBytes()))
		writeHTTPResponse(w, result.Code, Response{})
		return
	}
	if result.Code == ENGINE_NO_EQUIVALENT {
		logger.Info("Node hasn't equivalent for command: " + string(command.ToBytes()))
		writeHTTPResponse(w, result.Code, Response{})
		return
	}

	if len(result.Tokens) == 0 {
		logger.Error("Node return invalid result tokens size on command: " + string(command.ToBytes()))
		writeHTTPResponse(w, ENGINE_UNEXPECTED_ERROR, Response{})
		return
	}

	contractorsCount, err := strconv.Atoi(result.Tokens[0])
	if err != nil {
		logger.Error("Node return invalid token on command: " + string(command.ToBytes()) +
			". Details: " + err.Error())
		writeHTTPResponse(w, ENGINE_UNEXPECTED_ERROR, Response{})
		return
	}

	if contractorsCount == 0 {
		writeHTTPResponse(w, OK, Response{Count: contractorsCount})
		return
	}

	response := Response{Count: contractorsCount}
	for i := 0; i < contractorsCount; i++ {
		response.Contractors = append(response.Contractors, Record{
			ContractorID:        result.Tokens[i*2+1],
			ContractorAddresses: result.Tokens[i*2+2],
		})
	}

	writeHTTPResponse(w, OK, response)
}

func (handler *NodesHandler) trustLineByID() {
	if !ValidateInt(ContractorID) {
		logger.Error("Bad request: invalid contractorID parameter in get-by-id request")
		fmt.Println("Bad request: invalid contractorID parameter")
		return
	}
	if !ValidateInt(Equivalent) {
		logger.Error("Bad request: invalid equivalent parameter in get-by-id request")
		fmt.Println("Bad request: invalid equivalent parameter")
		return
	}

	command := NewCommand("GET:contractors/trust-lines/one/id", ContractorID, Equivalent)

	go handler.trustLineGetResult(command)
}

func (handler *NodesHandler) trustLineByAddresses() {
	if len(Addresses) == 0 {
		logger.Error("Bad request: there are no contractor addresses parameters in get-by-addresses request")
		fmt.Println("Bad request: there are no contractor addresses parameters")
		return
	}

	if !ValidateInt(Equivalent) {
		logger.Error("Bad request: invalid equivalent parameter in get-by-addresses request")
		fmt.Println("Bad request: invalid equivalent parameter")
		return
	}

	var addresses []string
	for idx := 0; idx < len(Addresses); idx++ {
		addressType, address := ValidateAddress(Addresses[idx])
		if addressType == "" {
			logger.Error("Bad request: invalid address parameter in get-by-addresses request")
			fmt.Println("Bad request: invalid address parameter")
			return
		}
		addresses = append(addresses, addressType, address)
	}

	addresses = append([]string{strconv.Itoa(len(Addresses))}, addresses...)
	addresses = append([]string{"GET:contractors/trust-lines/one/address"}, addresses...)
	addresses = append(addresses, []string{Equivalent}...)
	command := NewCommand(addresses...)

	go handler.trustLineGetResult(command)
}

func (handler *NodesHandler) trustLineGetResult(command *Command) {
	type TrustLine struct {
		ID                    string `json:"id"`
		State                 string `json:"state"`
		OwnKeysPresent        string `json:"own_keys_present"`
		ContractorKeysPresent string `json:"contractor_keys_present"`
		IncomingTrustAmount   string `json:"incoming_trust_amount"`
		OutgoingTrustAmount   string `json:"outgoing_trust_amount"`
		Balance               string `json:"balance"`
	}

	type Response struct {
		TrustLine TrustLine `json:"trust_line"`
	}

	err := handler.node.SendCommand(command)
	if err != nil {
		logger.Error("Can't send command: " + string(command.ToBytes()) + " to node. Details: " + err.Error())
		resultJSON := buildJSONResponse(COMMAND_TRANSFERRING_ERROR, Response{})
		fmt.Println(string(resultJSON))
		return
	}

	result, err := handler.node.GetResult(command, TRUST_LINE_RESULT_TIMEOUT)
	if err != nil {
		// Remote node is inaccessible
		logger.Error("Node is inaccessible during processing command: " +
			string(command.ToBytes()) + ". Details: " + err.Error())
		resultJSON := buildJSONResponse(NODE_IS_INACCESSIBLE, Response{})
		fmt.Println(string(resultJSON))
		return
	}

	if result.Code != OK && result.Code != ENGINE_NO_EQUIVALENT && result.Code != NODE_NOT_FOUND {
		logger.Error("Node return wrong command result: " + strconv.Itoa(result.Code) +
			" on command: " + string(command.ToBytes()))
		resultJSON := buildJSONResponse(result.Code, Response{})
		fmt.Println(string(resultJSON))
		return
	}
	if result.Code == ENGINE_NO_EQUIVALENT {
		logger.Info("Node hasn't equivalent for command: " + string(command.ToBytes()))
		resultJSON := buildJSONResponse(result.Code, Response{})
		fmt.Println(string(resultJSON))
		return
	}
	if result.Code == NODE_NOT_FOUND {
		logger.Info("Node hasn't TL with requested contractor for command " + string(command.ToBytes()))
		resultJSON := buildJSONResponse(result.Code, Response{})
		fmt.Println(string(resultJSON))
		return
	}

	if len(result.Tokens) < 4 {
		logger.Error("Node return invalid result tokens size on command: " + string(command.ToBytes()))
		resultJSON := buildJSONResponse(ENGINE_UNEXPECTED_ERROR, Response{})
		fmt.Println(string(resultJSON))
		return
	}

	// Contractors received well
	response := Response{TrustLine: TrustLine{
		ID:                    result.Tokens[0],
		State:                 result.Tokens[1],
		OwnKeysPresent:        result.Tokens[2],
		ContractorKeysPresent: result.Tokens[3],
		IncomingTrustAmount:   result.Tokens[4],
		OutgoingTrustAmount:   result.Tokens[5],
		Balance:               result.Tokens[6],
	}}
	resultJSON := buildJSONResponse(OK, response)
	fmt.Println(string(resultJSON))
}

func (handler *NodesHandler) GetTrustLineByID(w http.ResponseWriter, r *http.Request) {
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
	if !ValidateInt(contractorID) {
		logger.Error("Bad request: invalid contractorID parameter: " + url)
		w.WriteHeader(BAD_REQUEST)
		return
	}

	command := NewCommand(
		"GET:contractors/trust-lines/one/id", contractorID, equivalent)

	type TrustLine struct {
		ID                    string `json:"id"`
		State                 string `json:"state"`
		OwnKeysPresent        string `json:"own_keys_present"`
		ContractorKeysPresent string `json:"contractor_keys_present"`
		IncomingTrustAmount   string `json:"incoming_trust_amount"`
		OutgoingTrustAmount   string `json:"outgoing_trust_amount"`
		Balance               string `json:"balance"`
	}

	type Response struct {
		TrustLine TrustLine `json:"trust_line"`
	}

	err = handler.node.SendCommand(command)
	if err != nil {
		logger.Error("Can't send command: " + string(command.ToBytes()) + " to node. Details: " + err.Error())
		writeHTTPResponse(w, COMMAND_TRANSFERRING_ERROR, Response{})
		return
	}

	result, err := handler.node.GetResult(command, TRUST_LINE_RESULT_TIMEOUT)
	if err != nil {
		// Remote node is inaccessible
		logger.Error("Node is inaccessible during processing command: " +
			string(command.ToBytes()) + ". Details: " + err.Error())
		writeHTTPResponse(w, NODE_IS_INACCESSIBLE, Response{})
		return
	}

	if result.Code != OK && result.Code != ENGINE_NO_EQUIVALENT && result.Code != NODE_NOT_FOUND {
		logger.Error("Node return wrong command result: " + strconv.Itoa(result.Code) +
			" on command: " + string(command.ToBytes()))
		writeHTTPResponse(w, result.Code, Response{})
		return
	}
	if result.Code == ENGINE_NO_EQUIVALENT {
		logger.Info("Node hasn't equivalent for command: " + string(command.ToBytes()))
		writeHTTPResponse(w, result.Code, Response{})
		return
	}
	if result.Code == NODE_NOT_FOUND {
		logger.Info("Node hasn't TL with requested contractor for command " + string(command.ToBytes()))
		writeHTTPResponse(w, result.Code, Response{})
		return
	}

	if len(result.Tokens) < 4 {
		logger.Error("Node return invalid result tokens size on command: " + string(command.ToBytes()))
		writeHTTPResponse(w, ENGINE_UNEXPECTED_ERROR, Response{})
		return
	}

	// Contractors received well
	response := Response{TrustLine: TrustLine{
		ID:                    result.Tokens[0],
		State:                 result.Tokens[1],
		OwnKeysPresent:        result.Tokens[2],
		ContractorKeysPresent: result.Tokens[3],
		IncomingTrustAmount:   result.Tokens[4],
		OutgoingTrustAmount:   result.Tokens[5],
		Balance:               result.Tokens[6],
	}}
	writeHTTPResponse(w, OK, response)
}

func (handler *NodesHandler) GetTrustLineByAddress(w http.ResponseWriter, r *http.Request) {
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
	command := NewCommand(contractorAddresses...)

	type TrustLine struct {
		ID                    string `json:"id"`
		State                 string `json:"state"`
		OwnKeysPresent        string `json:"own_keys_present"`
		ContractorKeysPresent string `json:"contractor_keys_present"`
		IncomingTrustAmount   string `json:"incoming_trust_amount"`
		OutgoingTrustAmount   string `json:"outgoing_trust_amount"`
		Balance               string `json:"balance"`
	}

	type Response struct {
		TrustLine TrustLine `json:"trust_line"`
	}

	err = handler.node.SendCommand(command)
	if err != nil {
		logger.Error("Can't send command: " + string(command.ToBytes()) + " to node. Details: " + err.Error())
		writeHTTPResponse(w, COMMAND_TRANSFERRING_ERROR, Response{})
		return
	}

	result, err := handler.node.GetResult(command, TRUST_LINE_RESULT_TIMEOUT)
	if err != nil {
		// Remote node is inaccessible
		logger.Error("Node is inaccessible during processing command: " +
			string(command.ToBytes()) + ". Details: " + err.Error())
		writeHTTPResponse(w, NODE_IS_INACCESSIBLE, Response{})
		return
	}

	if result.Code != OK && result.Code != ENGINE_NO_EQUIVALENT && result.Code != NODE_NOT_FOUND {
		logger.Error("Node return wrong command result: " + strconv.Itoa(result.Code) +
			" on command: " + string(command.ToBytes()))
		writeHTTPResponse(w, result.Code, Response{})
		return
	}
	if result.Code == ENGINE_NO_EQUIVALENT {
		logger.Info("Node hasn't equivalent for command: " + string(command.ToBytes()))
		writeHTTPResponse(w, result.Code, Response{})
		return
	}
	if result.Code == NODE_NOT_FOUND {
		logger.Info("Node hasn't TL with requested contractor for command " + string(command.ToBytes()))
		writeHTTPResponse(w, result.Code, Response{})
		return
	}

	if len(result.Tokens) < 4 {
		logger.Error("Node return invalid result tokens size on command: " + string(command.ToBytes()))
		writeHTTPResponse(w, ENGINE_UNEXPECTED_ERROR, Response{})
		return
	}

	// Contractors received well
	response := Response{TrustLine: TrustLine{
		ID:                    result.Tokens[0],
		State:                 result.Tokens[1],
		OwnKeysPresent:        result.Tokens[2],
		ContractorKeysPresent: result.Tokens[3],
		IncomingTrustAmount:   result.Tokens[4],
		OutgoingTrustAmount:   result.Tokens[5],
		Balance:               result.Tokens[6],
	}}
	writeHTTPResponse(w, OK, response)
}

func (handler *NodesHandler) listEquivalents() {

	command := NewCommand("GET:equivalents")

	go handler.listEquivalentsGetResult(command)
}

func (handler *NodesHandler) listEquivalentsGetResult(command *Command) {
	type Response struct {
		Count       int      `json:"count"`
		Equivalents []string `json:"equivalents"`
	}

	err := handler.node.SendCommand(command)
	if err != nil {
		logger.Error("Can't send command: " + string(command.ToBytes()) + " to node. Details: " + err.Error())
		resultJSON := buildJSONResponse(COMMAND_TRANSFERRING_ERROR, Response{})
		fmt.Println(string(resultJSON))
		return
	}

	result, err := handler.node.GetResult(command, TRUST_LINE_RESULT_TIMEOUT)
	if err != nil {
		logger.Error("Node is inaccessible during processing command: " +
			string(command.ToBytes()) + ". Details: " + err.Error())
		resultJSON := buildJSONResponse(NODE_IS_INACCESSIBLE, Response{})
		fmt.Println(string(resultJSON))
		return
	}

	if result.Code != OK {
		logger.Error("Node return wrong command result: " + strconv.Itoa(result.Code) +
			" on command: " + string(command.ToBytes()))
		resultJSON := buildJSONResponse(result.Code, Response{})
		fmt.Println(string(resultJSON))
		return
	}

	if len(result.Tokens) == 0 {
		logger.Error("Node return invalid result tokens size on command: " + string(command.ToBytes()))
		resultJSON := buildJSONResponse(ENGINE_UNEXPECTED_ERROR, Response{})
		fmt.Println(string(resultJSON))
		return
	}

	// Equivalents received well
	equivalentsCount, err := strconv.Atoi(result.Tokens[0])
	if err != nil {
		logger.Error("Node return invalid token on command: " +
			string(command.ToBytes()) + ". Details: " + err.Error())
		resultJSON := buildJSONResponse(ENGINE_UNEXPECTED_ERROR, Response{})
		fmt.Println(string(resultJSON))
		return
	}

	if equivalentsCount == 0 {
		resultJSON := buildJSONResponse(OK, Response{Count: equivalentsCount})
		fmt.Println(string(resultJSON))
		return
	}

	response := Response{Count: equivalentsCount}
	for i := 0; i < equivalentsCount; i++ {
		response.Equivalents = append(response.Equivalents, result.Tokens[i+1])
	}
	resultJSON := buildJSONResponse(OK, response)
	fmt.Println(string(resultJSON))
}

func (handler *NodesHandler) ListEquivalents(w http.ResponseWriter, r *http.Request) {
	_, err := preprocessRequest(r)
	if err != nil {
		logger.Error("Bad request: invalid security parameters: " + err.Error())
		w.WriteHeader(BAD_REQUEST)
		return
	}

	command := NewCommand("GET:equivalents")

	type Response struct {
		Count       int      `json:"count"`
		Equivalents []string `json:"equivalents"`
	}

	err = handler.node.SendCommand(command)
	if err != nil {
		logger.Error("Can't send command: " + string(command.ToBytes()) + " to node. Details: " + err.Error())
		writeHTTPResponse(w, COMMAND_TRANSFERRING_ERROR, Response{})
		return
	}

	result, err := handler.node.GetResult(command, TRUST_LINE_RESULT_TIMEOUT)
	if err != nil {
		logger.Error("Node is inaccessible during processing command: " +
			string(command.ToBytes()) + ". Details: " + err.Error())
		writeHTTPResponse(w, NODE_IS_INACCESSIBLE, Response{})
		return
	}

	if result.Code != OK {
		logger.Error("Node return wrong command result: " + strconv.Itoa(result.Code) +
			" on command: " + string(command.ToBytes()))
		writeHTTPResponse(w, result.Code, Response{})
		return
	}

	if len(result.Tokens) == 0 {
		logger.Error("Node return invalid result tokens size on command: " + string(command.ToBytes()))
		writeHTTPResponse(w, ENGINE_UNEXPECTED_ERROR, Response{})
		return
	}

	// Equivalents received well
	equivalentsCount, err := strconv.Atoi(result.Tokens[0])
	if err != nil {
		logger.Error("Node return invalid token on command: " +
			string(command.ToBytes()) + ". Details: " + err.Error())
		writeHTTPResponse(w, ENGINE_UNEXPECTED_ERROR, Response{})
		return
	}

	if equivalentsCount == 0 {
		writeHTTPResponse(w, OK, Response{Count: equivalentsCount})
		return
	}

	response := Response{Count: equivalentsCount}
	for i := 0; i < equivalentsCount; i++ {
		response.Equivalents = append(response.Equivalents, result.Tokens[i+1])
	}
	writeHTTPResponse(w, OK, response)
}

func (handler *NodesHandler) totalBalance() {

	if !ValidateInt(Equivalent) {
		logger.Error("Bad request: invalid equivalent parameter in total-balance request")
		fmt.Println("Bad request: invalid equivalent parameter")
		return
	}

	command := NewCommand("GET:stats/balance/total", Equivalent)

	go handler.totalBalanceGetResult(command)
}

func (handler *NodesHandler) totalBalanceGetResult(command *Command) {

	type Response struct {
		TotalIncomingTrustAmount     string `json:"total_incoming_trust_amount"`
		TotalUsedIncomingTrustAmount string `json:"total_used_incoming_trust_amount"`
		TotalOutgoingTrustAmount     string `json:"total_outgoing_trust_amount"`
		TotalUsedOutgoingTrustAmount string `json:"total_used_outgoing_trust_amount"`
	}

	err := handler.node.SendCommand(command)
	if err != nil {
		logger.Error("Can't send command: " + string(command.ToBytes()) + " to node. Details: " + err.Error())
		resultJSON := buildJSONResponse(COMMAND_TRANSFERRING_ERROR, Response{})
		fmt.Println(string(resultJSON))
		return
	}

	result, err := handler.node.GetResult(command, STATS_RESULT_TIMEOUT)
	if err != nil {
		logger.Error("Node is inaccessible during processing command: " + string(command.ToBytes()) +
			". Details: " + err.Error())
		resultJSON := buildJSONResponse(NODE_IS_INACCESSIBLE, Response{})
		fmt.Println(string(resultJSON))
		return
	}

	if result.Code != OK && result.Code != ENGINE_NO_EQUIVALENT {
		logger.Error("Node return wrong command result: " + strconv.Itoa(result.Code) +
			" on command: " + string(command.ToBytes()))
		resultJSON := buildJSONResponse(result.Code, Response{})
		fmt.Println(string(resultJSON))
		return
	}
	if result.Code == ENGINE_NO_EQUIVALENT {
		logger.Info("Node hasn't equivalent for command: " + string(command.ToBytes()))
		resultJSON := buildJSONResponse(result.Code, Response{})
		fmt.Println(string(resultJSON))
	}

	if len(result.Tokens) < 4 {
		logger.Error("Node return invalid result tokens size on command: " + string(command.ToBytes()))
		resultJSON := buildJSONResponse(ENGINE_UNEXPECTED_ERROR, Response{})
		fmt.Println(string(resultJSON))
		return
	}

	// Contractors received well
	response := Response{
		TotalIncomingTrustAmount:     result.Tokens[0],
		TotalUsedIncomingTrustAmount: result.Tokens[1],
		TotalOutgoingTrustAmount:     result.Tokens[2],
		TotalUsedOutgoingTrustAmount: result.Tokens[3],
	}
	resultJSON := buildJSONResponse(OK, response)
	fmt.Println(string(resultJSON))
}

func (handler *NodesHandler) TotalBalance(w http.ResponseWriter, r *http.Request) {
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

	command := NewCommand("GET:stats/balance/total", equivalent)

	type Response struct {
		TotalIncomingTrustAmount     string `json:"total_incoming_trust_amount"`
		TotalUsedIncomingTrustAmount string `json:"total_used_incoming_trust_amount"`
		TotalOutgoingTrustAmount     string `json:"total_outgoing_trust_amount"`
		TotalUsedOutgoingTrustAmount string `json:"total_used_outgoing_trust_amount"`
	}

	err = handler.node.SendCommand(command)
	if err != nil {
		logger.Error("Can't send command: " + string(command.ToBytes()) + " to node. Details: " + err.Error())
		writeHTTPResponse(w, COMMAND_TRANSFERRING_ERROR, Response{})
		return
	}

	result, err := handler.node.GetResult(command, STATS_RESULT_TIMEOUT)
	if err != nil {
		logger.Error("Node is inaccessible during processing command: " + string(command.ToBytes()) +
			". Details: " + err.Error())
		writeHTTPResponse(w, NODE_IS_INACCESSIBLE, Response{})
		return
	}

	if result.Code != OK && result.Code != ENGINE_NO_EQUIVALENT {
		logger.Error("Node return wrong command result: " + strconv.Itoa(result.Code) +
			" on command: " + string(command.ToBytes()))
		writeHTTPResponse(w, result.Code, Response{})
		return
	}
	if result.Code == ENGINE_NO_EQUIVALENT {
		logger.Info("Node hasn't equivalent for command: " + string(command.ToBytes()))
		writeHTTPResponse(w, result.Code, Response{})
		return
	}

	if len(result.Tokens) < 4 {
		logger.Error("Node return invalid result tokens size on command: " + string(command.ToBytes()))
		writeHTTPResponse(w, ENGINE_UNEXPECTED_ERROR, Response{})
		return
	}

	// Contractors received well
	response := Response{
		TotalIncomingTrustAmount:     result.Tokens[0],
		TotalUsedIncomingTrustAmount: result.Tokens[1],
		TotalOutgoingTrustAmount:     result.Tokens[2],
		TotalUsedOutgoingTrustAmount: result.Tokens[3],
	}
	writeHTTPResponse(w, OK, response)
}
