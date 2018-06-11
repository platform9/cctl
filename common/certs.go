package common

import (
	"github.com/google/easypki/pkg/easypki"
	"io/ioutil"
	"log"
	"encoding/base64"
	"path"
	"github.com/google/easypki/pkg/store"
	"github.com/google/easypki/pkg/certificate"
	"crypto/x509"
)

const (
	CERT_NAME = "k8sca"
)

type router struct {
	PKI *easypki.EasyPKI
}

func base64FileContent(filename string) string {
	bytes, err := ioutil.ReadFile(filename)
	if err != nil {
		log.Fatal(err)
	}
	return base64.StdEncoding.EncodeToString(bytes)
}

func ReadCA() string {
	certFile := path.Join(CERT_NAME, "certs", CERT_NAME+".crt")
	return base64FileContent(certFile)
}

func ReadKey() string {
	keyFile := path.Join(CERT_NAME, "keys", CERT_NAME+".key")
	return base64FileContent(keyFile)
}

func CreateRootCA() {
	r := router{PKI: &easypki.EasyPKI{Store: &store.Local{}}}
	var bundle *certificate.Bundle
	req := &easypki.Request{
		Name: CERT_NAME,
		Template: &x509.Certificate{
			IsCA: true,
		},
	}
	if err := r.PKI.Sign(bundle, req); err != nil {
		log.Fatal(err)
	}
}