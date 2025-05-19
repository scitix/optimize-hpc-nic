
# optimize-hpc-nic

A high-performance tool for optimizing ring buffer settings on high-speed network interfaces (≥200G) in Linux environments. Written in Go with parallel processing capabilities, this tool automatically detects and configures network interfaces to use optimal ring buffer settings for improved network performance.

# Features

- High-Speed NIC Detection: Automatically identifies physical NICs with speeds of 200Gbps or higher
- Parallel Processing: Configures multiple NICs simultaneously for faster execution
 Multiple Operation Modes:
  - Query Mode: Display current ring buffer settings (default)
  - Set Mode: Configure ring buffers to optimal settings
  - Monitor Mode: Continuously monitor and maintain optimal settings
- Comprehensive Logging: Structured logging with rotation and size management
- Debian Package Support: Easy installation via .deb package
- Systemd Integration: Automatic startup and monitoring


# Installation

## Using Debian Package (Recommended)

```bash
# Install the package
sudo dpkg -i optimize-hpc-nic_1.0.0_amd64.deb
```

# Build the binary

```bash
make build
```

# Install the application

```bash
sudo make install
```

# Building Debian Packages

The project includes a script to create Debian packages for easy distribution:

``` bash

# Build the .deb package
make package
```

## Usage

```bash
Usage: optimize-hpc-nic [options]

Options:
  -q                   Query mode - show current settings (default)
  -s                   Set mode - optimize ring buffers once
  -m                   Monitor mode - continuously monitor and adjust settings
  -interval int        Monitoring interval in seconds (default: 300)
  -min-speed int       Minimum NIC speed in Mbps (default: 200000)
  -workers int         Maximum number of parallel workers (default: 5)
  -v                   Verbose output
  -log string          Log file path (default: /var/log/optimize-hpc-nic/optimize-hpc-nic.log)
```

## Examples

```bash
# Show currently optimized NICs (query mode)
root@ops# optimize-hpc-nic s
=== Configured High-Speed NICs (≥200G) ===
Interface       Speed(Mbps)  Driver          MAC Address          Ring Buffer(RX/TX)        Status
----------------------------------------------------------------------------------------------------
eth0            400000       mlx5_core       c4:70:bd:fe:2b:8c    8192/8192                 OPTIMIZED
eth1            400000       mlx5_core       c4:70:bd:f7:1d:8a    8192/8192                 OPTIMIZED
eth2            400000       mlx5_core       c4:70:bd:f6:fe:ca    8192/8192                 OPTIMIZED
eth3            400000       mlx5_core       c4:70:bd:f6:fe:92    8192/8192                 OPTIMIZED
eth4            400000       mlx5_core       c4:70:bd:f6:fe:d2    8192/8192                 OPTIMIZED
eth5            400000       mlx5_core       c4:70:bd:f6:fe:6a    8192/8192                 OPTIMIZED
eth6            400000       mlx5_core       c4:70:bd:f7:1d:6a    8192/8192                 OPTIMIZED
eth7            400000       mlx5_core       c4:70:bd:fe:2b:84    8192/8192                 OPTIMIZED

# Configure all high-speed NICs once
root@ops:~# optimize-hpc-nic -v -s
2025/05/19 13:55:55 [INFO] optimize-hpc-nic starting with mode: set
2025/05/19 13:55:55 [INFO] Configuring ring buffers for high-speed NICs
2025/05/19 13:55:55 [INFO] Found 8 high-speed physical NICs (≥200000Mbps)
2025/05/19 13:55:55 [INFO] Optimizing NIC: eth4 (Speed: 400000Mbps, Driver: mlx5_core)
2025/05/19 13:55:55 [INFO] Optimizing NIC: eth1 (Speed: 400000Mbps, Driver: mlx5_core)
2025/05/19 13:55:55 [INFO] Optimizing NIC: eth5 (Speed: 400000Mbps, Driver: mlx5_core)
2025/05/19 13:55:55 [INFO] Optimizing NIC: eth2 (Speed: 400000Mbps, Driver: mlx5_core)
2025/05/19 13:55:55 [INFO] Optimizing NIC: eth6 (Speed: 400000Mbps, Driver: mlx5_core)
2025/05/19 13:55:55 [INFO] Optimizing NIC: eth7 (Speed: 400000Mbps, Driver: mlx5_core)
2025/05/19 13:55:55 [INFO] Optimizing NIC: eth3 (Speed: 400000Mbps, Driver: mlx5_core)
2025/05/19 13:55:55 [INFO] Optimizing NIC: eth0 (Speed: 400000Mbps, Driver: mlx5_core)
2025/05/19 13:55:55 [INFO] eth4 already optimized (RX: 8192/8192, TX: 8192/8192)
2025/05/19 13:55:55 [INFO] eth1 already optimized (RX: 8192/8192, TX: 8192/8192)
2025/05/19 13:55:55 [INFO] eth5 already optimized (RX: 8192/8192, TX: 8192/8192)
2025/05/19 13:55:55 [INFO] eth2 already optimized (RX: 8192/8192, TX: 8192/8192)
2025/05/19 13:55:55 [INFO] eth6 already optimized (RX: 8192/8192, TX: 8192/8192)
2025/05/19 13:55:55 [INFO] eth7 already optimized (RX: 8192/8192, TX: 8192/8192)
2025/05/19 13:55:55 [INFO] eth3 already optimized (RX: 8192/8192, TX: 8192/8192)
2025/05/19 13:55:55 [INFO] eth0 already optimized (RX: 8192/8192, TX: 8192/8192)
2025/05/19 13:55:55 [INFO] Optimization complete: 0 of 8 NICs optimized

# Check service status
systemctl status optimize-hpc-nic
root@ops:~# systemctl status optimize-hpc-nic
● optimize-hpc-nic.service - Optimize and Monitor Ring Buffers for High-Performance NICs (≥200G)
     Loaded: loaded (/lib/systemd/system/optimize-hpc-nic.service; enabled; vendor preset: enabled)
     Active: active (running) since Mon 2025-05-19 12:05:04 CST; 1h 53min ago
   Main PID: 1559866 (optimize-hpc-ni)
      Tasks: 33 (limit: 629145)
     Memory: 9.0M
        CPU: 12.303s
     CGroup: /system.slice/optimize-hpc-nic.service
             └─1559866 /usr/local/bin/optimize-hpc-nic -m -interval 300

May 19 12:05:04 hercules-g86-188 systemd[1]: Started Optimize and Monitor Ring Buffers for High-Performance NICs (≥200G).

```
