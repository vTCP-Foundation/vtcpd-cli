package routes

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
	"github.com/vTCP-Foundation/vtcpd-cli/internal/common"
	"github.com/vTCP-Foundation/vtcpd-cli/internal/handler"
	"github.com/vTCP-Foundation/vtcpd-cli/internal/logger"
)

var (
	INFLUENCE_RESULT_TIMEOUT uint16 = 20 // seconds
)

func (router *RoutesHandler) SetTestingFlags(w http.ResponseWriter, r *http.Request) {
	_, err := preprocessRequest(r)
	if err != nil {
		logger.Error("Bad request: invalid security parameters: " + err.Error())
		w.WriteHeader(BAD_REQUEST)
		return
	}

	flags := mux.Vars(r)["flags"]
	if flags == "" {
		logger.Error("Bad request: invalid flags parameter in set test flag request")
		writeHTTPResponse(w, BAD_REQUEST, common.ControlResponse{})
		return
	}

	forbiddenNodeAddress := r.URL.Query().Get("forbidden_address")
	forbiddenAmount := r.URL.Query().Get("forbidden_amount")

	if forbiddenNodeAddress == "" {
		command := handler.NewCommand(
			"SET:subsystems_controller/flags", flags)
		err := router.nodeHandler.Node.SendCommand(command)
		if err != nil {
			logger.Error("Can't send command: " + string(command.ToBytes()) + " to node. Details: " + err.Error())
			writeHTTPResponse(w, NODE_IS_INACCESSIBLE, common.ControlResponse{})
			return
		}

		// Node is not responding to this command
		writeHTTPResponse(w, http.StatusOK, common.ControlResponse{})
		return
	}

	typeAndAddress := strings.Split(forbiddenNodeAddress, "-")
	if len(typeAndAddress) != 2 {
		logger.Error("Bad request: invalid forbidden_address parameter in set test flag request")
		writeHTTPResponse(w, BAD_REQUEST, common.ControlResponse{})
		return
	}

	if forbiddenAmount == "" {
		command := handler.NewCommand(
			"SET:subsystems_controller/flags", flags, typeAndAddress[0], typeAndAddress[1])
		err := router.nodeHandler.Node.SendCommand(command)
		if err != nil {
			logger.Error("Can't send command: " + string(command.ToBytes()) + " to node. Details: " + err.Error())
			writeHTTPResponse(w, NODE_IS_INACCESSIBLE, common.ControlResponse{})
			return
		}

		// Node is not responding to this command
		writeHTTPResponse(w, http.StatusOK, common.ControlResponse{})
		return
	}

	command := handler.NewCommand(
		"SET:subsystems_controller/flags", flags, typeAndAddress[0], typeAndAddress[1])

	err = router.nodeHandler.Node.SendCommand(command)
	if err != nil {
		logger.Error("Can't send command: " + string(command.ToBytes()) + " to node. Details: " + err.Error())
		writeHTTPResponse(w, NODE_IS_INACCESSIBLE, common.ControlResponse{})
		return
	}
	logger.Info("Command sent: " + string(command.ToBytes()))

	// Node is not responding to this command
	writeHTTPResponse(w, http.StatusOK, common.ControlResponse{})
}

func (router *RoutesHandler) SetSLInfluenceFlags(w http.ResponseWriter, r *http.Request) {
	_, err := preprocessRequest(r)
	if err != nil {
		logger.Error("Bad request: invalid security parameters: " + err.Error())
		w.WriteHeader(BAD_REQUEST)
		return
	}

	flags := mux.Vars(r)["flags"]
	if flags == "" {
		logger.Error("Bad request: invalid flags parameter in set test flag SL request")
		fmt.Println("Bad request: invalid flags parameter")
		return
	}

	firstParameter := r.URL.Query().Get("first_parameter")
	secondParameter := r.URL.Query().Get("second_parameter")
	thirdParameter := r.URL.Query().Get("third_parameter")

	command := handler.NewCommand(
		"SET:subsystems_controller/trust_lines_influence/flags", flags, firstParameter, secondParameter, thirdParameter)

	err = router.nodeHandler.Node.SendCommand(command)
	if err != nil {
		logger.Error("Can't send command: " + string(command.ToBytes()) + " to node. Details: " + err.Error())
		writeHTTPResponse(w, NODE_IS_INACCESSIBLE, common.ControlResponse{})
		return
	}

	// Node is not responding to this command
	writeHTTPResponse(w, http.StatusOK, common.ControlResponse{})
}

func (router *RoutesHandler) MakeNodeBusy(w http.ResponseWriter, r *http.Request) {

	interval := r.URL.Query().Get("interval")
	if interval == "" {
		interval = "60"
	}

	command := handler.NewCommand("TEST:make-node-busy", interval)

	err := router.nodeHandler.Node.SendCommand(command)
	if err != nil {
		logger.Error("Can't send command: " + string(command.ToBytes()) + " to node. Details: " + err.Error())
		writeHTTPResponse(w, NODE_IS_INACCESSIBLE, common.ControlResponse{})
		return
	}
	result, err := router.nodeHandler.Node.GetResult(command, INFLUENCE_RESULT_TIMEOUT)
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
