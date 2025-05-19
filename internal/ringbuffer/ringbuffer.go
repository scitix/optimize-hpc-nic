package ringbuffer

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"optimize-hpc-nic/internal/config"
	"optimize-hpc-nic/internal/logger"
	"optimize-hpc-nic/internal/nic"
)

// 网卡类型常量
const (
	NICTypeEthernet   = "Ethernet"
	NICTypeInfiniband = "Infiniband"
	NICTypeUnknown    = "Unknown"
)

// Result represents the result of an optimization operation
type Result struct {
	NIC       *nic.NIC
	Optimized bool
	Error     error
}

// Optimizer handles ring buffer optimization
type Optimizer struct {
	nicMgr  *nic.Manager
	log     *logger.Logger
	cfg     *config.Config
	ethtool ethtool
}

// New creates a new Optimizer
func New(nicMgr *nic.Manager, log *logger.Logger, cfg *config.Config) *Optimizer {
	return &Optimizer{
		nicMgr:  nicMgr,
		log:     log,
		cfg:     cfg,
		ethtool: &ethtoolWrapper{log: log},
	}
}

// ethtool interface defines methods for interacting with ethtool
type ethtool interface {
	SetRingBufferSettings(iface string, rx, tx int) error
}

// ethtoolWrapper wraps ethtool commands
type ethtoolWrapper struct {
	log *logger.Logger
}

// SetRingBufferSettings sets ring buffer settings
func (e *ethtoolWrapper) SetRingBufferSettings(iface string, rx, tx int) error {
	e.log.Debug("Setting ring buffer for %s: RX=%d, TX=%d", iface, rx, tx)
	cmd := fmt.Sprintf("ethtool -G %s rx %d tx %d", iface, rx, tx)
	_, err := system.RunCommand(cmd)
	return err
}

// OptimizeNIC optimizes a single NIC's ring buffer settings
func (o *Optimizer) OptimizeNIC(nic *nic.NIC) (bool, error) {
	// Skip Infiniband interfaces
	if nic.LinkType == NICTypeInfiniband {
		o.log.Info("Skipping Infiniband interface %s (not supported for ring buffer optimization)", nic.Name)
		return false, nil
	}

	// Check if already optimized
	if nic.IsOptimal {
		o.log.Debug("%s is already optimized (RX: %d/%d, TX: %d/%d)",
			nic.Name, nic.RXCurrent, nic.RXMax, nic.TXCurrent, nic.TXMax)
		return false, nil
	}

	// Skip if max values are not available
	if nic.RXMax <= 0 || nic.TXMax <= 0 {
		return false, fmt.Errorf("invalid max values for %s: RX=%d, TX=%d", nic.Name, nic.RXMax, nic.TXMax)
	}

	// Optimize the NIC
	err := o.ethtool.SetRingBufferSettings(nic.Name, nic.RXMax, nic.TXMax)
	if err != nil {
		return false, fmt.Errorf("failed to set ring buffer for %s: %v", nic.Name, err)
	}

	// Update NIC object to reflect new settings
	nic.RXCurrent = nic.RXMax
	nic.TXCurrent = nic.TXMax
	nic.IsOptimal = true

	return true, nil
}

// OptimizeAll optimizes all high-speed NICs
func (o *Optimizer) OptimizeAll(showAll bool) ([]*nic.NIC, error) {
	// Get all NICs
	nics, err := o.nicMgr.GetHighSpeedNICs()
	if err != nil {
		o.log.Error("Error getting NICs: %v", err)
		return nil, err
	}

	o.log.Info("Found %d high-speed physical NICs (≥%dMbps)", len(nics), o.cfg.MinSpeed)

	// 分类网卡
	var ethernetNICs []*nic.NIC
	var infinibandNICs []*nic.NIC

	for _, n := range nics {
		if n.LinkType == NICTypeInfiniband {
			o.log.Info("Skipping Infiniband interface %s (not supported for ring buffer optimization)", n.Name)
			infinibandNICs = append(infinibandNICs, n)
		} else {
			ethernetNICs = append(ethernetNICs, n)
		}
	}

	o.log.Info("Found %d Ethernet interfaces for optimization (skipped %d Infiniband interfaces)",
		len(ethernetNICs), len(infinibandNICs))

	if len(ethernetNICs) == 0 {
		o.log.Info("No Ethernet interfaces to optimize")

		if showAll {
			// 即使没有要优化的以太网接口，也显示所有网卡（包括Infiniband）
			fmt.Println("\n=== Configuration Results for High-Speed NICs (≥200G) ===")
			DisplayFormattedResults(nics)
		}

		return nil, nil
	}

	// 创建结果通道
	results := make(chan Result, len(ethernetNICs))

	// 创建等待组
	var wg sync.WaitGroup

	// 创建工作池
	workers := make(chan struct{}, o.cfg.MaxWorkers)

	// 处理每个Ethernet NIC
	for _, n := range ethernetNICs {
		wg.Add(1)
		workers <- struct{}{} // 获取工作者

		go func(n *nic.NIC) {
			defer wg.Done()
			defer func() { <-workers }() // 释放工作者

			o.log.Info("Optimizing Ethernet NIC: %s (Speed: %dMbps, Driver: %s)", n.Name, n.Speed, n.Driver)
			optimized, err := o.OptimizeNIC(n)
			results <- Result{NIC: n, Optimized: optimized, Error: err}
		}(n)
	}

	// 等待所有工作者完成
	go func() {
		wg.Wait()
		close(results)
	}()

	// 处理结果
	optimizedCount := 0
	var optimizedNICs []*nic.NIC
	var processedNICs []*nic.NIC // 处理过的以太网网卡

	for result := range results {
		n := result.NIC
		processedNICs = append(processedNICs, n)

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

	o.log.Info("Optimization complete: %d of %d Ethernet NICs optimized", optimizedCount, len(ethernetNICs))

	// 显示结果
	if showAll {
		fmt.Println("\n=== Configuration Results for High-Speed NICs (≥200G) ===")

		// 合并所有网卡列表以便显示
		var allNICs []*nic.NIC
		allNICs = append(allNICs, infinibandNICs...) // 先添加Infiniband网卡
		allNICs = append(allNICs, processedNICs...)  // 再添加处理过的以太网网卡

		DisplayFormattedResults(allNICs) // 显示所有网卡
	}

	return optimizedNICs, nil
}

// Query displays current ring buffer settings
func (o *Optimizer) Query() error {
	// 获取所有高速NIC
	nics, err := o.nicMgr.GetHighSpeedNICs()
	if err != nil {
		o.log.Error("Error querying NICs: %v", err)
		return err
	}

	// 分类网卡
	var ethernetNICs []*nic.NIC
	var infinibandNICs []*nic.NIC

	for _, n := range nics {
		if n.LinkType == NICTypeInfiniband {
			infinibandNICs = append(infinibandNICs, n)
		} else {
			ethernetNICs = append(ethernetNICs, n)
		}
	}

	o.log.Info("Found %d high-speed physical NICs (≥%dMbps): %d Ethernet, %d Infiniband",
		len(nics), o.cfg.MinSpeed, len(ethernetNICs), len(infinibandNICs))

	// 显示所有网卡的结果
	fmt.Println("\n=== Configuration Results for All High-Speed NICs (≥200G) ===")
	DisplayFormattedResults(nics) // 显示所有网卡，包括Infiniband

	return nil
}

// Monitor continuously checks and optimizes ring buffer settings
func (o *Optimizer) Monitor(interval int) error {
	o.log.Info("Starting monitor mode with interval: %d seconds", interval)

	for {
		o.log.Info("Checking ring buffer settings...")
		_, err := o.OptimizeAll(false) // 优化但不显示详细结果
		if err != nil {
			o.log.Error("Error during optimization: %v", err)
		}

		// 获取所有NIC以显示完整状态，包括Infiniband接口
		allNICs, err := o.nicMgr.GetHighSpeedNICs()
		if err != nil {
			o.log.Error("Error getting NICs: %v", err)
		} else {
			fmt.Println("\n=== Current Configuration of High-Speed NICs (≥200G) ===")
			DisplayFormattedResults(allNICs) // 显示所有网卡
		}

		o.log.Info("Sleeping for %d seconds...", interval)
		time.Sleep(time.Duration(interval) * time.Second)
	}
}

// DisplayFormattedResults formats and displays NIC information
func DisplayFormattedResults(nics []*nic.NIC) {
	// 打印表头
	fmt.Printf("%-15s %-12s %-10s %-15s %-20s %-25s %-15s\n",
		"Interface", "Speed(Mbps)", "Type", "Driver", "MAC Address", "Ring Buffer(RX/TX)", "Status")
	fmt.Println(strings.Repeat("-", 110))

	// 统计各类网卡
	ethernetCount := 0
	infinibandCount := 0
	optimizedCount := 0

	// 打印NIC
	for _, n := range nics {
		status := "SUB-OPTIMAL"

		if n.LinkType == NICTypeInfiniband {
			infinibandCount++
			status = "SKIPPED"  // Infiniband接口标记为已跳过
			ringBuffer := "N/A" // Infiniband接口无需显示环形缓冲区设置

			fmt.Printf("%-15s %-12d %-10s %-15s %-20s %-25s %-15s\n",
				n.Name, n.Speed, n.LinkType, n.Driver, n.MAC, ringBuffer, status)
		} else {
			ethernetCount++
			// Ethernet接口
			if n.IsOptimal {
				optimizedCount++
				status = "OPTIMIZED"
			}

			ringBuffer := fmt.Sprintf("%d/%d", n.RXCurrent, n.TXCurrent)
			fmt.Printf("%-15s %-12d %-10s %-15s %-20s %-25s %-15s\n",
				n.Name, n.Speed, n.LinkType, n.Driver, n.MAC, ringBuffer, status)
		}
	}

	if len(nics) == 0 {
		fmt.Println("No high-speed NICs found.")
	} else {
		// 添加摘要信息
		fmt.Println(strings.Repeat("-", 110))
		fmt.Printf("SUMMARY: Total: %d NICs | Ethernet: %d | Infiniband: %d | Optimized: %d\n",
			len(nics), ethernetCount, infinibandCount, optimizedCount)

		if infinibandCount > 0 {
			fmt.Println("NOTE: Infiniband interfaces are skipped as ring buffer optimization is not applicable")
		}
	}
}

// 提供一个简单的RunCommand接口以供系统命令执行，避免需要导入pkg/system包
var system struct {
	RunCommand func(cmd string) (string, error)
}
