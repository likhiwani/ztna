package tunnel

import (
	"encoding/json"
	"io"
	"net"
	"ztna-core/edge-api/rest_model"
	"ztna-core/ztna/tunnel/health"

	"github.com/openziti/sdk-golang/ziti"
	"github.com/openziti/sdk-golang/ziti/edge"
	"github.com/sirupsen/logrus"
)

type HostingContext interface {
	ServiceId() string
	ServiceName() string
	ListenOptions() *ziti.ListenOptions
	Dial(options map[string]interface{}) (net.Conn, bool, error)
	GetHealthChecks() []health.CheckDefinition
	GetInitialHealthState() (ziti.Precedence, uint16)
	OnClose()
	SetCloseCallback(func())
}

type HostControl interface {
	io.Closer
	UpdateCost(cost uint16) error
	UpdatePrecedence(precedence edge.Precedence) error
	UpdateCostAndPrecedence(cost uint16, precedence edge.Precedence) error
	SendHealthEvent(pass bool) error
}

type FabricProvider interface {
	PrepForUse(serviceId string)
	GetCurrentIdentity() (*rest_model.IdentityDetail, error)
	GetCurrentIdentityWithBackoff() (*rest_model.IdentityDetail, error)
	TunnelService(service Service, identity string, conn net.Conn, halfClose bool, appInfo []byte) error
	HostService(hostCtx HostingContext) (HostControl, error)
}

func AppDataToMap(appData []byte) (map[string]interface{}, error) {
	result := map[string]interface{}{}
	if len(appData) != 0 {
		if err := json.Unmarshal(appData, &result); err != nil {
			return nil, err
		}
	}
	return result, nil
}

func NewContextProvider(context ziti.Context) FabricProvider {
	return &contextProvider{
		Context: context,
	}
}

type contextProvider struct {
	ziti.Context
}

// GetCurrentIdentity implements FabricProvider.
// Subtle: this method shadows the method (Context).GetCurrentIdentity of contextProvider.Context.
func (cp *contextProvider) GetCurrentIdentity() (*rest_model.IdentityDetail, error) {
	panic("unimplemented")
}

// GetCurrentIdentityWithBackoff implements FabricProvider.
// Subtle: this method shadows the method (Context).GetCurrentIdentityWithBackoff of contextProvider.Context.
func (cp *contextProvider) GetCurrentIdentityWithBackoff() (*rest_model.IdentityDetail, error) {
	panic("unimplemented")
}

func (cp *contextProvider) PrepForUse(serviceId string) {
	if _, err := cp.Context.GetSession(serviceId); err != nil {
		logrus.WithError(err).Error("failed to acquire network session")
	} else {
		logrus.Debug("acquired network session")
	}
}

func (cp *contextProvider) TunnelService(service Service, identity string, conn net.Conn, halfClose bool, appData []byte) error {
	options := &ziti.DialOptions{
		ConnectTimeout: service.GetDialTimeout(),
		AppData:        appData,
		Identity:       identity,
	}

	zitiConn, err := cp.Context.DialWithOptions(service.GetName(), options)
	if err != nil {
		return err
	}

	Run(zitiConn, conn, halfClose)
	return nil
}

func (cp *contextProvider) HostService(hostCtx HostingContext) (HostControl, error) {
	logger := logrus.WithField("service", hostCtx.ServiceName())
	listener, err := cp.Context.ListenWithOptions(hostCtx.ServiceName(), hostCtx.ListenOptions())
	if err != nil {
		logger.WithError(err).Error("error listening for service")
		return nil, err
	}

	go cp.accept(listener, hostCtx)

	return listener, nil
}

func (cp *contextProvider) accept(listener edge.Listener, hostCtx HostingContext) {
	defer hostCtx.OnClose()

	logger := logrus.WithField("service", hostCtx.ServiceName())
	for {
		logger.Info("hosting service, waiting for connections")
		conn, err := listener.AcceptEdge()
		if err != nil {
			logger.WithError(err).Error("closing listener for service")
			return
		}

		options, err := AppDataToMap(conn.GetAppData())
		if err != nil {
			logger.WithError(err).Error("dial failed")
			conn.CompleteAcceptFailed(err)
			if closeErr := conn.Close(); closeErr != nil {
				logger.WithError(closeErr).Error("close of ziti connection failed")
			}
			continue
		}

		externalConn, halfClose, err := hostCtx.Dial(options)
		if err != nil {
			logger.WithError(err).Error("dial failed")
			conn.CompleteAcceptFailed(err)
			if closeErr := conn.Close(); closeErr != nil {
				logger.WithError(closeErr).Error("close of ziti connection failed")
			}
			continue
		}

		log.Infof("successful connection %v->%v", conn.LocalAddr(), conn.RemoteAddr())

		if err := conn.CompleteAcceptSuccess(); err != nil {
			logger.WithError(err).Error("complete accept success failed")

			if closeErr := conn.Close(); closeErr != nil {
				logger.WithError(closeErr).Error("close of ziti connection failed")
			}

			if closeErr := externalConn.Close(); closeErr != nil {
				logger.WithError(closeErr).Error("close of external connection failed")
			}
			continue
		}

		go Run(conn, externalConn, halfClose)
	}
}
