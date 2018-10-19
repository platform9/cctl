package machine

import (
	"fmt"
	"path/filepath"

	log "github.com/platform9/ssh-provider/pkg/logrus"

	"github.com/platform9/ssh-provider/pkg/controller"
	"github.com/platform9/ssh-provider/pkg/machine"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	clusterv1 "sigs.k8s.io/cluster-api/pkg/apis/cluster/v1alpha1"
)

type ClusterSecretConstants struct {
	SecretRefName string
	CertKey       string
	KeyKey        string
	CertPath      string
	KeyPath       string
}

var (
	EtcdCASecretConstants = ClusterSecretConstants{
		SecretRefName: "EtcdCASecret",
		CertKey:       "tls.crt",
		KeyKey:        "tls.key",
		CertPath:      "/etc/etcd/pki/ca.crt",
		KeyPath:       "/etc/etcd/pki/ca.key",
	}
	APIServerCASecretConstants = ClusterSecretConstants{
		SecretRefName: "APIServerCASecret",
		CertKey:       "tls.crt",
		KeyKey:        "tls.key",
		CertPath:      "/etc/kubernetes/pki/ca.crt",
		KeyPath:       "/etc/kubernetes/pki/ca.key",
	}
	FrontProxyCASecretConstants = ClusterSecretConstants{
		SecretRefName: "FrontProxyCASecret",
		CertKey:       "tls.crt",
		KeyKey:        "tls.key",
		CertPath:      "/etc/kubernetes/pki/front-proxy-ca.crt",
		KeyPath:       "/etc/kubernetes/pki/front-proxy-ca.key",
	}
	ServiceAccountKeySecretConstants = ClusterSecretConstants{
		SecretRefName: "ServiceAccountKeySecret",
		CertKey:       "publickey",
		KeyKey:        "privatekey",
		CertPath:      "/etc/kubernetes/pki/sa.pub",
		KeyPath:       "/etc/kubernetes/pki/sa.key",
	}
)

func (a *Actuator) writeMasterSecretsToMachine(cluster *clusterv1.Cluster, machineClient machine.Client) error {
	var err error

	clusterSpec, err := controller.GetClusterSpec(*cluster)
	if err != nil {
		return fmt.Errorf("unable to decode cluster spec: %v", err)
	}

	for clusterSecretConstants, secretRef := range map[ClusterSecretConstants]*corev1.LocalObjectReference{
		EtcdCASecretConstants:            clusterSpec.EtcdCASecret,
		APIServerCASecretConstants:       clusterSpec.APIServerCASecret,
		FrontProxyCASecretConstants:      clusterSpec.FrontProxyCASecret,
		ServiceAccountKeySecretConstants: clusterSpec.ServiceAccountKeySecret,
	} {
		if secretRef == nil {
			return fmt.Errorf("cluster spec secret ref %q is undefined", clusterSecretConstants.SecretRefName)
		}
		secret, err := a.kubeClient.CoreV1().Secrets(cluster.Namespace).Get(secretRef.Name, metav1.GetOptions{})
		if errors.IsNotFound(err) {
			// Secret does not exist; assume it will be created and populated
			// once this machine is created
		} else if err != nil {
			return fmt.Errorf("error getting secret %q: %v", secretRef.Name, err)
		} else {
			if err := writeSecretToMachine(machineClient, secret, clusterSecretConstants.CertKey, clusterSecretConstants.KeyKey, clusterSecretConstants.CertPath, clusterSecretConstants.KeyPath); err != nil {
				return fmt.Errorf("unable to write secret %q to machine: %v", secret.Name, err)
			}
			log.Printf("[secrets] wrote secret %q and key %q", clusterSecretConstants.CertPath, clusterSecretConstants.KeyPath)
		}
	}
	return nil
}

func writeSecretToMachine(machineClient machine.Client, secret *corev1.Secret, certKey, keyKey, certPath, keyPath string) error {
	cert, ok := secret.Data[certKey]
	if !ok {
		return fmt.Errorf("did not find key %q in secret %q", certKey, secret.Name)
	}
	key, ok := secret.Data[keyKey]
	if !ok {
		return fmt.Errorf("did not find key %q in secret %q", keyKey, secret.Name)
	}
	// TODO(dlipovetsky) Use same dir for cert and key
	certDir := filepath.Dir(certPath)
	if err := machineClient.MkdirAll(certDir, 0755); err != nil {
		return fmt.Errorf("unable to create cert dir %q on machine: %v", certDir, err)
	}
	keyDir := filepath.Dir(keyPath)
	if err := machineClient.MkdirAll(keyDir, 0755); err != nil {
		return fmt.Errorf("unable to create key dir %q on machine: %v", keyDir, err)
	}

	// Non root users will not have permission to write to /etc/ directly
	// Write cert and key to /tmp instead and then move the certs over to their respective paths
	tmpCertPath := fmt.Sprintf("/tmp/%s", certKey)
	tmpKeyPath := fmt.Sprintf("/tmp/%s", keyKey)
	if err := machineClient.WriteFile(tmpCertPath, 0644, cert); err != nil {
		return fmt.Errorf("unable to write cert to %q on machine: %v", tmpCertPath, err)
	}
	if err := machineClient.WriteFile(tmpKeyPath, 0600, key); err != nil {
		return fmt.Errorf("unable to write key to %q on machine: %v", tmpKeyPath, err)
	}
	// Copy cert and key from /tmp to its respective destination
	if err := machineClient.MoveFile(tmpCertPath, certPath); err != nil {
		return err
	}
	return machineClient.MoveFile(tmpKeyPath, keyPath)
}
