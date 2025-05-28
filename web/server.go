package web

import (
	"fmt"
	"net/http"

	"github.com/CyberwizD/Concurrent-Web-Content-Aggregator/internal/aggregator"
	"github.com/CyberwizD/Concurrent-Web-Content-Aggregator/pkg/config"
)

type APIServer struct {
	config     *config.APIConfig
	aggregator *aggregator.Aggregator
}

type Server struct {
	*http.Server
	aggregator *aggregator.Aggregator
	config     *config.WebConfig
}

func NewAPIServer(config *config.APIConfig, agg *aggregator.Aggregator) *APIServer {
	return &APIServer{
		config:     config,
		aggregator: agg,
	}
}

func (s *APIServer) Start() error {
	// TODO: Implement API server logic
	return nil
}

func NewServer(cfg *config.WebConfig, agg *aggregator.Aggregator) *Server {
	if cfg == nil {
		return nil //, fmt.Errorf("configuration cannot be nil")
	}

	if agg == nil {
		return nil //, fmt.Errorf("aggregator cannot be nil")
	}

	mux := http.NewServeMux()

	// Register handlers
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/index.html", http.StatusFound)
	})

	mux.HandleFunc("/index.html", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "web/index.html")
	})

	server := &Server{
		Server: &http.Server{
			Addr:    fmt.Sprintf(":%d", cfg.Port),
			Handler: mux,
		},
		aggregator: agg,
	}

	return server
}

// Start initializes and starts the web server
func (s *Server) Start() error {
	// TODO: Implement actual web server logic
	return nil
}

func (s *Server) Stop() error {
	if s.Server != nil {
		return s.Server.Close()
	}
	return nil //, fmt.Errorf("server is not running")
}
