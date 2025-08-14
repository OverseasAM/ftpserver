// Package sftp provides an SFTP connection layer
package sftp

import (
	"fmt"
	"net"
	"os"
	"strings"

	"github.com/pkg/sftp"
	"github.com/spf13/afero"
	"github.com/spf13/afero/sftpfs"
	"golang.org/x/crypto/ssh"

	"github.com/fclairamb/ftpserver/config/confpar"
)

// ConnectionError is returned when a connection occurs while connecting to the SFTP server
type ConnectionError struct {
	error
	Source error
}

func (err ConnectionError) Error() string {
	return fmt.Sprintf("Could not connect to SFTP host: %#v", err.Source)
}

// LoadFs loads a file system from an access description
func LoadFs(access *confpar.Access) (afero.Fs, error) {
	par := access.Params
	config := &ssh.ClientConfig{
		User: par["username"],
	}

	// Load host key if provided
	hostKeyPath, hasHostKey := par["hostKey"]
	var hostKey ssh.PublicKey
	if hasHostKey && hostKeyPath != "" {
		hostKeyBytes, err := os.ReadFile(hostKeyPath)
		if err != nil {
			return nil, &ConnectionError{Source: fmt.Errorf("unable to read host key: %w", err)}
		}
		hostKey, _, _, _, err = ssh.ParseAuthorizedKey(hostKeyBytes)
		if err != nil {
			return nil, &ConnectionError{Source: fmt.Errorf("unable to parse host key: %w", err)}
		}
	}

	// Load authmethod if provided
	authMethod, hasAuthMethod := par["method"]
	if hasAuthMethod && authMethod != "" {
		authMethod = strings.ToLower(authMethod)
	} else {
		authMethod = "password"
	}

	switch authMethod {
	case "publickey":
		key, err := os.ReadFile(par["privateKey"])
		if err != nil {
			return nil, &ConnectionError{Source: fmt.Errorf("unable to read private key: %w", err)}
		}
		var signer ssh.Signer
		passphrase, hasPassphrase := par["privateKeyPassphrase"]
		if hasPassphrase && passphrase != "" {
			signer, err = ssh.ParsePrivateKeyWithPassphrase(key, []byte(passphrase))
			if err != nil {
				return nil, &ConnectionError{Source: fmt.Errorf("unable to parse private key with passphrase: %w", err)}
			}
		} else {
			signer, err = ssh.ParsePrivateKey(key)
			if err != nil {
				return nil, &ConnectionError{Source: fmt.Errorf("unable to parse private key: %w", err)}
			}
		}
		config.Auth = []ssh.AuthMethod{
			ssh.PublicKeys(signer),
		}
		if hostKey != nil {
			config.HostKeyCallback = ssh.FixedHostKey(hostKey)
		} else {
			config.HostKeyCallback = func(hostname string, remote net.Addr, key ssh.PublicKey) error {
				return nil
			}
		}
	case "password", "":
		config.Auth = []ssh.AuthMethod{
			ssh.Password(par["password"]),
		}
		if hostKey != nil {
			config.HostKeyCallback = ssh.FixedHostKey(hostKey)
		} else {
			config.HostKeyCallback = func(hostname string, remote net.Addr, key ssh.PublicKey) error {
				return nil
			}
		}
	default:
		fmt.Println("SFTP auth method not set")
	}

	// Dial your ssh server.
	conn, errSSH := ssh.Dial("tcp", par["hostname"], config)
	if errSSH != nil {
		return nil, &ConnectionError{Source: errSSH}
	}

	client, errSftp := sftp.NewClient(conn)
	if errSftp != nil {
		return nil, &ConnectionError{Source: errSftp}
	}

	// Get the SFTP user's home directory
	/*
		homeDir, err := client.Getwd()
		if err != nil {
			return nil, &ConnectionError{Source: fmt.Errorf("unable to get SFTP home directory: %w", err)}
		}

		//fmt.Println("SFTP user's home directory:", homeDir)
	*/

	basePath, hasBasePath := par["basePath"]
	if hasBasePath && basePath != "" {
		fs := afero.NewBasePathFs(sftpfs.New(client), basePath)
		return fs, nil
	}

	return sftpfs.New(client), nil

}
