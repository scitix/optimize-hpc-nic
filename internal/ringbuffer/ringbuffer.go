package ringbuffer

import (
	"fmt"
	"strings"
	"sync"

	"optimize-hpc-nic/internal/config"
	"optimize-hpc-nic/internal/logger"
	"optimize-hpc-nic/internal/nic"
	"optimize-hpc-nic/pkg/system"
)

// Result represents the result of an optimization attempt
type Result struct {
	NIC       *nic.NIC
	Optimized bool
	Error     error
}

// Optimizer handles ring buffer optimization
type Optimizer struct {
	cfg     *config.Config
	log     *logger.Logger
	ethtool *system.Ethtool
	nicMgr  *nic.Manager
}

// NewOptimizer creates a new ring buffer optimizer
func NewOptimizer(cfg *config.Config, log *logger.Logger) *Optimizer {
	return &Optimizer{
		cfg:     cfg,
		log:     log,
		ethtool: system.NewEthtool(),
		nicMgr:  nic.NewManager(cfg.MinSpeed, log),
	}
}

// SetRingBuffer sets the ring buffer for a network interface
func (o *Optimizer) SetRingBuffer(nic *nic.NIC) error {
	return o.ethtool.SetRingBuffer(nic.Name, nic.RXMax, nic.TXMax)
}

// OptimizeNIC optimizes the ring buffer settings for a single NIC
func (o *Optimizer) OptimizeNIC(nic *nic.NIC) (bool, error) {
	// Check if optimization is needed
	if nic.RXCurrent == nic.RXMax && nic.TXCurrent == nic.TXMax {
		return false, nil
	}

	// Set the ring buffer
	err := o.SetRingBuffer(nic)
	if err != nil {
		return false, err
	}

	// Verify the settings
	rxCur, txCur, _, _, err := o.ethtool.GetRingBufferSettings(nic.Name)
	if err != nil {
		return false, err
	}

	// Update the current values
	nic.RXCurrent = rxCur
	nic.TXCurrent = txCur
	nic.IsOptimal = (rxCur == nic.RXMax && txCur == nic.TXMax)

	// Check if optimization was successful
	return nic.IsOptimal, nil
}

// OptimizeAll configures ring buffers for all high-speed NICs
func (o *Optimizer) OptimizeAll(showAll bool) ([]*nic.NIC, error) {
	// Get all NICs
	nics, err := o.nicMgr.GetHighSpeedNICs()
	if err != nil {
		o.log.Error("Error getting NICs: %v", err)
		return nil, err
	}

	o.log.Info("Found %d high-speed physical NICs (≥%dMbps)", len(nics), o.cfg.MinSpeed)

	// Create a channel for results
	results := make(chan Result, len(nics))

	// Create a wait group
	var wg sync.WaitGroup

	// Create a worker pool
	workers := make(chan struct{}, o.cfg.MaxWorkers)

	// Process each NIC
	for _, n := range nics {
		wg.Add(1)
		workers <- struct{}{} // Acquire a worker

		go func(nicObj *nic.NIC) {
			defer wg.Done()
			defer func() { <-workers }() // Release the worker

			o.log.Info("Optimizing NIC: %s (Speed: %dMbps, Driver: %s)", nicObj.Name, nicObj.Speed, nicObj.Driver)
			optimized, err := o.OptimizeNIC(nicObj)
			results <- Result{NIC: nicObj, Optimized: optimized, Error: err}
		}(n)
	}

	// Wait for all workers to finish
	go func() {
		wg.Wait()
		close(results)
	}()

	// Process results
	optimizedCount := 0
	var optimizedNICs []*nic.NIC
	var allNICs []*nic.NIC

	for result := range results {
		n := result.NIC
		allNICs = append(allNICs, n)

		if result.Error != nil {
			o.log.Error("Error optimizing %s: %v", n.Name, result.Error)
		} else if result.Optimized {
			o.log.Info("Successfully optimized %s (RX: %d, TX: %d)", n.Name, n.RXMax, n.TXMax)
			optimizedCount++
			optimizedNICs = append(optimizedNICs, n)
		} else {
			o.log.Info("%s already optimized (RX: %d/%d, TX: %d/%d)",
				n.Name, n.RXCurrent, n.RXMax, n.TXCurrent, n.TXMax)
			if n.IsOptimal {
				optimizedNICs = append(optimizedNICs, n)
			}
		}
	}

	o.log.Info("Optimization complete: %d of %d NICs optimized", optimizedCount, len(nics))

	// Display results
	if showAll {
		fmt.Println("\n=== Configuration Results for High-Speed NICs (≥200G) ===")
		DisplayFormattedResults(allNICs)
	}

	return optimizedNICs, nil
}

// DisplayResults displays the current ring buffer settings
func DisplayResults(nics []*nic.NIC, showAll bool) {
	fmt.Println("=== Configured High-Speed NICs (≥200G) ===")

	// Filter for optimized NICs if not showing all
	var displayNICs []*nic.NIC
	if showAll {
		displayNICs = nics
	} else {
		for _, nic := range nics {
			if nic.IsOptimal {
				displayNICs = append(displayNICs, nic)
			}
		}
	}

	DisplayFormattedResults(displayNICs)
}

// DisplayFormattedResults displays formatted results for a list of NICs
func DisplayFormattedResults(nics []*nic.NIC) {
	// Print header
	fmt.Printf("%-15s %-12s %-15s %-20s %-25s %-15s\n",
		"Interface", "Speed(Mbps)", "Driver", "MAC Address", "Ring Buffer(RX/TX)", "Status")
	fmt.Println(strings.Repeat("-", 100))

	// Print NICs
	for _, nic := range nics {
		status := "SUB-OPTIMAL"
		if nic.IsOptimal {
			status = "OPTIMIZED"
		}

		ringBuffer := fmt.Sprintf("%d/%d", nic.RXCurrent, nic.TXCurrent)
		fmt.Printf("%-15s %-12d %-15s %-20s %-25s %-15s\n",
			nic.Name, nic.Speed, nic.Driver, nic.MAC, ringBuffer, status)
	}

	if len(nics) == 0 {
		fmt.Println("No optimized high-speed NICs found.")
	}
}
