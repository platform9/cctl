package cmd

import (
	"crypto/x509"
	"encoding/base64"
	"fmt"
	"github.com/google/easypki/pkg/certificate"
	"github.com/google/easypki/pkg/easypki"
	"github.com/google/easypki/pkg/store"
	"github.com/google/uuid"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
	"os"
	"path"
)

const (
	CERT_NAME = "k8sca"
)

// clusterCmd represents the cluster command
var clusterCmd = &cobra.Command{
	Use:   "cluster",
	Short: "Creates clusterspec in the current directory",
	Run: func(cmd *cobra.Command, args []string) {
		clusterSpec := new(ClusterSpec)
		clusterSpec.Name = cmd.Flag("name").Value.String()
		clusterSpec.ServiceNetwork = cmd.Flag("serviceNetwork").Value.String()
		clusterSpec.PodNetwork = cmd.Flag("podNetwork").Value.String()
		clusterSpec.Vip = cmd.Flag("vip").Value.String()
		clusterSpec.Cacert = cmd.Flag("cacert").Value.String()
		clusterSpec.Cakey = cmd.Flag("cakey").Value.String()
		clusterSpec.Version = cmd.Flag("version").Value.String()
		clusterSpec.Token = uuid.New().String()
		if len(clusterSpec.Cacert) == 0 {
			createRootCA()
			clusterSpec.Cacert = readCA()
			clusterSpec.Cakey = readKey()
			os.RemoveAll(CERT_NAME)
		}
		bytes, _ := yaml.Marshal(clusterSpec)
		ioutil.WriteFile("./cluster-spec.yaml", bytes, 0600)
		fmt.Println("Cluster spec written in current dir!")
	},
}

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

func readCA() string {
	certFile := path.Join(CERT_NAME, "certs", CERT_NAME+".crt")
	return base64FileContent(certFile)
}

func readKey() string {
	keyFile := path.Join(CERT_NAME, "keys", CERT_NAME+".key")
	return base64FileContent(keyFile)
}

func createRootCA() {
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

func init() {
	createCmd.AddCommand(clusterCmd)
	clusterCmd.Flags().String("name", "example-cluster", "Name of the cluster")
	clusterCmd.Flags().String("serviceNetwork", "10.1.0.0/16", "Network CIDR for services e.g. 10.1.0.0/16")
	clusterCmd.Flags().String("podNetwork", "10.2.0.0/16", "Network CIDR for pods e.g. 10.2.0.0.16")
	clusterCmd.Flags().String("vip", "192.168.10.5", "VIP ip to be used for multi master setup")
	clusterCmd.Flags().String("cacert", "", "Base64 encoded CA cert for compoenents to trust")
	clusterCmd.Flags().String("cakey", "", "Base64 encoded CA key for signing certs")
	clusterCmd.Flags().String("version", "1.10.2", "Kubernetes version")
}
