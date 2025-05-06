package handler

import (
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/vTCP-Foundation/vtcpd-cli/src/logger"
)

type ControlMsgResponse struct {
	Status string `json:"status"`
	Msg    string `json:"msg"`
}

type ControlResponse struct{}

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

	writeHTTPResponse(w, OK, ControlMsgResponse{"ok", "Stop request received"})
}

func (handler *NodesHandler) RemoveOutdatedCryptoDataCommand() {
	// Command generation
	command := NewCommand("DELETE:outdated-crypto", "1")

	err := handler.node.SendCommand(command)
	if err != nil {
		logger.Error("Can't send command: " + string(command.ToBytes()) + " to node. Details: " + err.Error())
		resultJSON := buildJSONResponse(COMMAND_TRANSFERRING_ERROR, ControlResponse{})
		fmt.Println(string(resultJSON))
		return
	}

	result, err := handler.node.GetResult(command, DELETE_CRYPTO_DATA_TIMEOUT)
	if err != nil {
		logger.Error("Node is inaccessible during processing command: " +
			string(command.ToBytes()) + ". Details: " + err.Error())
		resultJSON := buildJSONResponse(NODE_IS_INACCESSIBLE, ControlResponse{})
		fmt.Println(string(resultJSON))
		return
	}

	if result.Code != OK {
		logger.Error("Node return wrong command result: " + strconv.Itoa(result.Code) +
			" on command: " + string(command.ToBytes()))
	}

	resultJSON := buildJSONResponse(result.Code, ControlResponse{})
	fmt.Println(string(resultJSON))
}

func (handler *NodesHandler) RemoveOutdatedCryptoData(w http.ResponseWriter, r *http.Request) {
	_, err := preprocessRequest(r)
	if err != nil {
		logger.Error("Bad request: invalid security parameters: " + err.Error())
		w.WriteHeader(BAD_REQUEST)
		return
	}

	vacuum := r.URL.Query().Get("vacuum")
	if vacuum == "" {
		vacuum = "0"
	}

	// Command generation
	command := NewCommand("DELETE:outdated-crypto", vacuum)

	err = handler.node.SendCommand(command)
	if err != nil {
		logger.Error("Can't send command: " + string(command.ToBytes()) + " to node. Details: " + err.Error())
		writeHTTPResponse(w, COMMAND_TRANSFERRING_ERROR, ControlResponse{})
		return
	}

	result, err := handler.node.GetResult(command, DELETE_CRYPTO_DATA_TIMEOUT)
	if err != nil {
		logger.Error("Node is inaccessible during processing command: " +
			string(command.ToBytes()) + ". Details: " + err.Error())
		writeHTTPResponse(w, NODE_IS_INACCESSIBLE, ControlResponse{})
		return
	}

	if result.Code != OK {
		logger.Error("Node return wrong command result: " + strconv.Itoa(result.Code) +
			" on command: " + string(command.ToBytes()))
	}

	writeHTTPResponse(w, result.Code, ControlResponse{})
}

func (handler *NodesHandler) RegenerateAllKeys(w http.ResponseWriter, r *http.Request) {
	_, err := preprocessRequest(r)
	if err != nil {
		logger.Error("Bad request: invalid security parameters: " + err.Error())
		w.WriteHeader(BAD_REQUEST)
		return
	}

	delayInt := 0

	delay := r.URL.Query().Get("delay")
	if delay == "" {
		delayInt = 5
	} else {
		delayInt, err = strconv.Atoi(delay)
		if err != nil {
			logger.Error("Bad request: invalid delay parameters: " + err.Error())
			w.WriteHeader(BAD_REQUEST)
			return
		}
	}

	var equivalents []string
	var contractors []string

	command := NewCommand("GET:equivalents")

	err = handler.node.SendCommand(command)
	if err != nil {
		logger.Error("Can't send command: " + string(command.ToBytes()) + " to node. Details: " + err.Error())
		writeHTTPResponse(w, COMMAND_TRANSFERRING_ERROR, ControlResponse{})
		return
	}

	resultEquivalents, err := handler.node.GetResult(command, SETTLEMENT_LINE_RESULT_TIMEOUT)
	if err != nil {
		logger.Error("Node is inaccessible during processing command: " +
			string(command.ToBytes()) + ". Details: " + err.Error())
		writeHTTPResponse(w, NODE_IS_INACCESSIBLE, ControlResponse{})
		return
	}

	if resultEquivalents.Code != OK {
		logger.Error("Node return wrong command result: " + strconv.Itoa(resultEquivalents.Code) +
			" on command: " + string(command.ToBytes()))
		writeHTTPResponse(w, resultEquivalents.Code, ControlResponse{})
		return
	}

	if len(resultEquivalents.Tokens) == 0 {
		logger.Error("Node return invalid result tokens size on command: " + string(command.ToBytes()))
		writeHTTPResponse(w, ENGINE_UNEXPECTED_ERROR, ControlResponse{})
		return
	}

	// Equivalents received well
	equivalentsCount, err := strconv.Atoi(resultEquivalents.Tokens[0])
	if err != nil {
		logger.Error("Node return invalid token on command: " +
			string(command.ToBytes()) + ". Details: " + err.Error())
		writeHTTPResponse(w, ENGINE_UNEXPECTED_ERROR, ControlResponse{})
		return
	}

	if equivalentsCount == 0 {
		logger.Info("There are no TL")
		writeHTTPResponse(w, resultEquivalents.Code, ControlResponse{})
		return
	}

	for i := 0; i < equivalentsCount; i++ {
		equivalents = append(equivalents, resultEquivalents.Tokens[i+1])
	}

	/////
	command = NewCommand("GET:contractors-all")

	err = handler.node.SendCommand(command)
	if err != nil {
		logger.Error("Can't send command: " + string(command.ToBytes()) + " to node. Details: " + err.Error())
		writeHTTPResponse(w, COMMAND_TRANSFERRING_ERROR, ControlResponse{})
		return
	}

	resultContractors, err := handler.node.GetResult(command, CHANNEL_RESULT_TIMEOUT)
	if err != nil {
		logger.Error("Node is inaccessible during processing command: " +
			string(command.ToBytes()) + ". Details: " + err.Error())
		writeHTTPResponse(w, NODE_IS_INACCESSIBLE, ControlResponse{})
		return
	}

	if resultContractors.Code != OK {
		logger.Error("Node return wrong command result: " + strconv.Itoa(resultContractors.Code) +
			" on command: " + string(command.ToBytes()))
		writeHTTPResponse(w, resultContractors.Code, ControlResponse{})
		return
	}

	if len(resultContractors.Tokens) == 0 {
		logger.Error("Node return invalid result tokens size on command: " + string(command.ToBytes()))
		writeHTTPResponse(w, ENGINE_UNEXPECTED_ERROR, ControlResponse{})
		return
	}

	// Channels received well
	channelsCount, err := strconv.Atoi(resultContractors.Tokens[0])
	if err != nil {
		logger.Error("Node return invalid token on command: " +
			string(command.ToBytes()) + ". Details: " + err.Error())
		writeHTTPResponse(w, ENGINE_UNEXPECTED_ERROR, ControlResponse{})
		return
	}

	if channelsCount == 0 {
		logger.Info("There are no contractors")
		writeHTTPResponse(w, resultEquivalents.Code, ControlResponse{})
		return
	}

	for i := 0; i < channelsCount; i++ {
		contractors = append(contractors, resultContractors.Tokens[i*2+1])
	}
	/////

	go handler.regenerateAllKeys(contractors, equivalents, delayInt)
	writeHTTPResponse(w, resultEquivalents.Code, ControlResponse{})
}

func (handler *NodesHandler) regenerateAllKeys(contractors []string, equivalents []string, delay int) {

	for _, contractor := range contractors {
		for _, equivalent := range equivalents {

			command := NewCommand(
				"SET:contractors/trust-line-keys", contractor, equivalent)

			err := handler.node.SendCommand(command)
			if err != nil {
				logger.Error("Can't send command: " + string(command.ToBytes()) + " to node. Details: " + err.Error())
				continue
			}

			result, err := handler.node.GetResult(command, SETTLEMENT_LINE_RESULT_TIMEOUT)
			if err != nil {
				logger.Error("Node is inaccessible during processing command: " +
					string(command.ToBytes()) + ". Details: " + err.Error())
				continue
			}

			if result.Code != OK {
				logger.Error("Node return wrong command result: " + strconv.Itoa(result.Code) +
					" on command: " + string(command.ToBytes()))
			}
			time.Sleep(time.Second * time.Duration(delay))
		}
	}
}
