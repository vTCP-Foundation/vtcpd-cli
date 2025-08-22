package routes

import (
	"encoding/json"
	"errors"
	"io"
	"net"
	"net/http"
	"strings"

	"github.com/vTCP-Foundation/vtcpd-cli/internal/conf"
	"github.com/vTCP-Foundation/vtcpd-cli/internal/handler"
	"github.com/vTCP-Foundation/vtcpd-cli/internal/logger"
)

var (
	// Response status codes
	// for the commands
	OK                         = 200
	CREATED                    = 201
	ACCEPTED                   = 202
	BAD_REQUEST                = 400
	NODE_NOT_FOUND             = 405
	SERVER_ERROR               = 500
	NODE_IS_INACCESSIBLE       = 503
	ENGINE_UNEXPECTED_ERROR    = 504
	COMMAND_TRANSFERRING_ERROR = 505
	ENGINE_NO_EQUIVALENT       = 604
)

type RoutesHandler struct {
	nodeHandler *handler.NodeHandler
}

func NewRoutesHandler(nodeHandler *handler.NodeHandler) *RoutesHandler {

	routesHandler := &RoutesHandler{
		nodeHandler: nodeHandler,
	}

	return routesHandler
}

func writeHTTPResponse(w http.ResponseWriter, statusCode int, data interface{}) {
	w.WriteHeader(statusCode)
	writeJSONResponse(data, w)
}

func writeJSONResponse(data interface{}, w http.ResponseWriter) {
	type Response struct {
		Data interface{} `json:"data"`
	}

	response := Response{Data: data}
	js, err := json.Marshal(response)
	if err != nil {
		logger.Error("Can't marshall data. Details are: " + err.Error())
		writeServerError("JSON forming error", w)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(js)
}

func writeServerError(message string, w http.ResponseWriter) {
	w.WriteHeader(SERVER_ERROR)
	w.Header().Set("Content-Type", "application/json")

	content := make(map[string]string)
	content["error"] = message

	js, _ := json.Marshal(content)
	w.Write(js)
}

func preprocessRequest(r *http.Request) (string, error) {
	url := ""
	if r.Method == "GET" {
		url = r.Method + ": " + r.URL.String()
	} else {
		bodyBytes, _ := io.ReadAll(r.Body)
		url = r.Method + ": " + r.URL.String() + "{ " + string(bodyBytes) + "}"
	}
	logger.Info(url)
	requesterIP := getRealAddr(r)
	logger.Info("Requester IP: " + requesterIP)
	if len(conf.Params.Security.AllowableIPs) > 0 {
		ipIsAllow := false
		for _, allowableIP := range conf.Params.Security.AllowableIPs {
			if allowableIP == requesterIP {
				ipIsAllow = true
				break
			}
		}
		if !ipIsAllow {
			return url, errors.New("IP " + requesterIP + " is not allow")
		}
	}
	apiKey := r.Header.Get("api-key")
	if conf.Params.Security.ApiKey != "" {
		if apiKey != conf.Params.Security.ApiKey {
			return url, errors.New("Invalid api-key " + apiKey)
		}
	}
	return url, nil
}

func getRealAddr(r *http.Request) string {
	remoteIP := ""
	// the default is the originating ip. but we try to find better options because this is almost
	// never the right IP
	if parts := strings.Split(r.RemoteAddr, ":"); len(parts) == 2 {
		remoteIP = parts[0]
	}
	// If we have a forwarded-for header, take the address from there
	if xff := strings.Trim(r.Header.Get("X-Forwarded-For"), ","); len(xff) > 0 {
		addrs := strings.Split(xff, ",")
		lastFwd := addrs[len(addrs)-1]
		if ip := net.ParseIP(lastFwd); ip != nil {
			remoteIP = ip.String()
		}
		// parse X-Real-Ip header
	} else if xri := r.Header.Get("X-Real-Ip"); len(xri) > 0 {
		if ip := net.ParseIP(xri); ip != nil {
			remoteIP = ip.String()
		}
	}

	return remoteIP
}
