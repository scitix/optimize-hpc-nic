package nic

import (
	"fmt"
	"os"
	"strings"

	"optimize-hpc-nic/internal/logger"
	"optimize-hpc-nic/pkg/system"
)

// NIC represents a network interface
type NIC struct {
	Name       string
	Speed      int
	Driver     string
	MAC        string
	RXCurrent  int
	TXCurrent  int
	RXMax      int
	TXMax      int
	IsPhysical bool
	IsOptimal  bool
}

// Manager handles NIC operations
type Manager struct {
	minSpeed int
	log      *logger.Logger
	ethtool  *system.Ethtool
}

// NewManager creates a new NIC manager
func NewManager(minSpeed int, log *logger.Logger) *Manager {
	return &Manager{
		minSpeed: minSpeed,
		log:      log,
		ethtool:  system.NewEthtool(),
	}
}

// GetAllInterfaces returns a list of all network interfaces
func (m *Manager) GetAllInterfaces() ([]string, error) {
	var interfaces []string

	// Open /sys/class/net directory
	files, err := os.ReadDir("/sys/class/net")
	if err != nil {
		return nil, fmt.Errorf("error reading network interfaces: %v", err)
	}

	// Filter out loopback interface
	for _, file := range files {
		name := file.Name()
		if name != "lo" {
			interfaces = append(interfaces, name)
		}
	}

	return interfaces, nil
}

// IsPhysicalNIC checks if a network interface is a physical device
func (m *Manager) IsPhysicalNIC(name string) bool {
	// Check if it's a virtual interface
	if _, err := os.Stat(fmt.Sprintf("/sys/devices/virtual/net/%s", name)); err == nil {
		return false
	}

	// Check if it has a physical device connection
	if _, err := os.Stat(fmt.Sprintf("/sys/class/net/%s/device", name)); err == nil {
		return true
	}

	// Check if it has a driver
	_, err := m.ethtool.GetDriverInfo(name)
	return err == nil
}

// GetNICSpeed returns the speed of a network interface in Mbps
func (m *Manager) GetNICSpeed(name string) (int, error) {
	// Try to get speed from ethtool
	speed, err := m.ethtool.GetSpeed(name)
	if err == nil && speed > 0 {
		return speed, nil
	}

	// Try to read from system file
	speedFile := fmt.Sprintf("/sys/class/net/%s/speed", name)
	if _, err := os.Stat(speedFile); err == nil {
		data, err := os.ReadFile(speedFile)
		if err == nil {
			var speedVal int
			if _, err := fmt.Sscanf(strings.TrimSpace(string(data)), "%d", &speedVal); err == nil {
				return speedVal, nil
			}
		}
	}

	return 0, fmt.Errorf("unable to determine speed for %s", name)
}

// GetNICMAC returns the MAC address of a network interface
func (m *Manager) GetNICMAC(name string) (string, error) {
	macFile := fmt.Sprintf("/sys/class/net/%s/address", name)
	data, err := os.ReadFile(macFile)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(data)), nil
}

// GetHighSpeedNICs returns a list of all high-speed physical NICs
func (m *Manager) GetHighSpeedNICs() ([]*NIC, error) {
	var nics []*NIC

	// Get all interfaces
	interfaces, err := m.GetAllInterfaces()
	if err != nil {
		return nil, err
	}

	// Process each interface
	for _, iface := range interfaces {
		if m.IsPhysicalNIC(iface) {
			nic := &NIC{
				Name:       iface,
				IsPhysical: true,
			}

			// Get speed
			speed, err := m.GetNICSpeed(iface)
			if err == nil {
				nic.Speed = speed
			}

			// Only add high-speed NICs
			if nic.Speed >= m.minSpeed {
				// Get MAC address
				mac, err := m.GetNICMAC(iface)
				if err == nil {
					nic.MAC = mac
				}

				// Get driver
				driver, err := m.ethtool.GetDriverInfo(iface)
				if err == nil {
					nic.Driver = driver
				}

				// Get ring buffer settings
				rxCur, txCur, rxMax, txMax, err := m.ethtool.GetRingBufferSettings(iface)
				if err == nil {
					nic.RXCurrent = rxCur
					nic.TXCurrent = txCur
					nic.RXMax = rxMax
					nic.TXMax = txMax
					nic.IsOptimal = (rxCur == rxMax && txCur == txMax)
				}

				nics = append(nics, nic)
			}
		}
	}

	return nics, nil
}