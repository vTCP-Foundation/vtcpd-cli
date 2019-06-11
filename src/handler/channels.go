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
	CHANNEL_RESULT_TIMEOUT uint16 = 20 // seconds
)

func (handler *NodesHandler) Channels() {

	if CommandType == "init" {
		handler.initChannel()

	} else if CommandType == "get" {
		handler.listChannels()

	} else if CommandType == "one" {
		handler.channelInfo()

	} else if CommandType == "set-addresses" {
		handler.setChannelAddresses()

	} else if CommandType == "set-crypto-key" {
		handler.setChannelCryptoKey()

	} else if CommandType == "regenerate-crypto-key" {
		handler.regenerateChannelCryptoKey()

	} else {
		logger.Error("Invalid channel command " + CommandType)
		fmt.Println("Invalid channel command")
		return
	}
}

func (handler *NodesHandler) initChannel() {
	if len(Addresses) == 0 {
		logger.Error("Bad request: there are no contractor addresses parameters in channel init request")
		fmt.Println("Bad request: there are no contractor addresses parameters")
		return
	}

	var addresses []string
	for idx := 0; idx < len(Addresses); idx++ {
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
		if !ValidateInt(ContractorID) {
			logger.Error("Bad request: invalid contractorID parameter in channel init request")
			fmt.Println("Bad request: invalid contractorID parameter")
			return
		}
		addresses = append(addresses, []string{ContractorID}...)
	}
	command := NewCommand(addresses...)

	go handler.initChannelGetResult(command)
}

func (handler *NodesHandler) initChannelGetResult(command *Command) {

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

func (handler *NodesHandler) InitChannel(w http.ResponseWriter, r *http.Request) {
	url := logRequest(r)

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

	cryptoKey := r.FormValue("crypto_key")

	// Command generation
	contractorAddresses = append([]string{strconv.Itoa(len(contractorAddresses) / 2)}, contractorAddresses...)
	contractorAddresses = append([]string{"INIT:contractors/channel"}, contractorAddresses...)
	if cryptoKey != "" {
		contractorChannelID := r.FormValue("contractor_id")
		if !ValidateInt(contractorChannelID) {
			logger.Error("Bad request: there are no contractor_id parameter: " + url)
			w.WriteHeader(BAD_REQUEST)
			return
		}
		contractorAddresses = append(contractorAddresses, []string{cryptoKey, contractorChannelID}...)
	}
	command := NewCommand(contractorAddresses...)

	type Response struct {
		ChannelID string `json:"channel_id"`
		CryptoKey string `json:"crypto_key"`
	}

	err := handler.node.SendCommand(command)
	if err != nil {
		logger.Error("Can't send command: " + string(command.ToBytes()) + " to node. Details: " + err.Error())
		writeHTTPResponse(w, COMMAND_TRANSFERRING_ERROR, Response{})
		return
	}

	result, err := handler.node.GetResult(command, CHANNEL_RESULT_TIMEOUT)
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

	if len(result.Tokens) < 2 {
		logger.Error("Node return invalid result tokens size on command: " + string(command.ToBytes()))
		writeHTTPResponse(w, ENGINE_UNEXPECTED_ERROR, Response{})
		return
	}

	writeHTTPResponse(w, OK, Response{
		ChannelID: result.Tokens[0],
		CryptoKey: result.Tokens[1]})
}

func (handler *NodesHandler) listChannels() {

	command := NewCommand("GET:contractors-all")

	go handler.listChannelsGetResult(command)
}

func (handler *NodesHandler) listChannelsGetResult(command *Command) {
	type Channel struct {
		ID        string `json:"channel_id"`
		Addresses string `json:"channel_addresses"`
	}

	type Response struct {
		Count    int       `json:"count"`
		Channels []Channel `json:"channels"`
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

	if len(result.Tokens) == 0 {
		logger.Error("Node return invalid result tokens size on command: " + string(command.ToBytes()))
		resultJSON := buildJSONResponse(ENGINE_UNEXPECTED_ERROR, Response{})
		fmt.Println(string(resultJSON))
		return
	}

	// Channels received well
	channelsCount, err := strconv.Atoi(result.Tokens[0])
	if err != nil {
		logger.Error("Node return invalid token on command: " +
			string(command.ToBytes()) + ". Details: " + err.Error())
		resultJSON := buildJSONResponse(ENGINE_UNEXPECTED_ERROR, Response{})
		fmt.Println(string(resultJSON))
		return
	}

	if channelsCount == 0 {
		resultJSON := buildJSONResponse(OK, Response{Count: channelsCount})
		fmt.Println(string(resultJSON))
		return
	}

	response := Response{Count: channelsCount}
	for i := 0; i < channelsCount; i++ {
		response.Channels = append(response.Channels, Channel{
			ID:        result.Tokens[i*2+1],
			Addresses: result.Tokens[i*2+2]})
	}
	resultJSON := buildJSONResponse(OK, response)
	fmt.Println(string(resultJSON))
}

func (handler *NodesHandler) ListChannels(w http.ResponseWriter, r *http.Request) {
	logRequest(r)

	command := NewCommand("GET:contractors-all")

	type Channel struct {
		ID        string `json:"channel_id"`
		Addresses string `json:"channel_addresses"`
	}

	type Response struct {
		Count    int       `json:"count"`
		Channels []Channel `json:"channels"`
	}

	err := handler.node.SendCommand(command)
	if err != nil {
		logger.Error("Can't send command: " + string(command.ToBytes()) + " to node. Details: " + err.Error())
		writeHTTPResponse(w, COMMAND_TRANSFERRING_ERROR, Response{})
		return
	}

	result, err := handler.node.GetResult(command, CHANNEL_RESULT_TIMEOUT)
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

	// Channels received well
	channelsCount, err := strconv.Atoi(result.Tokens[0])
	if err != nil {
		logger.Error("Node return invalid token on command: " +
			string(command.ToBytes()) + ". Details: " + err.Error())
		writeHTTPResponse(w, ENGINE_UNEXPECTED_ERROR, Response{})
		return
	}

	if channelsCount == 0 {
		writeHTTPResponse(w, OK, Response{Count: channelsCount})
		return
	}

	response := Response{Count: channelsCount}
	for i := 0; i < channelsCount; i++ {
		response.Channels = append(response.Channels, Channel{
			ID:        result.Tokens[i*2+1],
			Addresses: result.Tokens[i*2+2]})
	}
	writeHTTPResponse(w, OK, response)
}

func (handler *NodesHandler) channelInfo() {

	if !ValidateInt(ContractorID) {
		logger.Error("Bad request: invalid contractorID parameter in channels one request")
		fmt.Println("Bad request: invalid contractorID parameter")
		return
	}

	command := NewCommand("GET:channels/one", ContractorID)

	go handler.channelInfoGetResult(command)
}

func (handler *NodesHandler) channelInfoGetResult(command *Command) {
	type Response struct {
		ID                  string   `json:"channel_id"`
		Addresses           []string `json:"channel_addresses"`
		IsConfirmed         string   `json:"channel_confirmed"`
		CryptoKey           string   `json:"channel_crypto_key"`
		ContractorCryptoKey string   `json:"channel_contractor_crypto_key"`
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

	addressesCount, err := strconv.Atoi(result.Tokens[1])
	if err != nil {
		logger.Error("Node return invalid token on command: " + string(command.ToBytes()) + ". Details: " + err.Error())
		resultJSON := buildJSONResponse(ENGINE_UNEXPECTED_ERROR, Response{})
		fmt.Println(string(resultJSON))
		return
	}

	if addressesCount == 0 {
		logger.Error("Node return invalid addresses token on command: " + string(command.ToBytes()))
		resultJSON := buildJSONResponse(ENGINE_UNEXPECTED_ERROR, Response{})
		fmt.Println(string(resultJSON))
		return
	}

	var addresses []string
	for idx := 0; idx < addressesCount; idx++ {
		addresses = append(addresses, result.Tokens[idx+2])
	}

	response := Response{
		ID:                  result.Tokens[0],
		Addresses:           addresses,
		IsConfirmed:         result.Tokens[2+addressesCount],
		CryptoKey:           result.Tokens[2+addressesCount+1],
		ContractorCryptoKey: result.Tokens[2+addressesCount+2]}
	resultJSON := buildJSONResponse(OK, response)
	fmt.Println(string(resultJSON))
}

func (handler *NodesHandler) ChannelInfo(w http.ResponseWriter, r *http.Request) {
	url := logRequest(r)

	contractorID, isParamPresent := mux.Vars(r)["contractor_id"]
	if !isParamPresent {
		logger.Error("Bad request: missing contractor_id parameter: " + url)
		w.WriteHeader(BAD_REQUEST)
		return
	}

	if !ValidateInt(contractorID) {
		logger.Error("Bad request: invalid contractor_id parameter")
		w.WriteHeader(BAD_REQUEST)
		return
	}

	command := NewCommand("GET:channels/one", contractorID)

	type Response struct {
		ID                  string   `json:"channel_id"`
		Addresses           []string `json:"channel_addresses"`
		IsConfirmed         string   `json:"channel_confirmed"`
		CryptoKey           string   `json:"channel_crypto_key"`
		ContractorCryptoKey string   `json:"channel_contractor_crypto_key"`
	}

	err := handler.node.SendCommand(command)
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

	if len(result.Tokens) < 2 {
		logger.Error("Node return invalid result tokens size on command: " + string(command.ToBytes()))
		writeHTTPResponse(w, ENGINE_UNEXPECTED_ERROR, Response{})
		return
	}

	addressesCount, err := strconv.Atoi(result.Tokens[1])
	if err != nil {
		logger.Error("Node return invalid token on command: " + string(command.ToBytes()) + ". Details: " + err.Error())
		writeHTTPResponse(w, ENGINE_UNEXPECTED_ERROR, Response{})
		return
	}

	if addressesCount == 0 {
		logger.Error("Node return invalid addresses token on command: " + string(command.ToBytes()))
		writeHTTPResponse(w, ENGINE_UNEXPECTED_ERROR, Response{})
		return
	}

	var addresses []string
	for idx := 0; idx < addressesCount; idx++ {
		addresses = append(addresses, result.Tokens[idx+2])
	}

	response := Response{
		ID:                  result.Tokens[0],
		Addresses:           addresses,
		IsConfirmed:         result.Tokens[2+addressesCount],
		CryptoKey:           result.Tokens[2+addressesCount+1],
		ContractorCryptoKey: result.Tokens[2+addressesCount+2]}
	writeHTTPResponse(w, OK, response)
}

func (handler *NodesHandler) setChannelAddresses() {

	if !ValidateInt(ContractorID) {
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
	for idx := 0; idx < len(Addresses); idx++ {
		addressType, address := ValidateAddress(Addresses[idx])
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

func (handler *NodesHandler) setChannelAddressesGetResult(command *Command) {
	type Response struct{}

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
	}

	resultJSON := buildJSONResponse(result.Code, Response{})
	fmt.Println(string(resultJSON))
}

func (handler *NodesHandler) SetChannelAddresses(w http.ResponseWriter, r *http.Request) {
	url := logRequest(r)

	contractorID, isParamPresent := mux.Vars(r)["contractor_id"]
	if !isParamPresent {
		logger.Error("Bad request: missing contractor_id parameter: " + url)
		w.WriteHeader(BAD_REQUEST)
		return
	}

	if !ValidateInt(contractorID) {
		logger.Error("Bad request: invalid contractor_id parameter")
		w.WriteHeader(BAD_REQUEST)
		return
	}

	addresses := []string{}
	for key, values := range r.URL.Query() {
		if key != "contractor_address" {
			continue
		}

		for _, value := range values {
			typeAndAddress := strings.Split(value, "-")
			addresses = append(addresses, typeAndAddress[0])
			addresses = append(addresses, typeAndAddress[1])
		}

		break
	}
	if len(addresses) == 0 {
		logger.Error("Bad request: there are no contractor_addresses parameters: " + url)
		w.WriteHeader(BAD_REQUEST)
		return
	}

	// Command generation
	addresses = append([]string{"SET:channel/address", contractorID, strconv.Itoa(len(addresses) / 2)}, addresses...)
	command := NewCommand(addresses...)

	type Response struct{}

	err := handler.node.SendCommand(command)
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

func (handler *NodesHandler) setChannelCryptoKey() {

	if !ValidateInt(ContractorID) {
		logger.Error("Bad request: invalid contractorID parameter in channels set-crypto-key request")
		fmt.Println("Bad request: invalid contractorID parameter")
		return
	}

	var commandParams []string
	commandParams = append(commandParams, "SET:channel/crypto-key", ContractorID, CryptoKey)
	if ChannelIDOnContractorSide != "" {
		if !ValidateInt(ChannelIDOnContractorSide) {
			logger.Error("Bad request: invalid channel-id-on-contractor-side parameter in channels set-crypto-key request")
			fmt.Println("Bad request: invalid channel-id-on-contractor-side parameter")
			return
		}
		commandParams = append(commandParams, []string{ChannelIDOnContractorSide}...)
	}

	command := NewCommand(commandParams...)

	go handler.setChannelCryptoKeyGetResult(command)
}

func (handler *NodesHandler) setChannelCryptoKeyGetResult(command *Command) {
	type Response struct{}

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
	}

	resultJSON := buildJSONResponse(result.Code, Response{})
	fmt.Println(string(resultJSON))
}

func (handler *NodesHandler) SetChannelCryptoKey(w http.ResponseWriter, r *http.Request) {
	url := logRequest(r)

	contractorID, isParamPresent := mux.Vars(r)["contractor_id"]
	if !isParamPresent {
		logger.Error("Bad request: missing contractor_id parameter: " + url)
		w.WriteHeader(BAD_REQUEST)
		return
	}

	if !ValidateInt(contractorID) {
		logger.Error("Bad request: invalid contractor_id parameter")
		w.WriteHeader(BAD_REQUEST)
		return
	}

	cryptoKey := r.FormValue("crypto_key")
	var commandParams []string
	commandParams = append(commandParams, "SET:channel/crypto-key", contractorID, cryptoKey)

	channelIDOnContractorSide := r.FormValue("channel_id_on_contractor_side")
	if channelIDOnContractorSide != "" {
		if !ValidateInt(channelIDOnContractorSide) {
			logger.Error("Bad request: invalid channel_id_on_contractor_side parameter: " + url)
			w.WriteHeader(BAD_REQUEST)
			return
		}
		commandParams = append(commandParams, []string{channelIDOnContractorSide}...)
	}

	type Response struct{}

	// Command generation
	command := NewCommand(commandParams...)
	err := handler.node.SendCommand(command)
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

func (handler *NodesHandler) regenerateChannelCryptoKey() {

	if !ValidateInt(ContractorID) {
		logger.Error("Bad request: invalid contractor_id parameter in channels regenerate-crypto-key request")
		fmt.Println("Bad request: invalid contractor_id parameter")
		return
	}

	command := NewCommand("SET:channel/regenerate-crypto-key", ContractorID)

	go handler.regenerateChannelCryptoKeyGetResult(command)
}

func (handler *NodesHandler) regenerateChannelCryptoKeyGetResult(command *Command) {
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

func (handler *NodesHandler) RegenerateChannelCryptoKey(w http.ResponseWriter, r *http.Request) {
	url := logRequest(r)

	contractorID, isParamPresent := mux.Vars(r)["contractor_id"]
	if !isParamPresent {
		logger.Error("Bad request: missing contractor_id parameter: " + url)
		w.WriteHeader(BAD_REQUEST)
		return
	}

	if !ValidateInt(contractorID) {
		logger.Error("Bad request: invalid contractor_id parameter")
		w.WriteHeader(BAD_REQUEST)
		return
	}

	// Command generation
	command := NewCommand("SET:channel/regenerate-crypto-key", contractorID)

	type Response struct {
		ChannelID string `json:"channel_id"`
		CryptoKey string `json:"crypto_key"`
	}

	err := handler.node.SendCommand(command)
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

	if len(result.Tokens) < 2 {
		logger.Error("Node return invalid result tokens size on command: " + string(command.ToBytes()))
		writeHTTPResponse(w, ENGINE_UNEXPECTED_ERROR, Response{})
		return
	}

	writeHTTPResponse(w, OK, Response{
		ChannelID: result.Tokens[0],
		CryptoKey: result.Tokens[1]})
}
