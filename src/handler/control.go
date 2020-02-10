package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"

	"logger"
)

var (
	DELETE_CRYPTO_DATA_TIMEOUT uint16 = 20 // seconds
)

func (handler *NodesHandler) StopEverything(w http.ResponseWriter, r *http.Request) {
	_, err := preprocessRequest(r)
	if err != nil {
		logger.Error("Bad request: invalid security parameters: " + err.Error())
		w.WriteHeader(BAD_REQUEST)
		return
	}

	type Response struct {
		Status string `json:"status"`
		Msg    string `json:"msg"`
	}

	go func(nodesHandler *NodesHandler) {
		err := nodesHandler.StopNode()
		if err != nil {
			logger.Error("Can't stop node " + err.Error())
			fmt.Println("Can't stop node " + err.Error())
		} else {
			logger.Info("Node stopped")
			fmt.Println("Stopped")
		}
		err = nodesHandler.StopEventsMonitoring()
		if err != nil {
			logger.Error("Can't stop events-monitor. Details: " + err.Error())
		} else {
			logger.Info("Events-monitor stopped")
		}
		nodesHandler.ClearEventsMonitoringPID()
		os.Exit(0)
	}(handler)

	writeHTTPResponse(w, OK, Response{"ok", "Stop request received"})
}

func (handler *NodesHandler) SaveTrustLineInfo(w http.ResponseWriter, r *http.Request) {
	_, err := preprocessRequest(r)
	if err != nil {
		logger.Error("Bad request: invalid security parameters: " + err.Error())
		w.WriteHeader(BAD_REQUEST)
		return
	}

	type Response struct {
		Status string `json:"status"`
		Msg    string `json:"msg"`
	}

	flag := os.O_TRUNC | os.O_WRONLY | os.O_CREATE

	filename := "nodeTL.json"
	jsonFile, err := os.OpenFile(filename, flag, 0600)
	if err != nil {
		logger.Error("Can't open file for data")
		return
	}
	defer jsonFile.Close()

	handler.SaveTrustLineEquivalentInfo("101", jsonFile)
	handler.SaveTrustLineEquivalentInfo("104", jsonFile)
	handler.SaveTrustLineEquivalentInfo("1001", jsonFile)
	handler.SaveTrustLineEquivalentInfo("1002", jsonFile)
	handler.SaveTrustLineEquivalentInfo("2001", jsonFile)
	handler.SaveTrustLineEquivalentInfo("2002", jsonFile)

	writeHTTPResponse(w, OK, Response{"ok", "Data saved"})
}

func (handler *NodesHandler) SaveTrustLineEquivalentInfo(equivalent string, file *os.File) {
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

	type Result struct {
		Equivalent string      `json:"equivalent"`
		Count      int         `json:"count"`
		TrustLines []TrustLine `json:"trust_lines"`
	}

	err := handler.node.SendCommand(command)
	if err != nil {
		logger.Error("Can't send command: " + string(command.ToBytes()) + " to node. Details: " + err.Error())
		return
	}

	result, err := handler.node.GetResult(command, TRUST_LINE_RESULT_TIMEOUT)
	if err != nil {
		logger.Error("Node is inaccessible during processing command: " +
			string(command.ToBytes()) + ". Details: " + err.Error())
		return
	}

	if result.Code != OK && result.Code != ENGINE_NO_EQUIVALENT {
		logger.Error("Node return wrong command result: " + strconv.Itoa(result.Code) +
			" on command: " + string(command.ToBytes()))
		return
	}
	if result.Code == ENGINE_NO_EQUIVALENT {
		logger.Info("Node hasn't equivalent for command: " + string(command.ToBytes()))
		return
	}

	if len(result.Tokens) == 0 {
		logger.Error("Node return invalid result tokens size on command: " + string(command.ToBytes()))
		return
	}

	// Contractors received well
	contractorsCount, err := strconv.Atoi(result.Tokens[0])
	if err != nil {
		logger.Error("Node return invalid token on command: " +
			string(command.ToBytes()) + ". Details: " + err.Error())
		return
	}

	resultTR := Result{Equivalent: equivalent, Count: contractorsCount}
	for i := 0; i < contractorsCount; i++ {
		resultTR.TrustLines = append(resultTR.TrustLines, TrustLine{
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
	out, err := json.MarshalIndent(resultTR, "", "\t")
	if err != nil {
		logger.Error("Can't marshal data. Details: " + err.Error())
		return
	}
	symbolsWritten, err := file.WriteString(string(out))
	if symbolsWritten == 0 || err != nil {
		if err != nil {
			logger.Error("File logger: can't write nodeTL.json record. Details: " + err.Error())
		} else {
			logger.Error("File logger: can't write nodeTL.json record.")
		}
		return
	}
}

func (handler *NodesHandler) RemoveOutdatedCryptoDataCommand() {
	// Command generation
	command := NewCommand("DELETE:outdated-crypto")

	type Response struct{}

	err := handler.node.SendCommand(command)
	if err != nil {
		logger.Error("Can't send command: " + string(command.ToBytes()) + " to node. Details: " + err.Error())
		resultJSON := buildJSONResponse(COMMAND_TRANSFERRING_ERROR, Response{})
		fmt.Println(string(resultJSON))
		return
	}

	result, err := handler.node.GetResult(command, DELETE_CRYPTO_DATA_TIMEOUT)
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

func (handler *NodesHandler) RemoveOutdatedCryptoData(w http.ResponseWriter, r *http.Request) {
	_, err := preprocessRequest(r)
	if err != nil {
		logger.Error("Bad request: invalid security parameters: " + err.Error())
		w.WriteHeader(BAD_REQUEST)
		return
	}

	// Command generation
	command := NewCommand("DELETE:outdated-crypto")

	type Response struct{}

	err = handler.node.SendCommand(command)
	if err != nil {
		logger.Error("Can't send command: " + string(command.ToBytes()) + " to node. Details: " + err.Error())
		writeHTTPResponse(w, COMMAND_TRANSFERRING_ERROR, Response{})
		return
	}

	result, err := handler.node.GetResult(command, DELETE_CRYPTO_DATA_TIMEOUT)
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
