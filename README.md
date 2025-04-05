# Whosay - A Developer-Friendly System Monitor

![Whosay](https://img.shields.io/badge/version-0.1.0-blue)

Whosay is a lightweight, easy-to-use system monitoring tool designed specifically for developers. It provides real-time insights into your system's resources with a clean, modern terminal interface.

## Features

- **System Overview**: Get detailed information about your operating system, kernel, and hardware
- **Resource Monitoring**: Track CPU, memory, and disk usage in real-time
- **Process Management**: View and monitor running processes sorted by resource usage
- **Network Activity**: Monitor network interfaces and bandwidth usage
- **Docker Integration**: View and inspect running containers with resource metrics
- **Temperature Monitoring**: Keep an eye on CPU and system temperatures
- **Battery Information**: View battery status, health, and estimated time remaining
- **Watch Mode**: Continuous monitoring with automatic refreshing
- **JSON Output**: Export data in JSON format for integration with other tools

## Installation

### From Source

```bash
# Clone the repository
git clone https://github.com/tiwariParth/whosay.git

# Navigate to the project directory
cd whosay

# Build the binary
go build -o whosay

# Optionally, move to a directory in your PATH
sudo mv whosay /usr/local/bin/
```

## Usage

Whosay offers various commands to monitor different aspects of your system:

```bash
# Display basic system information
./whosay -sys

# Monitor CPU usage
./whosay -cpu

# Monitor memory usage
./whosay -mem

# Monitor disk usage
./whosay -disk

# Monitor network information
./whosay -net

# Monitor network traffic
./whosay -nettraffic

# View process information
./whosay -proc

# View docker containers
./whosay -docker

# Monitor container logs (specify container name or ID)
./whosay -container-logs <container-name>

# Display battery status (on laptops)
./whosay -battery

# Monitor temperature sensors
./whosay -temp

# View system logs
./whosay -logs

# Show all information
./whosay -all

# Enable continuous monitoring with watch mode
./whosay -all -watch

# Set refresh rate for watch mode (in seconds)
./whosay -all -watch -refresh 2
```

### Output Options

```bash
# Get JSON output instead of text display
./whosay -cpu -json

# Show more detailed information
./whosay -cpu -verbose

# Disable colors in output
./whosay -cpu -no-color
```

## Example Output

When you run Whosay, you'll see a beautifully formatted terminal UI that looks something like this:

```
 ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
 │ ℹ System │
 ┣───────────────────────────────────────────────┫
 │  OS:        Ubuntu 22.04                      │
 │  Hostname:  dev-machine                       │
 │  Kernel:    5.15.0-67-generic                 │
 │  Uptime:    2d 3h 45m                         │
 ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

 ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
 │ ⚙ CPU │
 ┣───────────────────────────────────────────────┫
 │  CPUs:           8 cores                      │
 │  Architecture:   x86_64 (64-bit)              │
 │  Usage:          32.5%                        │
 │                  ■■■■■□□□□□□□□□□□ 32.5%       │
 ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
```

## Understanding the Output

- **System Section**: Shows your OS details, hostname, and kernel version
- **CPU Section**: Displays core count, architecture, and current usage
- **Memory Section**: Shows total, used, and free memory with usage bar
- **Disk Section**: Indicates storage capacity and usage for your file systems
- **Process Section**: Lists the top processes consuming resources
- **Network Section**: Shows interface details and current connectivity
- **Docker Section**: Lists running containers with their resource usage

## Color Coding

Whosay uses color to help you quickly understand the state of your system:

- **Green**: Normal resource usage (0-60%)
- **Yellow**: Moderate resource usage (60-85%)
- **Red**: High resource usage (85-100%)

## Docker Container Monitoring

Monitor your Docker containers without switching tools:

```bash
# Get general Docker information
./whosay -docker

# Monitor logs from a specific container
./whosay -container-logs nginx -logs-limit 50
```

## Advanced Features

### Resource Usage Trends

When using watch mode, Whosay can show resource usage over time:

```bash
./whosay -cpu -mem -nettraffic -watch
```

### System Temperature Monitoring

Keep your system's health in check by monitoring temperatures:

```bash
./whosay -temp -watch
```

## Compatibility

Whosay works on:

- Linux (Ubuntu, Debian, Fedora, etc.)
- macOS
- Windows 
- WSL (Windows Subsystem for Linux)

## Contributing

Contributions are welcome! Feel free to open issues or submit pull requests on GitHub.

## License

This project is licensed under the MIT License - see the LICENSE file for details.

---

_"Stay resourceful with Whosay!" 🚀_
