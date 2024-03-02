package udp

import (
	"github.com/traefik/traefik/v3/pkg/needleware"
	"io"
	"net"

	"github.com/rs/zerolog/log"
)

// Proxy is a reverse-proxy implementation of the Handler interface.
type Proxy struct {
	// TODO: maybe optimize by pre-resolving it at proxy creation time
	target string
	needle needleware.Needle
}

// NewProxy creates a new Proxy.
func NewProxy(address string, needle needleware.Needle) (*Proxy, error) {
	return &Proxy{target: address, needle: needle}, nil
}

// ServeUDP implements the Handler interface.
func (p *Proxy) ServeUDP(conn *Conn) {
	log.Debug().Msgf("Handling UDP stream from %s to %s", conn.rAddr, p.target)

	// needed because of e.g. server.trackedConnection
	defer conn.Close()

	if p.needle != nil {
		remoteAddr := conn.rAddr.String()
		localAddr := p.target

		criteria, err := p.needle.NewUDPCriteria(remoteAddr, localAddr)
		if err == nil {
			// wait until the decision is made
			decision, _ := p.needle.Decide(criteria)
			defer p.needle.OnConnClose(decision)
			if decision.ConnRejected() {
				conn.Close()
				return
			}
		} else {
			log.Error().Err(err).Msgf("Failed to create criteria when serving UDP connection from %s to %s", remoteAddr, localAddr)
		}
	}

	connBackend, err := net.Dial("udp", p.target)
	if err != nil {
		log.Error().Err(err).Msg("Error while dialing backend")
		return
	}

	// maybe not needed, but just in case
	defer connBackend.Close()

	errChan := make(chan error)
	go connCopy(conn, connBackend, errChan)
	go connCopy(connBackend, conn, errChan)

	err = <-errChan
	if err != nil {
		log.Error().Err(err).Msg("Error while handling UDP stream")
	}

	<-errChan
}

func connCopy(dst io.WriteCloser, src io.Reader, errCh chan error) {
	// The buffer is initialized to the maximum UDP datagram size,
	// to make sure that the whole UDP datagram is read or written atomically (no data is discarded).
	buffer := make([]byte, maxDatagramSize)

	_, err := io.CopyBuffer(dst, src, buffer)
	errCh <- err

	if err := dst.Close(); err != nil {
		log.Debug().Err(err).Msg("Error while terminating UDP stream")
	}
}
