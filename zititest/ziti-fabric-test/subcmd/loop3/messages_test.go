package loop3

import (
	"bytes"
	"crypto/rand"
	"crypto/sha512"
	"reflect"
	"testing"
	"ztna-core/ztna/logtrace"
	loop3_pb "ztna-core/ztna/zititest/ziti-fabric-test/subcmd/loop3/pb"

	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/require"
)

type testPeer struct {
	bytes.Buffer
}

func (t *testPeer) Close() error {
	logtrace.LogWithFunctionName()
	return nil
}

func Test_MessageSerDeser(t *testing.T) {
	logtrace.LogWithFunctionName()
	req := require.New(t)
	data := make([]byte, 4192)
	_, err := rand.Read(data)
	req.NoError(err)

	hash := sha512.Sum512(data)

	block := &RandHashedBlock{
		Type:     BlockTypePlain,
		Sequence: 10,
		Hash:     hash[:],
		Data:     data,
	}

	testBuf := &testPeer{}

	p := &protocol{
		peer: testBuf,
		test: &loop3_pb.Test{
			Name: "test",
		},
	}

	req.NoError(block.Tx(p))

	readBlock := &RandHashedBlock{}
	req.NoError(readBlock.Rx(p))

	req.True(reflect.DeepEqual(block, readBlock), cmp.Diff(block, readBlock))

	data = make([]byte, 4192)
	_, err = rand.Read(data)
	req.NoError(err)

	hash = sha512.Sum512(data)

	block = &RandHashedBlock{
		Type:     BlockTypeLatencyRequest,
		Sequence: 10,
		Hash:     hash[:],
		Data:     data,
	}

	req.NoError(block.Tx(p))

	readBlock = &RandHashedBlock{}
	req.NoError(readBlock.Rx(p))

	req.Equal("", cmp.Diff(block, readBlock))
}
