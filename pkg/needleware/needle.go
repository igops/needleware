package needleware

import (
	"github.com/traefik/traefik/v3/pkg/needleware/client"
)

type Needle interface {
	NewTCPCriteria(remoteAddr string, localAddr string) (*client.DecisionCriteria, error)
	NewUDPCriteria(remoteAddr string, localAddr string) (*client.DecisionCriteria, error)
	Decide(criteria *client.DecisionCriteria) (*DecisionWrapper, error)
	OnConnClose(decision *DecisionWrapper)
}
