package test

import (
	"testing"

	"github.com/platform9/cctl/common"
)

func TestBTSecret(t *testing.T) {

	btSecret, _ := common.CreateBootstrapTokenSecret("bootstrap")

	if btSecret.ObjectMeta.Name != "bootstrap" {
		t.Error("Expected bootstrap, got ", btSecret.ObjectMeta.Name)
	}
}

func TestCASecret(t *testing.T) {
	caSecret, err := common.CreateCASecretDefault("apiserver-ca")
	if err != nil {
		t.Error("Error creating ca secret: ", err)
	}

	if caSecret.ObjectMeta.Name != "apiserver-ca" {
		t.Error("Expected apiserver-ca, got ", caSecret.ObjectMeta.Name)
	}

	if _, ok := caSecret.Data["tls.crt"]; !ok {
		t.Error("tls.crt not found.")
	}
	if _, ok := caSecret.Data["tls.key"]; !ok {
		t.Error("tls.key not found")
	}
}

func TestSASecret(t *testing.T) {
	saSecret, err := common.CreateSAKeySecretDefault("serviceaccount")
	if err != nil {
		t.Error("Error creating ca secret: ", err)
	}

	if saSecret.ObjectMeta.Name != "serviceaccount" {
		t.Error("Expected serviceaccount, got ", saSecret.ObjectMeta.Name)
	}

	if _, ok := saSecret.Data["privatekey"]; !ok {
		t.Error("privatekey not found.")
	}
	if _, ok := saSecret.Data["publickey"]; !ok {
		t.Error("publickey not found")
	}
}
