package handler

import (
	"errors"
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

// Parses the result from raw bytes sequence
// (often from the result.fifo file of some node)
//
// Returns result even if bytes sequence was not parsed correctly.
// In that case, the Error field of the result would be set.
func ResultFromRawInput(body []byte) *Result {
	UUID_HEX_LENGTH := 36
	if len(body) < UUID_HEX_LENGTH {
		return &Result{Error: errors.New("too short")}
	}

	uuidPart := body[:UUID_HEX_LENGTH]
	identifier, err := uuid.Parse(string(uuidPart))
	if err != nil {
		return &Result{
			UUID:  uuid.Nil,
			Error: errors.New("can't parse result UUID"),
		}
	}

	content := body[UUID_HEX_LENGTH+1:]
	contentWithoutTrailingSymbol := content[:len(content)-1]
	tokens := strings.Split(string(contentWithoutTrailingSymbol), string('\t'))
	code, err := strconv.Atoi(tokens[0])
	if err != nil {
		return &Result{
			UUID:  identifier,
			Error: errors.New("can't parse result code"),
		}
	}

	return &Result{
		UUID:   identifier,
		Code:   code,
		Tokens: tokens[1:],
	}
}
