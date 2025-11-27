# Installation Guide

This guide covers different ways to install AINAR on various platforms.

## Binary Installation

### Windows

#### Using ZIP Package

1. Download `ainar-windows-amd64.zip` from [Releases](https://github.com/HeavySnowJakarta/ainar/releases)
2. Extract the ZIP file
3. Run `install.bat` as Administrator to install as a Windows Service

#### Manual Installation

1. Extract `ainar.exe` to a directory (e.g., `C:\Program Files\AINAR\`)
2. Add the directory to your PATH
3. To install as a service:
   ```cmd
   sc create AINAR binPath= "C:\Program Files\AINAR\ainar.exe serve" start= auto
   sc start AINAR
   ```

### macOS

#### Using DMG

1. Download `ainar-macos-arm64.dmg` (Apple Silicon) or `ainar-macos-amd64.dmg` (Intel) from [Releases](https://github.com/HeavySnowJakarta/ainar/releases)
2. Open the DMG file
3. Drag `AINAR.app` to your Applications folder
4. To enable auto-start, copy the included plist file:
   ```bash
   cp /Volumes/AINAR/com.ainar.app.plist ~/Library/LaunchAgents/
   launchctl load ~/Library/LaunchAgents/com.ainar.app.plist
   ```

#### Using Homebrew (Coming Soon)

```bash
brew install ainar
```

### Linux

#### Debian/Ubuntu (DEB)

```bash
# Download the package
wget https://github.com/HeavySnowJakarta/ainar/releases/download/vX.Y.Z/ainar_X.Y.Z_amd64.deb

# Install
sudo dpkg -i ainar_X.Y.Z_amd64.deb

# The service is automatically enabled and started
sudo systemctl status ainar
```

#### RHEL/CentOS/Fedora (RPM)

```bash
# Download the package
wget https://github.com/HeavySnowJakarta/ainar/releases/download/vX.Y.Z/ainar-X.Y.Z-1.x86_64.rpm

# Install
sudo rpm -i ainar-X.Y.Z-1.x86_64.rpm

# Start the service
sudo systemctl enable --now ainar
```

#### Arch Linux (Coming Soon)

```bash
yay -S ainar
```

#### Manual Installation

1. Download the binary for your architecture
2. Make it executable: `chmod +x ainar-linux-amd64`
3. Move to a location in your PATH: `sudo mv ainar-linux-amd64 /usr/local/bin/ainar`
4. Create a systemd service (see below)

## Building from Source

### Prerequisites

- Go 1.22 or later
- Node.js 20 or later
- npm

### Build Steps

```bash
# Clone the repository
git clone https://github.com/HeavySnowJakarta/ainar.git
cd ainar

# Build (using the provided script)
./scripts/build.sh

# Or manually:
cd web && npm install && npm run build && cd ..
cp -r web/dist/* internal/webui/dist/
go build -o ainar ./cmd/ainar
```

### Cross-compilation

```bash
# Build for multiple platforms
./scripts/build.sh linux/amd64 linux/arm64 darwin/amd64 darwin/arm64 windows/amd64
```

## System Service

### systemd (Linux)

Create `/etc/systemd/system/ainar.service`:

```ini
[Unit]
Description=AINAR - AI Model Router
After=network.target

[Service]
Type=simple
User=ainar
Group=ainar
ExecStart=/usr/bin/ainar serve
Restart=always
RestartSec=5

[Install]
WantedBy=multi-user.target
```

Then:

```bash
sudo systemctl daemon-reload
sudo systemctl enable ainar
sudo systemctl start ainar
```

### launchd (macOS)

Create `~/Library/LaunchAgents/com.ainar.app.plist`:

```xml
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
  <key>Label</key>
  <string>com.ainar.app</string>
  <key>ProgramArguments</key>
  <array>
    <string>/usr/local/bin/ainar</string>
    <string>serve</string>
  </array>
  <key>RunAtLoad</key>
  <true/>
  <key>KeepAlive</key>
  <true/>
</dict>
</plist>
```

Then:

```bash
launchctl load ~/Library/LaunchAgents/com.ainar.app.plist
```

### Windows Service

```cmd
sc create AINAR binPath= "C:\path\to\ainar.exe serve" start= auto DisplayName= "AINAR - AI Model Router"
sc start AINAR
```

## Verification

After installation, verify AINAR is working:

```bash
# Check version
ainar version

# Test the API (after starting the server)
curl http://localhost:8080/v1/models \
  -H "Authorization: Bearer YOUR_KEY"
```
