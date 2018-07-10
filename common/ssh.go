package common

import (
	"fmt"
	"log"
	"net"

	pm "github.com/platform9/ssh-provider/provisionedmachine"
	"golang.org/x/crypto/ssh"
	corev1 "k8s.io/api/core/v1"
)

func SSHClient(m *pm.ProvisionedMachine, sshCredentials *corev1.Secret, insecureIgnoreHostKey bool) (*ssh.Client, error) {
	sshUsername, ok := sshCredentials.Data["username"]
	if !ok {
		return nil, fmt.Errorf("error reading SSH username")
	}
	sshPrivateKey, ok := sshCredentials.Data["ssh-privatekey"]
	if !ok {
		return nil, fmt.Errorf("error reading SSH private key")
	}
	signer, err := ssh.ParsePrivateKey(sshPrivateKey)
	if err != nil {
		return nil, fmt.Errorf("error parsing SSH private key: %s", err)
	}
	sshConfig := &ssh.ClientConfig{
		User: string(sshUsername),
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(signer),
		},
	}
	if insecureIgnoreHostKey {
		sshConfig.HostKeyCallback = ssh.InsecureIgnoreHostKey()
	} else {
		parsedKeys := make([]ssh.PublicKey, len(m.SSHConfig.PublicKeys))
		for i, key := range m.SSHConfig.PublicKeys {
			parsedKey, _, _, _, err := ssh.ParseAuthorizedKey([]byte(key))
			if err != nil {
				log.Fatalf("unable to parse host public key: %v", err)
			}
			parsedKeys[i] = parsedKey
		}
		sshConfig.HostKeyCallback = FixedHostKeys(parsedKeys)
	}
	client, err := ssh.Dial("tcp", fmt.Sprintf("%s:%d", m.SSHConfig.Host, m.SSHConfig.Port), sshConfig)
	if err != nil {
		log.Fatalf("unable to dial %s:%d: %s", m.SSHConfig.Host, m.SSHConfig.Port, err)
	}
	return client, nil
}

// FixedHostKeys is a version of ssh.FixedHostKey that checks a list of SSH public keys
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
