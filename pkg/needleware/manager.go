package needleware

import (
	"context"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/traefik/traefik/v3/pkg/config/runtime"
	"github.com/traefik/traefik/v3/pkg/logs"
	"math/rand"
	"time"
)

type Manager struct {
	logger  zerolog.Logger
	needles map[string]Needle
}

func NewManager() *Manager {
	return &Manager{
		needles: map[string]Needle{},
	}
}

func (m *Manager) BuildNeedles(rootCtx context.Context, conf *runtime.Configuration) {
	randSource := rand.NewSource(time.Now().UnixNano())

	for k, v := range conf.Needles {
		logger := log.Ctx(rootCtx).With().Str(logs.NeedleName, k).Logger()
		logger.Debug().Msg("building needle")

		parser := &needleConfParser{
			conf:   v,
			logger: logger,
		}
		client, ok := parser.buildClient()
		if !ok {
			continue
		}
		onTimeout, ok := parser.validOnTimeout()
		if !ok {
			continue
		}
		onError, ok := parser.validOnError()
		if !ok {
			continue
		}
		notifyOnClose, ok := parser.validNotifyOnClose()
		if !ok {
			continue
		}
		connTimeout, ok := parser.validConnTimeout()
		if !ok {
			continue
		}

		m.needles[k] = &BasicNeedle{
			client:        client,
			logger:        logger,
			connTimeout:   connTimeout,
			randSource:    randSource,
			onTimeout:     onTimeout,
			onError:       onError,
			notifyOnClose: notifyOnClose,
		}
	}
}

func (m *Manager) GetNeedle(needle string, metadata map[string]string) Needle {
	n := m.needles[needle]
	if n == nil {
		return nil
	}
	if metadata == nil {
		return n
	}
	return &NeedleWithMeta{
		needle: n,
		meta:   metadata,
	}
}
