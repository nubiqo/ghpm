# GitHub Profile Manager

A desktop application to manage multiple GitHub profiles with ease. Switch between different git configurations and SSH keys for different GitHub accounts.

## Features

- **Multiple Profile Management**: Store and switch between different GitHub profiles
- **Git Configuration**: Automatically set git username and email per profile
- **SSH Key Management**: Store and manage SSH keys per profile with automatic key switching
- **Profile Detection**: Detect current system git/SSH configuration and save as profile
- **Import/Export**: Share profiles between machines using JSON export/import
- **SSH Testing**: Test SSH connection to GitHub with current configuration
- **Secure Storage**: SSH keys are stored securely in user's home directory

## Installation

### Ubuntu/Debian

Download the latest `.deb` package from the [releases page](https://github.com/nubiqo/ghpm/releases) and install:

```bash
sudo dpkg -i github-profile-manager_*.deb
```

The application will be available in your applications menu or run from terminal:

```bash
github-profile-manager
```

### macOS

Download the appropriate package from the [releases page](https://github.com/nubiqo/ghpm/releases):

#### DMG Installation (Recommended)
- **Intel Macs**: Download `github-profile-manager-*-darwin-amd64.dmg`
- **Apple Silicon Macs**: Download `github-profile-manager-*-darwin-arm64.dmg`

1. Double-click the DMG file to mount it
2. Drag the application to your Applications folder
3. Launch from Applications or Spotlight

#### ZIP Installation
- **Intel Macs**: Download `github-profile-manager-*-darwin-amd64.zip`
- **Apple Silicon Macs**: Download `github-profile-manager-*-darwin-arm64.zip`

1. Extract the ZIP file
2. Move the `.app` bundle to Applications folder
3. Right-click and select "Open" the first time (to bypass Gatekeeper)

### Linux (Portable)

Download the tarball from the [releases page](https://github.com/nubiqo/ghpm/releases):

```bash
curl -L -o github-profile-manager.tar.xz https://github.com/nubiqo/ghpm/releases/latest/download/GitHubProfileManager-*-linux-amd64.tar.xz
tar -xf github-profile-manager.tar.xz
./github-profile-manager
```

### From Source

#### Prerequisites

- Go 1.22 or later
- Docker (for cross-compilation to macOS)
- System dependencies for Fyne:
  ```bash
  sudo apt-get install libgl1-mesa-dev xorg-dev
  ```

#### Build and Install

```bash
git clone https://github.com/nubiqo/ghpm.git
cd ghpm
make deps
make build-linux
sudo cp build/linux/github-profile-manager /usr/local/bin/
```

## Quick Start

1. **Launch the application** from your applications menu or terminal

2. **Detect current configuration** (recommended for first use):
   - Click "Detect Current" to create a profile from your existing git/SSH setup
   - Give it a meaningful name (e.g., "work", "personal")

3. **Add more profiles**:
   - Click "Add Profile" to manually create new profiles
   - Fill in git username, email, and optionally SSH keys
   - SSH keys can be loaded from existing files

4. **Switch between profiles**:
   - Select a profile from the list
   - Click "Switch Profile"
   - Confirm the operation

## Usage

### Managing Profiles

- **Add Profile**: Create new profile manually with git config and SSH keys
- **Detect Current**: Create profile from current system git/SSH configuration
- **Edit**: Modify existing profile details
- **Delete**: Remove profile (cannot delete active profile)
- **Import/Export**: Share profiles between machines using JSON files

### Profile Operations

- **Switch Profile**: Changes global git config and SSH keys to selected profile
- **Test SSH**: Verify SSH connection to GitHub works with current keys
- **Refresh**: Reload profiles from disk

### SSH Key Management

When switching profiles with SSH keys:
- Existing SSH files are moved to `~/.ssh/dump/[timestamp]/` for backup
- Profile SSH keys are written to `~/.ssh/id_rsa` and `~/.ssh/id_rsa.pub`
- Proper file permissions are automatically set

## Configuration

Profiles are stored in `~/.ghpm/` as individual JSON files. Each profile contains:
- Git username and email
- SSH private and public keys (if provided)
- Profile metadata

## Security Notes

- SSH keys are stored with appropriate file permissions (600 for private, 644 for public)
- Existing SSH keys are backed up before being replaced
- Profile files are stored with restricted permissions (600)

## Troubleshooting

### macOS Installation Issues

**"App cannot be opened because it is from an unidentified developer"**
1. Right-click the application and select "Open"
2. Click "Open" in the security dialog
3. Or go to System Preferences → Security & Privacy → General and click "Open Anyway"

**Alternative method:**
```bash
# Remove quarantine attribute
xattr -d com.apple.quarantine /Applications/GitHubProfileManager.app
```

### SSH Connection Issues

1. Ensure your public key is added to your GitHub account
2. Use "Test SSH" to verify connection
3. Check that SSH keys have correct permissions
4. Verify the SSH key format is correct

### Git Configuration Issues

- Verify git is installed and accessible from PATH
- Check current git config with: `git config --global --list`

### Profile Management Issues

- Ensure `~/.ghpm` directory has proper permissions
- Check application logs for detailed error messages

## Building

To build the application yourself:

### Prerequisites

- Go 1.22 or later
- Docker (for cross-compilation to macOS)
- System dependencies for Fyne (Linux):
  ```bash
  sudo apt-get install libgl1-mesa-dev xorg-dev
  ```

### Build Commands

```bash
# Install dependencies (including fyne-cross)
make deps

# Build for current platform
make build-linux

# Build for macOS (requires Docker)
make build-darwin

# Build for specific macOS architecture
make build-darwin-amd64  # Intel Macs
make build-darwin-arm64  # Apple Silicon Macs

# Create distribution packages
make package-linux       # .deb and .tar.xz
make package-darwin      # .dmg and .zip for both architectures

# Full release build (all platforms)
make release
```

## License

MIT License - see [LICENSE](LICENSE) file for details.