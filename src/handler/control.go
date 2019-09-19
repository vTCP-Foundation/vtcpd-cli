package handler

import (
	"fmt"
	"net/http"
	"os"

	"logger"
)

func (handler *NodesHandler) StopEverything(w http.ResponseWriter, r *http.Request) {
	_, err := preprocessRequest(r)
	if err != nil {
		logger.Error("Bad request: invalid security parameters: " + err.Error())
		w.WriteHeader(BAD_REQUEST)
		return
	}

	type Response struct {
		Status string `json:"status"`
		Msg    string `json:"msg"`
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

	writeHTTPResponse(w, OK, Response{"ok", "Stop request received"})
}
