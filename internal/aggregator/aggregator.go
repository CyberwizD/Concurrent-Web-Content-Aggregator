package aggregator

import (
	"github.com/CyberwizD/Concurrent-Web-Content-Aggregator/internal/coordinator"
	"github.com/CyberwizD/Concurrent-Web-Content-Aggregator/pkg/config"
)

type Aggregator struct {
	Coordinator *coordinator.Coordinator
	Config      *config.Config
}

type Result struct {
}

func New(cfg *config.Config, coord *coordinator.Coordinator) (*Aggregator, error) {
	if cfg == nil {
		return nil, coordinator.ErrInvalidConfig
	}

	if coord == nil {
		return nil, coordinator.ErrInvalidCoordinator
	}

	return &Aggregator{
		Config:      cfg,
		Coordinator: coord,
	}, nil
}
