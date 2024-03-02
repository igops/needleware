package needleware

import (
	"github.com/traefik/traefik/v3/pkg/needleware/client"
)

type NeedleWithMeta struct {
	needle Needle
	meta   map[string]string
}

func (n *NeedleWithMeta) NewTCPCriteria(remoteAddr string, localAddr string) (*client.DecisionCriteria, error) {
	criteria, err := n.needle.NewTCPCriteria(remoteAddr, localAddr)
	criteria.Metadata = n.meta
	return criteria, err
}

func (n *NeedleWithMeta) NewUDPCriteria(remoteAddr string, localAddr string) (*client.DecisionCriteria, error) {
	criteria, err := n.needle.NewUDPCriteria(remoteAddr, localAddr)
	criteria.Metadata = n.meta
	return criteria, err
}

func (n *NeedleWithMeta) Decide(criteria *client.DecisionCriteria) (*DecisionWrapper, error) {
	return n.needle.Decide(criteria)
}

func (n *NeedleWithMeta) OnConnClose(decision *DecisionWrapper) {
	n.needle.OnConnClose(decision)
}
