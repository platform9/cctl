package cmd

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strconv"
	"text/template"

	"github.com/ghodss/yaml"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	certutil "k8s.io/client-go/util/cert"

	"github.com/platform9/cctl/common"

	spv1 "github.com/platform9/ssh-provider/pkg/apis/sshprovider/v1alpha1"
	sputil "github.com/platform9/ssh-provider/pkg/controller"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	clusterv1 "sigs.k8s.io/cluster-api/pkg/apis/cluster/v1alpha1"
)

var forceDelete bool

// clusterCmd represents the cluster command
var clusterCmdCreate = &cobra.Command{
	Use:   "cluster",
	Short: "Creates clusterspec in the current directory",
	Run: func(cmd *cobra.Command, args []string) {
		routerID, err := strconv.Atoi(cmd.Flag("routerID").Value.String())
		if err != nil {
			log.Fatalf("Invalid routerId %v", err)
		}
		vip := cmd.Flag("vip").Value.String()
		servicesCIDR := cmd.Flag("serviceNetwork").Value.String()
		podsCIDR := cmd.Flag("podNetwork").Value.String()
		saPrivateKeyFile := cmd.Flag("saPrivateKey").Value.String()
		saPublicKeyFile := cmd.Flag("saPublicKey").Value.String()
		if (len(saPrivateKeyFile) == 0) != (len(saPublicKeyFile) == 0) {
			log.Fatalf("Must specify both saPrivateKey and saPublicKey")
		}
		caKeyFile := cmd.Flag("cakey").Value.String()
		caCertFile := cmd.Flag("cacert").Value.String()
		if (len(caKeyFile) == 0) != (len(caCertFile) == 0) {
			log.Fatalf("Must specify both caKeyFile and caCertFile")
		}

		newCommonCASecret := createCASecret(common.DefaultCommonCASecretName, caCertFile, caKeyFile)
		newServiceAccountKeySecret := createServiceAccountKeySecret(saPrivateKeyFile, saPublicKeyFile)
		newBootstrapTokenSecret := createBootstrapTokenSecret(common.DefaultBootstrapTokenSecretName)
		newCluster := createCluster(common.DefaultClusterName, podsCIDR, servicesCIDR, vip, routerID)

		if _, err := state.KubeClient.CoreV1().Secrets(common.DefaultNamespace).Create(newCommonCASecret); err != nil {
			log.Fatalf("Unable to create common CA secret: %v", err)
		}
		if _, err := state.KubeClient.CoreV1().Secrets(common.DefaultNamespace).Create(newServiceAccountKeySecret); err != nil {
			log.Fatalf("Unable to create service account secret: %v", err)
		}
		if _, err := state.KubeClient.CoreV1().Secrets(common.DefaultNamespace).Create(newBootstrapTokenSecret); err != nil {
			log.Fatalf("Unable to create bootstrap token secret: %v", err)
		}
		if _, err := state.ClusterClient.ClusterV1alpha1().Clusters(common.DefaultNamespace).Create(newCluster); err != nil {
			log.Fatalf("Unable to create cluster %q: %v", common.DefaultClusterName, err)
		}

		if err := state.PullFromAPIs(); err != nil {
			log.Fatalf("Unable to sync on-disk state: %v", err)
		}

		log.Println("Cluster created successfully.")
	},
}

func createCluster(clusterName, podsCIDR, servicesCIDR, vip string, routerID int) *clusterv1.Cluster {
	newCluster := clusterv1.Cluster{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Cluster",
			APIVersion: "cluster.k8s.io/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:              clusterName,
			Namespace:         common.DefaultNamespace,
			CreationTimestamp: metav1.Now(),
		},
		Spec: clusterv1.ClusterSpec{
			ClusterNetwork: clusterv1.ClusterNetworkingConfig{
				Services: clusterv1.NetworkRanges{
					CIDRBlocks: []string{
						servicesCIDR,
					},
				},
				Pods: clusterv1.NetworkRanges{
					CIDRBlocks: []string{
						podsCIDR,
					},
				},
				ServiceDomain: "cluster.local",
			},
		},
		Status: clusterv1.ClusterStatus{
			APIEndpoints: []clusterv1.APIEndpoint{
				clusterv1.APIEndpoint{
					Host: vip,
					Port: common.DefaultApiserverPort,
				},
			},
		},
	}

	spClusterSpec := spv1.ClusterSpec{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "sshprovider.platform9.com/v1alpha1",
			Kind:       "ClusterSpec",
		},
		APIServerCASecret: &corev1.LocalObjectReference{
			Name: common.DefaultCommonCASecretName,
		},
		EtcdCASecret: &corev1.LocalObjectReference{
			Name: common.DefaultCommonCASecretName,
		},
		FrontProxyCASecret: &corev1.LocalObjectReference{
			Name: common.DefaultCommonCASecretName,
		},
		ServiceAccountKeySecret: &corev1.LocalObjectReference{
			Name: common.DefaultServiceAccountKeySecretName,
		},
		BootstrapTokenSecret: &corev1.LocalObjectReference{
			Name: common.DefaultBootstrapTokenSecretName,
		},
		VIPConfiguration: &spv1.VIPConfiguration{
			IP:       vip,
			RouterID: routerID,
		},
	}

	spClusterStatus := spv1.ClusterStatus{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "sshprovider.platform9.com/v1alpha1",
			Kind:       "ClusterStatus",
		},
		EtcdMembers: []spv1.EtcdMember{},
	}

	sputil.PutClusterSpec(spClusterSpec, &newCluster)
	sputil.PutClusterStatus(spClusterStatus, &newCluster)

	return &newCluster
}

func createServiceAccountKeySecret(saPrivateKeyFile, saPublicKeyFile string) *corev1.Secret {
	sakSecret := corev1.Secret{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Secret",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:              "serviceaccount-key",
			Namespace:         common.DefaultNamespace,
			CreationTimestamp: metav1.Now(),
		},
		Data: make(map[string][]byte),
	}

	var privateKeyBytes []byte
	var publicKeyBytes []byte
	if len(saPrivateKeyFile) != 0 && len(saPublicKeyFile) != 0 {
		var err error
		privateKeyBytes, err = ioutil.ReadFile(saPrivateKeyFile)
		if err != nil {
			log.Fatalf("Unable to read service account private key %q: %v", saPrivateKeyFile, err)
		}
		publicKeyBytes, err = ioutil.ReadFile(saPublicKeyFile)
		if err != nil {
			log.Fatalf("Unable to read service account public key %q: %v", saPublicKeyFile, err)
		}
	} else {
		key, err := certutil.NewPrivateKey()
		if err != nil {
			log.Fatalf("Unable to create a service account private key: %v", err)
		}
		privateKeyBytes = certutil.EncodePrivateKeyPEM(key)
		publicKeyBytes, err = certutil.EncodePublicKeyPEM(&key.PublicKey)
		if err != nil {
			log.Fatalf("Unable to encode service account public key to PEM format: %v", err)
		}
	}

	sakSecret.Data["privatekey"] = privateKeyBytes
	sakSecret.Data["publickey"] = publicKeyBytes

	return &sakSecret
}

func createCASecret(secretName, certFilename, keyFilename string) *corev1.Secret {
	caSecret := corev1.Secret{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Secret",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:              secretName,
			Namespace:         common.DefaultNamespace,
			CreationTimestamp: metav1.Now(),
		},
		Data: make(map[string][]byte),
	}

	var certBytes []byte
	var keyBytes []byte
	if len(certFilename) != 0 && len(keyFilename) != 0 {
		var err error
		certBytes, err = ioutil.ReadFile(certFilename)
		if err != nil {
			log.Fatalf("Unable to read CA cert %q: %v", certFilename, err)
		}
		keyBytes, err = ioutil.ReadFile(keyFilename)
		if err != nil {
			log.Fatalf("Unable to read CA key %q: %v", keyFilename, err)
		}
	} else {
		cert, key, err := common.NewCertificateAuthority()
		if err != nil {
			log.Fatalf("Unable to create CA: %v", err)
		}
		certBytes = certutil.EncodeCertPEM(cert)
		keyBytes = certutil.EncodePrivateKeyPEM(key)
	}
	caSecret.Data["tls.crt"] = certBytes
	caSecret.Data["tls.key"] = keyBytes
	return &caSecret
}

func createBootstrapTokenSecret(name string) *corev1.Secret {
	btSecret := corev1.Secret{
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
	return &btSecret
}

var clusterCmdDelete = &cobra.Command{
	Use:   "cluster",
	Short: "Deletes a node to the cluster",
	Run: func(cmd *cobra.Command, args []string) {
		log.Println("Running cluster delete")

		cluster, err := state.ClusterClient.ClusterV1alpha1().Clusters(common.DefaultNamespace).Get(common.DefaultClusterName, metav1.GetOptions{})
		if err != nil {
			log.Fatalf("Unable to get cluster: %v", err)
		}

		machineList, err := state.ClusterClient.ClusterV1alpha1().Machines(common.DefaultNamespace).List(metav1.ListOptions{})
		if err != nil {
			log.Fatalf("Unable to list machines: %v", err)
		}

		if len(machineList.Items) > 0 {
			var machineNames []string
			for _, machine := range machineList.Items {
				machineNames = append(machineNames, machine.Name)
			}
			if forceDelete {
				log.Printf("Machines [%s] part of cluster. Deleting them from the state.", machineNames)
				for _, machine := range machineList.Items {
					if err := state.ClusterClient.ClusterV1alpha1().Machines(common.DefaultNamespace).Delete(machine.Name, &metav1.DeleteOptions{}); err != nil {
						if !apierrors.IsNotFound(err) {
							log.Fatalf("Unable to delete machine %q: %v", machine.Name, err)
						}
					}
				}
			} else {
				log.Fatalf("Machines [%s] part of cluster. Delete them before deleting the cluster.", machineNames)
			}
		}

		clusterProviderSpec, err := sputil.GetClusterSpec(*cluster)
		if err != nil {
			log.Fatalf("Unable to decode cluster spec: %v", err)
		}
		if clusterProviderSpec.APIServerCASecret != nil {
			if err := state.KubeClient.CoreV1().Secrets(common.DefaultNamespace).Delete(clusterProviderSpec.APIServerCASecret.Name, &metav1.DeleteOptions{}); err != nil {
				if !apierrors.IsNotFound(err) {
					log.Fatalf("Unable to delete API server CA secret: %v", err)
				}
			}
		}
		if clusterProviderSpec.EtcdCASecret != nil {
			if err := state.KubeClient.CoreV1().Secrets(common.DefaultNamespace).Delete(clusterProviderSpec.EtcdCASecret.Name, &metav1.DeleteOptions{}); err != nil {
				if !apierrors.IsNotFound(err) {
					log.Fatalf("Unable to delete etcd CA secret: %v", err)
				}
			}
		}
		if clusterProviderSpec.FrontProxyCASecret != nil {
			if err := state.KubeClient.CoreV1().Secrets(common.DefaultNamespace).Delete(clusterProviderSpec.FrontProxyCASecret.Name, &metav1.DeleteOptions{}); err != nil {
				if !apierrors.IsNotFound(err) {
					log.Fatalf("Unable to delete front proxy CA secret: %v", err)
				}
			}
		}
		if clusterProviderSpec.ServiceAccountKeySecret != nil {
			if err := state.KubeClient.CoreV1().Secrets(common.DefaultNamespace).Delete(clusterProviderSpec.ServiceAccountKeySecret.Name, &metav1.DeleteOptions{}); err != nil {
				if !apierrors.IsNotFound(err) {
					log.Fatalf("Unable to delete service account key secret: %v", err)
				}
			}
		}
		if clusterProviderSpec.BootstrapTokenSecret != nil {
			if err := state.KubeClient.CoreV1().Secrets(common.DefaultNamespace).Delete(clusterProviderSpec.BootstrapTokenSecret.Name, &metav1.DeleteOptions{}); err != nil {
				if !apierrors.IsNotFound(err) {
					log.Fatalf("Unable to delete bootstrap token secret: %v", err)
				}
			}
		}

		if err := state.ClusterClient.ClusterV1alpha1().Clusters(common.DefaultNamespace).Delete(cluster.Name, &metav1.DeleteOptions{}); err != nil {
			if !apierrors.IsNotFound(err) {
				log.Fatalf("Unable to delete cluster: %v", err)
			}
		}

		if err := state.PullFromAPIs(); err != nil {
			log.Fatalf("Unable to sync on-disk state: %v", err)
		}
		log.Println("Cluster deleted successfully")
	},
}

var clusterCmdGet = &cobra.Command{
	Use:   "cluster",
	Short: "Get the cluster details",
	Run: func(cmd *cobra.Command, args []string) {
		cluster, err := state.ClusterClient.ClusterV1alpha1().Clusters(common.DefaultNamespace).Get(common.DefaultClusterName, metav1.GetOptions{})
		if err != nil {
			log.Fatalf("Unable to get cluster: %v", err)
		}
		switch outputFmt {
		case "yaml":
			bytes, err := yaml.Marshal(cluster)
			if err != nil {
				log.Fatalf("Unable to marshal cluster spec file to yaml: %s", err)
			}
			os.Stdout.Write(bytes)
		case "json":
			bytes, err := json.Marshal(cluster)
			if err != nil {
				log.Fatalf("Unable to marshal cluster spec file to json: %s", err)
			}
			os.Stdout.Write(bytes)
		case "":
			// Pretty print cluster details
			clusterProviderSpec, err := sputil.GetClusterSpec(*cluster)
			if err != nil {
				log.Fatalf("Could not decode cluster provider spec: %v", err)
			}
			data := struct {
				Cluster             *clusterv1.Cluster
				ClusterProviderSpec *spv1.ClusterSpec
			}{
				Cluster:             cluster,
				ClusterProviderSpec: clusterProviderSpec,
			}
			t := template.Must(template.New("ClusterV1PrintTemplate").Parse(common.ClusterV1PrintTemplate))
			if err := t.Execute(os.Stdout, &data); err != nil {
				log.Fatalf("Could not pretty print cluster details: %s", err)
			}
		default:
			log.Fatalf("Unsupported output format %q", outputFmt)
		}
	},
}

var clusterCmdUpgrade = &cobra.Command{
	Use:   "cluster",
	Short: "Upgrade the cluster",
	Run: func(cmd *cobra.Command, args []string) {
		// Stub code
		fmt.Println("Running Upgrade cluster")
	},
}

var clusterCmdRecover = &cobra.Command{
	Use:   "cluster",
	Short: "Recover the cluster",
	Run: func(cmd *cobra.Command, args []string) {
		// Stub code
		fmt.Println("Running Recover cluster")
	},
}

var clusterCmdBackup = &cobra.Command{
	Use:   "cluster",
	Short: "Backup the cluster",
	Run: func(cmd *cobra.Command, args []string) {
		// Stub code
		fmt.Println("Running Backup cluster")
	},
}

func init() {
	createCmd.AddCommand(clusterCmdCreate)
	clusterCmdCreate.Flags().String("serviceNetwork", "10.1.0.0/16", "Network CIDR for services e.g. 10.1.0.0/16")
	clusterCmdCreate.Flags().String("podNetwork", "10.2.0.0/16", "Network CIDR for pods e.g. 10.2.0.0.16")
	clusterCmdCreate.Flags().String("vip", "", "Virtual IP to be used for multi master setup")
	clusterCmdCreate.Flags().String("routerID", "", "Virtual router ID for keepalived for multi master setup. Must be in the range [0, 254]. Must be unique within a single L2 network domain.")
	clusterCmdCreate.Flags().String("cacert", "", "Location of file containing CA cert for components to trust")
	clusterCmdCreate.Flags().String("cakey", "", "Location of file containing CA key for signing certs")
	clusterCmdCreate.Flags().String("saPrivateKey", "", "Location of file containing private key used for sigining service account tokens")
	clusterCmdCreate.Flags().String("saPublicKey", "", "Location of file containing public key used for sigining service account tokens")
	clusterCmdCreate.MarkFlagRequired("vip")
	clusterCmdCreate.MarkFlagRequired("routerID")
	//clusterCmdCreate.Flags().String("version", "1.10.2", "Kubernetes version")

	deleteCmd.AddCommand(clusterCmdDelete)
	clusterCmdDelete.Flags().BoolVar(&forceDelete, "force", false, "Force delete a cluster")

	getCmd.AddCommand(clusterCmdGet)
	upgradeCmd.AddCommand(clusterCmdUpgrade)
	recoverCmd.AddCommand(clusterCmdRecover)
	backupCmd.AddCommand(clusterCmdBackup)
}
