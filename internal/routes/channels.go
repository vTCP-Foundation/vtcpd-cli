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
	CHANNEL_RESULT_TIMEOUT uint16 = 20 // seconds
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

func (router *RoutesHandler) InitChannel(w http.ResponseWriter, r *http.Request) {
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
		if !handler.ValidateInt(contractorChannelID) {
			logger.Error("Bad request: there are no contractor_id parameter: " + url)
			w.WriteHeader(BAD_REQUEST)
			return
		}
		contractorAddresses = append(contractorAddresses, []string{cryptoKey, contractorChannelID}...)
	}
	command := handler.NewCommand(contractorAddresses...)

	err = router.nodeHandler.Node.SendCommand(command)
	if err != nil {
		logger.Error("Can't send command: " + string(command.ToBytes()) + " to node. Details: " + err.Error())
		writeHTTPResponse(w, COMMAND_TRANSFERRING_ERROR, ChannelInitResponse{})
		return
	}

	result, err := router.nodeHandler.Node.GetResult(command, CHANNEL_RESULT_TIMEOUT)
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

func (router *RoutesHandler) ListChannels(w http.ResponseWriter, r *http.Request) {
	_, err := preprocessRequest(r)
	if err != nil {
		logger.Error("Bad request: invalid security parameters: " + err.Error())
		w.WriteHeader(BAD_REQUEST)
		return
	}

	command := handler.NewCommand("GET:contractors-all")

	err = router.nodeHandler.Node.SendCommand(command)
	if err != nil {
		logger.Error("Can't send command: " + string(command.ToBytes()) + " to node. Details: " + err.Error())
		writeHTTPResponse(w, COMMAND_TRANSFERRING_ERROR, ChannelListResponse{})
		return
	}

	result, err := router.nodeHandler.Node.GetResult(command, CHANNEL_RESULT_TIMEOUT)
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
	for i := range channelsCount {
		response.Channels = append(response.Channels, ChannelListItem{
			ID:        result.Tokens[i*2+1],
			Addresses: result.Tokens[i*2+2]})
	}
	writeHTTPResponse(w, OK, response)
}

func (router *RoutesHandler) ChannelInfo(w http.ResponseWriter, r *http.Request) {
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

	if !handler.ValidateInt(contractorID) {
		logger.Error("Bad request: invalid contractor_id parameter")
		w.WriteHeader(BAD_REQUEST)
		return
	}

	command := handler.NewCommand("GET:channels/one", contractorID)

	err = router.nodeHandler.Node.SendCommand(command)
	if err != nil {
		logger.Error("Can't send command: " + string(command.ToBytes()) + " to node. Details: " + err.Error())
		writeHTTPResponse(w, COMMAND_TRANSFERRING_ERROR, ChannelInfoResponse{})
		return
	}

	result, err := router.nodeHandler.Node.GetResult(command, CHANNEL_RESULT_TIMEOUT)
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
	for idx := range addressesCount {
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

func (router *RoutesHandler) ChannelInfoByAddresses(w http.ResponseWriter, r *http.Request) {
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
	command := handler.NewCommand(contractorAddresses...)

	err = router.nodeHandler.Node.SendCommand(command)
	if err != nil {
		logger.Error("Can't send command: " + string(command.ToBytes()) + " to node. Details: " + err.Error())
		writeHTTPResponse(w, COMMAND_TRANSFERRING_ERROR, ChannelInfoByAddressResponse{})
		return
	}

	result, err := router.nodeHandler.Node.GetResult(command, CHANNEL_RESULT_TIMEOUT)
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

func (router *RoutesHandler) SetChannelAddresses(w http.ResponseWriter, r *http.Request) {
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

	if !handler.ValidateInt(contractorID) {
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
	command := handler.NewCommand(addresses...)

	err = router.nodeHandler.Node.SendCommand(command)
	if err != nil {
		logger.Error("Can't send command: " + string(command.ToBytes()) + " to node. Details: " + err.Error())
		writeHTTPResponse(w, COMMAND_TRANSFERRING_ERROR, ChannelResponse{})
		return
	}

	result, err := router.nodeHandler.Node.GetResult(command, CHANNEL_RESULT_TIMEOUT)
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

func (router *RoutesHandler) SetChannelCryptoKey(w http.ResponseWriter, r *http.Request) {
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

	if !handler.ValidateInt(contractorID) {
		logger.Error("Bad request: invalid contractor_id parameter")
		w.WriteHeader(BAD_REQUEST)
		return
	}

	cryptoKey := r.FormValue("crypto_key")
	var commandParams []string
	commandParams = append(commandParams, "SET:channel/crypto-key", contractorID, cryptoKey)

	channelIDOnContractorSide := r.FormValue("channel_id_on_contractor_side")
	if channelIDOnContractorSide != "" {
		if !handler.ValidateInt(channelIDOnContractorSide) {
			logger.Error("Bad request: invalid channel_id_on_contractor_side parameter: " + url)
			w.WriteHeader(BAD_REQUEST)
			return
		}
		commandParams = append(commandParams, []string{channelIDOnContractorSide}...)
	}

	// Command generation
	command := handler.NewCommand(commandParams...)
	err = router.nodeHandler.Node.SendCommand(command)
	if err != nil {
		logger.Error("Can't send command: " + string(command.ToBytes()) + " to node. Details: " + err.Error())
		writeHTTPResponse(w, COMMAND_TRANSFERRING_ERROR, ChannelResponse{})
		return
	}

	result, err := router.nodeHandler.Node.GetResult(command, CHANNEL_RESULT_TIMEOUT)
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

func (router *RoutesHandler) RegenerateChannelCryptoKey(w http.ResponseWriter, r *http.Request) {
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

	if !handler.ValidateInt(contractorID) {
		logger.Error("Bad request: invalid contractor_id parameter")
		w.WriteHeader(BAD_REQUEST)
		return
	}

	// Command generation
	command := handler.NewCommand("SET:channel/regenerate-crypto-key", contractorID)

	err = router.nodeHandler.Node.SendCommand(command)
	if err != nil {
		logger.Error("Can't send command: " + string(command.ToBytes()) + " to node. Details: " + err.Error())
		writeHTTPResponse(w, COMMAND_TRANSFERRING_ERROR, ChannelInitResponse{})
		return
	}

	result, err := router.nodeHandler.Node.GetResult(command, CHANNEL_RESULT_TIMEOUT)
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

func (router *RoutesHandler) RemoveChannel(w http.ResponseWriter, r *http.Request) {
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

	if !handler.ValidateInt(contractorID) {
		logger.Error("Bad request: invalid contractor_id parameter")
		w.WriteHeader(BAD_REQUEST)
		return
	}

	// Command generation
	command := handler.NewCommand("DELETE:channel/contractor-id", contractorID)

	err = router.nodeHandler.Node.SendCommand(command)
	if err != nil {
		logger.Error("Can't send command: " + string(command.ToBytes()) + " to node. Details: " + err.Error())
		writeHTTPResponse(w, COMMAND_TRANSFERRING_ERROR, ChannelResponse{})
		return
	}

	result, err := router.nodeHandler.Node.GetResult(command, CHANNEL_RESULT_TIMEOUT)
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
