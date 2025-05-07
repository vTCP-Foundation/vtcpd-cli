package handler

import (
	"fmt"
	"strconv"

	"github.com/vTCP-Foundation/vtcpd-cli/internal/logger"
)

type ControlMsgResponse struct {
	Status string `json:"status"`
	Msg    string `json:"msg"`
}

type ControlResponse struct{}

var (
	DELETE_CRYPTO_DATA_TIMEOUT uint16 = 20 // seconds
)

func (handler *NodeHandler) RemoveOutdatedCryptoDataCommand() {
	// Command generation
	command := NewCommand("DELETE:outdated-crypto", "1")

	err := handler.Node.SendCommand(command)
	if err != nil {
		logger.Error("Can't send command: " + string(command.ToBytes()) + " to node. Details: " + err.Error())
		resultJSON := buildJSONResponse(COMMAND_TRANSFERRING_ERROR, ControlResponse{})
		fmt.Println(string(resultJSON))
		return
	}

	result, err := handler.Node.GetResult(command, DELETE_CRYPTO_DATA_TIMEOUT)
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
