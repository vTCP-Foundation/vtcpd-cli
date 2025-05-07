package server

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/vTCP-Foundation/vtcpd-cli/internal/conf"
	"github.com/vTCP-Foundation/vtcpd-cli/internal/logger"
	"github.com/vTCP-Foundation/vtcpd-cli/internal/routes"
)

func InitTestNodeHandlerServer(r *routes.RoutesHandler) *mux.Router {

	router := mux.NewRouter()

	// router.HandleFunc("/api/v1/node/subsystems-controller/{flags}/", handler.SetTestingFlags).Methods("PUT")
	// router.HandleFunc("/api/v1/node/settlement-lines-influence/{flags}/", handler.SetSLInfluenceFlags).Methods("PUT")
	// router.HandleFunc("/api/v1/node/make-node-busy/", handler.MakeNodeBusy).Methods("PUT")

	http.Handle("/", router)
	logger.Info("Requests accepting started on " + conf.Params.TestHandler.HTTPInterface())
	return router
}
