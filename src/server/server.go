package server

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/vTCP-Foundation/vtcpd-cli/src/conf"
	"github.com/vTCP-Foundation/vtcpd-cli/src/handler"
	"github.com/vTCP-Foundation/vtcpd-cli/src/logger"
)

func InitNodesHandlerServer(handler *handler.NodesHandler) {

	router := mux.NewRouter()

	// Equivalents
	router.HandleFunc("/api/v1/node/equivalents/", handler.ListEquivalents).Methods("GET")

	// Channels
	router.HandleFunc("/api/v1/node/contractors/init-channel/", handler.InitChannel).Methods("POST")
	router.HandleFunc("/api/v1/node/contractors/channels/", handler.ListChannels).Methods("GET")
	router.HandleFunc("/api/v1/node/channels/{contractor_id}/", handler.ChannelInfo).Methods("GET")
	router.HandleFunc("/api/v1/node/channel-by-address/", handler.ChannelInfoByAddresses).Methods("GET")
	router.HandleFunc("/api/v1/node/channels/{contractor_id}/set-addresses/", handler.SetChannelAddresses).Methods("PUT")
	router.HandleFunc("/api/v1/node/channels/{contractor_id}/set-crypto-key/", handler.SetChannelCryptoKey).Methods("PUT")
	router.HandleFunc("/api/v1/node/channels/{contractor_id}/regenerate-crypto-key/", handler.RegenerateChannelCryptoKey).Methods("PUT")
	router.HandleFunc("/api/v1/node/channels/{contractor_id}/remove/", handler.RemoveChannel).Methods("DELETE")

	// Contractors
	router.HandleFunc("/api/v1/node/contractors/{equivalent}/", handler.ListContractors).Methods("GET")

	// Contractors / Settlement Lines
	router.HandleFunc("/api/v1/node/contractors/settlement-lines/{equivalent}/", handler.ListSettlementLines).Methods("GET")
	router.HandleFunc("/api/v1/node/contractors/settlement-lines/{offset}/{count}/{equivalent}/", handler.ListSettlementLinesPortions).Methods("GET")
	router.HandleFunc("/api/v1/node/contractors/settlement-lines/equivalents/all/", handler.ListSettlementLinesAllEquivalents).Methods("GET")
	router.HandleFunc("/api/v1/node/contractors/settlement-line-by-id/{equivalent}/", handler.GetSettlementLineByID).Methods("GET")
	router.HandleFunc("/api/v1/node/contractors/settlement-line-by-address/{equivalent}/", handler.GetSettlementLineByAddress).Methods("GET")
	router.HandleFunc("/api/v1/node/contractors/{contractor_id}/init-settlement-line/{equivalent}/", handler.InitSettlementLine).Methods("POST")
	router.HandleFunc("/api/v1/node/contractors/{contractor_id}/settlement-lines/{equivalent}/", handler.SetMaxPositiveBalance).Methods("PUT")
	router.HandleFunc("/api/v1/node/contractors/{contractor_id}/close-incoming-settlement-line/{equivalent}/", handler.ZeroOutMaxNegativeBalance).Methods("DELETE")
	router.HandleFunc("/api/v1/node/contractors/{contractor_id}/keys-sharing/{equivalent}/", handler.PublicKeysSharing).Methods("PUT")
	router.HandleFunc("/api/v1/node/contractors/{contractor_id}/remove-settlement-line/{equivalent}/", handler.RemoveSettlementLine).Methods("DELETE")
	router.HandleFunc("/api/v1/node/contractors/{contractor_id}/reset-settlement-line/{equivalent}/", handler.ResetSettlementLine).Methods("PUT")

	// Contractors / Transactions
	router.HandleFunc("/api/v1/node/contractors/transactions/{equivalent}/", handler.CreateTransaction).Methods("POST")
	router.HandleFunc("/api/v1/node/contractors/transactions/max/{equivalent}/", handler.BatchMaxFullyTransaction).Methods("GET")
	router.HandleFunc("/api/v1/node/transactions/{command_uuid}/", handler.GetTransactionByCommandUUID).Methods("GET")

	// Stats
	router.HandleFunc("/api/v1/node/stats/total-balance/{equivalent}/", handler.TotalBalance).Methods("GET")

	// History
	router.HandleFunc("/api/v1/node/history/transactions/payments/{offset}/{count}/{equivalent}/", handler.PaymentsHistory).Methods("GET")
	router.HandleFunc("/api/v1/node/history/transactions/payments-all/{offset}/{count}/", handler.PaymentsHistoryAllEquivalents).Methods("GET")
	router.HandleFunc("/api/v1/node/history/transactions/payments/additional/{offset}/{count}/{equivalent}/", handler.PaymentsAdditionalHistory).Methods("GET")
	router.HandleFunc("/api/v1/node/history/transactions/settlement-lines/{offset}/{count}/{equivalent}/", handler.SettlementLinesHistory).Methods("GET")
	router.HandleFunc("/api/v1/node/history/contractors/{offset}/{count}/{equivalent}/", handler.HistoryWithContractor).Methods("GET")

	// Optimization
	router.HandleFunc("/api/v1/node/remove-outdated-crypto/", handler.RemoveOutdatedCryptoData).Methods("DELETE")
	router.HandleFunc("/api/v1/node/regenerate-all-keys/", handler.RegenerateAllKeys).Methods("POST")

	// Control
	router.HandleFunc("/api/v1/ctrl/stop/", handler.StopEverything).Methods("POST")

	http.Handle("/", router)
	logger.Info("Requests accepting started on " + conf.Params.Handler.HTTPInterface())

}
