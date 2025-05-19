package system

import (
	"bufio"
	"fmt"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
)

// Ethtool provides access to ethtool functionality
type Ethtool struct{}

// NewEthtool creates a new Ethtool
func NewEthtool() *Ethtool {
	return &Ethtool{}
}

// GetDriverInfo returns the driver info for a network interface
func (e *Ethtool) GetDriverInfo(name string) (string, error) {
	cmd := exec.Command("ethtool", "-i", name)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", err
	}

	scanner := bufio.NewScanner(strings.NewReader(string(output)))
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "driver:") {
			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 2 {
				return strings.TrimSpace(parts[1]), nil
			}
		}
	}

	return "", fmt.Errorf("driver not found for %s", name)
}

// GetSpeed returns the speed of a network interface in Mbps
func (e *Ethtool) GetSpeed(name string) (int, error) {
	cmd := exec.Command("ethtool", name)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return 0, err
	}

	scanner := bufio.NewScanner(strings.NewReader(string(output)))
	for scanner.Scan() {
		line := scanner.Text()
		if strings.Contains(line, "Speed:") {
			// Extract the numeric part
			re := regexp.MustCompile(`(\d+)`)
			matches := re.FindStringSubmatch(line)
			if len(matches) > 1 {
				speed, err := strconv.Atoi(matches[1])
				if err == nil {
					return speed, nil
				}
			}
		}
	}

	return 0, fmt.Errorf("speed not found for %s", name)
}

// GetRingBufferSettings returns the current and maximum ring buffer settings
func (e *Ethtool) GetRingBufferSettings(name string) (rxCur, txCur, rxMax, txMax int, err error) {
	cmd := exec.Command("ethtool", "-g", name)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return 0, 0, 0, 0, err
	}

	// Parse the output
	scanner := bufio.NewScanner(strings.NewReader(string(output)))
	inPreset := false
	inCurrent := false

	for scanner.Scan() {
		line := scanner.Text()

		if strings.Contains(line, "Pre-set maximums:") {
			inPreset = true
			inCurrent = false
			continue
		} else if strings.Contains(line, "Current hardware settings:") {
			inPreset = false
			inCurrent = true
			continue
		}

		if inPreset && strings.HasPrefix(line, "RX:") {
			parts := strings.Fields(line)
			if len(parts) >= 2 {
				rxMax, _ = strconv.Atoi(parts[1])
			}
		} else if inPreset && strings.HasPrefix(line, "TX:") {
			parts := strings.Fields(line)
			if len(parts) >= 2 {
				txMax, _ = strconv.Atoi(parts[1])
			}
		} else if inCurrent && strings.HasPrefix(line, "RX:") {
			parts := strings.Fields(line)
			if len(parts) >= 2 {
				rxCur, _ = strconv.Atoi(parts[1])
			}
		} else if inCurrent && strings.HasPrefix(line, "TX:") {
			parts := strings.Fields(line)
			if len(parts) >= 2 {
				txCur, _ = strconv.Atoi(parts[1])
			}
		}
	}

	return rxCur, txCur, rxMax, txMax, nil
}

// SetRingBuffer sets the ring buffer for a network interface
func (e *Ethtool) SetRingBuffer(name string, rx, tx int) error {
	cmd := exec.Command("ethtool", "-G", name, "rx", fmt.Sprintf("%d", rx), "tx", fmt.Sprintf("%d", tx))
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to set ring buffer: %v, output: %s", err, output)
	}
	return nil
}
