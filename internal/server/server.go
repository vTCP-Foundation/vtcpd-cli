package server

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/vTCP-Foundation/vtcpd-cli/internal/conf"
	"github.com/vTCP-Foundation/vtcpd-cli/internal/logger"
	"github.com/vTCP-Foundation/vtcpd-cli/internal/routes"
)

func InitNodeHandlerServer(r *routes.RoutesHandler) *mux.Router {

	router := mux.NewRouter()

	// Equivalents
	router.HandleFunc("/api/v1/node/equivalents/", r.ListEquivalents).Methods("GET")

	// Channels
	router.HandleFunc("/api/v1/node/contractors/init-channel/", r.InitChannel).Methods("POST")
	router.HandleFunc("/api/v1/node/contractors/channels/", r.ListChannels).Methods("GET")
	router.HandleFunc("/api/v1/node/channels/{contractor_id}/", r.ChannelInfo).Methods("GET")
	router.HandleFunc("/api/v1/node/channel-by-address/", r.ChannelInfoByAddresses).Methods("GET")
	router.HandleFunc("/api/v1/node/channels/{contractor_id}/set-addresses/", r.SetChannelAddresses).Methods("PUT")
	router.HandleFunc("/api/v1/node/channels/{contractor_id}/set-crypto-key/", r.SetChannelCryptoKey).Methods("PUT")
	router.HandleFunc("/api/v1/node/channels/{contractor_id}/regenerate-crypto-key/", r.RegenerateChannelCryptoKey).Methods("PUT")
	router.HandleFunc("/api/v1/node/channels/{contractor_id}/remove/", r.RemoveChannel).Methods("DELETE")

	// Contractors
	router.HandleFunc("/api/v1/node/contractors/{equivalent}/", r.ListContractors).Methods("GET")

	// Contractors / Settlement Lines
	router.HandleFunc("/api/v1/node/contractors/settlement-lines/{equivalent}/", r.ListSettlementLines).Methods("GET")
	router.HandleFunc("/api/v1/node/contractors/settlement-lines/{offset}/{count}/{equivalent}/", r.ListSettlementLinesPortions).Methods("GET")
	router.HandleFunc("/api/v1/node/contractors/settlement-lines/equivalents/all/", r.ListSettlementLinesAllEquivalents).Methods("GET")
	router.HandleFunc("/api/v1/node/contractors/settlement-line-by-id/{equivalent}/", r.GetSettlementLineByID).Methods("GET")
	router.HandleFunc("/api/v1/node/contractors/settlement-line-by-address/{equivalent}/", r.GetSettlementLineByAddress).Methods("GET")
	router.HandleFunc("/api/v1/node/contractors/{contractor_id}/init-settlement-line/{equivalent}/", r.InitSettlementLine).Methods("POST")
	router.HandleFunc("/api/v1/node/contractors/{contractor_id}/settlement-lines/{equivalent}/", r.SetMaxPositiveBalance).Methods("PUT")
	router.HandleFunc("/api/v1/node/contractors/{contractor_id}/close-incoming-settlement-line/{equivalent}/", r.ZeroOutMaxNegativeBalance).Methods("DELETE")
	router.HandleFunc("/api/v1/node/contractors/{contractor_id}/keys-sharing/{equivalent}/", r.PublicKeysSharing).Methods("PUT")
	router.HandleFunc("/api/v1/node/contractors/{contractor_id}/remove-settlement-line/{equivalent}/", r.RemoveSettlementLine).Methods("DELETE")
	router.HandleFunc("/api/v1/node/contractors/{contractor_id}/reset-settlement-line/{equivalent}/", r.ResetSettlementLine).Methods("PUT")

	// Contractors / Transactions
	router.HandleFunc("/api/v1/node/contractors/transactions/{equivalent}/", r.CreateTransaction).Methods("POST")
	router.HandleFunc("/api/v1/node/contractors/transactions/max/{equivalent}/", r.BatchMaxFullyTransaction).Methods("GET")
	router.HandleFunc("/api/v1/node/transactions/{command_uuid}/", r.GetTransactionByCommandUUID).Methods("GET")

	// Stats
	router.HandleFunc("/api/v1/node/stats/total-balance/{equivalent}/", r.TotalBalance).Methods("GET")

	// History
	router.HandleFunc("/api/v1/node/history/transactions/payments/{offset}/{count}/{equivalent}/", r.PaymentsHistory).Methods("GET")
	router.HandleFunc("/api/v1/node/history/transactions/payments-all/{offset}/{count}/", r.PaymentsHistoryAllEquivalents).Methods("GET")
	router.HandleFunc("/api/v1/node/history/transactions/payments/additional/{offset}/{count}/{equivalent}/", r.PaymentsAdditionalHistory).Methods("GET")
	router.HandleFunc("/api/v1/node/history/transactions/settlement-lines/{offset}/{count}/{equivalent}/", r.SettlementLinesHistory).Methods("GET")
	router.HandleFunc("/api/v1/node/history/contractors/{offset}/{count}/{equivalent}/", r.HistoryWithContractor).Methods("GET")

	// Optimization
	router.HandleFunc("/api/v1/node/remove-outdated-crypto/", r.RemoveOutdatedCryptoData).Methods("DELETE")
	router.HandleFunc("/api/v1/node/regenerate-all-keys/", r.RegenerateAllKeys).Methods("POST")

	// Control
	router.HandleFunc("/api/v1/ctrl/stop/", r.StopEverything).Methods("POST")

	http.Handle("/", router)
	logger.Info("Requests accepting started on " + conf.Params.HTTP.HTTPInterface())
	return router
}
