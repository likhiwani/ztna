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

package udp_vconn

import (
	"io"
	"net"
	"sync/atomic"
	"time"
	"ztna-core/ztna/logtrace"

	"github.com/michaelquigley/pfxlog"
	"github.com/openziti/foundation/v2/mempool"
	"github.com/sirupsen/logrus"
)

type udpConn struct {
	readC       chan mempool.PooledBuffer
	closeNotify chan struct{}
	service     string
	srcAddr     net.Addr
	manager     *manager
	writeConn   UDPWriterTo
	lastUse     atomic.Value
	closed      atomic.Bool
	leftOver    []byte
	leftOverBuf mempool.PooledBuffer
}

func (conn *udpConn) Service() string {
	logtrace.LogWithFunctionName()
	return conn.service
}

func (conn *udpConn) Accept(buffer mempool.PooledBuffer) {
	logtrace.LogWithFunctionName()
	logrus.WithField("udpConnId", conn.srcAddr.String()).Debugf("udp->ziti: queuing")
	select {
	case conn.readC <- buffer:
	case <-conn.closeNotify:
		buffer.Release()
		logrus.WithField("udpConnId", conn.srcAddr.String()).Debugf("udp->ziti: closed, cancelling accept")
	}
}

func (conn *udpConn) markUsed() {
	logtrace.LogWithFunctionName()
	conn.lastUse.Store(time.Now())
}

func (conn *udpConn) GetLastUsed() time.Time {
	logtrace.LogWithFunctionName()
	val := conn.lastUse.Load()
	return val.(time.Time)
}

func (conn *udpConn) WriteTo(w io.Writer) (n int64, err error) {
	logtrace.LogWithFunctionName()
	var bytesWritten int64
	for {
		var buf mempool.PooledBuffer

		select {
		case buf = <-conn.readC:
		case <-conn.closeNotify:
			select {
			case buf = <-conn.readC:
			default:
			}
		}

		if buf == nil {
			return bytesWritten, io.EOF
		}

		payload := buf.GetPayload()
		pfxlog.Logger().WithField("udpConnId", conn.srcAddr.String()).Debugf("udp->ziti: %v bytes", len(payload))
		n, err := w.Write(payload)
		buf.Release()
		conn.markUsed()
		bytesWritten += int64(n)
		if err != nil {
			return bytesWritten, err
		}
	}
}

func (conn *udpConn) Read(b []byte) (n int, err error) {
	logtrace.LogWithFunctionName()
	leftOverLen := len(conn.leftOver)
	if leftOverLen > 0 {
		copy(b, conn.leftOver)
		if leftOverLen > len(b) {
			conn.leftOver = conn.leftOver[len(b):]
			conn.markUsed()
			return len(b), nil
		}

		conn.leftOver = nil
		conn.leftOverBuf.Release()
		conn.leftOverBuf = nil

		conn.markUsed()
		return leftOverLen, nil
	}

	var bytesWritten int

	var buf mempool.PooledBuffer

	select {
	case buf = <-conn.readC:
	case <-conn.closeNotify:
		select {
		case buf = <-conn.readC:
		default:
		}
	}

	if buf == nil {
		conn.markUsed()
		return bytesWritten, io.EOF
	}

	data := buf.GetPayload()
	dataLen := len(data)
	copy(b, data)
	if dataLen <= len(b) {
		buf.Release()
		conn.markUsed()
		return dataLen, nil
	}

	conn.leftOver = data[len(b):]
	conn.leftOverBuf = buf
	conn.markUsed()

	return len(b), nil
}

func (conn *udpConn) Write(b []byte) (int, error) {
	logtrace.LogWithFunctionName()
	pfxlog.Logger().WithField("udpConnId", conn.srcAddr.String()).Debugf("ziti->udp: %v bytes", len(b))
	// TODO: UDP chunking, MTU chunking?
	n, err := conn.writeConn.WriteTo(b, conn.srcAddr)
	conn.markUsed()
	return n, err
}

func (conn *udpConn) Close() error {
	logtrace.LogWithFunctionName()
	if conn.closed.CompareAndSwap(false, true) {
		close(conn.closeNotify)
		if err := conn.writeConn.Close(); err != nil {
			logrus.WithField("service", conn.service).
				WithField("src_addr", conn.srcAddr).
				WithError(err).Error("error while closing udp connection")
		}
	}

	return nil
}

func (conn *udpConn) LocalAddr() net.Addr {
	logtrace.LogWithFunctionName()
	return conn.writeConn.LocalAddr()
}

func (conn *udpConn) RemoteAddr() net.Addr {
	logtrace.LogWithFunctionName()
	return conn.srcAddr
}

func (conn *udpConn) SetDeadline(time.Time) error {
	logtrace.LogWithFunctionName()
	// ignore, since this is a shared connection
	return nil
}

func (conn *udpConn) SetReadDeadline(time.Time) error {
	logtrace.LogWithFunctionName()
	// ignore, since this is a shared connection
	return nil
}

func (conn *udpConn) SetWriteDeadline(time.Time) error {
	logtrace.LogWithFunctionName()
	// ignore, since this is a shared connection
	return nil
}
