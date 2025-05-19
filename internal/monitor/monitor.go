package monitor

import (
	"time"

	"optimize-hpc-nic/internal/config"
	"optimize-hpc-nic/internal/logger"
	"optimize-hpc-nic/internal/ringbuffer"
)

// Service is the monitoring service
type Service struct {
	cfg      *config.Config
	log      *logger.Logger
	optimizer *ringbuffer.Optimizer
	stopChan chan struct{}
}

// New creates a new monitoring service
func New(cfg *config.Config, log *logger.Logger) *Service {
	return &Service{
		cfg:      cfg,
		log:      log,
		optimizer: ringbuffer.NewOptimizer(cfg, log),
		stopChan: make(chan struct{}),
	}
}

// Start starts the monitoring service
func (s *Service) Start() {
	s.log.Info("Starting monitoring with interval: %d seconds", s.cfg.MonitorInterval)

	// Initial configuration
	s.optimizer.OptimizeAll(false)

	// Monitor loop
	ticker := time.NewTicker(time.Duration(s.cfg.MonitorInterval) * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			s.log.Info("Performing scheduled ring buffer check")
			s.checkAndOptimize()
		case <-s.stopChan:
			s.log.Info("Monitoring service stopped")
			return
		}
	}
}

// Stop stops the monitoring service
func (s *Service) Stop() {
	close(s.stopChan)
}

// checkAndOptimize checks and optimizes ring buffer settings
func (s *Service) checkAndOptimize() {
	// Optimize all NICs
	s.optimizer.OptimizeAll(false)
}