
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

# Requirements

- Linux operating system
- Ethtool package installed (apt-get install ethtool)
- Root privileges (for changing NIC settings)
- Go 1.18+ (for building from source)

# Installation

## Using Debian Package (Recommended)

```bash
# Install the package
sudo dpkg -i optimize-hpc-nic_1.0.0_amd64.deb

# Resolve dependencies if needed
sudo apt-get install -f
Building from Source
```

# Clone the repository

git clone <https://github.com/yourusername/optimize-hpc-nic.git>
cd optimize-hpc-nic

# Build the binary

make build

# Install the application

sudo make install
Usage
Command Line Options
java
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
Examples
bash

# Show currently optimized NICs (query mode)

sudo optimize-hpc-nic

# Configure all high-speed NICs once

sudo optimize-hpc-nic -s

# Start continuous monitoring (check every 2 minutes)

sudo optimize-hpc-nic -m -interval 120

# Configure NICs with verbose output

sudo optimize-hpc-nic -s -v
Systemd Service
When installed via the Debian package, a systemd service is automatically created and started:

bash

# Check service status

systemctl status optimize-hpc-nic

# Start the service

systemctl start optimize-hpc-nic

# Stop the service

systemctl stop optimize-hpc-nic

# View logs

journalctl -u optimize-hpc-nic
Project Structure
graphql
optimize-hpc-nic/
├── cmd/
│   └── optimize-hpc-nic/     # Application entry point
│       └── main.go
├── internal/
│   ├── config/               # Configuration management
│   ├── logger/               # Logging system
│   ├── monitor/              # Continuous monitoring service
│   ├── nic/                  # NIC management and detection
│   └── ringbuffer/           # Ring buffer operations
├── pkg/
│   └── system/               # System utilities (ethtool wrapper)
├── scripts/                  # Build and packaging scripts
├── go.mod                    # Go module file
├── go.sum                    # Go module checksums
├── Makefile                  # Build automation
└── optimize-hpc-nic.service  # Systemd service file
Building Debian Packages
The project includes a script to create Debian packages for easy distribution:

bash

# Build the .deb package

make package

# The package will be created in the build/ directory

Logging
Logs are written to /var/log/optimize-hpc-nic/optimize-hpc-nic.log by default. Log rotation is configured to:

Keep logs up to 50MB in size
Maintain 3 backup files
Compress old logs
Delete logs older than 28 days
Configuration Options
All configuration is done via command-line flags. There is no separate configuration file.

Key configuration values:

Minimum NIC Speed: 200,000 Mbps (200Gbps)
Default Monitoring Interval: 300 seconds (5 minutes)
Default Worker Count: 5 parallel workers
Known Issues and Limitations
The tool requires ethtool to be installed
Some ring buffer settings may require network service restart to take full effect
Only supports Linux systems with systemd
Contributing
Contributions are welcome! Please feel free to submit a Pull Request.

Fork the repository
Create your feature branch (git checkout -b feature/amazing-feature)
Commit your changes (git commit -m 'Add some amazing feature')
Push to the branch (git push origin feature/amazing-feature)
Open a Pull Request
License
This project is licensed under the MIT License - see the LICENSE file for details.

Acknowledgments
The Linux ethtool project
Go programming language and community
For bug reports and feature requests, please open an issue on the GitHub repository.
