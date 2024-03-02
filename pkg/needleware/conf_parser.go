package needleware

import (
	"github.com/rs/zerolog"
	"github.com/traefik/traefik/v3/pkg/config/runtime"
	"github.com/traefik/traefik/v3/pkg/needleware/client"
	"github.com/traefik/traefik/v3/pkg/needleware/client/pb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"strings"
	"time"
)

const (
	defaultClientType = "grpc"
	defaultTimeout    = 5 * time.Second
	defaultOnTimeout  = DecisionRefReject
	defaultOnError    = DecisionRefReject
)

func defaultNotifyOnClose() map[DecisionRef]bool {
	return map[DecisionRef]bool{
		DecisionRefAccept: true,
		DecisionRefReject: false,
	}
}

type needleConfParser struct {
	conf   *runtime.NeedleInfo
	logger zerolog.Logger
}

func (n *needleConfParser) buildClient() (client.Client, bool) {
	var clientType string
	cc := n.conf.Client
	if cc != nil && cc.Type != "" {
		clientType = cc.Type
	} else {
		clientType = defaultClientType
	}
	switch strings.ToLower(clientType) {
	case "grpc":
		return n.buildGRPCClient()
	default:
		n.logger.Error().Msgf("unknown client.type value: %s", clientType)
		return nil, false
	}
}

func (n *needleConfParser) buildGRPCClient() (client.Client, bool) {
	endpoint, ok := n.validEndpoint()
	if !ok {
		return nil, false
	}
	creds, ok := n.validGRPCCredentials()
	if !ok {
		return nil, false
	}

	grpcConn, err := grpc.Dial(endpoint, grpc.WithTransportCredentials(creds))
	if err != nil {
		n.logger.Error().Err(err).Msg("failed to create gRPC connection")
		return nil, false
	}

	return client.NewGRPCClient(pb.NewNeedlewareClient(grpcConn)), true
}

func (n *needleConfParser) validEndpoint() (string, bool) {
	endpoint := n.conf.Endpoint
	if endpoint == "" {
		n.logger.Error().Msg("endpoint is empty")
		return "", false
	}
	return endpoint, true
}

func (n *needleConfParser) validGRPCCredentials() (credentials.TransportCredentials, bool) {
	var tc credentials.TransportCredentials
	cc := n.conf.Client
	if cc != nil && cc.Auth != nil {
		var err error
		if cc.Auth.TlsCertFilePath == "" {
			n.logger.Error().Msg("client.auth.tlsCertFilePath is empty")
			return nil, false
		}
		tc, err = credentials.NewClientTLSFromFile(cc.Auth.TlsCertFilePath, "")
		if err != nil {
			n.logger.Error().Err(err).Msgf("failed to create TLS credentials from file %s", cc.Auth.TlsCertFilePath)
			return nil, false
		}
		return tc, true
	}
	return insecure.NewCredentials(), true
}

func (n *needleConfParser) validOnTimeout() (DecisionRef, bool) {
	decision := n.conf.Decision
	if decision == nil || decision.OnTimeout == "" {
		return defaultOnTimeout, true
	}
	switch strings.ToLower(decision.OnTimeout) {
	case "accept":
		return DecisionRefAccept, true
	case "reject":
		return DecisionRefReject, true
	}
	n.logger.Error().Msgf("unknown decision.onTimeout value: %s", decision.OnTimeout)
	return 0, false
}

func (n *needleConfParser) validOnError() (DecisionRef, bool) {
	decision := n.conf.Decision
	if decision == nil || decision.OnError == "" {
		return defaultOnError, true
	}
	switch strings.ToLower(decision.OnError) {
	case "accept":
		return DecisionRefAccept, true
	case "reject":
		return DecisionRefReject, true
	}
	n.logger.Error().Msgf("unknown decision.onError value: %s", n.conf.Decision.OnError)
	return 0, false
}

func (n *needleConfParser) validNotifyOnClose() (map[DecisionRef]bool, bool) {
	var m = make(map[DecisionRef]bool)
	if len(n.conf.NotifyConnClose) == 0 {
		return defaultNotifyOnClose(), true
	}

	for _, v := range n.conf.NotifyConnClose {
		switch strings.ToLower(v) {
		case "accept":
			m[DecisionRefAccept] = true
		case "reject":
			m[DecisionRefReject] = true
		default:
			n.logger.Error().Msgf("unknown notifyOnClose value: %s", v)
			return nil, false
		}
	}

	return m, true
}

func (n *needleConfParser) validConnTimeout() (time.Duration, bool) {
	cc := n.conf.Client
	if cc == nil || cc.Timeout == "" {
		return defaultTimeout, true
	}
	timeout := cc.Timeout
	duration, err := time.ParseDuration(timeout)
	if err != nil {
		n.logger.Error().Err(err).Msgf("invalid client.timeout value: %s", timeout)
		return 0, false
	}
	return duration, true
}
