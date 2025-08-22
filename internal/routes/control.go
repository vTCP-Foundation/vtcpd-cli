package routes

import (
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/vTCP-Foundation/vtcpd-cli/internal/common"
	"github.com/vTCP-Foundation/vtcpd-cli/internal/handler"
	"github.com/vTCP-Foundation/vtcpd-cli/internal/logger"
)

func (router *RoutesHandler) StopEverything(w http.ResponseWriter, r *http.Request) {
	_, err := preprocessRequest(r)
	if err != nil {
		logger.Error("Bad request: invalid security parameters: " + err.Error())
		w.WriteHeader(BAD_REQUEST)
		return
	}

	go func(nodesHandler *handler.NodeHandler) {
		err := nodesHandler.StopNode()
		if err != nil {
			logger.Error("Can't stop node " + err.Error())
			fmt.Println("Can't stop node " + err.Error())
		} else {
			logger.Info("Node stopped")
			fmt.Println("Stopped")
		}
		os.Exit(0)
	}(router.nodeHandler)

	writeHTTPResponse(w, OK, common.ControlMsgResponse{Status: "ok", Msg: "Stop request received"})
}

func (router *RoutesHandler) RemoveOutdatedCryptoData(w http.ResponseWriter, r *http.Request) {
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
	command := handler.NewCommand("DELETE:outdated-crypto", vacuum)

	err = router.nodeHandler.Node.SendCommand(command)
	if err != nil {
		logger.Error("Can't send command: " + string(command.ToBytes()) + " to node. Details: " + err.Error())
		writeHTTPResponse(w, COMMAND_TRANSFERRING_ERROR, common.ControlResponse{})
		return
	}

	result, err := router.nodeHandler.Node.GetResult(command, common.DELETE_CRYPTO_DATA_TIMEOUT)
	if err != nil {
		logger.Error("Node is inaccessible during processing command: " +
			string(command.ToBytes()) + ". Details: " + err.Error())
		writeHTTPResponse(w, NODE_IS_INACCESSIBLE, common.ControlResponse{})
		return
	}

	if result.Code != OK {
		logger.Error("Node return wrong command result: " + strconv.Itoa(result.Code) +
			" on command: " + string(command.ToBytes()))
	}

	writeHTTPResponse(w, result.Code, common.ControlResponse{})
}

func (router *RoutesHandler) RegenerateAllKeys(w http.ResponseWriter, r *http.Request) {
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

	command := handler.NewCommand("GET:equivalents")

	err = router.nodeHandler.Node.SendCommand(command)
	if err != nil {
		logger.Error("Can't send command: " + string(command.ToBytes()) + " to node. Details: " + err.Error())
		writeHTTPResponse(w, COMMAND_TRANSFERRING_ERROR, common.ControlResponse{})
		return
	}

	resultEquivalents, err := router.nodeHandler.Node.GetResult(command, common.CHANNEL_RESULT_TIMEOUT)
	if err != nil {
		logger.Error("Node is inaccessible during processing command: " +
			string(command.ToBytes()) + ". Details: " + err.Error())
		writeHTTPResponse(w, NODE_IS_INACCESSIBLE, common.ControlResponse{})
		return
	}

	if resultEquivalents.Code != OK {
		logger.Error("Node return wrong command result: " + strconv.Itoa(resultEquivalents.Code) +
			" on command: " + string(command.ToBytes()))
		writeHTTPResponse(w, resultEquivalents.Code, common.ControlResponse{})
		return
	}

	if len(resultEquivalents.Tokens) == 0 {
		logger.Error("Node return invalid result tokens size on command: " + string(command.ToBytes()))
		writeHTTPResponse(w, ENGINE_UNEXPECTED_ERROR, common.ControlResponse{})
		return
	}

	// Equivalents received well
	equivalentsCount, err := strconv.Atoi(resultEquivalents.Tokens[0])
	if err != nil {
		logger.Error("Node return invalid token on command: " +
			string(command.ToBytes()) + ". Details: " + err.Error())
		writeHTTPResponse(w, ENGINE_UNEXPECTED_ERROR, common.ControlResponse{})
		return
	}

	if equivalentsCount == 0 {
		logger.Info("There are no SL")
		writeHTTPResponse(w, resultEquivalents.Code, common.ControlResponse{})
		return
	}

	for i := range equivalentsCount {
		equivalents = append(equivalents, resultEquivalents.Tokens[i+1])
	}

	/////
	command = handler.NewCommand("GET:contractors-all")

	err = router.nodeHandler.Node.SendCommand(command)
	if err != nil {
		logger.Error("Can't send command: " + string(command.ToBytes()) + " to node. Details: " + err.Error())
		writeHTTPResponse(w, COMMAND_TRANSFERRING_ERROR, common.ControlResponse{})
		return
	}

	resultContractors, err := router.nodeHandler.Node.GetResult(command, common.CHANNEL_RESULT_TIMEOUT)
	if err != nil {
		logger.Error("Node is inaccessible during processing command: " +
			string(command.ToBytes()) + ". Details: " + err.Error())
		writeHTTPResponse(w, NODE_IS_INACCESSIBLE, common.ControlResponse{})
		return
	}

	if resultContractors.Code != OK {
		logger.Error("Node return wrong command result: " + strconv.Itoa(resultContractors.Code) +
			" on command: " + string(command.ToBytes()))
		writeHTTPResponse(w, resultContractors.Code, common.ControlResponse{})
		return
	}

	if len(resultContractors.Tokens) == 0 {
		logger.Error("Node return invalid result tokens size on command: " + string(command.ToBytes()))
		writeHTTPResponse(w, ENGINE_UNEXPECTED_ERROR, common.ControlResponse{})
		return
	}

	// Channels received well
	channelsCount, err := strconv.Atoi(resultContractors.Tokens[0])
	if err != nil {
		logger.Error("Node return invalid token on command: " +
			string(command.ToBytes()) + ". Details: " + err.Error())
		writeHTTPResponse(w, ENGINE_UNEXPECTED_ERROR, common.ControlResponse{})
		return
	}

	if channelsCount == 0 {
		logger.Info("There are no contractors")
		writeHTTPResponse(w, resultEquivalents.Code, common.ControlResponse{})
		return
	}

	for i := range channelsCount {
		contractors = append(contractors, resultContractors.Tokens[i*2+1])
	}
	/////

	go router.regenerateAllKeys(contractors, equivalents, delayInt)
	writeHTTPResponse(w, resultEquivalents.Code, common.ControlResponse{})
}

func (router *RoutesHandler) regenerateAllKeys(contractors []string, equivalents []string, delay int) {

	for _, contractor := range contractors {
		for _, equivalent := range equivalents {

			command := handler.NewCommand(
				"SET:contractors/trust-line-keys", contractor, equivalent)

			err := router.nodeHandler.Node.SendCommand(command)
			if err != nil {
				logger.Error("Can't send command: " + string(command.ToBytes()) + " to node. Details: " + err.Error())
				continue
			}

			result, err := router.nodeHandler.Node.GetResult(command, common.CHANNEL_RESULT_TIMEOUT)
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
