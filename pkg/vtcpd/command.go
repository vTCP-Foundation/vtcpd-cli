package vtcpd

import (
	"strings"

	"github.com/google/uuid"
)

type Command struct {
	UUID uuid.UUID
	Body string
}

func NewCommand(body ...string) *Command {
	return &Command{
		UUID: uuid.New(),
		Body: strings.Join(body, "\t"),
	}
}

func (c *Command) ToBytes() []byte {
	var b strings.Builder

	// Pre-allocate buffer with exact size needed
	b.Grow(uuidHexLength + 1 + len(c.Body) + 1)

	b.WriteString(c.UUID.String())
	b.WriteByte('\t')
	b.WriteString(c.Body)
	b.WriteByte('\n')
	return []byte(b.String())
}
