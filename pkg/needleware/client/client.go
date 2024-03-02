package client

import "context"

type Protocol int

const (
	ProtocolUDP Protocol = iota
	ProtocolTCP
)

type DecisionStatus int

const (
	StatusDecisionLoaded DecisionStatus = iota
	StatusDecisionError
	StatusDecisionTimeout
)

type DecisionCode int

const (
	DecisionConnAccepted DecisionCode = iota
	DecisionConnRejected
)

type Client interface {
	OnConnOpened(criteria *DecisionCriteria, ctx context.Context) *DecisionResponse
	OnConnClosed(connId int32, ctx context.Context) error
}

type DecisionCriteria struct {
	Protocol   Protocol
	ConnId     int32
	RemoteHost string
	RemotePort int32
	LocalHost  string
	LocalPort  int32
	Metadata   map[string]string
}

type Decision struct {
	Code DecisionCode
}

type DecisionResponse struct {
	Status   DecisionStatus
	Decision *Decision
	Err      error
}

func (r *DecisionResponse) Loaded() bool {
	return r.Status == StatusDecisionLoaded
}

func (r *DecisionResponse) Error() bool {
	return r.Status == StatusDecisionError
}

func (r *DecisionResponse) Timeout() bool {
	return r.Status == StatusDecisionTimeout
}

func (r *DecisionResponse) ConnAccepted() bool {
	return r.Decision != nil && r.Decision.Code == DecisionConnAccepted
}

func (r *DecisionResponse) ConnRejected() bool {
	return r.Decision != nil && r.Decision.Code == DecisionConnRejected
}
