package machine

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"strconv"
	"strings"

	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
)

// Client is used to perform actions on a machine, e.g., run commands and write
// files
type Client interface {
	RunCommand(cmd string) ([]byte, []byte, error)
	WriteFile(path string, mode os.FileMode, b []byte) error
	ReadFile(path string) ([]byte, error)
	MkdirAll(path string, mode os.FileMode) error
	MoveFile(srcFilePath, dstFilePath string) error
	CopyFile(srcFilePath, dstFilePath string) error
	Exists(filePath string) (bool, error)
}

type client struct {
	sshClient  *ssh.Client
	sftpClient *sftp.Client
}

const (
	runAsSudo = true
)

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
	stdOutPipe, err := session.StdoutPipe()
	if err != nil {
		return nil, nil, fmt.Errorf("unable to pipe stdout: %s", err)
	}
	stdErrPipe, err := session.StderrPipe()
	if err != nil {
		return nil, nil, fmt.Errorf("unable to pipe stderr: %s", err)
	}
	// Prepend sudo if runAsSudo set to true
	if runAsSudo {
		cmd = fmt.Sprintf("sudo %s", cmd)
	}
	err = session.Start(cmd)
	if err != nil {
		return nil, nil, fmt.Errorf("unable to run command: %s", err)
	}
	stdOut, err := ioutil.ReadAll(stdOutPipe)
	stdErr, err := ioutil.ReadAll(stdErrPipe)
	err = session.Wait()
	if err != nil {
		switch err.(type) {
		case *ssh.ExitError:
			return stdOut, stdErr, fmt.Errorf("command failed: %s", err)
		case *ssh.ExitMissingError:
			return stdOut, stdErr, fmt.Errorf("command failed (no exit status): %s", err)
		default:
			return stdOut, stdErr, fmt.Errorf("command failed: %s", err)
		}
	}
	return stdOut, stdErr, nil
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

// MkdirAll creates a directory named path, along with any necessary parents,
// and returns nil, or else returns an error.
func (c *client) MkdirAll(path string, mode os.FileMode) error {
	cmd := fmt.Sprintf("mkdir -p %s", path)
	_, _, err := c.RunCommand(cmd)
	if err != nil {
		return fmt.Errorf("unable to create directory %q: %s", path, err)
	}
	// Change directory permission. Note that mode needs to be
	// converted to bit (octet) representation for chmod consumption
	cmd = fmt.Sprintf("chmod %s %s", strconv.FormatUint(uint64(mode), 8), path)
	_, _, err = c.RunCommand(cmd)
	if err != nil {
		return fmt.Errorf("unable to set permissions to directory %q: %s", path, err)
	}
	return nil
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

// MoveFile moves file specifed by srcFilePath to dstFilePath,
// and returns nil, or else returns an error.
func (c *client) MoveFile(srcFilePath, dstFilePath string) error {
	cmd := fmt.Sprintf("mv -f %s %s", srcFilePath, dstFilePath)
	_, _, err := c.RunCommand(cmd)
	if err != nil {
		return fmt.Errorf("unable to move file from %q to %q: %s", srcFilePath, dstFilePath, err)
	}
	return nil
}

// CopyFile copies file specified by srcFilePath to dstFilePath
func (c *client) CopyFile(srcFilePath, dstFilePath string) error {
	cmd := fmt.Sprintf("cp -f %s %s", srcFilePath, dstFilePath)
	_, _, err := c.RunCommand(cmd)
	if err != nil {
		return fmt.Errorf("unable to copy file from %q to %q: %s", srcFilePath, dstFilePath, err)
	}
	return nil
}

// Exists checks if specified path exists
func (c *client) Exists(path string) (bool, error) {
	cmd := fmt.Sprintf("test -e %s && echo true || echo false", path)
	outputBytes, _, err := c.RunCommand(cmd)
	if err != nil {
		return false, fmt.Errorf("unable to check if path %q exists: %s", path, err)
	}
	outputString := strings.TrimSpace(string(outputBytes))
	if outputString == "true" {
		return true, nil
	}
	return false, nil
}
