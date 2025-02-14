package vtcpd

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/google/uuid"
)

type Result struct {
	UUID   uuid.UUID
	Code   int
	Tokens []string
	Error  error
}

// ResultFromRawInput parses a byte sequence into a Result struct.
// The expected format is:
// <36-byte UUID><tab><result code><tab>[optional tokens...]><trailing symbol>
//
// The function always returns a Result, even if parsing fails.
// In case of parsing errors, the Error field of the Result will be set and other fields may contain zero/nil values.
func ResultFromRawInput(body []byte) *Result {
	result := &Result{
		UUID: uuid.Nil, // Initialize with zero UUID
	}

	if len(body) < uuidHexLength {
		result.Error = errors.New("input too short to contain UUID")
		return result
	}

	// Parse UUID
	identifier, err := uuid.ParseBytes(body[:uuidHexLength])
	if err != nil {
		result.Error = fmt.Errorf("invalid UUID format: %w", err)
		return result
	}
	result.UUID = identifier

	// Check if there's content after UUID
	if len(body) <= uuidHexLength+1 {
		result.Error = errors.New("no content after UUID")
		return result
	}

	content := body[uuidHexLength+1:]
	if len(content) == 0 {
		result.Error = errors.New("empty content after UUID")
		return result
	}

	// Remove trailing symbol and split
	contentLen := len(content)
	tokens := strings.Split(string(content[:contentLen-1]), "\t")
	if len(tokens) == 0 {
		result.Error = errors.New("no tokens found after splitting")
		return result
	}

	// Parse code
	code, err := strconv.Atoi(tokens[0])
	if err != nil {
		result.Error = fmt.Errorf("invalid result code: %w", err)
		return result
	}

	result.Code = code
	result.Tokens = tokens[1:]
	return result
}
