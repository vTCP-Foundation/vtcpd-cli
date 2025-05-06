package handler

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
	"github.com/vTCP-Foundation/vtcpd-cli/internal/logger"
)

// --- Global structs for channels ---

type ChannelListItem struct {
	ID        string `json:"channel_id"`
	Addresses string `json:"channel_addresses"`
}

// --- Global API responses for channels ---

type ChannelInitResponse struct {
	ChannelID string `json:"channel_id"`
	CryptoKey string `json:"crypto_key"`
}

type ChannelInfoResponse struct {
	ID                  string   `json:"channel_id"`
	Addresses           []string `json:"channel_addresses"`
	IsConfirmed         string   `json:"channel_confirmed"`
	CryptoKey           string   `json:"channel_crypto_key"`
	ContractorCryptoKey string   `json:"channel_contractor_crypto_key"`
}

type ChannelListResponse struct {
	Count    int               `json:"count"`
	Channels []ChannelListItem `json:"channels"`
}

type ChannelInfoByAddressResponse struct {
	ID          string `json:"channel_id"`
	IsConfirmed string `json:"channel_confirmed"`
}

type ChannelResponse struct{}

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

	err := handler.node.SendCommand(command)
	if err != nil {
		logger.Error("Can't send command: " + string(command.ToBytes()) + " to node. Details: " + err.Error())
		resultJSON := buildJSONResponse(COMMAND_TRANSFERRING_ERROR, ChannelInitResponse{})
		fmt.Println(string(resultJSON))
		return
	}

	result, err := handler.node.GetResult(command, CHANNEL_RESULT_TIMEOUT)
	if err != nil {
		logger.Error("Node is inaccessible during processing command: " +
			string(command.ToBytes()) + ". Details: " + err.Error())
		resultJSON := buildJSONResponse(NODE_IS_INACCESSIBLE, ChannelInitResponse{})
		fmt.Println(string(resultJSON))
		return
	}

	if result.Code != OK {
		logger.Error("Node return wrong command result: " + strconv.Itoa(result.Code) +
			" on command: " + string(command.ToBytes()))
		resultJSON := buildJSONResponse(result.Code, ChannelInitResponse{})
		fmt.Println(string(resultJSON))
		return
	}

	if len(result.Tokens) < 2 {
		logger.Error("Node return invalid result tokens size on command: " + string(command.ToBytes()))
		resultJSON := buildJSONResponse(ENGINE_UNEXPECTED_ERROR, ChannelInitResponse{})
		fmt.Println(string(resultJSON))
		return
	}
	resultJSON := buildJSONResponse(result.Code, ChannelInitResponse{
		ChannelID: result.Tokens[0],
		CryptoKey: result.Tokens[1]})
	fmt.Println(string(resultJSON))
}

func (handler *NodesHandler) InitChannel(w http.ResponseWriter, r *http.Request) {
	url, err := preprocessRequest(r)
	if err != nil {
		logger.Error("Bad request: invalid security parameters: " + err.Error())
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

	err = handler.node.SendCommand(command)
	if err != nil {
		logger.Error("Can't send command: " + string(command.ToBytes()) + " to node. Details: " + err.Error())
		writeHTTPResponse(w, COMMAND_TRANSFERRING_ERROR, ChannelInitResponse{})
		return
	}

	result, err := handler.node.GetResult(command, CHANNEL_RESULT_TIMEOUT)
	if err != nil {
		logger.Error("Node is inaccessible during processing command: " +
			string(command.ToBytes()) + ". Details: " + err.Error())
		writeHTTPResponse(w, NODE_IS_INACCESSIBLE, ChannelInitResponse{})
		return
	}

	if result.Code != OK {
		logger.Error("Node return wrong command result: " + strconv.Itoa(result.Code) +
			" on command: " + string(command.ToBytes()))
		writeHTTPResponse(w, result.Code, ChannelInitResponse{})
		return
	}

	if len(result.Tokens) < 2 {
		logger.Error("Node return invalid result tokens size on command: " + string(command.ToBytes()))
		writeHTTPResponse(w, ENGINE_UNEXPECTED_ERROR, ChannelInitResponse{})
		return
	}

	writeHTTPResponse(w, OK, ChannelInitResponse{
		ChannelID: result.Tokens[0],
		CryptoKey: result.Tokens[1]})
}

func (handler *NodesHandler) listChannels() {

	command := NewCommand("GET:contractors-all")

	go handler.listChannelsGetResult(command)
}

func (handler *NodesHandler) listChannelsGetResult(command *Command) {
	err := handler.node.SendCommand(command)
	if err != nil {
		logger.Error("Can't send command: " + string(command.ToBytes()) + " to node. Details: " + err.Error())
		resultJSON := buildJSONResponse(COMMAND_TRANSFERRING_ERROR, ChannelListResponse{})
		fmt.Println(string(resultJSON))
		return
	}

	result, err := handler.node.GetResult(command, CHANNEL_RESULT_TIMEOUT)
	if err != nil {
		logger.Error("Node is inaccessible during processing command: " +
			string(command.ToBytes()) + ". Details: " + err.Error())
		resultJSON := buildJSONResponse(NODE_IS_INACCESSIBLE, ChannelListResponse{})
		fmt.Println(string(resultJSON))
		return
	}

	if result.Code != OK {
		logger.Error("Node return wrong command result: " + strconv.Itoa(result.Code) +
			" on command: " + string(command.ToBytes()))
		resultJSON := buildJSONResponse(result.Code, ChannelListResponse{})
		fmt.Println(string(resultJSON))
		return
	}

	if len(result.Tokens) == 0 {
		logger.Error("Node return invalid result tokens size on command: " + string(command.ToBytes()))
		resultJSON := buildJSONResponse(ENGINE_UNEXPECTED_ERROR, ChannelListResponse{})
		fmt.Println(string(resultJSON))
		return
	}

	// Channels received well
	channelsCount, err := strconv.Atoi(result.Tokens[0])
	if err != nil {
		logger.Error("Node return invalid token on command: " +
			string(command.ToBytes()) + ". Details: " + err.Error())
		resultJSON := buildJSONResponse(ENGINE_UNEXPECTED_ERROR, ChannelListResponse{})
		fmt.Println(string(resultJSON))
		return
	}

	if channelsCount == 0 {
		resultJSON := buildJSONResponse(OK, ChannelListResponse{Count: channelsCount})
		fmt.Println(string(resultJSON))
		return
	}

	response := ChannelListResponse{Count: channelsCount}
	for i := 0; i < channelsCount; i++ {
		response.Channels = append(response.Channels, ChannelListItem{
			ID:        result.Tokens[i*2+1],
			Addresses: result.Tokens[i*2+2]})
	}
	resultJSON := buildJSONResponse(OK, response)
	fmt.Println(string(resultJSON))
}

func (handler *NodesHandler) ListChannels(w http.ResponseWriter, r *http.Request) {
	_, err := preprocessRequest(r)
	if err != nil {
		logger.Error("Bad request: invalid security parameters: " + err.Error())
		w.WriteHeader(BAD_REQUEST)
		return
	}

	command := NewCommand("GET:contractors-all")

	err = handler.node.SendCommand(command)
	if err != nil {
		logger.Error("Can't send command: " + string(command.ToBytes()) + " to node. Details: " + err.Error())
		writeHTTPResponse(w, COMMAND_TRANSFERRING_ERROR, ChannelListResponse{})
		return
	}

	result, err := handler.node.GetResult(command, CHANNEL_RESULT_TIMEOUT)
	if err != nil {
		logger.Error("Node is inaccessible during processing command: " +
			string(command.ToBytes()) + ". Details: " + err.Error())
		writeHTTPResponse(w, NODE_IS_INACCESSIBLE, ChannelListResponse{})
		return
	}

	if result.Code != OK {
		logger.Error("Node return wrong command result: " + strconv.Itoa(result.Code) +
			" on command: " + string(command.ToBytes()))
		writeHTTPResponse(w, result.Code, ChannelListResponse{})
		return
	}

	if len(result.Tokens) == 0 {
		logger.Error("Node return invalid result tokens size on command: " + string(command.ToBytes()))
		writeHTTPResponse(w, ENGINE_UNEXPECTED_ERROR, ChannelListResponse{})
		return
	}

	// Channels received well
	channelsCount, err := strconv.Atoi(result.Tokens[0])
	if err != nil {
		logger.Error("Node return invalid token on command: " +
			string(command.ToBytes()) + ". Details: " + err.Error())
		writeHTTPResponse(w, ENGINE_UNEXPECTED_ERROR, ChannelListResponse{})
		return
	}

	if channelsCount == 0 {
		writeHTTPResponse(w, OK, ChannelListResponse{Count: channelsCount})
		return
	}

	response := ChannelListResponse{Count: channelsCount}
	for i := 0; i < channelsCount; i++ {
		response.Channels = append(response.Channels, ChannelListItem{
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
	err := handler.node.SendCommand(command)
	if err != nil {
		logger.Error("Can't send command: " + string(command.ToBytes()) + " to node. Details: " + err.Error())
		resultJSON := buildJSONResponse(COMMAND_TRANSFERRING_ERROR, ChannelInfoResponse{})
		fmt.Println(string(resultJSON))
		return
	}

	result, err := handler.node.GetResult(command, CHANNEL_RESULT_TIMEOUT)
	if err != nil {
		logger.Error("Node is inaccessible during processing command: " +
			string(command.ToBytes()) + ". Details: " + err.Error())
		resultJSON := buildJSONResponse(NODE_IS_INACCESSIBLE, ChannelInfoResponse{})
		fmt.Println(string(resultJSON))
		return
	}

	if result.Code != OK {
		logger.Error("Node return wrong command result: " + strconv.Itoa(result.Code) +
			" on command: " + string(command.ToBytes()))
		resultJSON := buildJSONResponse(result.Code, ChannelInfoResponse{})
		fmt.Println(string(resultJSON))
		return
	}

	if len(result.Tokens) < 2 {
		logger.Error("Node return invalid result tokens size on command: " + string(command.ToBytes()))
		resultJSON := buildJSONResponse(ENGINE_UNEXPECTED_ERROR, ChannelInfoResponse{})
		fmt.Println(string(resultJSON))
		return
	}

	addressesCount, err := strconv.Atoi(result.Tokens[1])
	if err != nil {
		logger.Error("Node return invalid token on command: " + string(command.ToBytes()) + ". Details: " + err.Error())
		resultJSON := buildJSONResponse(ENGINE_UNEXPECTED_ERROR, ChannelInfoResponse{})
		fmt.Println(string(resultJSON))
		return
	}

	if addressesCount == 0 {
		logger.Error("Node return invalid addresses token on command: " + string(command.ToBytes()))
		resultJSON := buildJSONResponse(ENGINE_UNEXPECTED_ERROR, ChannelInfoResponse{})
		fmt.Println(string(resultJSON))
		return
	}

	var addresses []string
	for idx := 0; idx < addressesCount; idx++ {
		addresses = append(addresses, result.Tokens[idx+2])
	}

	response := ChannelInfoResponse{
		ID:                  result.Tokens[0],
		Addresses:           addresses,
		IsConfirmed:         result.Tokens[2+addressesCount],
		CryptoKey:           result.Tokens[2+addressesCount+1],
		ContractorCryptoKey: result.Tokens[2+addressesCount+2]}
	resultJSON := buildJSONResponse(OK, response)
	fmt.Println(string(resultJSON))
}

func (handler *NodesHandler) ChannelInfo(w http.ResponseWriter, r *http.Request) {
	url, err := preprocessRequest(r)
	if err != nil {
		logger.Error("Bad request: invalid security parameters: " + err.Error())
		w.WriteHeader(BAD_REQUEST)
		return
	}

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

	err = handler.node.SendCommand(command)
	if err != nil {
		logger.Error("Can't send command: " + string(command.ToBytes()) + " to node. Details: " + err.Error())
		writeHTTPResponse(w, COMMAND_TRANSFERRING_ERROR, ChannelInfoResponse{})
		return
	}

	result, err := handler.node.GetResult(command, SETTLEMENT_LINE_RESULT_TIMEOUT)
	if err != nil {
		logger.Error("Node is inaccessible during processing command: " +
			string(command.ToBytes()) + ". Details: " + err.Error())
		writeHTTPResponse(w, NODE_IS_INACCESSIBLE, ChannelInfoResponse{})
		return
	}

	if result.Code != OK {
		logger.Error("Node return wrong command result: " + strconv.Itoa(result.Code) +
			" on command: " + string(command.ToBytes()))
		writeHTTPResponse(w, result.Code, ChannelInfoResponse{})
		return
	}

	if len(result.Tokens) < 2 {
		logger.Error("Node return invalid result tokens size on command: " + string(command.ToBytes()))
		writeHTTPResponse(w, ENGINE_UNEXPECTED_ERROR, ChannelInfoResponse{})
		return
	}

	addressesCount, err := strconv.Atoi(result.Tokens[1])
	if err != nil {
		logger.Error("Node return invalid token on command: " + string(command.ToBytes()) + ". Details: " + err.Error())
		writeHTTPResponse(w, ENGINE_UNEXPECTED_ERROR, ChannelInfoResponse{})
		return
	}

	if addressesCount == 0 {
		logger.Error("Node return invalid addresses token on command: " + string(command.ToBytes()))
		writeHTTPResponse(w, ENGINE_UNEXPECTED_ERROR, ChannelInfoResponse{})
		return
	}

	var addresses []string
	for idx := 0; idx < addressesCount; idx++ {
		addresses = append(addresses, result.Tokens[idx+2])
	}

	response := ChannelInfoResponse{
		ID:                  result.Tokens[0],
		Addresses:           addresses,
		IsConfirmed:         result.Tokens[2+addressesCount],
		CryptoKey:           result.Tokens[2+addressesCount+1],
		ContractorCryptoKey: result.Tokens[2+addressesCount+2]}
	writeHTTPResponse(w, OK, response)
}

func (handler *NodesHandler) channelInfoByAddresses() {

	if len(Addresses) == 0 {
		logger.Error("Bad request: there are no contractor addresses parameters in one-by-addresses request")
		fmt.Println("Bad request: there are no contractor addresses parameters")
		return
	}

	var addresses []string
	for idx := 0; idx < len(Addresses); idx++ {
		addressType, address := ValidateAddress(Addresses[idx])
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

func (handler *NodesHandler) channelInfoByAddressesGetResult(command *Command) {
	err := handler.node.SendCommand(command)
	if err != nil {
		logger.Error("Can't send command: " + string(command.ToBytes()) + " to node. Details: " + err.Error())
		resultJSON := buildJSONResponse(COMMAND_TRANSFERRING_ERROR, ChannelInfoByAddressResponse{})
		fmt.Println(string(resultJSON))
		return
	}

	result, err := handler.node.GetResult(command, CHANNEL_RESULT_TIMEOUT)
	if err != nil {
		logger.Error("Node is inaccessible during processing command: " +
			string(command.ToBytes()) + ". Details: " + err.Error())
		resultJSON := buildJSONResponse(NODE_IS_INACCESSIBLE, ChannelInfoByAddressResponse{})
		fmt.Println(string(resultJSON))
		return
	}

	if result.Code != OK {
		logger.Error("Node return wrong command result: " + strconv.Itoa(result.Code) +
			" on command: " + string(command.ToBytes()))
		resultJSON := buildJSONResponse(result.Code, ChannelInfoByAddressResponse{})
		fmt.Println(string(resultJSON))
		return
	}

	if len(result.Tokens) < 2 {
		logger.Error("Node return invalid result tokens size on command: " + string(command.ToBytes()))
		resultJSON := buildJSONResponse(ENGINE_UNEXPECTED_ERROR, ChannelInfoByAddressResponse{})
		fmt.Println(string(resultJSON))
		return
	}

	response := ChannelInfoByAddressResponse{
		ID:          result.Tokens[0],
		IsConfirmed: result.Tokens[1]}
	resultJSON := buildJSONResponse(OK, response)
	fmt.Println(string(resultJSON))
}

func (handler *NodesHandler) ChannelInfoByAddresses(w http.ResponseWriter, r *http.Request) {
	url, err := preprocessRequest(r)
	if err != nil {
		logger.Error("Bad request: invalid security parameters: " + err.Error())
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
	contractorAddresses = append([]string{"GET:channels/one/address"}, contractorAddresses...)
	command := NewCommand(contractorAddresses...)

	err = handler.node.SendCommand(command)
	if err != nil {
		logger.Error("Can't send command: " + string(command.ToBytes()) + " to node. Details: " + err.Error())
		writeHTTPResponse(w, COMMAND_TRANSFERRING_ERROR, ChannelInfoByAddressResponse{})
		return
	}

	result, err := handler.node.GetResult(command, SETTLEMENT_LINE_RESULT_TIMEOUT)
	if err != nil {
		logger.Error("Node is inaccessible during processing command: " +
			string(command.ToBytes()) + ". Details: " + err.Error())
		writeHTTPResponse(w, NODE_IS_INACCESSIBLE, ChannelInfoByAddressResponse{})
		return
	}

	if result.Code != OK {
		logger.Error("Node return wrong command result: " + strconv.Itoa(result.Code) +
			" on command: " + string(command.ToBytes()))
		writeHTTPResponse(w, result.Code, ChannelInfoByAddressResponse{})
		return
	}

	if len(result.Tokens) < 2 {
		logger.Error("Node return invalid result tokens size on command: " + string(command.ToBytes()))
		writeHTTPResponse(w, ENGINE_UNEXPECTED_ERROR, ChannelInfoByAddressResponse{})
		return
	}

	response := ChannelInfoByAddressResponse{
		ID:          result.Tokens[0],
		IsConfirmed: result.Tokens[1]}
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
	err := handler.node.SendCommand(command)
	if err != nil {
		logger.Error("Can't send command: " + string(command.ToBytes()) + " to node. Details: " + err.Error())
		resultJSON := buildJSONResponse(COMMAND_TRANSFERRING_ERROR, ChannelResponse{})
		fmt.Println(string(resultJSON))
		return
	}

	result, err := handler.node.GetResult(command, CHANNEL_RESULT_TIMEOUT)
	if err != nil {
		logger.Error("Node is inaccessible during processing command: " +
			string(command.ToBytes()) + ". Details: " + err.Error())
		resultJSON := buildJSONResponse(NODE_IS_INACCESSIBLE, ChannelResponse{})
		fmt.Println(string(resultJSON))
		return
	}

	if result.Code != OK {
		logger.Error("Node return wrong command result: " + strconv.Itoa(result.Code) +
			" on command: " + string(command.ToBytes()))
	}

	resultJSON := buildJSONResponse(result.Code, ChannelResponse{})
	fmt.Println(string(resultJSON))
}

func (handler *NodesHandler) SetChannelAddresses(w http.ResponseWriter, r *http.Request) {
	url, err := preprocessRequest(r)
	if err != nil {
		logger.Error("Bad request: invalid security parameters: " + err.Error())
		w.WriteHeader(BAD_REQUEST)
		return
	}

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

	err = handler.node.SendCommand(command)
	if err != nil {
		logger.Error("Can't send command: " + string(command.ToBytes()) + " to node. Details: " + err.Error())
		writeHTTPResponse(w, COMMAND_TRANSFERRING_ERROR, ChannelResponse{})
		return
	}

	result, err := handler.node.GetResult(command, SETTLEMENT_LINE_RESULT_TIMEOUT)
	if err != nil {
		logger.Error("Node is inaccessible during processing command: " +
			string(command.ToBytes()) + ". Details: " + err.Error())
		writeHTTPResponse(w, NODE_IS_INACCESSIBLE, ChannelResponse{})
		return
	}

	if result.Code != OK {
		logger.Error("Node return wrong command result: " + strconv.Itoa(result.Code) +
			" on command: " + string(command.ToBytes()))
	}

	writeHTTPResponse(w, result.Code, ChannelResponse{})
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
	err := handler.node.SendCommand(command)
	if err != nil {
		logger.Error("Can't send command: " + string(command.ToBytes()) + " to node. Details: " + err.Error())
		resultJSON := buildJSONResponse(COMMAND_TRANSFERRING_ERROR, ChannelResponse{})
		fmt.Println(string(resultJSON))
		return
	}

	result, err := handler.node.GetResult(command, CHANNEL_RESULT_TIMEOUT)
	if err != nil {
		logger.Error("Node is inaccessible during processing command: " +
			string(command.ToBytes()) + ". Details: " + err.Error())
		resultJSON := buildJSONResponse(NODE_IS_INACCESSIBLE, ChannelResponse{})
		fmt.Println(string(resultJSON))
		return
	}

	if result.Code != OK {
		logger.Error("Node return wrong command result: " + strconv.Itoa(result.Code) +
			" on command: " + string(command.ToBytes()))
	}

	resultJSON := buildJSONResponse(result.Code, ChannelResponse{})
	fmt.Println(string(resultJSON))
}

func (handler *NodesHandler) SetChannelCryptoKey(w http.ResponseWriter, r *http.Request) {
	url, err := preprocessRequest(r)
	if err != nil {
		logger.Error("Bad request: invalid security parameters: " + err.Error())
		w.WriteHeader(BAD_REQUEST)
		return
	}

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

	// Command generation
	command := NewCommand(commandParams...)
	err = handler.node.SendCommand(command)
	if err != nil {
		logger.Error("Can't send command: " + string(command.ToBytes()) + " to node. Details: " + err.Error())
		writeHTTPResponse(w, COMMAND_TRANSFERRING_ERROR, ChannelResponse{})
		return
	}

	result, err := handler.node.GetResult(command, SETTLEMENT_LINE_RESULT_TIMEOUT)
	if err != nil {
		logger.Error("Node is inaccessible during processing command: " +
			string(command.ToBytes()) + ". Details: " + err.Error())
		writeHTTPResponse(w, NODE_IS_INACCESSIBLE, ChannelResponse{})
		return
	}

	if result.Code != OK {
		logger.Error("Node return wrong command result: " + strconv.Itoa(result.Code) +
			" on command: " + string(command.ToBytes()))
	}

	writeHTTPResponse(w, result.Code, ChannelResponse{})
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
	err := handler.node.SendCommand(command)
	if err != nil {
		logger.Error("Can't send command: " + string(command.ToBytes()) + " to node. Details: " + err.Error())
		resultJSON := buildJSONResponse(COMMAND_TRANSFERRING_ERROR, ChannelInitResponse{})
		fmt.Println(string(resultJSON))
		return
	}

	result, err := handler.node.GetResult(command, CHANNEL_RESULT_TIMEOUT)
	if err != nil {
		logger.Error("Node is inaccessible during processing command: " +
			string(command.ToBytes()) + ". Details: " + err.Error())
		resultJSON := buildJSONResponse(NODE_IS_INACCESSIBLE, ChannelInitResponse{})
		fmt.Println(string(resultJSON))
		return
	}

	if result.Code != OK {
		logger.Error("Node return wrong command result: " + strconv.Itoa(result.Code) +
			" on command: " + string(command.ToBytes()))
		resultJSON := buildJSONResponse(result.Code, ChannelInitResponse{})
		fmt.Println(string(resultJSON))
		return
	}

	if len(result.Tokens) < 2 {
		logger.Error("Node return invalid result tokens size on command: " + string(command.ToBytes()))
		resultJSON := buildJSONResponse(ENGINE_UNEXPECTED_ERROR, ChannelInitResponse{})
		fmt.Println(string(resultJSON))
		return
	}
	resultJSON := buildJSONResponse(result.Code, ChannelInitResponse{
		ChannelID: result.Tokens[0],
		CryptoKey: result.Tokens[1]})
	fmt.Println(string(resultJSON))
}

func (handler *NodesHandler) RegenerateChannelCryptoKey(w http.ResponseWriter, r *http.Request) {
	url, err := preprocessRequest(r)
	if err != nil {
		logger.Error("Bad request: invalid security parameters: " + err.Error())
		w.WriteHeader(BAD_REQUEST)
		return
	}

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

	err = handler.node.SendCommand(command)
	if err != nil {
		logger.Error("Can't send command: " + string(command.ToBytes()) + " to node. Details: " + err.Error())
		writeHTTPResponse(w, COMMAND_TRANSFERRING_ERROR, ChannelInitResponse{})
		return
	}

	result, err := handler.node.GetResult(command, CHANNEL_RESULT_TIMEOUT)
	if err != nil {
		logger.Error("Node is inaccessible during processing command: " +
			string(command.ToBytes()) + ". Details: " + err.Error())
		writeHTTPResponse(w, NODE_IS_INACCESSIBLE, ChannelInitResponse{})
		return
	}

	if result.Code != OK {
		logger.Error("Node return wrong command result: " + strconv.Itoa(result.Code) +
			" on command: " + string(command.ToBytes()))
		writeHTTPResponse(w, result.Code, ChannelInitResponse{})
		return
	}

	if len(result.Tokens) < 2 {
		logger.Error("Node return invalid result tokens size on command: " + string(command.ToBytes()))
		writeHTTPResponse(w, ENGINE_UNEXPECTED_ERROR, ChannelInitResponse{})
		return
	}

	writeHTTPResponse(w, OK, ChannelInitResponse{
		ChannelID: result.Tokens[0],
		CryptoKey: result.Tokens[1]})
}

func (handler *NodesHandler) removeChannel() {

	if !ValidateInt(ContractorID) {
		logger.Error("Bad request: invalid contractor_id parameter in channels remove request")
		fmt.Println("Bad request: invalid contractor_id parameter")
		return
	}

	command := NewCommand("DELETE:channel/contractor-id", ContractorID)

	go handler.removeChannelGetResult(command)
}

func (handler *NodesHandler) removeChannelGetResult(command *Command) {
	err := handler.node.SendCommand(command)
	if err != nil {
		logger.Error("Can't send command: " + string(command.ToBytes()) + " to node. Details: " + err.Error())
		resultJSON := buildJSONResponse(COMMAND_TRANSFERRING_ERROR, ChannelResponse{})
		fmt.Println(string(resultJSON))
		return
	}

	result, err := handler.node.GetResult(command, CHANNEL_RESULT_TIMEOUT)
	if err != nil {
		logger.Error("Node is inaccessible during processing command: " +
			string(command.ToBytes()) + ". Details: " + err.Error())
		resultJSON := buildJSONResponse(NODE_IS_INACCESSIBLE, ChannelResponse{})
		fmt.Println(string(resultJSON))
		return
	}

	if result.Code != OK {
		logger.Error("Node return wrong command result: " + strconv.Itoa(result.Code) +
			" on command: " + string(command.ToBytes()))
	}

	resultJSON := buildJSONResponse(result.Code, ChannelResponse{})
	fmt.Println(string(resultJSON))
}

func (handler *NodesHandler) RemoveChannel(w http.ResponseWriter, r *http.Request) {
	url, err := preprocessRequest(r)
	if err != nil {
		logger.Error("Bad request: invalid security parameters: " + err.Error())
		w.WriteHeader(BAD_REQUEST)
		return
	}

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
	command := NewCommand("DELETE:channel/contractor-id", contractorID)

	err = handler.node.SendCommand(command)
	if err != nil {
		logger.Error("Can't send command: " + string(command.ToBytes()) + " to node. Details: " + err.Error())
		writeHTTPResponse(w, COMMAND_TRANSFERRING_ERROR, ChannelResponse{})
		return
	}

	result, err := handler.node.GetResult(command, CHANNEL_RESULT_TIMEOUT)
	if err != nil {
		logger.Error("Node is inaccessible during processing command: " +
			string(command.ToBytes()) + ". Details: " + err.Error())
		writeHTTPResponse(w, NODE_IS_INACCESSIBLE, ChannelResponse{})
		return
	}

	if result.Code != OK {
		logger.Error("Node return wrong command result: " + strconv.Itoa(result.Code) +
			" on command: " + string(command.ToBytes()))
	}

	writeHTTPResponse(w, result.Code, ChannelResponse{})
}
