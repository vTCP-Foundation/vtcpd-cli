package handler

import (
	"fmt"
	"strconv"

	"github.com/vTCP-Foundation/vtcpd-cli/internal/common"
	"github.com/vTCP-Foundation/vtcpd-cli/internal/logger"
)

func (handler *NodeHandler) Channels() {

	if CommandType == "init" {
		handler.initChannel()

	} else if CommandType == "get" {
		handler.listChannels()

	} else if CommandType == "one" {
		handler.channelInfo()

	} else if CommandType == "one-by-address" {
		handler.channelInfoByAddresses()

	} else if CommandType == "set-addresses" {
		handler.setChannelAddresses()

	} else if CommandType == "set-crypto-key" {
		handler.setChannelCryptoKey()

	} else if CommandType == "regenerate-crypto-key" {
		handler.regenerateChannelCryptoKey()

	} else if CommandType == "remove" {
		handler.removeChannel()

	} else {
		logger.Error("Invalid channel command " + CommandType)
		fmt.Println("Invalid channel command")
		return
	}
}

func (handler *NodeHandler) initChannel() {
	if len(Addresses) == 0 {
		logger.Error("Bad request: there are no contractor addresses parameters in channel init request")
		fmt.Println("Bad request: there are no contractor addresses parameters")
		return
	}

	var addresses []string
	for idx := range len(Addresses) {
		addressType, address := common.ValidateAddress(Addresses[idx])
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
		if !common.ValidateInt(ContractorID) {
			logger.Error("Bad request: invalid contractorID parameter in channel init request")
			fmt.Println("Bad request: invalid contractorID parameter")
			return
		}
		addresses = append(addresses, []string{ContractorID}...)
	}
	command := NewCommand(addresses...)

	go handler.initChannelGetResult(command)
}

func (handler *NodeHandler) initChannelGetResult(command *Command) {

	err := handler.Node.SendCommand(command)
	if err != nil {
		logger.Error("Can't send command: " + string(command.ToBytes()) + " to node. Details: " + err.Error())
		resultJSON := buildJSONResponse(COMMAND_TRANSFERRING_ERROR, common.ChannelInitResponse{})
		fmt.Println(string(resultJSON))
		return
	}

	result, err := handler.Node.GetResult(command, common.CHANNEL_RESULT_TIMEOUT)
	if err != nil {
		logger.Error("Node is inaccessible during processing command: " +
			string(command.ToBytes()) + ". Details: " + err.Error())
		resultJSON := buildJSONResponse(NODE_IS_INACCESSIBLE, common.ChannelInitResponse{})
		fmt.Println(string(resultJSON))
		return
	}

	if result.Code != OK {
		logger.Error("Node return wrong command result: " + strconv.Itoa(result.Code) +
			" on command: " + string(command.ToBytes()))
		resultJSON := buildJSONResponse(result.Code, common.ChannelInitResponse{})
		fmt.Println(string(resultJSON))
		return
	}

	if len(result.Tokens) < 2 {
		logger.Error("Node return invalid result tokens size on command: " + string(command.ToBytes()))
		resultJSON := buildJSONResponse(ENGINE_UNEXPECTED_ERROR, common.ChannelInitResponse{})
		fmt.Println(string(resultJSON))
		return
	}
	resultJSON := buildJSONResponse(result.Code, common.ChannelInitResponse{
		ChannelID: result.Tokens[0],
		CryptoKey: result.Tokens[1]})
	fmt.Println(string(resultJSON))
}

func (handler *NodeHandler) listChannels() {

	command := NewCommand("GET:contractors-all")

	go handler.listChannelsGetResult(command)
}

func (handler *NodeHandler) listChannelsGetResult(command *Command) {
	err := handler.Node.SendCommand(command)
	if err != nil {
		logger.Error("Can't send command: " + string(command.ToBytes()) + " to node. Details: " + err.Error())
		resultJSON := buildJSONResponse(COMMAND_TRANSFERRING_ERROR, common.ChannelListResponse{})
		fmt.Println(string(resultJSON))
		return
	}

	result, err := handler.Node.GetResult(command, common.CHANNEL_RESULT_TIMEOUT)
	if err != nil {
		logger.Error("Node is inaccessible during processing command: " +
			string(command.ToBytes()) + ". Details: " + err.Error())
		resultJSON := buildJSONResponse(NODE_IS_INACCESSIBLE, common.ChannelListResponse{})
		fmt.Println(string(resultJSON))
		return
	}

	if result.Code != OK {
		logger.Error("Node return wrong command result: " + strconv.Itoa(result.Code) +
			" on command: " + string(command.ToBytes()))
		resultJSON := buildJSONResponse(result.Code, common.ChannelListResponse{})
		fmt.Println(string(resultJSON))
		return
	}

	if len(result.Tokens) == 0 {
		logger.Error("Node return invalid result tokens size on command: " + string(command.ToBytes()))
		resultJSON := buildJSONResponse(ENGINE_UNEXPECTED_ERROR, common.ChannelListResponse{})
		fmt.Println(string(resultJSON))
		return
	}

	// Channels received well
	channelsCount, err := strconv.Atoi(result.Tokens[0])
	if err != nil {
		logger.Error("Node return invalid token on command: " +
			string(command.ToBytes()) + ". Details: " + err.Error())
		resultJSON := buildJSONResponse(ENGINE_UNEXPECTED_ERROR, common.ChannelListResponse{})
		fmt.Println(string(resultJSON))
		return
	}

	if channelsCount == 0 {
		resultJSON := buildJSONResponse(OK, common.ChannelListResponse{Count: channelsCount})
		fmt.Println(string(resultJSON))
		return
	}

	response := common.ChannelListResponse{Count: channelsCount}
	for i := range channelsCount {
		response.Channels = append(response.Channels, common.ChannelListItem{
			ID:        result.Tokens[i*2+1],
			Addresses: result.Tokens[i*2+2]})
	}
	resultJSON := buildJSONResponse(OK, response)
	fmt.Println(string(resultJSON))
}

func (handler *NodeHandler) channelInfo() {

	if !common.ValidateInt(ContractorID) {
		logger.Error("Bad request: invalid contractorID parameter in channels one request")
		fmt.Println("Bad request: invalid contractorID parameter")
		return
	}

	command := NewCommand("GET:channels/one", ContractorID)

	go handler.channelInfoGetResult(command)
}

func (handler *NodeHandler) channelInfoGetResult(command *Command) {
	err := handler.Node.SendCommand(command)
	if err != nil {
		logger.Error("Can't send command: " + string(command.ToBytes()) + " to node. Details: " + err.Error())
		resultJSON := buildJSONResponse(COMMAND_TRANSFERRING_ERROR, common.ChannelInfoResponse{})
		fmt.Println(string(resultJSON))
		return
	}

	result, err := handler.Node.GetResult(command, common.CHANNEL_RESULT_TIMEOUT)
	if err != nil {
		logger.Error("Node is inaccessible during processing command: " +
			string(command.ToBytes()) + ". Details: " + err.Error())
		resultJSON := buildJSONResponse(NODE_IS_INACCESSIBLE, common.ChannelInfoResponse{})
		fmt.Println(string(resultJSON))
		return
	}

	if result.Code != OK {
		logger.Error("Node return wrong command result: " + strconv.Itoa(result.Code) +
			" on command: " + string(command.ToBytes()))
		resultJSON := buildJSONResponse(result.Code, common.ChannelInfoResponse{})
		fmt.Println(string(resultJSON))
		return
	}

	if len(result.Tokens) < 2 {
		logger.Error("Node return invalid result tokens size on command: " + string(command.ToBytes()))
		resultJSON := buildJSONResponse(ENGINE_UNEXPECTED_ERROR, common.ChannelInfoResponse{})
		fmt.Println(string(resultJSON))
		return
	}

	addressesCount, err := strconv.Atoi(result.Tokens[1])
	if err != nil {
		logger.Error("Node return invalid token on command: " + string(command.ToBytes()) + ". Details: " + err.Error())
		resultJSON := buildJSONResponse(ENGINE_UNEXPECTED_ERROR, common.ChannelInfoResponse{})
		fmt.Println(string(resultJSON))
		return
	}

	if addressesCount == 0 {
		logger.Error("Node return invalid addresses token on command: " + string(command.ToBytes()))
		resultJSON := buildJSONResponse(ENGINE_UNEXPECTED_ERROR, common.ChannelInfoResponse{})
		fmt.Println(string(resultJSON))
		return
	}

	var addresses []string
	for idx := range addressesCount {
		addresses = append(addresses, result.Tokens[idx+2])
	}

	response := common.ChannelInfoResponse{
		ID:                  result.Tokens[0],
		Addresses:           addresses,
		IsConfirmed:         result.Tokens[2+addressesCount],
		CryptoKey:           result.Tokens[2+addressesCount+1],
		ContractorCryptoKey: result.Tokens[2+addressesCount+2]}
	resultJSON := buildJSONResponse(OK, response)
	fmt.Println(string(resultJSON))
}

func (handler *NodeHandler) channelInfoByAddresses() {

	if len(Addresses) == 0 {
		logger.Error("Bad request: there are no contractor addresses parameters in one-by-addresses request")
		fmt.Println("Bad request: there are no contractor addresses parameters")
		return
	}

	var addresses []string
	for idx := range len(addresses) {
		addressType, address := common.ValidateAddress(Addresses[idx])
		if addressType == "" {
			logger.Error("Bad request: invalid address parameter in one-by-addresses request")
			fmt.Println("Bad request: invalid address parameter")
			return
		}
		addresses = append(addresses, addressType, address)
	}

	addresses = append([]string{strconv.Itoa(len(Addresses))}, addresses...)
	addresses = append([]string{"GET:channels/one/address"}, addresses...)
	addresses = append(addresses, []string{Equivalent}...)
	command := NewCommand(addresses...)

	go handler.channelInfoByAddressesGetResult(command)
}

func (handler *NodeHandler) channelInfoByAddressesGetResult(command *Command) {
	err := handler.Node.SendCommand(command)
	if err != nil {
		logger.Error("Can't send command: " + string(command.ToBytes()) + " to node. Details: " + err.Error())
		resultJSON := buildJSONResponse(COMMAND_TRANSFERRING_ERROR, common.ChannelInfoByAddressResponse{})
		fmt.Println(string(resultJSON))
		return
	}

	result, err := handler.Node.GetResult(command, common.CHANNEL_RESULT_TIMEOUT)
	if err != nil {
		logger.Error("Node is inaccessible during processing command: " +
			string(command.ToBytes()) + ". Details: " + err.Error())
		resultJSON := buildJSONResponse(NODE_IS_INACCESSIBLE, common.ChannelInfoByAddressResponse{})
		fmt.Println(string(resultJSON))
		return
	}

	if result.Code != OK {
		logger.Error("Node return wrong command result: " + strconv.Itoa(result.Code) +
			" on command: " + string(command.ToBytes()))
		resultJSON := buildJSONResponse(result.Code, common.ChannelInfoByAddressResponse{})
		fmt.Println(string(resultJSON))
		return
	}

	if len(result.Tokens) < 2 {
		logger.Error("Node return invalid result tokens size on command: " + string(command.ToBytes()))
		resultJSON := buildJSONResponse(ENGINE_UNEXPECTED_ERROR, common.ChannelInfoByAddressResponse{})
		fmt.Println(string(resultJSON))
		return
	}

	response := common.ChannelInfoByAddressResponse{
		ID:          result.Tokens[0],
		IsConfirmed: result.Tokens[1]}
	resultJSON := buildJSONResponse(OK, response)
	fmt.Println(string(resultJSON))
}

func (handler *NodeHandler) setChannelAddresses() {

	if !common.ValidateInt(ContractorID) {
		logger.Error("Bad request: invalid contractorID parameter in channels set-addresses request")
		fmt.Println("Bad request: invalid contractorID parameter")
		return
	}

	if len(Addresses) == 0 {
		logger.Error("Bad request: there are no contractor addresses parameters in channel set-addresses request")
		fmt.Println("Bad request: there are no contractor addresses parameters")
		return
	}

	var addresses []string
	for idx := range len(Addresses) {
		addressType, address := common.ValidateAddress(Addresses[idx])
		if addressType == "" {
			logger.Error("Bad request: invalid address parameter in channel set-addresses request")
			fmt.Println("Bad request: invalid address parameter")
			return
		}
		addresses = append(addresses, addressType, address)
	}

	addresses = append([]string{"SET:channel/address", ContractorID, strconv.Itoa(len(Addresses))}, addresses...)
	command := NewCommand(addresses...)

	go handler.setChannelAddressesGetResult(command)
}

func (handler *NodeHandler) setChannelAddressesGetResult(command *Command) {
	err := handler.Node.SendCommand(command)
	if err != nil {
		logger.Error("Can't send command: " + string(command.ToBytes()) + " to node. Details: " + err.Error())
		resultJSON := buildJSONResponse(COMMAND_TRANSFERRING_ERROR, common.ChannelResponse{})
		fmt.Println(string(resultJSON))
		return
	}

	result, err := handler.Node.GetResult(command, common.CHANNEL_RESULT_TIMEOUT)
	if err != nil {
		logger.Error("Node is inaccessible during processing command: " +
			string(command.ToBytes()) + ". Details: " + err.Error())
		resultJSON := buildJSONResponse(NODE_IS_INACCESSIBLE, common.ChannelResponse{})
		fmt.Println(string(resultJSON))
		return
	}

	if result.Code != OK {
		logger.Error("Node return wrong command result: " + strconv.Itoa(result.Code) +
			" on command: " + string(command.ToBytes()))
	}

	resultJSON := buildJSONResponse(result.Code, common.ChannelResponse{})
	fmt.Println(string(resultJSON))
}

func (handler *NodeHandler) setChannelCryptoKey() {

	if !common.ValidateInt(ContractorID) {
		logger.Error("Bad request: invalid contractorID parameter in channels set-crypto-key request")
		fmt.Println("Bad request: invalid contractorID parameter")
		return
	}

	var commandParams []string
	commandParams = append(commandParams, "SET:channel/crypto-key", ContractorID, CryptoKey)
	if ChannelIDOnContractorSide != "" {
		if !common.ValidateInt(ChannelIDOnContractorSide) {
			logger.Error("Bad request: invalid channel-id-on-contractor-side parameter in channels set-crypto-key request")
			fmt.Println("Bad request: invalid channel-id-on-contractor-side parameter")
			return
		}
		commandParams = append(commandParams, []string{ChannelIDOnContractorSide}...)
	}

	command := NewCommand(commandParams...)

	go handler.setChannelCryptoKeyGetResult(command)
}

func (handler *NodeHandler) setChannelCryptoKeyGetResult(command *Command) {
	err := handler.Node.SendCommand(command)
	if err != nil {
		logger.Error("Can't send command: " + string(command.ToBytes()) + " to node. Details: " + err.Error())
		resultJSON := buildJSONResponse(COMMAND_TRANSFERRING_ERROR, common.ChannelResponse{})
		fmt.Println(string(resultJSON))
		return
	}

	result, err := handler.Node.GetResult(command, common.CHANNEL_RESULT_TIMEOUT)
	if err != nil {
		logger.Error("Node is inaccessible during processing command: " +
			string(command.ToBytes()) + ". Details: " + err.Error())
		resultJSON := buildJSONResponse(NODE_IS_INACCESSIBLE, common.ChannelResponse{})
		fmt.Println(string(resultJSON))
		return
	}

	if result.Code != OK {
		logger.Error("Node return wrong command result: " + strconv.Itoa(result.Code) +
			" on command: " + string(command.ToBytes()))
	}

	resultJSON := buildJSONResponse(result.Code, common.ChannelResponse{})
	fmt.Println(string(resultJSON))
}

func (handler *NodeHandler) regenerateChannelCryptoKey() {

	if !common.ValidateInt(ContractorID) {
		logger.Error("Bad request: invalid contractor_id parameter in channels regenerate-crypto-key request")
		fmt.Println("Bad request: invalid contractor_id parameter")
		return
	}

	command := NewCommand("SET:channel/regenerate-crypto-key", ContractorID)

	go handler.regenerateChannelCryptoKeyGetResult(command)
}

func (handler *NodeHandler) regenerateChannelCryptoKeyGetResult(command *Command) {
	err := handler.Node.SendCommand(command)
	if err != nil {
		logger.Error("Can't send command: " + string(command.ToBytes()) + " to node. Details: " + err.Error())
		resultJSON := buildJSONResponse(COMMAND_TRANSFERRING_ERROR, common.ChannelInitResponse{})
		fmt.Println(string(resultJSON))
		return
	}

	result, err := handler.Node.GetResult(command, common.CHANNEL_RESULT_TIMEOUT)
	if err != nil {
		logger.Error("Node is inaccessible during processing command: " +
			string(command.ToBytes()) + ". Details: " + err.Error())
		resultJSON := buildJSONResponse(NODE_IS_INACCESSIBLE, common.ChannelInitResponse{})
		fmt.Println(string(resultJSON))
		return
	}

	if result.Code != OK {
		logger.Error("Node return wrong command result: " + strconv.Itoa(result.Code) +
			" on command: " + string(command.ToBytes()))
		resultJSON := buildJSONResponse(result.Code, common.ChannelInitResponse{})
		fmt.Println(string(resultJSON))
		return
	}

	if len(result.Tokens) < 2 {
		logger.Error("Node return invalid result tokens size on command: " + string(command.ToBytes()))
		resultJSON := buildJSONResponse(ENGINE_UNEXPECTED_ERROR, common.ChannelInitResponse{})
		fmt.Println(string(resultJSON))
		return
	}
	resultJSON := buildJSONResponse(result.Code, common.ChannelInitResponse{
		ChannelID: result.Tokens[0],
		CryptoKey: result.Tokens[1]})
	fmt.Println(string(resultJSON))
}

func (handler *NodeHandler) removeChannel() {

	if !common.ValidateInt(ContractorID) {
		logger.Error("Bad request: invalid contractor_id parameter in channels remove request")
		fmt.Println("Bad request: invalid contractor_id parameter")
		return
	}

	command := NewCommand("DELETE:channel/contractor-id", ContractorID)

	go handler.removeChannelGetResult(command)
}

func (handler *NodeHandler) removeChannelGetResult(command *Command) {
	err := handler.Node.SendCommand(command)
	if err != nil {
		logger.Error("Can't send command: " + string(command.ToBytes()) + " to node. Details: " + err.Error())
		resultJSON := buildJSONResponse(COMMAND_TRANSFERRING_ERROR, common.ChannelResponse{})
		fmt.Println(string(resultJSON))
		return
	}

	result, err := handler.Node.GetResult(command, common.CHANNEL_RESULT_TIMEOUT)
	if err != nil {
		logger.Error("Node is inaccessible during processing command: " +
			string(command.ToBytes()) + ". Details: " + err.Error())
		resultJSON := buildJSONResponse(NODE_IS_INACCESSIBLE, common.ChannelResponse{})
		fmt.Println(string(resultJSON))
		return
	}

	if result.Code != OK {
		logger.Error("Node return wrong command result: " + strconv.Itoa(result.Code) +
			" on command: " + string(command.ToBytes()))
	}

	resultJSON := buildJSONResponse(result.Code, common.ChannelResponse{})
	fmt.Println(string(resultJSON))
}
