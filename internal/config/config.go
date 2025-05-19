package config

import (
	"flag"
)

const (
	// Mode constants
	ModeQuery   = "query"
	ModeSet     = "set"
	ModeMonitor = "monitor"

	// Default values
	DefaultMinSpeed        = 200000 // 200G in Mbps
	DefaultMonitorInterval = 300    // seconds
	DefaultMaxWorkers      = 5
	DefaultLogFile         = "/var/log/optimize-hpc-nic/optimize-hpc-nic.log"
	DefaultLogMaxSize      = 50     // MB
	DefaultLogMaxBackups   = 3
	DefaultLogMaxAge       = 28     // days
)

// Config holds all configuration options
type Config struct {
	// Mode settings
	Mode            string
	MonitorInterval int
	MinSpeed        int
	MaxWorkers      int
	Verbose         bool

	// Logging settings
	LogFile       string
	LogMaxSize    int
	LogMaxBackups int
	LogMaxAge     int
}

// ParseFlags parses command line flags and returns a Config
func ParseFlags() *Config {
	cfg := &Config{
		Mode:            ModeQuery,
		MinSpeed:        DefaultMinSpeed,
		MonitorInterval: DefaultMonitorInterval,
		MaxWorkers:      DefaultMaxWorkers,
		LogFile:         DefaultLogFile,
		LogMaxSize:      DefaultLogMaxSize,
		LogMaxBackups:   DefaultLogMaxBackups,
		LogMaxAge:       DefaultLogMaxAge,
	}

	// Define flags
	setMode := flag.Bool("s", false, "Set ring buffer mode (optimize NICs)")
	monitorMode := flag.Bool("m", false, "Monitor ring buffer settings continuously")
	queryMode := flag.Bool("q", false, "Query current ring buffer settings (default)")
	flag.IntVar(&cfg.MonitorInterval, "interval", DefaultMonitorInterval, "Monitor interval in seconds")
	flag.IntVar(&cfg.MinSpeed, "min-speed", DefaultMinSpeed, "Minimum NIC speed in Mbps")
	flag.IntVar(&cfg.MaxWorkers, "workers", DefaultMaxWorkers, "Maximum number of parallel workers")
	flag.BoolVar(&cfg.Verbose, "v", false, "Verbose output")
	flag.StringVar(&cfg.LogFile, "log", DefaultLogFile, "Log file path")

	// Parse flags
	flag.Parse()

	// Determine mode
	if *monitorMode {
		cfg.Mode = ModeMonitor
	} else if *setMode {
		cfg.Mode = ModeSet
	} else if *queryMode || (!*monitorMode && !*setMode) {
		cfg.Mode = ModeQuery
	}

	return cfg
}