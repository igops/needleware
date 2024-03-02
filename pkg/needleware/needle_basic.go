package needleware

import (
	"context"
	"fmt"
	"github.com/rs/zerolog"
	"github.com/traefik/traefik/v3/pkg/needleware/client"
	"math/rand"
	"time"
)

type DecisionRef int

const (
	DecisionRefAccept DecisionRef = iota
	DecisionRefReject
)

type BasicNeedle struct {
	client        client.Client
	logger        zerolog.Logger
	connTimeout   time.Duration
	randSource    rand.Source
	onTimeout     DecisionRef
	onError       DecisionRef
	notifyOnClose map[DecisionRef]bool
}

func (n *BasicNeedle) NewTCPCriteria(remoteAddr string, localAddr string) (*client.DecisionCriteria, error) {
	return n.newCriteria(remoteAddr, localAddr, client.ProtocolTCP)
}

func (n *BasicNeedle) NewUDPCriteria(remoteAddr string, localAddr string) (*client.DecisionCriteria, error) {
	return n.newCriteria(remoteAddr, localAddr, client.ProtocolUDP)
}

func (n *BasicNeedle) Decide(criteria *client.DecisionCriteria) (*DecisionWrapper, error) {
	ctx, cancel := context.WithTimeout(context.Background(), n.connTimeout)
	defer cancel()

	decisionResponse := n.client.OnConnOpened(criteria, ctx)

	switch decisionResponse.Status {
	case client.StatusDecisionLoaded:
		return n.decideOnLoaded(criteria, decisionResponse)

	case client.StatusDecisionError:
		return n.decideOnError(criteria, decisionResponse)

	case client.StatusDecisionTimeout:
		return n.decideOnTimeout(criteria)
	}
	return nil, fmt.Errorf("should never happen: unknown decision status %d; please validate it in the client code", decisionResponse.Status)
}

func (n *BasicNeedle) OnConnClose(decision *DecisionWrapper) {
	if n.notifyOnClose[DecisionRefAccept] && decision.ConnAccepted() ||
		n.notifyOnClose[DecisionRefReject] && decision.ConnRejected() {
		// try delivering the event no matter how long it takes
		go func() {
			err := n.client.OnConnClosed(decision.Criteria.ConnId, context.Background())
			if err != nil {
				n.logger.Error().Err(err).Msgf("Cannot deliver OnConnClosed event")
			}
		}()
	}
}

func (n *BasicNeedle) newCriteria(remoteAddr string, localAddr string, protocol client.Protocol) (*client.DecisionCriteria, error) {
	remoteHost, remotePort, err := parseHostPort(remoteAddr)
	if err != nil {
		return nil, err
	}
	localHost, localPort, err := parseHostPort(localAddr)
	if err != nil {
		return nil, err
	}
	return &client.DecisionCriteria{
		RemoteHost: remoteHost,
		RemotePort: remotePort,
		LocalHost:  localHost,
		LocalPort:  localPort,
		Protocol:   protocol,
		ConnId:     n.generateConnId(),
	}, nil
}

func (n *BasicNeedle) decideOnLoaded(criteria *client.DecisionCriteria, decision *client.DecisionResponse) (*DecisionWrapper, error) {
	if decision.ConnAccepted() {
		n.logger.Debug().Msgf("Connection from %s:%d to %s:%d accepted",
			criteria.RemoteHost, criteria.RemotePort, criteria.LocalHost, criteria.LocalPort)
		return &DecisionWrapper{
			Status:       client.StatusDecisionLoaded,
			DecisionCode: client.DecisionConnAccepted,
			Criteria:     criteria,
		}, nil
	}

	if decision.ConnRejected() {
		n.logger.Debug().Msgf("Connection from %s:%d to %s:%d rejected",
			criteria.RemoteHost, criteria.RemotePort, criteria.LocalHost, criteria.LocalPort)
		return &DecisionWrapper{
			Status:       client.StatusDecisionLoaded,
			DecisionCode: client.DecisionConnRejected,
			Criteria:     criteria,
		}, nil
	}

	return nil, fmt.Errorf("should never happen: unknown decision code %d; please validate it in the client code", decision.Decision.Code)
}

func (n *BasicNeedle) decideOnError(criteria *client.DecisionCriteria, decision *client.DecisionResponse) (*DecisionWrapper, error) {
	n.logger.Error().Err(decision.Err).Msgf("Cannot load decision")
	if n.onError == DecisionRefAccept {
		return &DecisionWrapper{
			Status:       client.StatusDecisionError,
			DecisionCode: client.DecisionConnAccepted,
			Criteria:     criteria,
		}, nil
	}
	if n.onError == DecisionRefReject {
		return &DecisionWrapper{
			Status:       client.StatusDecisionError,
			DecisionCode: client.DecisionConnRejected,
			Criteria:     criteria,
		}, nil
	}
	return nil, fmt.Errorf("should never happen: unknown onError code %d; please validate it while creating the needle", n.onError)
}

func (n *BasicNeedle) decideOnTimeout(criteria *client.DecisionCriteria) (*DecisionWrapper, error) {
	n.logger.Debug().Msgf("Decision timeout")
	if n.onTimeout == DecisionRefAccept {
		return &DecisionWrapper{
			Status:       client.StatusDecisionTimeout,
			DecisionCode: client.DecisionConnAccepted,
			Criteria:     criteria,
		}, nil
	}
	if n.onTimeout == DecisionRefReject {
		return &DecisionWrapper{
			Status:       client.StatusDecisionTimeout,
			DecisionCode: client.DecisionConnRejected,
			Criteria:     criteria,
		}, nil
	}
	return nil, fmt.Errorf("should never happen: unknown onTimeout code %d; please validate it while creating the needle", n.onError)
}

func (n *BasicNeedle) generateConnId() int32 {
	// 10 digits random int from MaxInt32 / 2 to MaxInt32
	return int32(1<<30 + rand.New(n.randSource).Intn(1<<30))
}
