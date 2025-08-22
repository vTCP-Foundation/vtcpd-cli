package handler

import (
	"fmt"
	"strconv"

	"github.com/vTCP-Foundation/vtcpd-cli/internal/common"
	"github.com/vTCP-Foundation/vtcpd-cli/internal/logger"
)

func (handler *NodeHandler) SettlementLines() {
	if CommandType == "init" {
		handler.initSettlementLine()
	} else if CommandType == "set" {
		handler.setMaxPositiveBalance()
	} else if CommandType == "close-incoming" {
		handler.zeroOutMaxNegativeBalance()
	} else if CommandType == "share-keys" {
		handler.shareKeysSettlementLine()
	} else if CommandType == "delete" {
		handler.deleteSettlementLine()
	} else if CommandType == "reset" {
		handler.resetSettlementLine()
	} else if CommandType == "get" {
		handler.listSettlementLinesPortions()
	} else if CommandType == "get-contractors" {
		handler.listContractors()
	} else if CommandType == "get-by-id" {
		handler.settlementLineByID()
	} else if CommandType == "get-by-addresses" {
		handler.settlementLineByAddresses()
	} else if CommandType == "equivalents" {
		handler.listEquivalents()
	} else if CommandType == "total-balance" {
		handler.totalBalance()
	} else {
		logger.Error("Invalid settlement-line command " + CommandType)
		fmt.Println("Invalid settlement-line command")
		return
	}
}

func (handler *NodeHandler) initSettlementLine() {
	if !common.ValidateInt(ContractorID) {
		logger.Error("Bad request: invalid contractorID parameter in open request")
		fmt.Println("Bad request: invalid contractorID parameter")
		return
	}
	if !common.ValidateInt(Equivalent) {
		logger.Error("Bad request: invalid equivalent parameter in open request")
		fmt.Println("Bad request: invalid equivalent parameter")
		return
	}

	command := NewCommand("INIT:contractors/trust-line", ContractorID, Equivalent)

	go handler.actionSettlementLineGetResult(command)
}

func (handler *NodeHandler) setMaxPositiveBalance() {
	if !common.ValidateInt(ContractorID) {
		logger.Error("Bad request: invalid contractorID parameter in set request")
		fmt.Println("Bad request: invalid contractorID parameter")
		return
	}
	if !common.ValidateSettlementLineAmount(Amount) {
		logger.Error("Bad request: invalid amount parameter in set request")
		fmt.Println("Bad request: invalid amount parameter")
		return
	}
	if !common.ValidateInt(Equivalent) {
		logger.Error("Bad request: invalid equivalent parameter in set request")
		fmt.Println("Bad request: invalid equivalent parameter")
		return
	}

	command := NewCommand("SET:contractors/trust-lines", ContractorID, Amount, Equivalent)

	go handler.actionSettlementLineGetResult(command)
}

func (handler *NodeHandler) zeroOutMaxNegativeBalance() {
	if !common.ValidateInt(ContractorID) {
		logger.Error("Bad request: invalid contractorID parameter in close-incoming request")
		fmt.Println("Bad request: invalid contractorID parameter")
		return
	}
	if !common.ValidateInt(Equivalent) {
		logger.Error("Bad request: invalid equivalent parameter in close-incoming request")
		fmt.Println("Bad request: invalid equivalent parameter")
		return
	}

	command := NewCommand("DELETE:contractors/incoming-trust-line", ContractorID, Equivalent)

	go handler.actionSettlementLineGetResult(command)
}

func (handler *NodeHandler) shareKeysSettlementLine() {
	if !common.ValidateInt(ContractorID) {
		logger.Error("Bad request: invalid contractorID parameter in share-keys request")
		fmt.Println("Bad request: invalid contractorID parameter")
		return
	}
	if !common.ValidateInt(Equivalent) {
		logger.Error("Bad request: invalid equivalent parameter in share-keys request")
		fmt.Println("Bad request: invalid equivalent parameter")
		return
	}

	command := NewCommand("SET:contractors/trust-line-keys", ContractorID, Equivalent)

	go handler.actionSettlementLineGetResult(command)
}

func (handler *NodeHandler) deleteSettlementLine() {
	if !common.ValidateInt(ContractorID) {
		logger.Error("Bad request: invalid contractorID parameter in delete request")
		fmt.Println("Bad request: invalid contractorID parameter")
		return
	}
	if !common.ValidateInt(Equivalent) {
		logger.Error("Bad request: invalid equivalent parameter in delete request")
		fmt.Println("Bad request: invalid equivalent parameter")
		return
	}

	command := NewCommand("DELETE:contractors/trust-line", ContractorID, Equivalent)

	go handler.actionSettlementLineGetResult(command)
}

func (handler *NodeHandler) resetSettlementLine() {
	if !common.ValidateInt(ContractorID) {
		logger.Error("Bad request: invalid contractorID parameter in reset request")
		fmt.Println("Bad request: invalid contractorID parameter")
		return
	}
	if !common.ValidateInt(Equivalent) {
		logger.Error("Bad request: invalid equivalent parameter in reset request")
		fmt.Println("Bad request: invalid equivalent parameter")
		return
	}

	if !common.ValidateInt(AuditNumber) {
		logger.Error("Bad request: invalid audit_number parameter in reset request")
		fmt.Println("Bad request: invalid audit_number parameter")
		return
	}

	if !common.ValidateSettlementLineAmount(MaxNegativeBalance) {
		logger.Error("Bad request: invalid max_negative_balance parameter in reset request")
		fmt.Println("Bad request: invalid max_negative_balance parameter")
		return
	}

	if !common.ValidateSettlementLineAmount(MaxPositiveBalance) {
		logger.Error("Bad request: invalid max_positive_balance parameterin reset request")
		fmt.Println("Bad request: invalid max_positive_balance parameter")
		return
	}

	if Balance == "" {
		logger.Error("Bad request: invalid balance parameter in reset request")
		fmt.Println("Bad request: invalid balance parameter")
		return
	}

	command := NewCommand(
		"SET:contractors/trust-lines/reset", ContractorID, AuditNumber,
		MaxNegativeBalance, MaxPositiveBalance, Balance, Equivalent)

	go handler.actionSettlementLineGetResult(command)
}

func (handler *NodeHandler) actionSettlementLineGetResult(command *Command) {
	err := handler.Node.SendCommand(command)
	if err != nil {
		logger.Error("Can't send command: " + string(command.ToBytes()) + " to node. Details: " + err.Error())
		resultJSON := buildJSONResponse(COMMAND_TRANSFERRING_ERROR, common.ActionResponse{})
		fmt.Println(string(resultJSON))
		return
	}

	result, err := handler.Node.GetResult(command, common.SETTLEMENT_LINE_RESULT_TIMEOUT)
	if err != nil {
		logger.Error("Node is inaccessible during processing command: " +
			string(command.ToBytes()) + ". Details: " + err.Error())
		resultJSON := buildJSONResponse(NODE_IS_INACCESSIBLE, common.ActionResponse{})
		fmt.Println(string(resultJSON))
		return
	}

	if result.Code != OK {
		logger.Error("Node return wrong command result: " + strconv.Itoa(result.Code) +
			" on command: " + string(command.ToBytes()))
	}

	resultJSON := buildJSONResponse(result.Code, common.ActionResponse{})
	fmt.Println(string(resultJSON))
}

func (handler *NodeHandler) listSettlementLinesPortions() {
	if Offset == "" {
		Offset = common.DEFAULT_SETTLEMENT_LINES_OFFSET
	} else if !common.ValidateInt(Offset) {
		logger.Error("Bad request: invalid offset parameter in get request")
		fmt.Println("Bad request: invalid offset parameter")
		return
	}

	if Count == "" {
		Count = common.DFEAULT_SETTLEMENT_LINES_COUNT
	} else if !common.ValidateInt(Count) {
		logger.Error("Bad request: invalid count parameter in get request")
		fmt.Println("Bad request: invalid count parameter")
		return
	}

	if !common.ValidateInt(Equivalent) {
		logger.Error("Bad request: invalid equivalent parameter in get request")
		fmt.Println("Bad request: invalid equivalent parameter")
		return
	}

	command := NewCommand("GET:contractors/trust-lines", Offset, Count, Equivalent)

	go handler.listSettlementLinesResult(command)
}

func (handler *NodeHandler) listSettlementLinesResult(command *Command) {
	err := handler.Node.SendCommand(command)
	if err != nil {
		logger.Error("Can't send command: " + string(command.ToBytes()) + " to node. Details: " + err.Error())
		resultJSON := buildJSONResponse(COMMAND_TRANSFERRING_ERROR, common.SettlementLineListResponse{})
		fmt.Println(string(resultJSON))
		return
	}

	result, err := handler.Node.GetResult(command, common.SETTLEMENT_LINE_RESULT_TIMEOUT)
	if err != nil {
		logger.Error("Node is inaccessible during processing command: " +
			string(command.ToBytes()) + ". Details: " + err.Error())
		resultJSON := buildJSONResponse(NODE_IS_INACCESSIBLE, common.SettlementLineListResponse{})
		fmt.Println(string(resultJSON))
		return
	}

	if result.Code != OK && result.Code != ENGINE_NO_EQUIVALENT {
		logger.Error("Node returned wrong command result: " + strconv.Itoa(result.Code) +
			" on command: " + string(command.ToBytes()))
		resultJSON := buildJSONResponse(result.Code, common.SettlementLineListResponse{})
		fmt.Println(string(resultJSON))
		return
	}
	if result.Code == ENGINE_NO_EQUIVALENT {
		logger.Info("Node hasn't equivalent for command: " + string(command.ToBytes()))
		resultJSON := buildJSONResponse(result.Code, common.SettlementLineListResponse{})
		fmt.Println(string(resultJSON))
		return
	}

	if len(result.Tokens) == 0 {
		logger.Error("Node returned invalid result tokens size on command: " + string(command.ToBytes()))
		resultJSON := buildJSONResponse(ENGINE_UNEXPECTED_ERROR, common.SettlementLineListResponse{})
		fmt.Println(string(resultJSON))
		return
	}

	contractorsCount, err := strconv.Atoi(result.Tokens[0])
	if err != nil {
		logger.Error("Node returned invalid token on command: " +
			string(command.ToBytes()) + ". Details: " + err.Error())
		resultJSON := buildJSONResponse(ENGINE_UNEXPECTED_ERROR, common.SettlementLineListResponse{})
		fmt.Println(string(resultJSON))
		return
	}

	if contractorsCount == 0 {
		resultJSON := buildJSONResponse(OK, common.SettlementLineListResponse{Count: contractorsCount})
		fmt.Println(string(resultJSON))
		return
	}

	response := common.SettlementLineListResponse{Count: contractorsCount}
	response.SettlementLines = make([]common.SettlementLineListItem, 0, contractorsCount)

	for i := range contractorsCount {
		response.SettlementLines = append(response.SettlementLines, common.SettlementLineListItem{
			ID:                    result.Tokens[i*8+1],
			Contractor:            result.Tokens[i*8+2],
			State:                 result.Tokens[i*8+3],
			OwnKeysPresent:        result.Tokens[i*8+4],
			ContractorKeysPresent: result.Tokens[i*8+5],
			MaxNegativeBalance:    result.Tokens[i*8+6],
			MaxPositiveBalance:    result.Tokens[i*8+7],
			Balance:               result.Tokens[i*8+8],
		})
	}
	resultJSON := buildJSONResponse(OK, response)
	fmt.Println(string(resultJSON))
}

func (handler *NodeHandler) listContractors() {

	if !common.ValidateInt(Equivalent) {
		logger.Error("Bad request: invalid equivalent parameter in get-contractors request")
		fmt.Println("Bad request: invalid equivalent parameter")
		return
	}

	command := NewCommand("GET:contractors", Equivalent)

	go handler.listContractorsResult(command)
}

func (handler *NodeHandler) listContractorsResult(command *Command) {
	err := handler.Node.SendCommand(command)
	if err != nil {
		logger.Error("Can't send command: " + string(command.ToBytes()) + " to node. Details: " + err.Error())
		resultJSON := buildJSONResponse(COMMAND_TRANSFERRING_ERROR, common.ContractorsListResponse{})
		fmt.Println(string(resultJSON))
		return
	}

	result, err := handler.Node.GetResult(command, common.CONTRACTORS_RESULT_TIMEOUT)
	if err != nil {
		logger.Error("Node is inaccessible during processing command: " +
			string(command.ToBytes()) + ". Details: " + err.Error())
		resultJSON := buildJSONResponse(NODE_IS_INACCESSIBLE, common.ContractorsListResponse{})
		fmt.Println(string(resultJSON))
		return
	}

	if result.Code != OK && result.Code != ENGINE_NO_EQUIVALENT {
		logger.Error("Node return wrong command result: " + strconv.Itoa(result.Code) +
			" on command: " + string(command.ToBytes()))
		resultJSON := buildJSONResponse(result.Code, common.ContractorsListResponse{})
		fmt.Println(string(resultJSON))
		return
	}
	if result.Code == ENGINE_NO_EQUIVALENT {
		logger.Info("Node hasn't equivalent for command: " + string(command.ToBytes()))
		resultJSON := buildJSONResponse(result.Code, common.ContractorsListResponse{})
		fmt.Println(string(resultJSON))
		return
	}

	if len(result.Tokens) == 0 {
		logger.Error("Node return invalid result tokens size on command: " + string(command.ToBytes()))
		resultJSON := buildJSONResponse(ENGINE_UNEXPECTED_ERROR, common.ContractorsListResponse{})
		fmt.Println(string(resultJSON))
		return
	}

	contractorsCount, err := strconv.Atoi(result.Tokens[0])
	if err != nil {
		logger.Error("Node return invalid token on command: " + string(command.ToBytes()) +
			". Details: " + err.Error())
		resultJSON := buildJSONResponse(ENGINE_UNEXPECTED_ERROR, common.ContractorsListResponse{})
		fmt.Println(string(resultJSON))
		return
	}

	if contractorsCount == 0 {
		resultJSON := buildJSONResponse(OK, common.ContractorsListResponse{Count: contractorsCount})
		fmt.Println(string(resultJSON))
		return
	}

	response := common.ContractorsListResponse{Count: contractorsCount}
	for i := range contractorsCount {
		response.Contractors = append(response.Contractors, common.ContractorInfo{
			ContractorID:        result.Tokens[i*2+1],
			ContractorAddresses: result.Tokens[i*2+2],
		})
	}

	resultJSON := buildJSONResponse(OK, response)
	fmt.Println(string(resultJSON))
}

func (handler *NodeHandler) settlementLineByID() {
	if !common.ValidateInt(ContractorID) {
		logger.Error("Bad request: invalid contractorID parameter in get-by-id request")
		fmt.Println("Bad request: invalid contractorID parameter")
		return
	}
	if !common.ValidateInt(Equivalent) {
		logger.Error("Bad request: invalid equivalent parameter in get-by-id request")
		fmt.Println("Bad request: invalid equivalent parameter")
		return
	}

	command := NewCommand("GET:contractors/trust-lines/one/id", ContractorID, Equivalent)

	go handler.settlementLineGetResult(command)
}

func (handler *NodeHandler) settlementLineByAddresses() {
	if len(Addresses) == 0 {
		logger.Error("Bad request: there are no contractor addresses parameters in get-by-addresses request")
		fmt.Println("Bad request: there are no contractor addresses parameters")
		return
	}

	if !common.ValidateInt(Equivalent) {
		logger.Error("Bad request: invalid equivalent parameter in get-by-addresses request")
		fmt.Println("Bad request: invalid equivalent parameter")
		return
	}

	var addresses []string
	for idx := range len(Addresses) {
		addressType, address := common.ValidateAddress(Addresses[idx])
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

	go handler.settlementLineGetResult(command)
}

func (handler *NodeHandler) settlementLineGetResult(command *Command) {
	err := handler.Node.SendCommand(command)
	if err != nil {
		logger.Error("Can't send command: " + string(command.ToBytes()) + " to node. Details: " + err.Error())
		resultJSON := buildJSONResponse(COMMAND_TRANSFERRING_ERROR, common.SettlementLineDetailResponse{})
		fmt.Println(string(resultJSON))
		return
	}

	result, err := handler.Node.GetResult(command, common.SETTLEMENT_LINE_RESULT_TIMEOUT)
	if err != nil {
		logger.Error("Node is inaccessible during processing command: " +
			string(command.ToBytes()) + ". Details: " + err.Error())
		resultJSON := buildJSONResponse(NODE_IS_INACCESSIBLE, common.SettlementLineDetailResponse{})
		fmt.Println(string(resultJSON))
		return
	}

	if result.Code != OK && result.Code != ENGINE_NO_EQUIVALENT && result.Code != NODE_NOT_FOUND {
		logger.Error("Node return wrong command result: " + strconv.Itoa(result.Code) +
			" on command: " + string(command.ToBytes()))
		resultJSON := buildJSONResponse(result.Code, common.SettlementLineDetailResponse{})
		fmt.Println(string(resultJSON))
		return
	}
	if result.Code == ENGINE_NO_EQUIVALENT {
		logger.Info("Node hasn't equivalent for command: " + string(command.ToBytes()))
		resultJSON := buildJSONResponse(result.Code, common.SettlementLineDetailResponse{})
		fmt.Println(string(resultJSON))
		return
	}
	if result.Code == NODE_NOT_FOUND {
		logger.Info("Node hasn't TL with requested contractor for command " + string(command.ToBytes()))
		resultJSON := buildJSONResponse(result.Code, common.SettlementLineDetailResponse{})
		fmt.Println(string(resultJSON))
		return
	}

	if len(result.Tokens) < 4 {
		logger.Error("Node return invalid result tokens size on command: " + string(command.ToBytes()))
		resultJSON := buildJSONResponse(ENGINE_UNEXPECTED_ERROR, common.SettlementLineDetailResponse{})
		fmt.Println(string(resultJSON))
		return
	}

	response := common.SettlementLineDetailResponse{SettlementLine: common.SettlementLineDetail{
		ID:                    result.Tokens[0],
		State:                 result.Tokens[1],
		OwnKeysPresent:        result.Tokens[2],
		ContractorKeysPresent: result.Tokens[3],
		AuditNumber:           result.Tokens[4],
		MaxNegativeBalance:    result.Tokens[5],
		MaxPositiveBalance:    result.Tokens[6],
		Balance:               result.Tokens[7],
	}}
	resultJSON := buildJSONResponse(OK, response)
	fmt.Println(string(resultJSON))
}

func (handler *NodeHandler) listEquivalents() {

	command := NewCommand("GET:equivalents")

	go handler.listEquivalentsGetResult(command)
}

func (handler *NodeHandler) listEquivalentsGetResult(command *Command) {
	err := handler.Node.SendCommand(command)
	if err != nil {
		logger.Error("Can't send command: " + string(command.ToBytes()) + " to node. Details: " + err.Error())
		resultJSON := buildJSONResponse(COMMAND_TRANSFERRING_ERROR, common.EquivalentsListResponse{})
		fmt.Println(string(resultJSON))
		return
	}

	result, err := handler.Node.GetResult(command, common.SETTLEMENT_LINE_RESULT_TIMEOUT)
	if err != nil {
		logger.Error("Node is inaccessible during processing command: " +
			string(command.ToBytes()) + ". Details: " + err.Error())
		resultJSON := buildJSONResponse(NODE_IS_INACCESSIBLE, common.EquivalentsListResponse{})
		fmt.Println(string(resultJSON))
		return
	}

	if result.Code != OK {
		logger.Error("Node return wrong command result: " + strconv.Itoa(result.Code) +
			" on command: " + string(command.ToBytes()))
		resultJSON := buildJSONResponse(result.Code, common.EquivalentsListResponse{})
		fmt.Println(string(resultJSON))
		return
	}

	if len(result.Tokens) == 0 {
		logger.Error("Node return invalid result tokens size on command: " + string(command.ToBytes()))
		resultJSON := buildJSONResponse(ENGINE_UNEXPECTED_ERROR, common.EquivalentsListResponse{})
		fmt.Println(string(resultJSON))
		return
	}

	equivalentsCount, err := strconv.Atoi(result.Tokens[0])
	if err != nil {
		logger.Error("Node return invalid token on command: " +
			string(command.ToBytes()) + ". Details: " + err.Error())
		resultJSON := buildJSONResponse(ENGINE_UNEXPECTED_ERROR, common.EquivalentsListResponse{})
		fmt.Println(string(resultJSON))
		return
	}

	if equivalentsCount == 0 {
		resultJSON := buildJSONResponse(OK, common.EquivalentsListResponse{Count: equivalentsCount})
		fmt.Println(string(resultJSON))
		return
	}

	response := common.EquivalentsListResponse{Count: equivalentsCount}
	for i := range equivalentsCount {
		response.Equivalents = append(response.Equivalents, result.Tokens[i+1])
	}
	resultJSON := buildJSONResponse(OK, response)
	fmt.Println(string(resultJSON))
}

func (handler *NodeHandler) totalBalance() {

	if !common.ValidateInt(Equivalent) {
		logger.Error("Bad request: invalid equivalent parameter in total-balance request")
		fmt.Println("Bad request: invalid equivalent parameter")
		return
	}

	command := NewCommand("GET:stats/balance/total", Equivalent)

	go handler.totalBalanceGetResult(command)
}

func (handler *NodeHandler) totalBalanceGetResult(command *Command) {

	err := handler.Node.SendCommand(command)
	if err != nil {
		logger.Error("Can't send command: " + string(command.ToBytes()) + " to node. Details: " + err.Error())
		resultJSON := buildJSONResponse(COMMAND_TRANSFERRING_ERROR, common.TotalBalanceResponse{})
		fmt.Println(string(resultJSON))
		return
	}

	result, err := handler.Node.GetResult(command, common.STATS_RESULT_TIMEOUT)
	if err != nil {
		logger.Error("Node is inaccessible during processing command: " + string(command.ToBytes()) +
			". Details: " + err.Error())
		resultJSON := buildJSONResponse(NODE_IS_INACCESSIBLE, common.TotalBalanceResponse{})
		fmt.Println(string(resultJSON))
		return
	}

	if result.Code != OK && result.Code != ENGINE_NO_EQUIVALENT {
		logger.Error("Node return wrong command result: " + strconv.Itoa(result.Code) +
			" on command: " + string(command.ToBytes()))
		resultJSON := buildJSONResponse(result.Code, common.TotalBalanceResponse{})
		fmt.Println(string(resultJSON))
		return
	}
	if result.Code == ENGINE_NO_EQUIVALENT {
		logger.Info("Node hasn't equivalent for command: " + string(command.ToBytes()))
		resultJSON := buildJSONResponse(result.Code, common.TotalBalanceResponse{})
		fmt.Println(string(resultJSON))
	}

	if len(result.Tokens) < 4 {
		logger.Error("Node return invalid result tokens size on command: " + string(command.ToBytes()))
		resultJSON := buildJSONResponse(ENGINE_UNEXPECTED_ERROR, common.TotalBalanceResponse{})
		fmt.Println(string(resultJSON))
		return
	}

	response := common.TotalBalanceResponse{
		TotalMaxNegativeBalance: result.Tokens[0],
		TotalNegativeBalance:    result.Tokens[1],
		TotalMaxPositiveBalance: result.Tokens[2],
		TotalPositiveBalance:    result.Tokens[3],
	}
	resultJSON := buildJSONResponse(OK, response)
	fmt.Println(string(resultJSON))
}
