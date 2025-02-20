package mesh

import (
	"runtime"
	"testing"
	"time"
	"ztna-core/ztna/controller/event"
	"ztna-core/ztna/logtrace"

	"github.com/openziti/foundation/v2/versions"

	"github.com/stretchr/testify/assert"
)

func Test_checkState_ReadonlyFalseWhenAllVersionsMatch(t *testing.T) {
	logtrace.LogWithFunctionName()
	m := &impl{
		Peers: map[string]*Peer{
			"1": {Version: testVersion("1"), Address: "1"},
			"2": {Version: testVersion("1"), Address: "2"},
		},
		version: NewVersionProviderTest(),
	}

	m.updateClusterState()
	assert.Equal(t, false, m.readonly.Load(), "Expected readonly to be false, got ", m.readonly.Load())
}

func Test_checkState_ReadonlyTrueWhenAllVersionsDoNotMatch(t *testing.T) {
	logtrace.LogWithFunctionName()
	m := &impl{
		Peers: map[string]*Peer{
			"1": {Version: testVersion("dne"), Address: "1"},
			"2": {Version: testVersion("dne"), Address: "2"},
		},
		version: NewVersionProviderTest(),
	}

	m.updateClusterState()
	assert.Equal(t, true, m.readonly.Load(), "Expected readonly to be true, got ", m.readonly.Load())
}

func Test_checkState_ReadonlySetToFalseWhenPreviouslyTrueAndAllVersionsNowMatch(t *testing.T) {
	logtrace.LogWithFunctionName()
	m := &impl{
		Peers: map[string]*Peer{
			"1": {Version: testVersion("1"), Address: "1"},
			"2": {Version: testVersion("1"), Address: "2"},
		},
		version: NewVersionProviderTest(),
	}
	m.readonly.Store(true)

	m.updateClusterState()
	assert.Equal(t, false, m.readonly.Load(), "Expected readonly to be false, got ", m.readonly.Load())
}

func Test_AddPeer_PassesReadonlyWhenVersionsMatch(t *testing.T) {
	logtrace.LogWithFunctionName()
	m := &impl{
		Peers:           map[string]*Peer{},
		version:         NewVersionProviderTest(),
		eventDispatcher: event.DispatcherMock{},
	}

	p := &Peer{Version: testVersion("1")}

	assert.NoError(t, m.PeerConnected(p, true))
	assert.Equal(t, false, m.readonly.Load(), "Expected readonly to be false, got ", m.readonly.Load())
}

func Test_AddPeer_TurnsReadonlyWhenVersionsDoNotMatch(t *testing.T) {
	logtrace.LogWithFunctionName()
	m := &impl{
		Peers:           map[string]*Peer{},
		version:         NewVersionProviderTest(),
		eventDispatcher: event.DispatcherMock{},
	}

	p := &Peer{Version: testVersion("dne")}

	assert.NoError(t, m.PeerConnected(p, true))
	assert.Equal(t, true, m.readonly.Load(), "Expected readonly to be true, got ", m.readonly.Load())
}

func Test_RemovePeer_StaysReadonlyWhenDeletingPeerAndStillHasMismatchedVersions(t *testing.T) {
	logtrace.LogWithFunctionName()
	m := &impl{
		Peers: map[string]*Peer{
			"1": {Version: testVersion("dne"), Address: "1"},
			"2": {Version: testVersion("dne"), Address: "2"},
		},
		version:         NewVersionProviderTest(),
		eventDispatcher: event.DispatcherMock{},
	}
	m.readonly.Store(true)

	m.PeerDisconnected(m.Peers["1"])
	assert.Equal(t, true, m.readonly.Load(), "Expected readonly to be true, got ", m.readonly.Load())
}

func Test_RemovePeer_RemovesReadonlyWhenDeletingPeerWithNoOtherMismatches(t *testing.T) {
	logtrace.LogWithFunctionName()
	m := &impl{
		Peers: map[string]*Peer{
			"1": {Version: testVersion("dne"), Address: "1"},
			"2": {Version: testVersion("1"), Address: "2"},
		},
		version:         NewVersionProviderTest(),
		eventDispatcher: event.DispatcherMock{},
	}
	m.readonly.Store(true)

	m.PeerDisconnected(m.Peers["1"])
	assert.Equal(t, false, m.readonly.Load(), "Expected readonly to be false, got ", m.readonly.Load())
}

func Test_RemovePeer_RemovesReadonlyWhenDeletingLastPeer(t *testing.T) {
	logtrace.LogWithFunctionName()
	m := impl{
		Peers: map[string]*Peer{
			"1": {Version: testVersion("dne"), Address: "1"},
		},
		version:         NewVersionProviderTest(),
		eventDispatcher: event.DispatcherMock{},
	}
	m.readonly.Store(true)

	m.PeerDisconnected(m.Peers["1"])
	assert.Equal(t, false, m.readonly.Load(), "Expected readonly to be false, got ", m.readonly.Load())
}

func testVersion(v string) *versions.VersionInfo {
	logtrace.LogWithFunctionName()
	return &versions.VersionInfo{Version: v}
}

type VersionProviderTest struct {
}

func (v VersionProviderTest) Branch() string {
	logtrace.LogWithFunctionName()
	return "local"
}

func (v VersionProviderTest) EncoderDecoder() versions.VersionEncDec {
	logtrace.LogWithFunctionName()
	return &versions.StdVersionEncDec
}

func (v VersionProviderTest) Version() string {
	logtrace.LogWithFunctionName()
	return "1"
}

func (v VersionProviderTest) BuildDate() string {
	logtrace.LogWithFunctionName()
	return time.Now().String()
}

func (v VersionProviderTest) Revision() string {
	logtrace.LogWithFunctionName()
	return ""
}

func (v VersionProviderTest) AsVersionInfo() *versions.VersionInfo {
	logtrace.LogWithFunctionName()
	return &versions.VersionInfo{
		Version:   v.Version(),
		Revision:  v.Revision(),
		BuildDate: v.BuildDate(),
		OS:        runtime.GOOS,
		Arch:      runtime.GOARCH,
	}
}

func NewVersionProviderTest() versions.VersionProvider {
	logtrace.LogWithFunctionName()
	return &VersionProviderTest{}
}
