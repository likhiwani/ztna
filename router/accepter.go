package router

import (
	"ztna-core/ztna/logtrace"
	"ztna-core/ztna/router/forwarder"
	"ztna-core/ztna/router/xlink"

	"github.com/sirupsen/logrus"
)

func newXlinkAccepter(f *forwarder.Forwarder) xlink.Acceptor {
	logtrace.LogWithFunctionName()
	return &xlinkAccepter{
		forwarder: f,
	}
}

func (self *xlinkAccepter) Accept(xlink xlink.Xlink) error {
	logtrace.LogWithFunctionName()
	if err := self.forwarder.RegisterLink(xlink); err != nil {
		return err
	}
	logrus.WithField("linkId", xlink.Id()).
		WithField("destId", xlink.DestinationId()).
		WithField("iteration", xlink.Iteration()).
		WithField("dialed", xlink.IsDialed()).
		Info("accepted new link")
	return nil
}

type xlinkAccepter struct {
	forwarder *forwarder.Forwarder
}
