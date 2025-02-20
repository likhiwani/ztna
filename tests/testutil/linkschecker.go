package testutil

import (
	"fmt"
	"sync"
	"time"

	"ztna-core/ztna/common/handler_common"
	"ztna-core/ztna/common/pb/ctrl_pb"
	"ztna-core/ztna/logtrace"

	"github.com/openziti/channel/v3"
	"github.com/openziti/foundation/v2/errorz"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

type TestLink struct {
	Id         string
	Src        string
	Dest       string
	FaultCount int
	Valid      bool
}

type LinkStateChecker struct {
	errorC chan error
	links  map[string]*TestLink
	req    *require.Assertions
	sync.Mutex
}

func (self *LinkStateChecker) reportError(err error) {
	logtrace.LogWithFunctionName()
	select {
	case self.errorC <- err:
	default:
	}
}

func (self *LinkStateChecker) HandleLink(msg *channel.Message, ch channel.Channel) {
	logtrace.LogWithFunctionName()
	self.Lock()
	defer self.Unlock()

	routerLinks := &ctrl_pb.RouterLinks{}
	if err := proto.Unmarshal(msg.Body, routerLinks); err != nil {
		self.reportError(err)
	}

	for _, link := range routerLinks.Links {
		testLink, ok := self.links[link.Id]
		if !ok {
			self.links[link.Id] = &TestLink{
				Id:    link.Id,
				Src:   ch.Id(),
				Dest:  link.DestRouterId,
				Valid: true,
			}
		} else {
			if testLink.Src != ch.Id() {
				self.reportError(fmt.Errorf("source router change for link %v => %v", testLink.Src, ch.Id()))
			}
			if testLink.Dest != link.DestRouterId {
				self.reportError(fmt.Errorf("dest router change for link %v => %v", testLink.Dest, link.DestRouterId))
			}
			testLink.Valid = true
		}
	}
}

func (self *LinkStateChecker) HandleFault(msg *channel.Message, _ channel.Channel) {
	logtrace.LogWithFunctionName()
	self.Lock()
	defer self.Unlock()

	fault := &ctrl_pb.Fault{}
	if err := proto.Unmarshal(msg.Body, fault); err != nil {
		select {
		case self.errorC <- err:
		default:
		}
	}

	if fault.Subject == ctrl_pb.FaultSubject_LinkFault || fault.Subject == ctrl_pb.FaultSubject_LinkDuplicate {
		if link, found := self.links[fault.Id]; found {
			link.FaultCount++
			link.Valid = false
		} else {
			self.reportError(errors.Errorf("no link with Id %s found", fault.Id))
		}
	}
}

func (self *LinkStateChecker) HandleOther(msg *channel.Message, _ channel.Channel) {
	logtrace.LogWithFunctionName()
	//  -33 = reconnect ping
	//    5 = heartbeat
	// 1007 = metrics message
	if msg.ContentType == -33 || msg.ContentType == 5 || msg.ContentType == 1007 {
		logrus.Debug("ignoring heartbeats, reconnect pings and metrics")
		return
	}

	self.reportError(errors.Errorf("unhandled msg of type %v received", msg.ContentType))
}

func (self *LinkStateChecker) RequireNoErrors() {
	logtrace.LogWithFunctionName()
	var errList errorz.MultipleErrors

	done := false
	for !done {
		select {
		case err := <-self.errorC:
			errList = append(errList, err)
		default:
			done = true
		}
	}

	if len(errList) > 0 {
		self.req.NoError(errList)
	}
}

func (self *LinkStateChecker) RequireOneActiveLink() *TestLink {
	logtrace.LogWithFunctionName()
	self.Lock()
	defer self.Unlock()

	var activeLink *TestLink

	for _, link := range self.links {
		if link.Valid {
			self.req.Nil(activeLink, "more than one active link found")
			activeLink = link
		}
	}
	self.req.NotNil(activeLink, "no active link found")
	return activeLink
}

func NewLinkChecker(assertions *require.Assertions) *LinkStateChecker {
	logtrace.LogWithFunctionName()
	checker := &LinkStateChecker{
		errorC: make(chan error, 4),
		links:  map[string]*TestLink{},
		req:    assertions,
	}
	return checker
}

func StartLinkTest(checker *LinkStateChecker, id string, uf channel.UnderlayFactory, assertions *require.Assertions) channel.Channel {
	logtrace.LogWithFunctionName()
	bindHandler := func(binding channel.Binding) error {
		binding.AddReceiveHandlerF(channel.AnyContentType, checker.HandleOther)
		binding.AddReceiveHandlerF(int32(ctrl_pb.ContentType_VerifyRouterType), func(msg *channel.Message, ch channel.Channel) {
			handler_common.SendSuccess(msg, ch, "link success")
		})
		binding.AddReceiveHandlerF(int32(ctrl_pb.ContentType_RouterLinksType), checker.HandleLink)
		binding.AddReceiveHandlerF(int32(ctrl_pb.ContentType_FaultType), checker.HandleFault)
		return nil
	}

	timeoutUF := NewTimeoutUnderlayFactory(uf, 2*time.Second)
	ch, err := channel.NewChannel(id, timeoutUF, channel.BindHandlerF(bindHandler), channel.DefaultOptions())
	assertions.NoError(err)
	return ch
}
