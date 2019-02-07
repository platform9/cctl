package secret

import (
	"fmt"
	"io/ioutil"

	"github.com/platform9/cctl/common"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	certutil "k8s.io/client-go/util/cert"
)

func CreateCASecretDefault(secretName string) (*corev1.Secret, error) {
	return CreateCASecret(secretName, "", "")
}

func CreateSAKeySecretDefault(secretName string) (*corev1.Secret, error) {
	return CreateSAKeySecret(secretName, "", "")
}

func CreateCASecret(secretName, certFilename, keyFilename string) (*corev1.Secret, error) {
	caSecret := createSecret(secretName)

	var certBytes []byte
	var keyBytes []byte
	if len(certFilename) != 0 && len(keyFilename) != 0 {
		var err error
		certBytes, err = ioutil.ReadFile(certFilename)
		if err != nil {
			return nil, fmt.Errorf("unable to read CA cert %q: %v", certFilename, err)
		}
		keyBytes, err = ioutil.ReadFile(keyFilename)
		if err != nil {
			return nil, fmt.Errorf("unable to read CA key %q: %v", keyFilename, err)
		}
	} else {
		var err error
		certBytes, keyBytes, err = generateCertPair()
		if err != nil {
			return nil, fmt.Errorf("unable to generate cert pair: %v", err)
		}

	}
	caSecret.Data["tls.crt"] = certBytes
	caSecret.Data["tls.key"] = keyBytes
	return caSecret, nil
}

func CreateSAKeySecret(secretName, saPrivateKeyFile, saPublicKeyFile string) (*corev1.Secret, error) {
	sakSecret := createSecret(secretName)

	var privateKeyBytes []byte
	var publicKeyBytes []byte
	if len(saPrivateKeyFile) != 0 && len(saPublicKeyFile) != 0 {
		var err error
		privateKeyBytes, err = ioutil.ReadFile(saPrivateKeyFile)
		if err != nil {
			return nil, fmt.Errorf("unable to read service account private key %q: %v", saPrivateKeyFile, err)
		}
		publicKeyBytes, err = ioutil.ReadFile(saPublicKeyFile)
		if err != nil {
			return nil, fmt.Errorf("unable to read service account public key %q: %v", saPublicKeyFile, err)
		}
	} else {
		var err error
		privateKeyBytes, publicKeyBytes, err = generateKeyPair()
		if err != nil {
			return nil, fmt.Errorf("unable to generate key pair: %v", err)
		}
	}

	sakSecret.Data["privatekey"] = privateKeyBytes
	sakSecret.Data["publickey"] = publicKeyBytes

	return sakSecret, nil
}

func CreateBootstrapTokenSecret(secretName string) (*corev1.Secret, error) {
	btSecret := createSecret(secretName)
	return btSecret, nil
}

func generateCertPair() ([]byte, []byte, error) {
	var certBytes, keyBytes []byte
	cert, key, err := common.NewCertificateAuthority()
	if err != nil {
		return nil, nil, fmt.Errorf("unable to create CA: %v", err)
	}
	certBytes = certutil.EncodeCertPEM(cert)
	keyBytes = certutil.EncodePrivateKeyPEM(key)
	return certBytes, keyBytes, nil
}

func generateKeyPair() ([]byte, []byte, error) {
	var privateKeyBytes, publicKeyBytes []byte
	key, err := certutil.NewPrivateKey()
	if err != nil {
		return nil, nil, fmt.Errorf("unable to create a service account private key: %v", err)
	}
	privateKeyBytes = certutil.EncodePrivateKeyPEM(key)
	publicKeyBytes, err = certutil.EncodePublicKeyPEM(&key.PublicKey)
	if err != nil {
		return nil, nil, fmt.Errorf("unable to encode service account public key to PEM format: %v", err)
	}
	return privateKeyBytes, publicKeyBytes, nil
}

func createSecret(name string) *corev1.Secret {
	secret := corev1.Secret{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Secret",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:              name,
			Namespace:         common.DefaultNamespace,
			CreationTimestamp: metav1.Now(),
		},
		Data: make(map[string][]byte),
	}
	return &secret
}
