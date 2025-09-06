

<p align="center">
    <img src="images/logo.png" alt="SSHM Logo" width="120" />
</p>

# 🚀 SSHM - SSH Manager

[![Go](https://img.shields.io/badge/Go-1.23+-00ADD8?style=for-the-badge&logo=go)](https://golang.org/)
[![Release](https://img.shields.io/github/v/release/Gu1llaum-3/sshm?style=for-the-badge)](https://github.com/Gu1llaum-3/sshm/releases)
[![License](https://img.shields.io/github/license/Gu1llaum-3/sshm?style=for-the-badge)](LICENSE)
[![Platform](https://img.shields.io/badge/platform-Linux%20%7C%20macOS%20%7C%20Windows-lightgrey?style=for-the-badge)](https://github.com/Gu1llaum-3/sshm/releases)

> **A modern, interactive SSH Manager for your terminal** 🔥

SSHM is a beautiful command-line tool that transforms how you manage and connect to your SSH hosts. Built with Go and featuring an intuitive TUI interface, it makes SSH connection management effortless and enjoyable.

<p align="center">
    <a href="images/sshm.gif" target="_blank">
        <img src="images/sshm.gif" alt="Demo SSHM Terminal" width="800" />
    </a>
    <br>
    <em>🖱️ Click on the image to view in full size</em>
</p>

## ✨ Features

### 🎯 **Core Features**
- **🎨 Beautiful TUI Interface** - Navigate your SSH hosts with an elegant, interactive terminal UI
- **⚡ Quick Connect** - Connect to any host instantly
- **🔄 Port Forwarding** - Easy setup for Local, Remote, and Dynamic (SOCKS) forwarding
- **📝Easy Management** - Add, edit, and manage SSH configurations seamlessly
- **🏷️ Tag Support** - Organize your hosts with custom tags for better categorization
- **🔍 Smart Search** - Find hosts quickly with built-in filtering and search
- **🔒 Secure** - Works directly with your existing `~/.ssh/config` file
- **📁 Custom Config Support** - Use any SSH configuration file with the `-c` flag
- **📂 SSH Include Support** - Full support for SSH Include directives to organize configurations across multiple files
- **⚙️ SSH Options Support** - Add any SSH configuration option through intuitive forms
- **🔄 Automatic Conversion** - Seamlessly converts between command-line and config formats

### 🛠️ **Management Operations**
- **Add new SSH hosts** with interactive forms
- **Edit existing configurations** in-place
- **Delete hosts** with confirmation prompts
- **Port forwarding setup** with intuitive interface for Local (-L), Remote (-R), and Dynamic (-D) forwarding
- **Backup configurations** automatically before changes
- **Validate settings** to prevent configuration errors
- **ProxyJump support** for secure connection tunneling through bastion hosts
- **SSH Options management** - Add any SSH option with automatic format conversion
- **Full SSH compatibility** - Maintains compatibility with standard SSH tools

### 🎮 **User Experience**
- **Zero configuration** - Works out of the box with your existing SSH setup
- **Keyboard shortcuts** for power users
- **Cross-platform** - Supports Linux, macOS (Intel & Apple Silicon), and Windows
- **Lightweight** - Single binary with no dependencies

## 🚀 Quick Start

### Installation

**Unix/Linux/macOS (One-line install):**
```bash
curl -sSL https://raw.githubusercontent.com/Gu1llaum-3/sshm/main/install/unix.sh | bash
```

**Windows (PowerShell):**
```powershell
irm https://raw.githubusercontent.com/Gu1llaum-3/sshm/main/install/windows.ps1 | iex
```

**Alternative methods:**

*Linux/macOS:*
```bash
# Download specific release
wget https://github.com/Gu1llaum-3/sshm/releases/latest/download/sshm-linux-amd64.tar.gz

# Extract and install
tar -xzf sshm-linux-amd64.tar.gz
sudo mv sshm-linux-amd64 /usr/local/bin/sshm
```

*Windows:*
```powershell
# Download and extract
Invoke-WebRequest -Uri "https://github.com/Gu1llaum-3/sshm/releases/latest/download/sshm-windows-amd64.zip" -OutFile "sshm-windows-amd64.zip"
Expand-Archive sshm-windows-amd64.zip -DestinationPath C:\tools\
# Add C:\tools to your PATH environment variable
```

## 📖 Usage

### Interactive Mode

Launch SSHM without arguments to enter the beautiful TUI interface:

```bash
sshm
```

**Navigation:**
- `↑/↓` or `j/k` - Navigate hosts
- `Enter` - Connect to selected host
- `a` - Add new host
- `e` - Edit selected host
- `d` - Delete selected host
- `f` - Port forwarding setup
- `q` - Quit
- `/` - Search/filter hosts

**Sorting & Filtering:**
- `s` - Switch between sorting modes (name ↔ last login)
- `n` - Sort by **name** (alphabetical)
- `r` - Sort by **recent** (last login time)
- `Tab` - Cycle between filtering modes
- Filter by **name** (default) - Search through host names
- Filter by **last login** - Sort and filter by most recently used connections

The interactive forms will guide you through configuration:
- **Hostname/IP** - Server address
- **Username** - SSH user
- **Port** - SSH port (default: 22)
- **Identity File** - Private key path
- **ProxyJump** - Jump server for connection tunneling
- **SSH Options** - Additional SSH options in `-o` format (e.g., `-o Compression=yes -o ServerAliveInterval=60`)
- **Tags** - Comma-separated tags for organization

### Port Forwarding

SSHM provides an intuitive interface for setting up SSH port forwarding. Press `f` while selecting a host to open the port forwarding setup:

**Forward Types:**
- **Local (-L)** - Forward a local port to a remote host/port through the SSH connection
  - Example: Access a remote database on `localhost:5432` via local port `15432`
  - Use case: `ssh -L 15432:localhost:5432 server` → Database accessible on `localhost:15432`

- **Remote (-R)** - Forward a remote port back to a local host/port
  - Example: Expose local web server on remote host's port `8080`
  - Use case: `ssh -R 8080:localhost:3000 server` → Local app accessible from remote host's port 8080
  - ⚠️ **Requirements for external access:**
    - **SSH Server Config**: Add `GatewayPorts yes` to `/etc/ssh/sshd_config` and restart SSH service
    - **Firewall**: Open the remote port in the server's firewall (`ufw allow 8080` or equivalent)
    - **Port Availability**: Ensure the remote port is not already in use
    - **Bind Address**: Use `0.0.0.0` for external access, `127.0.0.1` for local-only

- **Dynamic (-D)** - Create a SOCKS proxy for secure browsing
  - Example: Route web traffic through the SSH connection
  - Use case: `ssh -D 1080 server` → Configure browser to use `localhost:1080` as SOCKS proxy
  - ⚠️ **Configuration requirements:**
    - **Browser Setup**: Configure SOCKS v5 proxy in browser settings
    - **DNS**: Enable "Proxy DNS when using SOCKS v5" for full privacy
    - **Applications**: Only SOCKS-aware applications will use the proxy
    - **Bind Address**: Use `127.0.0.1` for security (local access only)

**Port Forwarding Interface:**
- Choose forward type with ←/→ arrow keys
- Configure ports and addresses with guided forms
- Optional bind address configuration (defaults to 127.0.0.1)
- Real-time validation of port numbers and addresses
- Connect automatically with configured forwarding options

**Troubleshooting Port Forwarding:**

*Remote Forwarding Issues:*
```bash
# Error: "remote port forwarding failed for listen port X"
# Solutions:
1. Check if port is already in use: ssh server "netstat -tln | grep :X"
2. Use a different port that's available
3. Enable GatewayPorts in SSH config for external access
```

*SSH Server Configuration for Remote Forwarding:*
```bash
# Edit SSH daemon config on the server:
sudo nano /etc/ssh/sshd_config

# Add or uncomment:
GatewayPorts yes

# Restart SSH service:
sudo systemctl restart sshd  # Ubuntu/Debian/CentOS 7+
# OR
sudo service ssh restart     # Older systems
```

*Firewall Configuration:*
```bash
# Ubuntu/Debian (UFW):
sudo ufw allow [port_number]

# CentOS/RHEL/Rocky (firewalld):
sudo firewall-cmd --add-port=[port_number]/tcp --permanent
sudo firewall-cmd --reload

# Check if port is accessible:
telnet [server_ip] [port_number]
```

*Dynamic Forwarding (SOCKS) Browser Setup:*
```
Firefox: about:preferences → Network Settings
- Manual proxy configuration
- SOCKS Host: localhost, Port: [your_port]
- SOCKS v5: ✓
- Proxy DNS when using SOCKS v5: ✓

Chrome: Launch with proxy
chrome --proxy-server="socks5://localhost:[your_port]"
```

### CLI Usage

SSHM provides both command-line operations and an interactive TUI interface:

```bash
# Launch interactive TUI mode for browsing and connecting to hosts
sshm

# Launch TUI with custom SSH config file
sshm -c /path/to/custom/ssh_config

# Add a new host using interactive form
sshm add

# Add a new host with pre-filled hostname
sshm add hostname

# Add a new host with custom SSH config file
sshm add hostname -c /path/to/custom/ssh_config

# Edit an existing host configuration
sshm edit my-server

# Edit host with custom SSH config file
sshm edit my-server -c /path/to/custom/ssh_config

# Show version information
sshm --version

# Show help and available commands
sshm --help
```

### Backup Configuration

SSHM automatically creates backups of your SSH configuration files before making any changes to ensure your configurations are safe.

**Backup Location:**
- **Unix/Linux/macOS**: `~/.config/sshm/backups/` (or `$XDG_CONFIG_HOME/sshm/backups/` if set)
- **Windows**: `%APPDATA%\sshm\backups\` (fallback: `%USERPROFILE%\.config\sshm\backups\`)

**Key Features:**
- Automatic backup before any modification
- One backup per file (overwrites previous backup)
- Stored separately to avoid SSH Include conflicts
- Easy manual recovery if needed

**Quick Recovery:**
```bash
# Unix/Linux/macOS
cp ~/.config/sshm/backups/config.backup ~/.ssh/config

# Windows
copy "%APPDATA%\sshm\backups\config.backup" "%USERPROFILE%\.ssh\config"
```

### Configuration File Options

By default, SSHM uses the standard SSH configuration file at `~/.ssh/config`. You can specify a different configuration file using the `-c` flag:

```bash
# Use custom config file in TUI mode
sshm -c /path/to/custom/ssh_config

# Use custom config file with commands
sshm add hostname -c /path/to/custom/ssh_config
sshm edit hostname -c /path/to/custom/ssh_config
```

### Platform-Specific Notes

**Windows:**
- SSHM works with the built-in OpenSSH client (Windows 10/11)
- Configuration file location: `%USERPROFILE%\.ssh\config`
- Compatible with WSL SSH configurations
- Supports the same SSH options as Unix systems

**Unix/Linux/macOS:**
- Standard SSH configuration file: `~/.ssh/config`
- Full compatibility with OpenSSH features
- Preserves file permissions automatically

## 🏗️ Configuration

SSHM works directly with your standard SSH configuration file (`~/.ssh/config`). It adds special comment tags for enhanced functionality while maintaining full compatibility with standard SSH tools.

### SSH Include Support

SSHM fully supports SSH Include directives, allowing you to organize your SSH configurations across multiple files. This is particularly useful for managing large numbers of hosts or organizing configurations by environment, project, or team.

**Include Examples:**
```ssh
# Main ~/.ssh/config file
Host personal-server
    HostName personal.example.com
    User myuser

# Include work-related configurations
Include work-servers.conf

# Include all configurations from a directory
Include projects/*

# Include with relative paths
Include ~/.ssh/configs/production.conf
```

**Organization Examples:**

*work-servers.conf:*
```ssh
# Tags: work, production
Host prod-web-01
    HostName 10.0.1.10
    User deploy
    ProxyJump bastion.company.com

# Tags: work, staging  
Host staging-api
    HostName staging-api.company.com
    User developer
```

*projects/client-alpha.conf:*
```ssh
# Tags: client, development
Host client-alpha-dev
    HostName dev.client-alpha.com
    User admin
    Port 2222
```

**Example configuration:**
Include ~/.ssh/conf.d/*

```ssh
# Tags: production, web, frontend
Host web-prod-01
    HostName 192.168.1.10
    User deploy
    Port 22
    IdentityFile ~/.ssh/production_key
    Compression yes
    ServerAliveInterval 60

# Tags: development, database
Host db-dev
    HostName dev-db.company.com
    User admin
    Port 2222
    IdentityFile ~/.ssh/dev_key
    StrictHostKeyChecking no
    UserKnownHostsFile /dev/null

# Tags: production, backend
Host backend-prod
    HostName 10.0.1.50
    User app
    Port 22
    ProxyJump bastion.company.com
    IdentityFile ~/.ssh/production_key
    Compression yes
    ServerAliveInterval 300
    BatchMode yes
```

### Supported SSH Options

SSHM supports all standard SSH configuration options:

**Built-in Fields:**
- `HostName` - Server hostname or IP address
- `User` - Username for SSH connection
- `Port` - SSH port number
- `IdentityFile` - Path to private key file
- `ProxyJump` - Jump server for connection tunneling (e.g., `user@jumphost:port`)
- `Tags` - Custom tags (SSHM extension)

**Additional SSH Options:**
You can add any valid SSH option using the "SSH Options" field in the interactive forms. Enter them in command-line format (e.g., `-o Compression=yes -o ServerAliveInterval=60`) and SSHM will automatically convert them to the proper SSH config format.

**Common SSH Options:**
- `Compression` - Enable/disable compression (`yes`/`no`)
- `ServerAliveInterval` - Interval in seconds for keepalive messages
- `ServerAliveCountMax` - Maximum number of keepalive messages
- `StrictHostKeyChecking` - Host key verification (`yes`/`no`/`ask`)
- `UserKnownHostsFile` - Path to known hosts file
- `BatchMode` - Disable interactive prompts (`yes`/`no`)
- `ConnectTimeout` - Connection timeout in seconds
- `ControlMaster` - Connection multiplexing (`yes`/`no`/`auto`)
- `ControlPath` - Path for control socket
- `ControlPersist` - Keep connection alive duration
- `ForwardAgent` - Forward SSH agent (`yes`/`no`)
- `LocalForward` - Local port forwarding (e.g., `8080:localhost:80`)
- `RemoteForward` - Remote port forwarding
- `DynamicForward` - SOCKS proxy port forwarding

**Example usage in forms:**
```
SSH Options: -o Compression=yes -o ServerAliveInterval=60 -o StrictHostKeyChecking=no
```

This will be automatically converted to:
```ssh
    Compression yes
    ServerAliveInterval 60
    StrictHostKeyChecking no
```

## 🛠️ Development

### Prerequisites

- Go 1.23+ 
- Git

### Build from Source

```bash
# Clone the repository
git clone https://github.com/Gu1llaum-3/sshm.git
cd sshm

# Build the binary
go build -o sshm .

# Run
./sshm
```

### Project Structure

```
sshm/
├── main.go             # Application entry point
├── cmd/                # CLI commands (Cobra)
│   ├── root.go         # Root command and interactive mode
│   ├── add.go          # Add host command
│   ├── edit.go         # Edit host command
│   └── search.go       # Search command
├── internal/
│   ├── config/         # SSH configuration management
│   │   └── ssh.go      # Config parsing and manipulation
│   ├── history/        # Connection history tracking
│   │   └── history.go  # History management and last login tracking
│   ├── ui/             # Terminal UI components (Bubble Tea)
│   │   ├── tui.go      # Main TUI interface and program setup
│   │   ├── model.go    # Core TUI model and state
│   │   ├── update.go   # Message handling and state updates
│   │   ├── view.go     # UI rendering and layout
│   │   ├── table.go    # Host list table component
│   │   ├── add_form.go # Add host form interface
│   │   ├── edit_form.go# Edit host form interface
│   │   ├── styles.go   # Lip Gloss styling definitions
│   │   ├── sort.go     # Sorting and filtering logic
│   │   └── utils.go    # UI utility functions
│   └── validation/     # Input validation
│       └── ssh.go      # SSH config validation
├── images/             # Documentation assets
│   ├── logo.png        # Project logo
│   └── sshm.gif        # Demo animation
├── install/            # Installation scripts
│   ├── unix.sh         # Unix/Linux/macOS installer
│   └── README.md       # Installation guide
├── .github/            # GitHub configuration
│   ├── copilot-instructions.md # Development guidelines
│   └── workflows/      # CI/CD pipelines
│       └── build.yml   # Multi-platform builds
├── go.mod              # Go module definition
├── go.sum              # Go module checksums
├── LICENSE             # MIT license
└── README.md           # Project documentation
```

### Dependencies

- [Cobra](https://github.com/spf13/cobra) - CLI framework
- [Bubble Tea](https://github.com/charmbracelet/bubbletea) - TUI framework
- [Bubbles](https://github.com/charmbracelet/bubbles) - TUI components
- [Lipgloss](https://github.com/charmbracelet/lipgloss) - Styling

## 📦 Releases

Automated releases are built for multiple platforms:

| Platform | Architecture | Download |
|----------|-------------|----------|
| Linux | AMD64 | [sshm-linux-amd64.tar.gz](https://github.com/Gu1llaum-3/sshm/releases/latest/download/sshm-linux-amd64.tar.gz) |
| Linux | ARM64 | [sshm-linux-arm64.tar.gz](https://github.com/Gu1llaum-3/sshm/releases/latest/download/sshm-linux-arm64.tar.gz) |
| macOS | Intel | [sshm-darwin-amd64.tar.gz](https://github.com/Gu1llaum-3/sshm/releases/latest/download/sshm-darwin-amd64.tar.gz) |
| macOS | Apple Silicon | [sshm-darwin-arm64.tar.gz](https://github.com/Gu1llaum-3/sshm/releases/latest/download/sshm-darwin-arm64.tar.gz) |
| Windows | AMD64 | [sshm-windows-amd64.zip](https://github.com/Gu1llaum-3/sshm/releases/latest/download/sshm-windows-amd64.zip) |
| Windows | ARM64 | [sshm-windows-arm64.zip](https://github.com/Gu1llaum-3/sshm/releases/latest/download/sshm-windows-arm64.zip) |

## 🤝 Contributing

Contributions are welcome! Please feel free to submit a Pull Request. For major changes, please open an issue first to discuss what you would like to change.

### Development Workflow

1. **Fork** the repository
2. **Create** a feature branch (`git checkout -b feature/amazing-feature`)
3. **Commit** your changes (`git commit -m 'Add amazing feature'`)
4. **Push** to the branch (`git push origin feature/amazing-feature`)
5. **Open** a Pull Request

## 📝 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🙏 Acknowledgments

- [Charm](https://charm.sh/) for the amazing TUI libraries
- [Cobra](https://cobra.dev/) for the excellent CLI framework
- [@yimeng](https://github.com/yimeng) for contributing SSH Include directive support
- The Go community for building such fantastic tools

---

<div align="center">

**Made with ❤️ by [Guillaume](https://github.com/Gu1llaum-3)**

⭐ **Star this repo if you found it useful!** ⭐

</div>
