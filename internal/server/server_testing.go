package server

import (
	"github.com/gorilla/mux"
	"github.com/vTCP-Foundation/vtcpd-cli/internal/conf"
	"github.com/vTCP-Foundation/vtcpd-cli/internal/logger"
	"github.com/vTCP-Foundation/vtcpd-cli/internal/routes"
)

func InitTestNodeHandlerServer(r *routes.RoutesHandler) *mux.Router {

	router := mux.NewRouter()

	router.HandleFunc("/api/v1/node/subsystems-controller/{flags}/", r.SetTestingFlags).Methods("PUT")
	router.HandleFunc("/api/v1/node/settlement-lines-influence/{flags}/", r.SetSLInfluenceFlags).Methods("PUT")
	router.HandleFunc("/api/v1/node/make-node-busy/", r.MakeNodeBusy).Methods("PUT")

	logger.Info("Testing requests accepting started on " + conf.Params.HTTPTesting.HTTPInterface())
	return router
}
