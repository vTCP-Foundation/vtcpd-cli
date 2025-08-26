package common

import (
	"fmt"
	"math"
	"strings"
)

// ParseAndValidateRealRate converts a real decimal rate to native (value, shift) format
func ParseAndValidateRealRate(realRate string, equivalentFrom string, equivalentTo string) (string, int16, error) {
	// Check for maximum 16 fractional digits
	parts := strings.Split(realRate, ".")
	if len(parts) == 2 && len(parts[1]) > 16 {
		return "", 0, fmt.Errorf("real_rate has more than 16 fractional digits")
	}

	// Parse the decimal value
	value, shift, err := NormalizeDecimal(realRate)
	if err != nil {
		return "", 0, fmt.Errorf("invalid real_rate format: %v", err)
	}

	// Apply scale difference
	decimalsFrom := DecimalsMap[equivalentFrom]
	decimalsTo := DecimalsMap[equivalentTo]
	adjustedShift := int64(shift) + int64(decimalsFrom-decimalsTo)

	// Validate int16 range
	if adjustedShift < math.MinInt16 || adjustedShift > math.MaxInt16 {
		return "", 0, fmt.Errorf("shift is out of int16 range after scale adjustment")
	}

	return value, int16(adjustedShift), nil
}

// NormalizeDecimal converts decimal string to (value, shift) format
func NormalizeDecimal(decimal string) (string, int16, error) {
	// Remove leading/trailing whitespace
	decimal = strings.TrimSpace(decimal)

	// Handle empty string
	if decimal == "" {
		return "0", 0, nil
	}

	// Handle negative sign
	negative := false
	if strings.HasPrefix(decimal, "-") {
		negative = true
		decimal = decimal[1:]
	} else if strings.HasPrefix(decimal, "+") {
		decimal = decimal[1:]
	}

	// Split by decimal point
	parts := strings.Split(decimal, ".")
	if len(parts) > 2 {
		return "", 0, fmt.Errorf("invalid decimal format")
	}

	var integerPart, fractionalPart string
	if len(parts) == 1 {
		integerPart = parts[0]
		fractionalPart = ""
	} else {
		integerPart = parts[0]
		fractionalPart = parts[1]
	}

	// Handle empty integer part
	if integerPart == "" {
		integerPart = "0"
	}

	// Remove leading zeros from integer part
	integerPart = strings.TrimLeft(integerPart, "0")
	if integerPart == "" {
		integerPart = "0"
	}

	// Remove trailing zeros from fractional part
	fractionalPart = strings.TrimRight(fractionalPart, "0")

	// Build the value
	value := integerPart + fractionalPart
	shift := len(fractionalPart)

	// Remove leading zeros from final value
	value = strings.TrimLeft(value, "0")
	if value == "" {
		value = "0"
		shift = 0
	}

	// Add negative sign if needed
	if negative && value != "0" {
		value = "-" + value
	}

	return value, int16(shift), nil
}

// ComputeRealRateString computes the real decimal rate from native (value, shift) format
func ComputeRealRateString(value string, shift int16, equivalentFrom string, equivalentTo string) string {
	// Reverse scale adjustment
	decimalsFrom := DecimalsMap[equivalentFrom]
	decimalsTo := DecimalsMap[equivalentTo]
	originalShift := int(shift) - (decimalsFrom - decimalsTo)

	return ApplyShiftToValue(value, originalShift)
}

// ApplyShiftToValue applies decimal shift to a value string
func ApplyShiftToValue(value string, shift int) string {
	if value == "0" {
		return "0"
	}

	negative := false
	if strings.HasPrefix(value, "-") {
		negative = true
		value = value[1:]
	}

	// If shift is 0, return as is
	if shift == 0 {
		if negative {
			return "-" + value
		}
		return value
	}

	// If shift is positive, we need to place decimal point within the number
	if shift > 0 {
		if len(value) <= shift {
			// Need to add leading zeros
			zerosNeeded := shift - len(value) + 1
			result := "0." + strings.Repeat("0", zerosNeeded-1) + value
			if negative {
				return "-" + result
			}
			return result
		} else {
			// Place decimal point within the number
			integerPart := value[:len(value)-shift]
			fractionalPart := value[len(value)-shift:]
			result := integerPart + "." + fractionalPart
			if negative {
				return "-" + result
			}
			return result
		}
	} else {
		// shift is negative, add zeros to the right
		result := value + strings.Repeat("0", -shift)
		if negative {
			return "-" + result
		}
		return result
	}
}