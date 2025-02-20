package dns

import (
	"net"
	"ztna-core/ztna/logtrace"

	cmap "github.com/orcaman/concurrent-map/v2"
)

func NewRefCountingResolver(resolver Resolver) Resolver {
	logtrace.LogWithFunctionName()
	return &RefCountingResolver{
		names:   cmap.New[int](),
		wrapped: resolver,
	}
}

type RefCountingResolver struct {
	names   cmap.ConcurrentMap[string, int]
	wrapped Resolver
}

func (self *RefCountingResolver) Lookup(ip net.IP) (string, error) {
	logtrace.LogWithFunctionName()
	return self.wrapped.Lookup(ip)
}

func (self *RefCountingResolver) LookupIP(hostname string) (net.IP, bool) {
	logtrace.LogWithFunctionName()
	return self.wrapped.LookupIP(hostname)
}

func (self *RefCountingResolver) AddDomain(name string, cb func(string) (net.IP, error)) error {
	logtrace.LogWithFunctionName()
	return self.wrapped.AddDomain(name, cb)
}

func (self *RefCountingResolver) RemoveDomain(name string) {
	logtrace.LogWithFunctionName()
	self.wrapped.RemoveDomain(name)
}

func (self *RefCountingResolver) AddHostname(s string, ip net.IP) error {
	logtrace.LogWithFunctionName()
	err := self.wrapped.AddHostname(s, ip)
	if err != nil {
		self.names.Upsert(s, 1, func(exist bool, valueInMap int, newValue int) int {
			if exist {
				return valueInMap + 1
			}
			return 1
		})
	}
	return err
}

func (self *RefCountingResolver) RemoveHostname(s string) net.IP {
	logtrace.LogWithFunctionName()
	val := self.names.Upsert(s, 1, func(exist bool, valueInMap int, newValue int) int {
		if exist {
			return valueInMap - 1
		}
		return 0
	})

	if val == 0 {
		self.names.Remove(s)
		return self.wrapped.RemoveHostname(s)
	}
	return nil
}

func (self *RefCountingResolver) Cleanup() error {
	logtrace.LogWithFunctionName()
	return self.wrapped.Cleanup()
}
