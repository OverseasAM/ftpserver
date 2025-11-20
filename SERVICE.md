# FTP Server - Service Installation Guide

The FTP server now supports running as a system service using the `github.com/kardianos/service` package.

## Usage

### Running Directly (Development)
```bash
# Run with default configuration
./ftpserver

# Run with custom configuration
./ftpserver -conf /path/to/config.json

# Generate configuration file only
./ftpserver -conf-only
```

### Service Management

#### Install as System Service

**Important:** Always specify an absolute path for the configuration file when installing as a service.

```bash
# Install with custom configuration (RECOMMENDED)
sudo ./ftpserver -service install -conf /etc/ftpserver/ftpserver.json

# Or for user-specific installation
sudo ./ftpserver -service install -conf /home/youruser/ftpserver.json
```

If you install without specifying `-conf`, the service will create `ftpserver.json` in the service's working directory. Check the service logs to find the exact location:

```bash
# On Linux with systemd
sudo journalctl -u ftpserver -n 50

# On Linux with other init systems, check /var/log/
```

#### Control the Service
```bash
# Start the service
sudo ./ftpserver -service start

# Stop the service
sudo ./ftpserver -service stop

# Restart the service
sudo ./ftpserver -service restart

# Uninstall the service
sudo ./ftpserver -service uninstall
```

## Architecture

The codebase is organized as follows:

- **main.go** - Service management and command-line interface
- **service_impl.go** - Service interface implementation (Start/Stop methods)
- **main_logic.go** - Core FTP server logic (runServer function)

This separation keeps service management code separate from the core server functionality, making the code more maintainable and testable.

## Supported Platforms

The `github.com/kardianos/service` package supports:
- Linux (systemd, upstart, SysV)
- Windows (Service Manager)
- macOS (launchd)

The service will automatically use the appropriate init system for your platform.
