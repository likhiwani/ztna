package xgress

import (
	"io"
	"sync"
	"ztna-core/ztna/logtrace"
)

type CloseHelper struct {
	closer    io.Closer
	closeLock sync.Mutex
}

func (self *CloseHelper) Init(closer io.Closer) {
	logtrace.LogWithFunctionName()
	if self == nil {
		return
	}

	self.closeLock.Lock()
	defer self.closeLock.Unlock()
	self.closer = closer
}

func (self *CloseHelper) Close() error {
	logtrace.LogWithFunctionName()
	if self == nil {
		return nil
	}

	self.closeLock.Lock()
	defer self.closeLock.Unlock()

	if self.closer != nil {
		result := self.closer.Close()
		self.closer = nil
		return result
	}
	return nil
}
