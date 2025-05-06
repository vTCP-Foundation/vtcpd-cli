package handler

import (
	"strings"

	"github.com/google/uuid"
)

type Command struct {
	UUID uuid.UUID
	Body string
}

func NewCommand(body ...string) *Command {
	command := ""
	for _, token := range body {
		command += token + "\t"
	}

	// Removing trailing "\t"
	command = command[:len(command)-1]

	return &Command{
		UUID: uuid.Must(uuid.New(), nil),
		Body: command,
	}
}

func NewCommandWithUUID(transactionUUID uuid.UUID, body ...string) *Command {
	command := ""
	for _, token := range body {
		command += token + "\t"
	}

	// Removing trailing "\t"
	command = command[:len(command)-1]

	return &Command{
		UUID: transactionUUID,
		Body: command,
	}
}

func (c *Command) ToBytes() []byte {
	command := c.UUID.String()
	tokens := strings.Split(c.Body, "\t")
	for _, token := range tokens {
		command += string('\t') + token
	}
	command += string('\n')

	return []byte(command)
}
