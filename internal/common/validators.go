package common

import (
	"math/big"
	"strconv"
	"strings"

	"github.com/google/uuid"
)

var (
	bigZero = big.NewInt(0)
)

func ValidateSettlementLineAmount(amount string) bool {
	if amount == "" {
		return false
	}

	parsedAmount := &big.Int{}
	parsedAmount, ok := parsedAmount.SetString(amount, 10)
	if !ok {
		return false
	}

	if parsedAmount.Cmp(bigZero) == -1 {
		return false
	}

	return true
}

func ValidateAddress(value string) (string, string) {
	typeAndAddress := strings.SplitN(value, ":", 2)
	if len(typeAndAddress) < 2 {
		return "", ""
	}
	clientAddressType := ""
	switch typeAndAddress[0] {
	case "ipv4":
		clientAddressType = "12"
	case "gns":
		clientAddressType = "41"
	default:
		clientAddressType = ""
	}
	return clientAddressType, typeAndAddress[1]
}

func ValidateInt(value string) bool {
	if _, err := strconv.Atoi(value); err != nil {
		return false
	}
	return true
}

func ValidateUUID(identifier string) bool {
	_, err := uuid.Parse(identifier)
	return err == nil
}
