package needleware

import "github.com/traefik/traefik/v3/pkg/needleware/client"

type DecisionWrapper struct {
	Status       client.DecisionStatus
	DecisionCode client.DecisionCode
	Criteria     *client.DecisionCriteria
}

func (dw *DecisionWrapper) ConnAccepted() bool {
	return dw.DecisionCode == client.DecisionConnAccepted
}

func (dw *DecisionWrapper) ConnRejected() bool {
	return dw.DecisionCode == client.DecisionConnRejected
}
