package machine

import (
	"bytes"
	"fmt"
	"net"
	"os"

	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
)

// Client is used to perform actions on a machine, e.g., run commands and write
// files
type Client interface {
	RunCommand(cmd string) ([]byte, []byte, error)
	WriteFile(path string, mode os.FileMode, b []byte) error
	ReadFile(path string) ([]byte, error)
}

type client struct {
	sshClient  *ssh.Client
	sftpClient *sftp.Client
}

// NewClient creates a new Client that can be used to perform action on a
// machine
func NewClient(host string, port int, username string, privateKey string, publicKeys []string, insecureIgnoreHostKey bool) (Client, error) {
	signer, err := ssh.ParsePrivateKey([]byte(privateKey))
	if err != nil {
		return nil, fmt.Errorf("error parsing private key: %s", err)
	}
	sshConfig := &ssh.ClientConfig{
		User: string(username),
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(signer),
		},
	}
	if insecureIgnoreHostKey {
		sshConfig.HostKeyCallback = ssh.InsecureIgnoreHostKey()
	} else {
		parsedKeys := make([]ssh.PublicKey, len(publicKeys))
		for i, key := range publicKeys {
			parsedKey, _, _, _, err := ssh.ParseAuthorizedKey([]byte(key))
			if err != nil {
				return nil, fmt.Errorf("unable to parse host public key: %v", err)
			}
			parsedKeys[i] = parsedKey
		}
		sshConfig.HostKeyCallback = FixedHostKeys(parsedKeys)
	}
	sshClient, err := ssh.Dial("tcp", fmt.Sprintf("%s:%d", host, port), sshConfig)
	if err != nil {
		return nil, fmt.Errorf("unable to dial %s:%d: %s", host, port, err)
	}
	sftpClient, err := sftp.NewClient(sshClient)
	return &client{
		sshClient:  sshClient,
		sftpClient: sftpClient,
	}, nil
}

// RunCommand runs a command on the machine and returns stdout and stderr
// separately
func (c *client) RunCommand(cmd string) ([]byte, []byte, error) {
	session, err := c.sshClient.NewSession()
	if err != nil {
		return nil, nil, fmt.Errorf("unable to create session: %s", err)
	}
	var stdOutBuf bytes.Buffer
	var stdErrBuf bytes.Buffer
	session.Stdout = &stdOutBuf
	session.Stderr = &stdErrBuf
	err = session.Run(cmd)
	if err != nil {
		switch err.(type) {
		case *ssh.ExitError:
			return nil, nil, fmt.Errorf("command failed: %s", err)
		case *ssh.ExitMissingError:
			return nil, nil, fmt.Errorf("command failed: %s", err)
		default:
			return nil, nil, fmt.Errorf("command failed: %s", err)
		}
	}
	return stdOutBuf.Bytes(), stdErrBuf.Bytes(), nil
}

// WriteFile writes a file to the machine
func (c *client) WriteFile(path string, mode os.FileMode, b []byte) error {
	f, err := c.sftpClient.Create(path)
	if err != nil {
		return fmt.Errorf("unable to create file: %s", err)
	}
	defer f.Close()
	_, err = f.Write(b)
	if err != nil {
		return fmt.Errorf("write failed: %s", err)
	}
	err = f.Chmod(mode)
	if err != nil {
		return fmt.Errorf("chmod failed: %s", err)
	}
	return nil
}

// ReadFile reads a file from the machine
func (c *client) ReadFile(path string) ([]byte, error) {
	f, err := c.sftpClient.Open(path)
	if err != nil {
		return nil, fmt.Errorf("unable to open file: %s", err)
	}
	defer f.Close()
	var w bytes.Buffer
	_, err = f.WriteTo(&w)
	if err != nil {
		return nil, fmt.Errorf("read failed: %s", err)
	}
	return w.Bytes(), nil
}

// FixedHostKeys is a version of ssh.FixedHostKey that checks a list of SSH
// public keys
func FixedHostKeys(keys []ssh.PublicKey) ssh.HostKeyCallback {
	callbacks := make([]ssh.HostKeyCallback, len(keys))
	for i, expectedKey := range keys {
		callbacks[i] = ssh.FixedHostKey(expectedKey)
	}

	return func(hostname string, remote net.Addr, actualKey ssh.PublicKey) error {
		for _, callback := range callbacks {
			err := callback(hostname, remote, actualKey)
			if err == nil {
				return nil
			}
		}
		return fmt.Errorf("host key does not match any expected keys")
	}
}
