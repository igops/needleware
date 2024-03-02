package runtime

import "github.com/traefik/traefik/v3/pkg/config/dynamic"

type NeedleInfo struct {
	*dynamic.Needle          // dynamic configuration
	Err             []string `json:"error,omitempty"` // initialization error
	// Status reports whether the router is disabled, in a warning state, or all good (enabled).
	// If not in "enabled" state, the reason for it should be in the list of Err.
	// It is the caller's responsibility to set the initial status.
	Status               string   `json:"status,omitempty"`
	UsedByTCPMiddlewares []string `json:"usedByTCPMiddlewares,omitempty"` //
	UsedByUDPServices    []string `json:"UsedByUDPServices,omitempty"`    //
}
