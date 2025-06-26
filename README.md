# GitHub Profile Manager (GHPM)

A desktop application for managing multiple GitHub profiles with different Git configurations and SSH keys. Easily switch between personal, work, or client GitHub accounts with a single click.

![GitHub release (latest by date)](https://img.shields.io/github/v/release/yourusername/ghpm)
![Go Version](https://img.shields.io/badge/Go-1.22%2B-blue)
![License](https://img.shields.io/badge/license-MIT-green)

## Features

- **Multiple Profile Management**: Store and manage unlimited GitHub profiles
- **Git Configuration**: Automatically switches git username and email
- **SSH Key Management**: Securely stores and switches SSH keys per profile
- **Profile Detection**: Detect and save current system configuration
- **Import/Export**: Share profiles between machines
- **SSH Testing**: Built-in SSH connection testing to GitHub
- **Secure Storage**: Profiles stored locally in `~/.ghpm`
- **Cross-Platform**: Works on Linux, macOS (Intel & ARM)

## Installation

### Download Pre-built Binaries

Download the latest release from the [Releases page](https://github.com/yourusername/ghpm/releases).

#### Linux

**Debian/Ubuntu (deb package):**
```bash
wget https://github.com/yourusername/ghpm/releases/download/v1.0.0/ghpm_1.0.0_amd64.deb
sudo dpkg -i ghpm_1.0.0_amd64.deb
```

**Other Linux distributions (tar.xz):**
```bash
wget https://github.com/yourusername/ghpm/releases/download/v1.0.0/ghpm_1.0.0_linux_amd64.tar.xz
tar -xf ghpm_1.0.0_linux_amd64.tar.xz
sudo mv ghpm /usr/local/bin/
```

#### macOS

**Intel Macs:**
```bash
# Download the .dmg file and double-click to install
# Or use the command line:
wget https://github.com/yourusername/ghpm/releases/download/v1.0.0/ghpm_1.0.0_darwin_amd64.dmg
hdiutil attach ghpm_1.0.0_darwin_amd64.dmg
cp -r /Volumes/GHPM/GHPM.app /Applications/
hdiutil detach /Volumes/GHPM
```

**Apple Silicon (M1/M2/M3):**
```bash
# Download the ARM version
wget https://github.com/yourusername/ghpm/releases/download/v1.0.0/ghpm_1.0.0_darwin_arm64.dmg
hdiutil attach ghpm_1.0.0_darwin_arm64.dmg
cp -r /Volumes/GHPM/GHPM.app /Applications/
hdiutil detach /Volumes/GHPM
```

### Build from Source

#### Prerequisites

- Go 1.22 or higher
- Make
- Git

**Linux additional requirements:**
```bash
sudo apt-get install libgl1-mesa-dev xorg-dev  # Debian/Ubuntu
sudo yum install mesa-libGL-devel libXcursor-devel libXrandr-devel libXinerama-devel libXi-devel  # RHEL/Fedora
```

#### Building

```bash
# Clone the repository
git clone https://github.com/yourusername/ghpm.git
cd ghpm

# Install dependencies
go mod download

# Build for your platform
make build

# Or build for a specific platform
make build-linux
make build-darwin-amd64  # macOS Intel
make build-darwin-arm64  # macOS ARM

# Run directly
./ghpm
```

## Usage Guide

### First Launch

When you first launch GHPM, the main window will show:
- **Current Configuration**: Displays the active Git configuration and SSH status
- **Profile List**: Shows all saved profiles (initially empty)
- **Action Buttons**: Various operations you can perform

### Creating Your First Profile

#### Option 1: Detect Current Configuration

1. Click **"Detect Current"** button
2. GHPM will read your current Git configuration and SSH keys
3. Enter a name for the profile (e.g., "Personal", "Work")
4. Review the detected configuration
5. Click **"Create"** to save the profile

#### Option 2: Manual Creation

1. Click **"Add Profile"** button
2. Fill in the required fields:
   - **Profile Name**: A unique identifier (e.g., "Work Account")
   - **Git Username**: Your GitHub username
   - **Git Email**: Your GitHub email
3. (Optional) Add SSH keys:
   - Click **"Select Private Key"** to choose your private key file
   - Click **"Select Public Key"** to choose your public key file
4. Click **"Save"** to create the profile

### Switching Profiles

1. Select a profile from the list
2. Click **"Switch Profile"**
3. Confirm the switch in the dialog
4. GHPM will:
   - Update your global Git configuration
   - Replace SSH keys in `~/.ssh/` (if profile has SSH keys)
   - Move existing SSH files to `~/.ssh/dump/[timestamp]/`
   - Mark the profile as active

### Managing Profiles

#### Edit a Profile
1. Select the profile from the list
2. Click **"Edit"**
3. Modify the desired fields
4. Click **"Save"**

#### Delete a Profile
1. Select the profile from the list
2. Click **"Delete"**
3. Confirm deletion (active profiles cannot be deleted)

#### Import a Profile
1. Click **"Import"**
2. Select a `.json` profile file
3. The profile will be added to your list

#### Export a Profile
1. Select the profile from the list
2. Click **"Export"**
3. Choose the destination folder
4. The profile will be saved as `[profile-name].json`

### Testing SSH Connection

1. Ensure you have an active profile with SSH keys
2. Click **"Test SSH"**
3. GHPM will attempt to connect to GitHub
4. Results will show success or error details

### Refresh

Click **"Refresh"** to reload profiles and update the current status display.

## Configuration

### Profile Storage

Profiles are stored in `~/.ghpm/` as individual JSON files:
```
~/.ghpm/
├── personal.json
├── work.json
└── client.json
```

### Profile Structure

Each profile contains:
```json
{
  "name": "Work Account",
  "git_username": "john-work",
  "git_email": "john@company.com",
  "ssh_private_key": "-----BEGIN OPENSSH PRIVATE KEY-----...",
  "ssh_public_key": "ssh-rsa AAAAB3NzaC1yc2...",
  "is_active": true,
  "created_from": "manual"
}
```

### SSH Key Backup

When switching profiles with SSH keys, existing SSH files are moved to:
```
~/.ssh/dump/[timestamp]/
```

This ensures you never lose existing SSH configurations.

## Security Considerations

- **Local Storage**: All profiles are stored locally on your machine
- **File Permissions**: Profile files are created with 600 permissions (read/write for owner only)
- **SSH Keys**: Private keys are stored with 600 permissions when written to `~/.ssh/`
- **No Network Access**: GHPM never sends your data anywhere except when testing SSH connections to GitHub

## Troubleshooting

### SSH Test Fails

1. **Check Key Permissions**: Ensure `~/.ssh/id_rsa` has 600 permissions
2. **Verify Key Format**: Keys should be in OpenSSH format
3. **GitHub Key Registration**: Ensure your public key is added to your GitHub account
4. **Network Issues**: Check your internet connection and firewall settings

### Profile Not Switching

1. **Check Git Installation**: Ensure Git is installed and in your PATH
2. **Permissions**: Verify you have write access to `~/.gitconfig`
3. **Active Profile**: Ensure the profile you're switching to isn't already active

### Cannot Delete Profile

- Active profiles cannot be deleted
- Switch to another profile first, then delete

### Import Fails

- Ensure the JSON file is valid
- Check that the profile name doesn't already exist
- Verify file permissions

## Building Packages

### Linux Packages

```bash
# Build deb package
make package-linux

# Build tar.xz archive
make package-linux
```

### macOS Packages

```bash
# Build DMG for Intel
make package-darwin-amd64

# Build DMG for ARM
make package-darwin-arm64
```

## Development

### Project Structure

```
ghpm/
├── main.go          # Application entry point
├── ui.go            # User interface implementation
├── config.go        # Configuration management
├── profile.go       # Profile data structures
├── git.go           # Git and SSH operations
├── Makefile         # Build automation
├── .github/         # GitHub Actions workflows
└── README.md        # This file
```

### Contributing

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

### Running Tests

```bash
go test ./...
```

## License

This project is licensed under the MIT License - see the LICENSE file for details.

## Acknowledgments

- Built with [Fyne](https://fyne.io/) - Cross-platform GUI framework for Go
- Icon assets from [Fyne's theme package](https://developer.fyne.io/explore/icons)

## Support

- **Issues**: [GitHub Issues](https://github.com/yourusername/ghpm/issues)
- **Discussions**: [GitHub Discussions](https://github.com/yourusername/ghpm/discussions)

---

Made with ❤️ for developers who juggle multiple GitHub accounts