/*
	Copyright NetFoundry Inc.

	Licensed under the Apache License, Version 2.0 (the "License");
	you may not use this file except in compliance with the License.
	You may obtain a copy of the License at

	https://www.apache.org/licenses/LICENSE-2.0

	Unless required by applicable law or agreed to in writing, software
	distributed under the License is distributed on an "AS IS" BASIS,
	WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
	See the License for the specific language governing permissions and
	limitations under the License.
*/

package intercept

import (
	"encoding/json"
	"errors"
	"io"
	"net"
	"strings"
	"time"
	"ztna-core/ztna/logtrace"

	"github.com/michaelquigley/pfxlog"
)

// record type
const (
	A   = 1
	MX  = 15
	TXT = 16
	SRV = 33
)

// response code
const (
	NOERROR  = 0
	FORMERR  = 1
	SERVFAIL = 2
	NXDOMAIN = 3
	NOTIMP   = 4
	REFUSED  = 5
)

type DnsQuestion struct {
	Name string `json:"name"`
	Type int    `json:"type"`
}

type DnsAnswer struct {
	Name     string `json:"name"`
	Type     int    `json:"type"`
	TTL      int    `json:"ttl"`
	Data     string `json:"data"`
	Port     uint16 `json:"port,omitempty"`
	Priority uint16 `json:"priority,omitempty"`
	Weight   uint16 `json:"weight,omitempty"`
}

type DnsMessage struct {
	Id       int            `json:"id"`
	Status   int            `json:"status"`
	Question []*DnsQuestion `json:"question"`
	Answer   []*DnsAnswer   `json:"answer"`
	Comment  string         `json:"comment,omitempty"`
}

type resolvConnAddr struct {
	service string
}

func (r *resolvConnAddr) Network() string {
	logtrace.LogWithFunctionName()
	return "ziti-resolve"
}

func (r *resolvConnAddr) String() string {
	logtrace.LogWithFunctionName()
	return r.service
}

type resolvConn struct {
	ctx *hostingContext

	respQueue chan *DnsMessage
	closed    bool
}

func newResolvConn(hostCtx *hostingContext) (net.Conn, bool, error) {
	logtrace.LogWithFunctionName()
	log := pfxlog.Logger().WithField("service", hostCtx.service.Name)
	log.Infof("starting resolver connection")
	r := &resolvConn{ctx: hostCtx, respQueue: make(chan *DnsMessage, 16)}
	return r, false, nil
}

func (r *resolvConn) Read(b []byte) (n int, err error) {
	logtrace.LogWithFunctionName()

	msg, ok := <-r.respQueue
	if !ok {
		return 0, io.EOF
	}

	msgBytes, err := json.Marshal(msg)
	if err != nil {
		return 0, err
	}
	if len(msgBytes) > len(b) {
		return 0, errors.New("short buffer")
	}
	return copy(b, msgBytes), nil
}

func (r *resolvConn) Write(b []byte) (int, error) {
	logtrace.LogWithFunctionName()
	log := pfxlog.Logger().WithField("service", r.ctx.service.Name)
	dnsMessage := &DnsMessage{}
	var matchName string
	dnsMatch := false
	var q *DnsQuestion

	err := json.Unmarshal(b, dnsMessage)
	if err != nil {
		dnsMessage.Status = FORMERR
		goto done
	}

	if len(dnsMessage.Question) != 1 {
		dnsMessage.Status = FORMERR
		goto done
	}

	q = dnsMessage.Question[0]

	matchName = strings.TrimSuffix(q.Name, ".")
	log.WithField("name", matchName).WithField("type", q.Type).Info("resolving")
	for _, allowed := range r.ctx.config.GetAllowedAddresses() {
		if allowed.Allows(matchName) {
			dnsMatch = true
			break
		}
	}

	if !dnsMatch {
		dnsMessage.Status = NXDOMAIN
		goto done
	}

	switch q.Type {
	case SRV:
		query := strings.SplitN(q.Name, ".", 3)
		if len(query) < 3 {
			dnsMessage.Status = FORMERR
			goto done
		}
		_, srvs, err := net.LookupSRV(query[0][1:], query[1][1:], query[2])
		if err != nil {
			dnsMessage.Comment = err.Error()
			dnsMessage.Status = SERVFAIL
			goto done
		}

		dnsMessage.Status = NOERROR
		for _, srv := range srvs {
			ans := &DnsAnswer{
				Name:     q.Name,
				Type:     q.Type,
				Data:     srv.Target,
				Port:     srv.Port,
				Priority: srv.Priority,
				Weight:   srv.Weight,
				TTL:      86400,
			}
			dnsMessage.Answer = append(dnsMessage.Answer, ans)
		}

	case TXT:
		txts, err := net.LookupTXT(q.Name)
		if err != nil {
			dnsMessage.Comment = err.Error()
			dnsMessage.Status = SERVFAIL
			goto done
		}
		dnsMessage.Status = NOERROR
		for _, txt := range txts {
			ans := &DnsAnswer{
				Name: q.Name,
				Type: q.Type,
				Data: txt,
				TTL:  86400,
			}
			dnsMessage.Answer = append(dnsMessage.Answer, ans)
		}

	case MX:
		mxs, err := net.LookupMX(q.Name)
		log.Infof("resolved %d MX recs, err=%v", len(mxs), err)
		if err != nil {
			dnsMessage.Comment = err.Error()
			dnsMessage.Status = SERVFAIL
			goto done
		}
		dnsMessage.Status = NOERROR
		for _, mx := range mxs {
			ans := &DnsAnswer{
				Name:     q.Name,
				Type:     q.Type,
				Data:     mx.Host,
				Priority: mx.Pref,
				TTL:      86400,
			}
			dnsMessage.Answer = append(dnsMessage.Answer, ans)
		}

	default:
		dnsMessage.Status = NOTIMP
	}

done:
	r.respQueue <- dnsMessage
	return len(b), nil
}

func (r *resolvConn) Close() error {
	logtrace.LogWithFunctionName()
	if r.closed {
		return nil
	}
	r.closed = true
	log := pfxlog.Logger().WithField("service", r.ctx.ServiceName())
	log.Infof("resolver connection closed")
	close(r.respQueue)
	return nil
}

func (r *resolvConn) LocalAddr() net.Addr {
	logtrace.LogWithFunctionName()
	return &resolvConnAddr{
		service: r.ctx.ServiceName(),
	}
}

func (r *resolvConn) RemoteAddr() net.Addr {
	logtrace.LogWithFunctionName()
	return &resolvConnAddr{
		service: r.ctx.ServiceName(),
	}
}

func (r *resolvConn) SetDeadline(t time.Time) error {
	logtrace.LogWithFunctionName()
	pfxlog.Logger().Warn("not implemented")
	return nil
}

func (r *resolvConn) SetReadDeadline(t time.Time) error {
	logtrace.LogWithFunctionName()
	pfxlog.Logger().Warn("not implemented")
	return nil
}

func (r *resolvConn) SetWriteDeadline(t time.Time) error {
	logtrace.LogWithFunctionName()
	pfxlog.Logger().Warn("not implemented")
	return nil
}
