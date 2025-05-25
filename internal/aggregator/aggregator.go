package aggregator

import (
	"github.com/CyberwizD/Concurrent-Web-Content-Aggregator/internal/coordinator"
	"github.com/CyberwizD/Concurrent-Web-Content-Aggregator/pkg/config"
)

type Aggregator struct {
	Coordinator *coordinator.Coordinator
	Config      *config.Config
	Results     []Result
}

func (a *Aggregator) New(cfg *config.Config, coord *coordinator.Coordinator) {

}
