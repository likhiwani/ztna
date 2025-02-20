package dns

import (
	"net"
	"ztna-core/ztna/logtrace"

	"github.com/michaelquigley/pfxlog"
)

type dummy struct{}

func (d dummy) AddHostname(_ string, _ net.IP) error {
	logtrace.LogWithFunctionName()
	pfxlog.Logger().Warnf("dummy resolver does not store hostname/ip mappings")
	return nil
}

func (d dummy) AddDomain(_ string, _ func(string) (net.IP, error)) error {
	logtrace.LogWithFunctionName()
	pfxlog.Logger().Warnf("dummy resolver does not store hostname/ip mappings")
	return nil
}

func (d dummy) Lookup(_ net.IP) (string, error) {
	logtrace.LogWithFunctionName()
	pfxlog.Logger().Warnf("dummy resolver does not store hostname/ip mappings")
	return "", nil
}

func (d dummy) LookupIP(_ string) (net.IP, bool) {
	logtrace.LogWithFunctionName()
	pfxlog.Logger().Warnf("dummy resolver does not store hostname/ip mappings")
	return nil, false
}

func (d dummy) RemoveHostname(_ string) net.IP {
	logtrace.LogWithFunctionName()
	return nil
}

func (d dummy) RemoveDomain(_ string) {
	logtrace.LogWithFunctionName()
}

func (d dummy) Cleanup() error {
	logtrace.LogWithFunctionName()
	return nil
}

func NewDummyResolver() Resolver {
	logtrace.LogWithFunctionName()
	pfxlog.Logger().Warnf("dummy resolver does not store hostname/ip mappings")
	return &dummy{}
}
