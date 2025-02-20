package eid

import (
	"crypto/rand"
	"encoding/binary"
	logtrace "ztna-core/ztna/logtrace"

	"github.com/teris-io/shortid"
)

const Alphabet = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ.-"

var idGenerator *shortid.Shortid

func init() {
	logtrace.LogWithFunctionName()
	buf := make([]byte, 8)
	_, _ = rand.Read(buf)
	seed := binary.LittleEndian.Uint64(buf)
	idGenerator = shortid.MustNew(0, Alphabet, seed)
}

func New() string {
	logtrace.LogWithFunctionName()
	id, _ := idGenerator.Generate()
	return id
}
