package tcpneedle

import (
	"context"
	"github.com/rs/zerolog"
	"github.com/traefik/traefik/v3/pkg/middlewares"
	"github.com/traefik/traefik/v3/pkg/needleware"
	"github.com/traefik/traefik/v3/pkg/tcp"
)

const typeName = "NeedleTCP"

type needleTCP struct {
	name   string
	next   tcp.Handler
	needle needleware.Needle
	logger *zerolog.Logger
}

// New creates Needle middleware.
func New(ctx context.Context, next tcp.Handler, needle needleware.Needle, name string) (tcp.Handler, error) {
	logger := middlewares.GetLogger(ctx, name, typeName)
	logger.Debug().Msg("Creating middleware")

	return &needleTCP{
		name:   name,
		next:   next,
		needle: needle,
		logger: logger,
	}, nil
}

// ServeTCP serves the given TCP connection.
func (i *needleTCP) ServeTCP(conn tcp.WriteCloser) {
	remoteAddr := conn.RemoteAddr().String()
	localAddr := conn.LocalAddr().String()

	criteria, err := i.needle.NewTCPCriteria(remoteAddr, localAddr)
	if err == nil {
		// wait until the decision is made
		decision, _ := i.needle.Decide(criteria)
		defer i.needle.OnConnClose(decision)
		if decision.ConnRejected() {
			conn.Close()
			return
		}
	} else {
		i.logger.Error().Err(err).Msgf("Failed to create criteria when serving TCP connection from %s to %s", remoteAddr, localAddr)
	}

	i.next.ServeTCP(conn)
}
