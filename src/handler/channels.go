package handler

import (
	"logger"
	"fmt"
	"strconv"
)

var (
	CHANNEL_RESULT_TIMEOUT uint16 = 20 // seconds
)

func (handler *NodesHandler) Channels() {

	if CommandType == "init" {
		handler.initChannel()

	} else {
		logger.Error("Invalid channel command " + CommandType)
		fmt.Println("Invalid channel command")
		return
	}
}

func (handler *NodesHandler)initChannel() {
	if len(Addresses) == 0 {
		logger.Error("Bad request: there are no contractor addresses parameters in channel init request")
		fmt.Println("Bad request: there are no contractor addresses parameters")
		return
	}

	var addresses []string
	for idx:= 0; idx < len(Addresses); idx++ {
		addressType, address := ValidateAddress(Addresses[idx])
		if addressType == "" {
			logger.Error("Bad request: invalid address parameter in channel init request")
			fmt.Println("Bad request: invalid address parameter")
			return
		}
		addresses = append(addresses, addressType, address)
	}

	addresses = append([]string{strconv.Itoa(len(Addresses))}, addresses...)
	addresses = append([]string{"INIT:contractors/channel"}, addresses...)
	if CryptoKey != "" {
		addresses = append(addresses, []string{CryptoKey}...)
	}
	command := NewCommand(addresses...)

	go handler.initChannelGetResult(command)
}

func (handler *NodesHandler)initChannelGetResult(command *Command) {

	type Response struct {
		ChannelID string `json:"channel_id"`
		CryptoKey string `json:"crypto_key"`
	}

	err := handler.node.SendCommand(command)
	if err != nil {
		logger.Error("Can't send command: " + string(command.ToBytes()) + " to node. Details: " + err.Error())
		resultJSON := buildJSONResponse(COMMAND_TRANSFERRING_ERROR, Response{})
		fmt.Println(string(resultJSON))
		return
	}

	result, err := handler.node.GetResult(command, CHANNEL_RESULT_TIMEOUT)
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

	if len(result.Tokens) < 2 {
		logger.Error("Node return invalid result tokens size on command: " + string(command.ToBytes()))
		resultJSON := buildJSONResponse(ENGINE_UNEXPECTED_ERROR, Response{})
		fmt.Println(string(resultJSON))
		return
	}
	resultJSON := buildJSONResponse(result.Code, Response{
		ChannelID: result.Tokens[0],
		CryptoKey: result.Tokens[1]})
	fmt.Println(string(resultJSON))
}
