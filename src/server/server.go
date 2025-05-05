package server

import (
	"github.com/gorilla/mux"
	"github.com/vTCP-Foundation/vtcpd-cli/src/conf"
	"github.com/vTCP-Foundation/vtcpd-cli/src/handler"
	"github.com/vTCP-Foundation/vtcpd-cli/src/logger"
	"net/http"
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

	// Contractors / Trust Lines
	router.HandleFunc("/api/v1/node/contractors/trust-lines/{equivalent}/", handler.ListTrustLines).Methods("GET")
	router.HandleFunc("/api/v1/node/contractors/trust-lines/{offset}/{count}/{equivalent}/", handler.ListTrustLinesPortions).Methods("GET")
	router.HandleFunc("/api/v1/node/contractors/trust-lines/equivalents/all/", handler.ListTrustLinesAllEquivalents).Methods("GET")
	router.HandleFunc("/api/v1/node/contractors/trust-line-by-id/{equivalent}/", handler.GetTrustLineByID).Methods("GET")
	router.HandleFunc("/api/v1/node/contractors/trust-line-by-address/{equivalent}/", handler.GetTrustLineByAddress).Methods("GET")
	router.HandleFunc("/api/v1/node/contractors/{contractor_id}/init-trust-line/{equivalent}/", handler.InitTrustLine).Methods("POST")
	router.HandleFunc("/api/v1/node/contractors/{contractor_id}/trust-lines/{equivalent}/", handler.SetTrustLine).Methods("PUT")
	router.HandleFunc("/api/v1/node/contractors/{contractor_id}/close-incoming-trust-line/{equivalent}/", handler.CloseIncomingTrustLine).Methods("DELETE")
	router.HandleFunc("/api/v1/node/contractors/{contractor_id}/keys-sharing/{equivalent}/", handler.PublicKeysSharing).Methods("PUT")
	router.HandleFunc("/api/v1/node/contractors/{contractor_id}/remove-trust-line/{equivalent}/", handler.RemoveTrustLine).Methods("DELETE")
	router.HandleFunc("/api/v1/node/contractors/{contractor_id}/reset-trust-line/{equivalent}/", handler.ResetTrustLine).Methods("PUT")

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
	router.HandleFunc("/api/v1/node/history/transactions/trust-lines/{offset}/{count}/{equivalent}/", handler.TrustLinesHistory).Methods("GET")
	router.HandleFunc("/api/v1/node/history/contractors/{offset}/{count}/{equivalent}/", handler.HistoryWithContractor).Methods("GET")

	// Optimization
	router.HandleFunc("/api/v1/node/remove-outdated-crypto/", handler.RemoveOutdatedCryptoData).Methods("DELETE")
	router.HandleFunc("/api/v1/node/regenerate-all-keys/", handler.RegenerateAllKeys).Methods("POST")

	// Control
	router.HandleFunc("/api/v1/ctrl/stop/", handler.StopEverything).Methods("POST")
	// save all TL data in nodeTL.json file. used only for internal comparisons
	router.HandleFunc("/api/v1/ctrl/save/", handler.SaveTrustLineInfo).Methods("POST")

	http.Handle("/", router)
	logger.Info("Requests accepting started on " + conf.Params.Handler.HTTPInterface())

	return
}
