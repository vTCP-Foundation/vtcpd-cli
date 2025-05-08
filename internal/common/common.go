package common

// --- Global conts for command timeouts ---
var (
	CHANNEL_RESULT_TIMEOUT          uint16 = 20 // seconds
	SETTLEMENT_LINE_RESULT_TIMEOUT  uint16 = 20 // seconds
	CONTRACTORS_RESULT_TIMEOUT      uint16 = 20
	STATS_RESULT_TIMEOUT            uint16 = 20 // seconds
	DEFAULT_SETTLEMENT_LINES_OFFSET        = "0"
	DFEAULT_SETTLEMENT_LINES_COUNT         = "10000"
	PAYMENT_OPERATION_TIMEOUT       uint16 = 60
	MAX_FLOW_FIRST_TIMEOUT          uint16 = 30
	MAX_FLOW_FULLY_TIMEOUT          uint16 = 60
	COMMAND_UUID_TIMEOUT            uint16 = 20
	HISTORY_RESULT_TIMEOUT          uint16 = 20 // seconds
	DELETE_CRYPTO_DATA_TIMEOUT      uint16 = 20 // seconds
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

// --- Global settlement lines structs ---

type SettlementLineListItem struct {
	ID                    string `json:"contractor_id"`
	Contractor            string `json:"contractor"`
	State                 string `json:"state"`
	OwnKeysPresent        string `json:"own_keys_present"`
	ContractorKeysPresent string `json:"contractor_keys_present"`
	MaxNegativeBalance    string `json:"max_negative_balance"`
	MaxPositiveBalance    string `json:"max_positive_balance"`
	Balance               string `json:"balance"`
}

type SettlementLineDetail struct {
	ID                    string `json:"id"` // todo : contractor_id
	State                 string `json:"state"`
	OwnKeysPresent        string `json:"own_keys_present"`
	ContractorKeysPresent string `json:"contractor_keys_present"`
	AuditNumber           string `json:"audit_number"`
	MaxNegativeBalance    string `json:"max_negative_balance"`
	MaxPositiveBalance    string `json:"max_positive_balance"`
	Balance               string `json:"balance"`
}

type EquivalentStatistics struct {
	Eq              string                   `json:"equivalent"`
	Count           int                      `json:"count"`
	SettlementLines []SettlementLineListItem `json:"settlement_lines"`
}

type ContractorInfo struct {
	ContractorID        string `json:"contractor_id"`
	ContractorAddresses string `json:"contractor_addresses"`
}

// --- Global API responses for settlement lines ---

// ActionResponse used for operations that return only status
type ActionResponse struct{}

type SettlementLineListResponse struct {
	Count           int                      `json:"count"`
	SettlementLines []SettlementLineListItem `json:"settlement_lines"`
}

type SettlementLineDetailResponse struct {
	SettlementLine SettlementLineDetail `json:"settlement_line"`
}

type AllEquivalentsResponse struct {
	Count       int                    `json:"count"`
	Equivalents []EquivalentStatistics `json:"equivalents"`
}

type ContractorsListResponse struct {
	Count       int              `json:"count"`
	Contractors []ContractorInfo `json:"contractors"`
}

type EquivalentsListResponse struct {
	Count       int      `json:"count"`
	Equivalents []string `json:"equivalents"`
}

type TotalBalanceResponse struct {
	TotalMaxNegativeBalance string `json:"total_max_negative_balance"`
	TotalNegativeBalance    string `json:"total_negative_balance"`
	TotalMaxPositiveBalance string `json:"total_max_positive_balance"`
	TotalPositiveBalance    string `json:"total_positive_balance"`
}

// --- Global structs for payments ---

type MaxFlowRecord struct {
	ContractorAddressType string `json:"address_type"`
	ContractorAddress     string `json:"contractor_address"`
	MaxAmount             string `json:"max_amount"`
}

// --- Global API responses for payments ---

type MaxFlowResponse struct {
	Count   int             `json:"count"`
	Records []MaxFlowRecord `json:"records"`
}

type MaxFlowPartialResponse struct {
	State   int             `json:"state"`
	Count   int             `json:"count"`
	Records []MaxFlowRecord `json:"records"`
}

type PaymentResponse struct {
	TransactionUUID string `json:"transaction_uuid"`
}

type GetTransactionByCommandUUIDResponse struct {
	Count           int    `json:"count"`
	TransactionUUID string `json:"transaction_uuid"`
}

// --- Global structs for history ---

type SettlementLineHistoryRecord struct {
	TransactionUUID           string `json:"transaction_uuid"`
	UnixTimestampMicroseconds string `json:"unix_timestamp_microseconds"`
	Contractor                string `json:"contractor"`
	OperationDirection        string `json:"operation_direction"`
	Amount                    string `json:"amount"`
}

type PaymentHistoryRecord struct {
	TransactionUUID           string `json:"transaction_uuid"`
	UnixTimestampMicroseconds string `json:"unix_timestamp_microseconds"`
	Contractor                string `json:"contractor"`
	OperationDirection        string `json:"operation_direction"`
	Amount                    string `json:"amount"`
	BalanceAfterOperation     string `json:"balance_after_operation"`
	Payload                   string `json:"payload"`
}

type PaymentAllEquivalentsHistoryRecord struct {
	Equivalent                string `json:"equivalent"`
	TransactionUUID           string `json:"transaction_uuid"`
	UnixTimestampMicroseconds string `json:"unix_timestamp_microseconds"`
	Contractor                string `json:"contractor"`
	OperationDirection        string `json:"operation_direction"`
	Amount                    string `json:"amount"`
	BalanceAfterOperation     string `json:"balance_after_operation"`
	Payload                   string `json:"payload"`
}

type ContractorOperationHistoryRecord struct {
	RecordType                string `json:"record_type"` // "payment" or "trustline"
	TransactionUUID           string `json:"transaction_uuid"`
	UnixTimestampMicroseconds string `json:"unix_timestamp_microseconds"`
	OperationDirection        string `json:"operation_direction"`
	Amount                    string `json:"amount"`
	BalanceAfterOperation     string `json:"balance_after_operation"` // Can be "0" для trustline
	Payload                   string `json:"payload"`                 // Can be "" для trustline
}

type AdditionalPaymentHistoryRecord struct {
	TransactionUUID           string `json:"transaction_uuid"`
	UnixTimestampMicroseconds string `json:"unix_timestamp_microseconds"`
	OperationDirection        string `json:"operation_direction"`
	Amount                    string `json:"amount"`
}

// --- Global API responses for history ---

type SettlementLineHistoryResponse struct {
	Count   int                           `json:"count"`
	Records []SettlementLineHistoryRecord `json:"records"`
}

type PaymentHistoryResponse struct {
	Count   int                    `json:"count"`
	Records []PaymentHistoryRecord `json:"records"`
}

type PaymentAllEquivalentsHistoryResponse struct {
	Count   int                                  `json:"count"`
	Records []PaymentAllEquivalentsHistoryRecord `json:"records"`
}

type ContractorOperationsHistoryResponse struct {
	Count   int                                `json:"count"`
	Records []ContractorOperationHistoryRecord `json:"records"`
}

type AdditionalPaymentHistoryResponse struct {
	Count   int                              `json:"count"`
	Records []AdditionalPaymentHistoryRecord `json:"records"`
}

// --- Global API responses for control

type ControlMsgResponse struct {
	Status string `json:"status"`
	Msg    string `json:"msg"`
}

type ControlResponse struct{}
