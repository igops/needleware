package server

import (
	"context"
	"github.com/traefik/traefik/v3/pkg/needleware"

	"github.com/rs/zerolog/log"
	"github.com/traefik/traefik/v3/pkg/config/runtime"
	"github.com/traefik/traefik/v3/pkg/config/static"
	"github.com/traefik/traefik/v3/pkg/metrics"
	"github.com/traefik/traefik/v3/pkg/server/middleware"
	tcpmiddleware "github.com/traefik/traefik/v3/pkg/server/middleware/tcp"
	"github.com/traefik/traefik/v3/pkg/server/router"
	tcprouter "github.com/traefik/traefik/v3/pkg/server/router/tcp"
	udprouter "github.com/traefik/traefik/v3/pkg/server/router/udp"
	"github.com/traefik/traefik/v3/pkg/server/service"
	tcpsvc "github.com/traefik/traefik/v3/pkg/server/service/tcp"
	udpsvc "github.com/traefik/traefik/v3/pkg/server/service/udp"
	"github.com/traefik/traefik/v3/pkg/tcp"
	"github.com/traefik/traefik/v3/pkg/tls"
	"github.com/traefik/traefik/v3/pkg/udp"
)

// RouterFactory the factory of TCP/UDP routers.
type RouterFactory struct {
	entryPointsTCP []string
	entryPointsUDP []string

	managerFactory  *service.ManagerFactory
	metricsRegistry metrics.Registry

	pluginBuilder middleware.PluginsBuilder

	chainBuilder *middleware.ChainBuilder
	tlsManager   *tls.Manager

	dialerManager *tcp.DialerManager

	cancelPrevState func()
}

// NewRouterFactory creates a new RouterFactory.
func NewRouterFactory(staticConfiguration static.Configuration, managerFactory *service.ManagerFactory, tlsManager *tls.Manager,
	chainBuilder *middleware.ChainBuilder, pluginBuilder middleware.PluginsBuilder, metricsRegistry metrics.Registry, dialerManager *tcp.DialerManager,
) *RouterFactory {
	var entryPointsTCP, entryPointsUDP []string
	for name, cfg := range staticConfiguration.EntryPoints {
		protocol, err := cfg.GetProtocol()
		if err != nil {
			// Should never happen because Traefik should not start if protocol is invalid.
			log.Error().Err(err).Msg("Invalid protocol")
		}

		if protocol == "udp" {
			entryPointsUDP = append(entryPointsUDP, name)
		} else {
			entryPointsTCP = append(entryPointsTCP, name)
		}
	}

	return &RouterFactory{
		entryPointsTCP:  entryPointsTCP,
		entryPointsUDP:  entryPointsUDP,
		managerFactory:  managerFactory,
		metricsRegistry: metricsRegistry,
		tlsManager:      tlsManager,
		chainBuilder:    chainBuilder,
		pluginBuilder:   pluginBuilder,
		dialerManager:   dialerManager,
	}
}

// CreateRouters creates new TCPRouters and UDPRouters.
func (f *RouterFactory) CreateRouters(rtConf *runtime.Configuration) (map[string]*tcprouter.Router, map[string]udp.Handler) {
	if f.cancelPrevState != nil {
		f.cancelPrevState()
	}

	var ctx context.Context
	ctx, f.cancelPrevState = context.WithCancel(context.Background())

	// HTTP
	serviceManager := f.managerFactory.Build(rtConf)

	middlewaresBuilder := middleware.NewBuilder(rtConf.Middlewares, serviceManager, f.pluginBuilder)

	routerManager := router.NewManager(rtConf, serviceManager, middlewaresBuilder, f.chainBuilder, f.metricsRegistry, f.tlsManager)

	handlersNonTLS := routerManager.BuildHandlers(ctx, f.entryPointsTCP, false)
	handlersTLS := routerManager.BuildHandlers(ctx, f.entryPointsTCP, true)

	serviceManager.LaunchHealthCheck(ctx)

	// Needles
	needlewareManager := needleware.NewManager()
	needlewareManager.BuildNeedles(ctx, rtConf)

	// TCP
	svcTCPManager := tcpsvc.NewManager(rtConf, f.dialerManager)

	middlewaresTCPBuilder := tcpmiddleware.NewBuilder(rtConf.TCPMiddlewares, needlewareManager)

	rtTCPManager := tcprouter.NewManager(rtConf, svcTCPManager, middlewaresTCPBuilder, handlersNonTLS, handlersTLS, f.tlsManager)
	routersTCP := rtTCPManager.BuildHandlers(ctx, f.entryPointsTCP)

	// UDP
	svcUDPManager := udpsvc.NewManager(rtConf, needlewareManager)
	rtUDPManager := udprouter.NewManager(rtConf, svcUDPManager)
	routersUDP := rtUDPManager.BuildHandlers(ctx, f.entryPointsUDP)

	rtConf.PopulateUsedBy()

	return routersTCP, routersUDP
}
