package idgeneration

import (
	"github.com/gofrs/uuid/v5"
	"github.com/rs/xid"
)

type IdGenerator struct{}

func CreateIdGenerator() *IdGenerator {
	return &IdGenerator{}
}

func (g *IdGenerator) GenerateUUIDv7() string {
	return uuid.Must(uuid.NewV7()).String()
}

func (g *IdGenerator) GenerateConfirmationCode() string {
	return xid.New().String()
}
