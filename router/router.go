/*
	(c) Copyright NetFoundry Inc.

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

package router

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path"
	"plugin"
	"runtime/debug"
	"sync/atomic"
	"time"
	"ztna-core/ztna/common"
	"ztna-core/ztna/controller/command"
	"ztna-core/ztna/logtrace"
	"ztna-core/ztna/router/state"

	"github.com/openziti/foundation/v2/rate"

	"ztna-core/ztna/common/config"
	"ztna-core/ztna/router/link"

	"github.com/openziti/foundation/v2/debugz"
	"github.com/openziti/foundation/v2/goroutines"
	"github.com/openziti/xweb/v2"
	metrics2 "github.com/rcrowley/go-metrics"

	"ztna-core/ztna/common/health"
	fabricMetrics "ztna-core/ztna/common/metrics"
	"ztna-core/ztna/common/pb/ctrl_pb"
	"ztna-core/ztna/common/profiler"
	"ztna-core/ztna/router/env"
	"ztna-core/ztna/router/forwarder"
	"ztna-core/ztna/router/handler_ctrl"
	"ztna-core/ztna/router/handler_link"
	"ztna-core/ztna/router/handler_xgress"
	"ztna-core/ztna/router/xgress"
	"ztna-core/ztna/router/xgress_proxy"
	"ztna-core/ztna/router/xgress_proxy_udp"
	"ztna-core/ztna/router/xgress_transport"
	"ztna-core/ztna/router/xgress_transport_udp"
	"ztna-core/ztna/router/xlink"
	"ztna-core/ztna/router/xlink_transport"

	gosundheit "github.com/AppsFlyer/go-sundheit"
	"github.com/AppsFlyer/go-sundheit/checks"
	"github.com/michaelquigley/pfxlog"
	"github.com/openziti/channel/v3"
	"github.com/openziti/foundation/v2/concurrenz"
	"github.com/openziti/foundation/v2/errorz"
	"github.com/openziti/foundation/v2/versions"
	"github.com/openziti/identity"
	"github.com/openziti/metrics"
	"github.com/openziti/transport/v2"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"google.golang.org/protobuf/proto"
	"gopkg.in/yaml.v3"
)

type Router struct {
	config              *Config
	ctrls               env.NetworkControllers
	ctrlBindhandler     channel.BindHandler
	faulter             *forwarder.Faulter
	forwarder           *forwarder.Forwarder
	xrctrls             []env.Xrctrl
	xlinkFactories      map[string]xlink.Factory
	xlinkListeners      []xlink.Listener
	xlinkDialers        []xlink.Dialer
	xlinkRegistry       xlink.Registry
	xgressListeners     []xgress.Listener
	linkDialerPool      goroutines.Pool
	rateLimiterPool     goroutines.Pool
	ctrlRateLimiter     rate.AdaptiveRateLimitTracker
	metricsRegistry     metrics.UsageRegistry
	shutdownC           chan struct{}
	shutdownDoneC       chan struct{}
	isShutdown          atomic.Bool
	metricsReporter     metrics.Handler
	versionProvider     versions.VersionProvider
	debugOperations     map[byte]func(c *bufio.ReadWriter) error
	stateManager        state.Manager
	rdmEnabled          *config.Value[bool]
	xwebs               []xweb.Instance
	xwebFactoryRegistry xweb.Registry
	agentBindHandlers   []channel.BindHandler
}

func (self *Router) GetRouterId() *identity.TokenId {
	logtrace.LogWithFunctionName()
	return self.config.Id
}

func (self *Router) GetNetworkControllers() env.NetworkControllers {
	logtrace.LogWithFunctionName()
	return self.ctrls
}

func (self *Router) GetDialerCfg() map[string]xgress.OptionsData {
	logtrace.LogWithFunctionName()
	return self.config.Dialers
}

func (self *Router) GetXlinkDialers() []xlink.Dialer {
	logtrace.LogWithFunctionName()
	return self.xlinkDialers
}

func (self *Router) GetXrctrls() []env.Xrctrl {
	logtrace.LogWithFunctionName()
	return self.xrctrls
}

func (self *Router) GetTraceHandler() *channel.TraceHandler {
	logtrace.LogWithFunctionName()
	return self.config.Trace.Handler
}

func (self *Router) GetXlinkRegistry() xlink.Registry {
	logtrace.LogWithFunctionName()
	return self.xlinkRegistry
}

func (self *Router) GetCloseNotify() <-chan struct{} {
	logtrace.LogWithFunctionName()
	return self.shutdownC
}

func (self *Router) GetMetricsRegistry() metrics.UsageRegistry {
	logtrace.LogWithFunctionName()
	return self.metricsRegistry
}

func (self *Router) RenderJsonConfig() (string, error) {
	logtrace.LogWithFunctionName()
	jsonMap, err := config.ToJsonCompatibleMap(self.config.src)
	delete(jsonMap, FlagsCfgMapKey)
	if err != nil {
		return "", err
	}
	b, err := json.Marshal(jsonMap)
	return string(b), err
}

func (self *Router) GetChannel(controllerId string) channel.Channel {
	logtrace.LogWithFunctionName()
	return self.ctrls.GetCtrlChannel(controllerId)
}

func (self *Router) DefaultRequestTimeout() time.Duration {
	logtrace.LogWithFunctionName()
	return self.config.Ctrl.DefaultRequestTimeout
}

func (self *Router) GetHeartbeatOptions() env.HeartbeatOptions {
	logtrace.LogWithFunctionName()
	return self.config.Ctrl.Heartbeats
}

func (self *Router) GetStateManager() state.Manager {
	logtrace.LogWithFunctionName()
	return self.stateManager
}

func (self *Router) GetRouterDataModel() *common.RouterDataModel {
	logtrace.LogWithFunctionName()
	return self.stateManager.RouterDataModel()
}

func (self *Router) IsRouterDataModelEnabled() bool {
	logtrace.LogWithFunctionName()
	return self.rdmEnabled.Load()
}

func (self *Router) GetRouterDataModelEnabledConfig() *config.Value[bool] {
	logtrace.LogWithFunctionName()
	return self.rdmEnabled
}

func (self *Router) GetConnectEventsConfig() *env.ConnectEventsConfig {
	logtrace.LogWithFunctionName()
	return &self.config.ConnectEvents
}

func Create(cfg *Config, versionProvider versions.VersionProvider) *Router {
	logtrace.LogWithFunctionName()
	closeNotify := make(chan struct{})

	if cfg.Metrics.IntervalAgeThreshold != 0 {
		metrics.SetIntervalAgeThreshold(cfg.Metrics.IntervalAgeThreshold)
		logrus.Infof("set interval age threshold to '%v'", cfg.Metrics.IntervalAgeThreshold)
	}
	env.IntervalSize = cfg.Metrics.ReportInterval
	metricsRegistry := metrics.NewUsageRegistry(cfg.Id.Token, map[string]string{}, closeNotify)
	xgress.InitMetrics(metricsRegistry)

	linkDialerPoolConfig := goroutines.PoolConfig{
		QueueSize:   uint32(cfg.Forwarder.LinkDial.QueueLength),
		MinWorkers:  0,
		MaxWorkers:  uint32(cfg.Forwarder.LinkDial.WorkerCount),
		IdleTime:    30 * time.Second,
		CloseNotify: closeNotify,
		PanicHandler: func(err interface{}) {
			pfxlog.Logger().WithField(logrus.ErrorKey, err).WithField("backtrace", string(debug.Stack())).Error("panic during link dial")
		},
	}

	fabricMetrics.ConfigureGoroutinesPoolMetrics(&linkDialerPoolConfig, metricsRegistry, "pool.link.dialer")

	linkDialerPool, err := goroutines.NewPool(linkDialerPoolConfig)
	if err != nil {
		panic(errors.Wrap(err, "error creating link dialer pool"))
	}

	router := &Router{
		config:              cfg,
		metricsRegistry:     metricsRegistry,
		shutdownC:           closeNotify,
		shutdownDoneC:       make(chan struct{}),
		versionProvider:     versionProvider,
		debugOperations:     map[byte]func(c *bufio.ReadWriter) error{},
		xwebFactoryRegistry: xweb.NewRegistryMap(),
		linkDialerPool:      linkDialerPool,
		ctrlRateLimiter:     command.NewAdaptiveRateLimitTracker(cfg.Ctrl.RateLimit, metricsRegistry, closeNotify),
		rdmEnabled:          config.NewConfigValue[bool](),
	}

	router.ctrls = env.NewNetworkControllers(cfg.Ctrl.DefaultRequestTimeout, router.connectToController, &cfg.Ctrl.Heartbeats)
	router.stateManager = state.NewManager(router)

	router.xlinkRegistry = link.NewLinkRegistry(router)
	router.faulter = forwarder.NewFaulter(router.ctrls, cfg.Forwarder.FaultTxInterval, closeNotify)
	router.forwarder = forwarder.NewForwarder(metricsRegistry, router.faulter, cfg.Forwarder, closeNotify)
	router.forwarder.StartScanner(router.ctrls)

	xgress.InitPayloadIngester(closeNotify)
	xgress.InitAcker(router.forwarder, metricsRegistry, closeNotify)
	xgress.InitRetransmitter(router.forwarder, router.forwarder, metricsRegistry, closeNotify)

	router.ctrlBindhandler, err = handler_ctrl.NewBindHandler(router, router.forwarder, router)
	if err != nil {
		panic(err)
	}

	return router
}

func (self *Router) RegisterXrctrl(x env.Xrctrl) error {
	logtrace.LogWithFunctionName()
	if err := self.config.Configure(x); err != nil {
		return err
	}
	if x.Enabled() {
		self.xrctrls = append(self.xrctrls, x)
	}
	return nil
}

func (self *Router) GetVersionInfo() versions.VersionProvider {
	logtrace.LogWithFunctionName()
	return self.versionProvider
}

func (self *Router) GetConfig() *Config {
	logtrace.LogWithFunctionName()
	return self.config
}

func (self *Router) Start() error {
	logtrace.LogWithFunctionName()
	if err := os.MkdirAll(self.config.Ctrl.DataDir, 0700); err != nil {
		logrus.WithField("dir", self.config.Ctrl.DataDir).WithError(err).Error("failed to initialize data directory")
		return err
	}

	self.showOptions()
	if err := self.initRateLimiterPool(); err != nil {
		return err
	}
	self.startProfiling()

	healthChecker, err := self.initializeHealthChecks()
	if err != nil {
		logrus.WithError(err).Fatalf("failed to create health checker")
	}

	if err := self.RegisterXWebHandlerFactory(health.NewHealthCheckApiFactory(healthChecker)); err != nil {
		logrus.WithError(err).Fatalf("failed to create health checks api factory")
	}

	if err := self.registerComponents(); err != nil {
		return err
	}

	if err := self.registerPlugins(); err != nil {
		return err
	}

	self.startXlinkDialers()
	self.startXlinkListeners()
	self.setDefaultDialerBindings()
	self.startXgressListeners()

	for _, web := range self.xwebs {
		go web.Run()
	}

	if err = self.startControlPlane(); err != nil {
		return err
	}
	return nil
}

func (self *Router) Shutdown() error {
	logtrace.LogWithFunctionName()
	var errs []error
	if self.isShutdown.CompareAndSwap(false, true) {
		if err := self.ctrls.Close(); err != nil {
			errs = append(errs, err)
		}

		close(self.shutdownC)

		for _, xlinkListener := range self.xlinkListeners {
			if err := xlinkListener.Close(); err != nil {
				errs = append(errs, err)
			}
		}

		self.xlinkRegistry.Shutdown()

		for _, xgressListener := range self.xgressListeners {
			if err := xgressListener.Close(); err != nil {
				errs = append(errs, err)
			}
		}

		for _, web := range self.xwebs {
			go web.Shutdown()
		}

		close(self.shutdownDoneC)
	}
	if len(errs) == 0 {
		return nil
	}
	if len(errs) == 1 {
		return errs[0]
	}
	return errorz.MultipleErrors(errs)
}

func (self *Router) Run() error {
	logtrace.LogWithFunctionName()
	if err := self.Start(); err != nil {
		return err
	}

	<-self.shutdownDoneC
	return nil
}

func (self *Router) showOptions() {
	logtrace.LogWithFunctionName()
	if output, err := json.Marshal(self.config.Ctrl.Options); err == nil {
		pfxlog.Logger().Infof("ctrl = %s", string(output))
	} else {
		logrus.Fatalf("unable to display options (%v)", err)
	}

	if output, err := json.Marshal(self.config.Metrics); err == nil {
		pfxlog.Logger().Infof("metrics = %s", string(output))
	} else {
		logrus.Fatalf("unable to display options (%v)", err)
	}
}

func (self *Router) startProfiling() {
	logtrace.LogWithFunctionName()
	if self.config.Profile.Memory.Path != "" {
		go profiler.NewMemoryWithShutdown(self.config.Profile.Memory.Path, self.config.Profile.Memory.Interval, self.shutdownC).Run()
	}
	if self.config.Profile.CPU.Path != "" {
		if cpu, err := profiler.NewCPUWithShutdown(self.config.Profile.CPU.Path, self.shutdownC); err == nil {
			go cpu.Run()
		} else {
			logrus.Errorf("unexpected error launching cpu profiling (%v)", err)
		}
	}
	go newRouterMonitor(self.forwarder, self.shutdownC).Monitor()
}

func (self *Router) initRateLimiterPool() error {
	logtrace.LogWithFunctionName()
	rateLimiterPoolConfig := goroutines.PoolConfig{
		QueueSize:   uint32(self.forwarder.Options.RateLimiter.QueueLength),
		MinWorkers:  0,
		MaxWorkers:  uint32(self.forwarder.Options.RateLimiter.WorkerCount),
		IdleTime:    30 * time.Second,
		CloseNotify: self.GetCloseNotify(),
		PanicHandler: func(err interface{}) {
			pfxlog.Logger().WithField(logrus.ErrorKey, err).WithField("backtrace", string(debug.Stack())).Error("panic during rate limited operation")
		},
	}

	fabricMetrics.ConfigureGoroutinesPoolMetrics(&rateLimiterPoolConfig, self.GetMetricsRegistry(), "pool.rate_limiter")

	rateLimiterPool, err := goroutines.NewPool(rateLimiterPoolConfig)
	if err != nil {
		return errors.Wrap(err, "error creating rate limited pool")
	}

	self.rateLimiterPool = rateLimiterPool
	return nil
}

func (self *Router) GetLinkDialerPool() goroutines.Pool {
	logtrace.LogWithFunctionName()
	return self.linkDialerPool
}

func (self *Router) GetRateLimiterPool() goroutines.Pool {
	logtrace.LogWithFunctionName()
	return self.rateLimiterPool
}

func (self *Router) GetCtrlRateLimiter() rate.AdaptiveRateLimitTracker {
	logtrace.LogWithFunctionName()
	return self.ctrlRateLimiter
}

func (self *Router) registerComponents() error {
	logtrace.LogWithFunctionName()
	self.xlinkFactories = make(map[string]xlink.Factory)
	xlinkAccepter := newXlinkAccepter(self.forwarder)
	xlinkChAccepter := handler_link.NewBindHandlerFactory(
		self.ctrls,
		self.forwarder,
		&self.config.Link.Heartbeats,
		self.metricsRegistry,
		self.xlinkRegistry,
	)

	linkTransportConfig := map[interface{}]interface{}{}
	for k, v := range self.config.Transport {
		linkTransportConfig[k] = v
	}
	linkTransportConfig[transport.KeyCachedProxyConfiguration] = self.config.Proxy

	self.xlinkFactories["transport"] = xlink_transport.NewFactory(xlinkAccepter, xlinkChAccepter, linkTransportConfig, self.xlinkRegistry, self.metricsRegistry)

	xgress.GlobalRegistry().Register("proxy", xgress_proxy.NewFactory(self.config.Id, self.ctrls, self.config.Transport))
	xgress.GlobalRegistry().Register("proxy_udp", xgress_proxy_udp.NewFactory(self.ctrls))
	xgress.GlobalRegistry().Register("transport", xgress_transport.NewFactory(self.config.Id, self.ctrls, self.config.Transport))
	xgress.GlobalRegistry().Register("transport_udp", xgress_transport_udp.NewFactory(self.config.Id, self.ctrls))

	if err := self.RegisterXweb(xweb.NewDefaultInstance(self.xwebFactoryRegistry, self.config.Id)); err != nil {
		return err
	}

	if v, ok := self.xlinkRegistry.(env.Xrctrl); ok {
		if err := self.RegisterXrctrl(v); err != nil {
			return err
		}
	}

	return nil
}

func (self *Router) registerPlugins() error {
	logtrace.LogWithFunctionName()
	for _, pluginPath := range self.config.Plugins {
		goPlugin, err := plugin.Open(pluginPath)
		if err != nil {
			return errors.Wrapf(err, "router unable to load plugin at path %v", pluginPath)
		}
		initializeSymbol, err := goPlugin.Lookup("Initialize")
		if err != nil {
			return errors.Wrapf(err, "router plugin at %v does not contain Initialize symbol", pluginPath)
		}
		initialize, ok := initializeSymbol.(func(*Router) error)
		if !ok {
			return errors.Errorf("router plugin at %v exports Initialize symbol, but it is not of type 'func(router *router.Router) error'", pluginPath)
		}
		if err := initialize(self); err != nil {
			return errors.Wrapf(err, "error initializing router plugin at %v", pluginPath)
		}
	}
	return nil
}

func (self *Router) startXlinkDialers() {
	logtrace.LogWithFunctionName()
	for _, lmap := range self.config.Link.Dialers {
		binding := "transport"
		if bindingVal, ok := lmap["binding"]; ok {
			bindingName := fmt.Sprintf("%v", bindingVal)
			if len(bindingName) > 0 {
				binding = bindingName
			}
		}

		if factory, found := self.xlinkFactories[binding]; found {
			dialer, err := factory.CreateDialer(self.config.Id, self.forwarder, lmap)
			if err != nil {
				logrus.Fatalf("error creating Xlink dialer (%v)", err)
			}
			self.xlinkDialers = append(self.xlinkDialers, dialer)
			logrus.Infof("started Xlink dialer with binding [%s]", binding)
		}
	}
}

func (self *Router) startXlinkListeners() {
	logtrace.LogWithFunctionName()
	for _, lmap := range self.config.Link.Listeners {
		binding := "transport"
		if bindingVal, ok := lmap["binding"]; ok {
			bindingName := fmt.Sprintf("%v", bindingVal)
			if len(bindingName) > 0 {
				binding = bindingName
			}
		}

		if factory, found := self.xlinkFactories[binding]; found {
			lmap[transport.KeyProtocol] = "ziti-link"
			listener, err := factory.CreateListener(self.config.Id, self.forwarder, lmap)
			if err != nil {
				logrus.Fatalf("error creating Xlink listener (%v)", err)
			}
			if err := listener.Listen(); err != nil {
				logrus.Fatalf("error listening on Xlink (%v)", err)
			}
			self.xlinkListeners = append(self.xlinkListeners, listener)
			logrus.Infof("started Xlink listener with binding [%s] advertising [%s]", binding, listener.GetAdvertisement())
		}
	}
}

func (self *Router) setDefaultDialerBindings() {
	logtrace.LogWithFunctionName()
	if len(self.xlinkDialers) == 1 && len(self.xlinkListeners) == 1 && self.xlinkDialers[0].GetBinding() == "" {
		self.xlinkDialers[0].AdoptBinding(self.xlinkListeners[0])
	}
}

func (self *Router) startXgressListeners() {
	logtrace.LogWithFunctionName()
	for _, binding := range self.config.Listeners {
		factory, err := xgress.GlobalRegistry().Factory(binding.name)
		if err != nil {
			logrus.Fatalf("error getting xgress factory [%s] (%v)", binding.name, err)
		}
		listener, err := factory.CreateListener(binding.options)
		if err != nil {
			logrus.Fatalf("error creating xgress listener [%s] (%v)", binding.name, err)
		}
		self.xgressListeners = append(self.xgressListeners, listener)

		var address string
		if addressVal, found := binding.options["address"]; found {
			address = addressVal.(string)
		}

		err = listener.Listen(address,
			handler_xgress.NewBindHandler(
				handler_xgress.NewReceiveHandler(self.forwarder),
				handler_xgress.NewCloseHandler(self.ctrls, self.forwarder),
				self.forwarder,
			),
		)
		if err != nil {
			logrus.Fatalf("error listening [%s] (%v)", binding.name, err)
		}
		logrus.Infof("created xgress listener [%s] at [%s]", binding.name, address)
	}
}

func (self *Router) startControlPlane() error {
	logtrace.LogWithFunctionName()
	endpoints, err := self.getInitialCtrlEndpoints()
	if err != nil {
		return err
	}

	log := pfxlog.Logger()
	log.Infof("router configured with %v controller endpoints", len(endpoints))

	self.ctrls.UpdateControllerEndpoints(endpoints)

	self.metricsReporter = fabricMetrics.NewControllersReporter(self.ctrls)
	self.metricsRegistry.StartReporting(self.metricsReporter, self.config.Metrics.ReportInterval, self.config.Metrics.MessageQueueSize)

	if self.config.Ctrl.StartupTimeout > 0 {
		time.AfterFunc(self.config.Ctrl.StartupTimeout, func() {
			if !self.isShutdown.Load() && len(self.ctrls.GetAll()) == 0 {
				if os.Getenv("STACKDUMP_ON_FAILED_STARTUP") == "true" {
					debugz.DumpStack()
				}
				pfxlog.Logger().Fatal("unable to connect to any controllers before timeout")
			}
		})
	}

	_ = self.ctrls.AnyValidCtrlChannel()
	for _, x := range self.xrctrls {
		if err := x.Run(self); err != nil {
			return err
		}
	}

	return nil
}

func (self *Router) connectToController(addr transport.Address, bindHandler channel.BindHandler) error {
	logtrace.LogWithFunctionName()
	attributes := map[int32][]byte{}

	version, err := self.versionProvider.EncoderDecoder().Encode(self.versionProvider.AsVersionInfo())

	if err != nil {
		return fmt.Errorf("error with version header information value: %v", err)
	}

	attributes[channel.HelloVersionHeader] = version

	listeners := &ctrl_pb.Listeners{}
	for _, listener := range self.xlinkListeners {
		listeners.Listeners = append(listeners.Listeners, &ctrl_pb.Listener{
			Address:      listener.GetAdvertisement(),
			Protocol:     listener.GetLinkProtocol(),
			CostTags:     listener.GetLinkCostTags(),
			Groups:       listener.GetGroups(),
			LocalBinding: listener.GetLocalBinding(),
		})
	}

	if len(listeners.Listeners) > 0 {
		if buf, err := proto.Marshal(listeners); err != nil {
			return errors.Wrap(err, "unable to marshal Listeners")
		} else {
			attributes[int32(ctrl_pb.ControlHeaders_ListenersHeader)] = buf
		}
	}

	routerMeta := &ctrl_pb.RouterMetadata{
		Capabilities: []ctrl_pb.RouterCapability{
			ctrl_pb.RouterCapability_LinkManagement,
		},
	}

	if buf, err := proto.Marshal(routerMeta); err != nil {
		return errors.Wrap(err, "unable to router metadata")
	} else {
		attributes[int32(ctrl_pb.ControlHeaders_RouterMetadataHeader)] = buf
	}

	var channelRef concurrenz.AtomicValue[channel.Channel]
	reconnectHandler := func() {
		if ch := channelRef.Load(); ch != nil {
			for _, x := range self.xrctrls {
				go x.NotifyOfReconnect(ch)
			}
			self.ctrls.NotifyOfReconnect(ch.Id())
		}
	}

	if "" != self.config.Ctrl.LocalBinding {
		logrus.Debugf("Using local interface %s to dial controller", self.config.Ctrl.LocalBinding)
	}
	dialer := channel.NewReconnectingDialer(channel.ReconnectingDialerConfig{
		Identity:         self.config.Id,
		Endpoint:         addr,
		LocalBinding:     self.config.Ctrl.LocalBinding,
		Headers:          attributes,
		ReconnectHandler: reconnectHandler,
		TransportConfig: transport.Configuration{
			transport.KeyProtocol:                 "ziti-ctrl",
			transport.KeyCachedProxyConfiguration: self.config.Proxy,
		},
	})

	bindHandler = channel.BindHandlers(bindHandler, self.ctrlBindhandler)
	ch, err := channel.NewChannel("ctrl", dialer, bindHandler, self.config.Ctrl.Options)
	if err != nil {
		return fmt.Errorf("error connecting ctrl (%v)", err)
	}
	channelRef.Store(ch)

	// If there are multiple controllers we may have to catch up the controllers that connected later
	// with things that have already happened because we had state from other controllers, such as
	// links
	reconnectHandler()

	return nil
}

func (self *Router) initializeHealthChecks() (gosundheit.Health, error) {
	logtrace.LogWithFunctionName()
	checkConfig := self.config.HealthChecks
	logrus.Infof("starting health check with ctrl ping initially after %v, then every %v, timing out after %v",
		checkConfig.CtrlPingCheck.InitialDelay, checkConfig.CtrlPingCheck.Interval, checkConfig.CtrlPingCheck.Timeout)

	h := gosundheit.New()
	ctrlPinger := &controllerPinger{
		router: self,
	}
	ctrlPingCheck, err := checks.NewPingCheck("controllerPing", ctrlPinger)
	if err != nil {
		return nil, err
	}

	err = h.RegisterCheck(ctrlPingCheck,
		gosundheit.ExecutionPeriod(checkConfig.CtrlPingCheck.Interval),
		gosundheit.ExecutionTimeout(checkConfig.CtrlPingCheck.Timeout),
		gosundheit.InitiallyPassing(false),
		gosundheit.InitialDelay(checkConfig.CtrlPingCheck.InitialDelay),
	)

	if err != nil {
		return nil, err
	}

	err = h.RegisterCheck(&linkHealthCheck{router: self, minLinks: checkConfig.LinkCheck.MinLinks},
		gosundheit.ExecutionPeriod(checkConfig.LinkCheck.Interval),
		gosundheit.ExecutionTimeout(5*time.Second),
		gosundheit.InitiallyPassing(checkConfig.LinkCheck.MinLinks == 0),
		gosundheit.InitialDelay(checkConfig.LinkCheck.InitialDelay),
	)

	if err != nil {
		return nil, err
	}

	return h, nil
}

func (self *Router) RegisterXweb(x xweb.Instance) error {
	logtrace.LogWithFunctionName()
	if err := self.config.Configure(x); err != nil {
		return err
	}
	if x.Enabled() {
		self.xwebs = append(self.xwebs, x)
	}
	return nil
}

func (self *Router) RegisterXWebHandlerFactory(x xweb.ApiHandlerFactory) error {
	logtrace.LogWithFunctionName()
	return self.xwebFactoryRegistry.Add(x)
}

func (self *Router) getInitialCtrlEndpoints() ([]string, error) {
	logtrace.LogWithFunctionName()
	log := pfxlog.Logger()
	if self.config.Ctrl.DataDir == "" {
		return nil, errors.New("ctrl DataDir not configured")
	}

	endpointsFile := path.Join(self.config.Ctrl.DataDir, "endpoints")

	var endpoints []string

	if _, err := os.Stat(endpointsFile); err != nil && errors.Is(err, fs.ErrNotExist) {
		log.Infof("controller endpoints file [%v] doesn't exist. Using initial endpoints from config", endpointsFile)
		for _, ep := range self.config.Ctrl.InitialEndpoints {
			endpoints = append(endpoints, ep.String())
		}
		return endpoints, nil
	}

	log.Infof("loading controller endpoints from [%v]", endpointsFile)

	b, err := os.ReadFile(endpointsFile)
	if err != nil {
		return nil, err
	}

	endpointCfg := &endpointConfig{}

	if err = yaml.Unmarshal(b, endpointCfg); err != nil {
		return nil, err
	}

	endpoints = endpointCfg.Endpoints

	if len(endpoints) == 0 {
		return nil, errors.Errorf("no controller endpoints found in [%v], consider deleting file", endpointsFile)
	}

	return endpoints, nil
}

func (self *Router) UpdateCtrlEndpoints(endpoints []string) {
	logtrace.LogWithFunctionName()
	log := pfxlog.Logger().WithField("endpoints", endpoints).WithField("filepath", self.config.Ctrl.DataDir)
	if changed := self.ctrls.UpdateControllerEndpoints(endpoints); changed {
		log.Info("Attempting to save file")
		endpointsFile := path.Join(self.config.Ctrl.DataDir, "endpoints")

		configData := map[string]interface{}{
			"Endpoints": endpoints,
		}

		if data, err := yaml.Marshal(configData); err != nil {
			log.WithError(err).Error("unable to marshal updated controller endpoints to yaml")
		} else if err = os.WriteFile(endpointsFile, data, 0600); err != nil {
			log.WithError(err).Error("unable to write updated controller endpoints to file")
		}
	}
}

func (self *Router) UpdateLeader(leaderId string) {
	logtrace.LogWithFunctionName()
	self.ctrls.UpdateLeader(leaderId)
}

type connectionToggle interface {
	Disconnect() error
	Reconnect() error
}

type controllerPinger struct {
	router *Router
}

func (self *controllerPinger) PingContext(context.Context) error {
	logtrace.LogWithFunctionName()
	ctrls := self.router.ctrls.GetAll()

	if len(ctrls) == 0 {
		return errors.New("no control channels established yet")
	}

	hasGoodConn := false

	for _, ctrl := range ctrls {
		if !ctrl.IsUnresponsive() {
			hasGoodConn = true
		}
	}

	if hasGoodConn {
		return nil
	}
	return errors.New("control channels are slow")
}

type endpointConfig struct {
	Endpoints []string `yaml:"Endpoints"`
}

type linkConnDetail struct {
	LocalAddr  string `json:"localAddr"`
	RemoteAddr string `json:"remoteAddr"`
}

type linkDetail struct {
	LinkId       string                    `json:"linkId"`
	DestRouterId string                    `json:"destRouterId"`
	Latency      *float64                  `json:"latency,omitempty"`
	Addresses    map[string]linkConnDetail `json:"addresses"`
}

type linkHealthCheck struct {
	router   *Router
	minLinks int
}

func (self *linkHealthCheck) Name() string {
	logtrace.LogWithFunctionName()
	return "link.health"
}

func (self *linkHealthCheck) Execute(ctx context.Context) (details interface{}, err error) {
	logtrace.LogWithFunctionName()
	var links []*linkDetail

	iter := self.router.xlinkRegistry.Iter()
	done := false

	for !done {
		var currentLink xlink.Xlink
		select {
		case nextLink, ok := <-iter:
			if !ok {
				done = true
			}
			currentLink = nextLink
		case <-ctx.Done():
			done = true
		}

		if currentLink != nil {
			detail := &linkDetail{
				LinkId:       currentLink.Id(),
				DestRouterId: currentLink.DestinationId(),
				Addresses:    map[string]linkConnDetail{},
			}
			for _, addr := range currentLink.GetAddresses() {
				detail.Addresses[addr.Id] = linkConnDetail{
					LocalAddr:  addr.LocalAddr,
					RemoteAddr: addr.RemoteAddr,
				}
			}
			latencyMetric := self.router.metricsRegistry.Histogram("link." + currentLink.Id() + ".latency")
			if latencyMetric != nil {
				latency := latencyMetric.(metrics2.Histogram).Mean()
				detail.Latency = &latency
			}
			links = append(links, detail)
		}
	}
	if len(links) < self.minLinks {
		return links, errors.Errorf("link count %v less than configured minimum of %v", len(links), self.minLinks)
	}
	return links, nil
}
