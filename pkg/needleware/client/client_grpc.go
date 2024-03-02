package client

import (
	"context"
	"fmt"
	"github.com/traefik/traefik/v3/pkg/needleware/client/pb"
	"google.golang.org/grpc/status"
)

type GRPCClient struct {
	client pb.NeedlewareClient
}

func NewGRPCClient(client pb.NeedlewareClient) *GRPCClient {
	return &GRPCClient{
		client: client,
	}
}

func (c *GRPCClient) OnConnOpened(criteria *DecisionCriteria, ctx context.Context) *DecisionResponse {
	var protocol pb.Protocol
	switch criteria.Protocol {
	case ProtocolUDP:
		protocol = pb.Protocol_UDP
	case ProtocolTCP:
		protocol = pb.Protocol_TCP
	default:
		return &DecisionResponse{
			Status: StatusDecisionError,
			Err:    fmt.Errorf("unknown protocol %d", criteria.Protocol),
		}
	}

	var metadata *pb.Metadata = nil
	if criteria.Metadata != nil {
		metadata = &pb.Metadata{
			Data: criteria.Metadata,
		}
	}

	response, err := c.client.OnConnOpened(ctx, &pb.Connection{
		Id: &pb.ConnectionId{
			Value: criteria.ConnId,
		},
		Protocol: protocol,
		RemoteAddress: &pb.Address{
			Host: criteria.RemoteHost,
			Port: criteria.RemotePort,
		},
		LocalAddress: &pb.Address{
			Host: criteria.LocalHost,
			Port: criteria.LocalPort,
		},
		Metadata: metadata,
	})

	if err != nil {
		return &DecisionResponse{
			Status: func() DecisionStatus {
				if status.Code(err) == status.Code(context.DeadlineExceeded) {
					return StatusDecisionTimeout
				}
				return StatusDecisionError
			}(),
		}
	}

	switch response.GetCode() {
	case pb.DecisionCode_ACCEPT:
		return &DecisionResponse{
			Status: StatusDecisionLoaded,
			Decision: &Decision{
				Code: DecisionConnAccepted,
			},
		}
	case pb.DecisionCode_REJECT:
		return &DecisionResponse{
			Status: StatusDecisionLoaded,
			Decision: &Decision{
				Code: DecisionConnRejected,
			},
		}
	}

	return &DecisionResponse{
		Status: StatusDecisionError,
		Err:    fmt.Errorf("unknown decision code %d", response.GetCode()),
	}
}

func (c *GRPCClient) OnConnClosed(connId int32, ctx context.Context) error {
	_, err := c.client.OnConnClosed(ctx, &pb.ConnectionId{
		Value: connId,
	})
	return err
}
