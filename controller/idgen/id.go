package idgen

import (
	"crypto/rand"
	"encoding/binary"
	"math/big"
	logtrace "ztna-core/ztna/logtrace"

	"github.com/dineshappavoo/basex"
	"github.com/google/uuid"
	"github.com/teris-io/shortid"
)

const Alphabet = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ.-"

var idGenerator Generator

func NewGenerator() Generator {
	logtrace.LogWithFunctionName()
	buf := make([]byte, 8)
	_, _ = rand.Read(buf)
	seed := binary.LittleEndian.Uint64(buf)
	return &shortIdGenerator{
		Shortid: shortid.MustNew(0, Alphabet, seed),
	}
}

func init() {
	logtrace.LogWithFunctionName()
	idGenerator = NewGenerator()
}

func New() string {
	logtrace.LogWithFunctionName()
	id, _ := idGenerator.NextId()
	return id
}

type Generator interface {
	NextId() (string, error)
}

type shortIdGenerator struct {
	*shortid.Shortid
}

func (self *shortIdGenerator) NextId() (string, error) {
	logtrace.LogWithFunctionName()
	for {
		id, err := self.Generate()
		if err != nil {
			return "", err
		}
		if id[0] != '-' && id[0] != '.' {
			return id, nil
		}
	}
}

func NewUUIDString() string {
	logtrace.LogWithFunctionName()
	id := uuid.New()
	v := &big.Int{}
	v.SetBytes(id[:])
	result, _ := basex.EncodeInt(v)
	return result
}
