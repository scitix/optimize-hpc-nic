package main

import (
	"os"
	"os/signal"
	"syscall"

	"optimize-hpc-nic/internal/config"
	"optimize-hpc-nic/internal/logger"
	"optimize-hpc-nic/internal/monitor"
	"optimize-hpc-nic/internal/nic"
	"optimize-hpc-nic/internal/ringbuffer"
)

func main() {
	// Parse command-line arguments
	cfg := config.ParseFlags()

	// Initialize logger
	log := logger.New(cfg.LogFile, cfg.LogMaxSize, cfg.LogMaxBackups, cfg.LogMaxAge, cfg.Verbose)
	defer log.Close()

	log.Info("optimize-hpc-nic starting with mode: %s", cfg.Mode)

	// Create signal channel for graceful shutdown
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	// Execute mode-specific operations
	switch cfg.Mode {
	case config.ModeMonitor:
		log.Info("Starting monitoring mode with interval: %d seconds", cfg.MonitorInterval)
		monitorService := monitor.New(cfg, log)
		go func() {
			<-sigs
			log.Info("Received shutdown signal, stopping service")
			monitorService.Stop()
		}()
		monitorService.Start()

	case config.ModeSet:
		log.Info("Configuring ring buffers for high-speed NICs")
		optimizer := ringbuffer.NewOptimizer(cfg, log)
		optimizer.OptimizeAll(true)

	case config.ModeQuery:
		nicManager := nic.NewManager(cfg.MinSpeed, log)
		nics, err := nicManager.GetHighSpeedNICs()
		if err != nil {
			log.Error("Failed to get NICs: %v", err)
			os.Exit(1)
		}

		ringbuffer.DisplayResults(nics, false)
	}
}
