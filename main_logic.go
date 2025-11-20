package main

import (
	"io"
	"io/ioutil"
	"os"
	"os/signal"
	"syscall"
	"time"

	ftpserver "github.com/fclairamb/ftpserverlib"
	"github.com/fclairamb/go-log"
	gkwrap "github.com/fclairamb/go-log/gokit"
	gokit "github.com/go-kit/log"

	"github.com/fclairamb/ftpserver/config"
	"github.com/fclairamb/ftpserver/server"
)

var (
	ftpServer *ftpserver.FtpServer
	driver    *server.Server
)

func getAbsolutePath(path string) (string, error) {
	// If already absolute, return as-is
	if len(path) > 0 && (path[0] == '/' || (len(path) > 2 && path[1] == ':')) {
		return path, nil
	}

	// Get current working directory
	cwd, err := os.Getwd()
	if err != nil {
		return "", err
	}

	// Join with current directory
	return cwd + string(os.PathSeparator) + path, nil
}

func getDefaultConfigPath() string {
	// On Windows, use ProgramData; on Unix, use current directory
	if os.PathSeparator == '\\' {
		// Windows: C:\ProgramData\ftpserver\ftpserver.json
		programData := os.Getenv("ProgramData")
		if programData == "" {
			programData = "C:\\ProgramData"
		}
		configDir := programData + "\\ftpserver"
		// Try to create the directory
		os.MkdirAll(configDir, 0755)
		return configDir + "\\ftpserver.json"
	}
	// Unix/Linux: ./ftpserver.json
	return "ftpserver.json"
}

// runServer contains the core FTP server logic
func runServer(confFile string, onlyConf bool, logger log.Logger) error {
	logger.Info("FTP server", "version", BuildVersion, "date", BuildDate, "commit", Commit)

	autoCreate := onlyConf

	// The general idea here is that if you start it without any arg, you're probably doing a local quick&dirty run
	// possibly on a windows machine, so we're better of just using a default file name and create the file.
	if confFile == "" {
		// Use a platform-appropriate default location
		confFile = getDefaultConfigPath()
		autoCreate = true
	}

	// Get absolute path for logging
	absConfFile, err := getAbsolutePath(confFile)
	if err != nil {
		logger.Warn("Could not get absolute path", "confFile", confFile, "err", err)
		absConfFile = confFile
	}

	if autoCreate {
		if _, err := os.Stat(confFile); err != nil && os.IsNotExist(err) {
			logger.Warn("No conf file, creating one", "confFile", absConfFile)

			if err := ioutil.WriteFile(confFile, confFileContent(), 0600); err != nil { //nolint: gomnd
				logger.Error("Couldn't create conf file", "confFile", absConfFile, "err", err)
				return err
			}
			logger.Info("Created config file", "confFile", absConfFile)
		}
	}

	logger.Info("Using config file", "confFile", absConfFile)

	conf, errConfig := config.NewConfig(confFile, logger)
	if errConfig != nil {
		logger.Error("Can't load conf", "err", errConfig)
		return errConfig
	}

	// Now is a good time to open a logging file
	if conf.Content.Logging.File != "" {
		writer, err := os.OpenFile(conf.Content.Logging.File, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0600) //nolint:gomnd

		if err != nil {
			logger.Error("Can't open log file", "err", err)
			return err
		}

		logger = gkwrap.NewWrap(gokit.NewLogfmtLogger(io.MultiWriter(writer, os.Stdout))).With(
			"ts", gokit.DefaultTimestampUTC,
			"caller", gokit.DefaultCaller,
		)
	}

	// Loading the driver
	var errNewServer error
	driver, errNewServer = server.NewServer(conf, logger.With("component", "driver"))

	if errNewServer != nil {
		logger.Error("Could not load the driver", "err", errNewServer)
		return errNewServer
	}

	// Instantiating the server by passing our driver implementation
	ftpServer = ftpserver.NewFtpServer(driver)

	// Overriding the server default silent logger by a sub-logger (component: server)
	ftpServer.Logger = logger.With("component", "server")

	// Preparing the SIGTERM handling
	go signalHandler()

	// Blocking call, behaving similarly to the http.ListenAndServe
	if onlyConf {
		logger.Warn("Only creating conf")
		return nil
	}

	if err := ftpServer.ListenAndServe(); err != nil {
		logger.Error("Problem listening", "err", err)
		return err
	}

	// We wait at most 1 minute for all clients to disconnect
	if err := driver.WaitGracefully(time.Minute); err != nil {
		ftpServer.Logger.Warn("Problem stopping server", "err", err)
	}

	return nil
}

func stop() {
	if driver != nil {
		driver.Stop()
	}

	if ftpServer != nil {
		if err := ftpServer.Stop(); err != nil {
			ftpServer.Logger.Error("Problem stopping server", "err", err)
		}
	}
}

func signalHandler() {
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGTERM)
	signal.Notify(ch, syscall.SIGHUP)

	for {
		sig := <-ch
		if sig == syscall.SIGHUP {
			if driver != nil {
				err := driver.ReloadConfig()
				if err != nil {
					ftpServer.Logger.Warn("Error reloading config ", err)
				} else {
					ftpServer.Logger.Info("Successfully reloaded config")
				}
			}
		}
		if sig == syscall.SIGTERM {
			stop()
			break
		}
	}
}

func confFileContent() []byte {
	str := `{
  "version": 1,
  "accesses": [
    {
      "user": "test",
      "pass": "test",
      "fs": "os",
      "params": {
        "basePath": "/tmp"
      }
    }
  ],
  "passive_transfer_port_range": {
    "start": 2122,
    "end": 2130
  }
}`

	return []byte(str)
}
