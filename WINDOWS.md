# FTP Server on Windows 11

## Configuration File Location

The FTP server will automatically create and use the config file in:
```
C:\ProgramData\ftpserver\ftpserver.json
```

This location is used by default when the service is installed without specifying a config file.

## Finding Your Current Config File

If the service is already running, check the Windows Event Viewer or service logs to see where the config file was created:

1. Open **Event Viewer** (Win + X → Event Viewer)
2. Navigate to: **Windows Logs → Application**
3. Look for events from source **ftpserver**
4. Find the log entry that says "Using config file" or "Created config file"

Alternatively, you can check the service configuration:
```powershell
sc qc ftpserver
```

## Managing the Service on Windows

### Open PowerShell or Command Prompt as Administrator

All service commands require Administrator privileges.

### Install the Service

**Recommended: Specify the config file path explicitly**
```powershell
# Install with config in ProgramData (default location)
.\ftpserver.exe -service install -conf "C:\ProgramData\ftpserver\ftpserver.json"

# Or install with config in a custom location
.\ftpserver.exe -service install -conf "C:\ftpserver\config.json"
```

### Control the Service

```powershell
# Start the service
.\ftpserver.exe -service start
# OR use Windows Service Manager
net start ftpserver

# Stop the service
.\ftpserver.exe -service stop
# OR
net stop ftpserver

# Restart the service
.\ftpserver.exe -service restart

# Uninstall the service
.\ftpserver.exe -service stop
.\ftpserver.exe -service uninstall
```

### Alternative: Using Windows Services Manager

1. Press Win + R, type `services.msc`, press Enter
2. Find "FTP Server" in the list
3. Right-click to Start, Stop, or Restart
4. Right-click → Properties to configure startup type

## Editing the Configuration

1. **Find the config file** at `C:\ProgramData\ftpserver\ftpserver.json`
2. **Stop the service** before editing
3. **Edit the file** with your preferred text editor (run as Administrator)
4. **Start the service** again

Example config location in File Explorer:
```
C:\ProgramData\ftpserver\ftpserver.json
```

Note: The `ProgramData` folder is hidden by default. You can:
- Type the path directly in the address bar, or
- Enable "Show hidden files" in File Explorer View options

## Current Service Configuration

If you've already installed the service and need to change the config file path:

1. Uninstall the current service:
   ```powershell
   .\ftpserver.exe -service stop
   .\ftpserver.exe -service uninstall
   ```

2. Reinstall with the desired config path:
   ```powershell
   .\ftpserver.exe -service install -conf "C:\ProgramData\ftpserver\ftpserver.json"
   ```

3. Start the service:
   ```powershell
   .\ftpserver.exe -service start
   ```

## Troubleshooting

### Service won't start
- Check Event Viewer for errors
- Verify the config file exists and is valid JSON
- Ensure the FTP port (default 2121) is not in use
- Check Windows Firewall settings

### Can't find the config file
- Run the service once and check Event Viewer logs
- The log will show: "Using config file: [full path]"

### Permission issues
- Ensure you're running PowerShell as Administrator
- Check that the service has read/write permissions to the config directory
