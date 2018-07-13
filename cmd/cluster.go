package cmd

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"os"
	"strconv"
	"text/template"

	"github.com/ghodss/yaml"
	sshproviderv1 "github.com/platform9/ssh-provider/sshproviderconfig/v1alpha1"
	"github.com/spf13/cobra"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/platform9/pf9-clusteradm/common"
	"github.com/platform9/pf9-clusteradm/statefileutil"
	certutil "k8s.io/client-go/util/cert"
	clusterv1 "sigs.k8s.io/cluster-api/pkg/apis/cluster/v1alpha1"
)

var forceDelete bool
var outputFmt string

// clusterCmd represents the cluster command
var clusterCmdCreate = &cobra.Command{
	Use:   "cluster",
	Short: "Creates clusterspec in the current directory",
	Run: func(cmd *cobra.Command, args []string) {
		spv1Codec, err := sshproviderv1.NewCodec()
		if err != nil {
			log.Fatalf("Could not initialize codec for internal types: %v", err)
		}
		routerID, err := strconv.Atoi(cmd.Flag("routerID").Value.String())
		vip := cmd.Flag("vip").Value.String()

		if err != nil {
			log.Fatalf("Invalid routerId %v", err)
		}

		sshClusterProviderConfig := sshproviderv1.SSHClusterProviderConfig{
			TypeMeta: v1.TypeMeta{
				APIVersion: "sshproviderconfig/v1alpha1",
				Kind:       "SSHClusterProviderConfig",
			},
			VIPConfiguration: &sshproviderv1.VIPConfiguration{
				IP:       net.ParseIP(vip),
				RouterID: routerID,
			},
		}

		providerConfig, err := spv1Codec.EncodeToProviderConfig(&sshClusterProviderConfig)
		if err != nil {
			log.Fatal(err)
		}

		sshClusterProviderStatus := sshproviderv1.SSHClusterProviderStatus{
			EtcdMembers: []sshproviderv1.EtcdMember{},
		}
		providerStatus, err := spv1Codec.EncodeToProviderStatus(&sshClusterProviderStatus)
		if err != nil {
			log.Fatal(err)
		}

		cluster := clusterv1.Cluster{
			TypeMeta: v1.TypeMeta{
				Kind:       "Cluster",
				APIVersion: "cluster.k8s.io/v1alpha1",
			},
			ObjectMeta: v1.ObjectMeta{
				Name:              cmd.Flag("name").Value.String(),
				CreationTimestamp: v1.Now(),
			},
			Spec: clusterv1.ClusterSpec{
				ClusterNetwork: clusterv1.ClusterNetworkingConfig{
					Services: clusterv1.NetworkRanges{
						CIDRBlocks: []string{
							cmd.Flag("serviceNetwork").Value.String(),
						},
					},
					Pods: clusterv1.NetworkRanges{
						CIDRBlocks: []string{
							cmd.Flag("podNetwork").Value.String(),
						},
					},
					ServiceDomain: "cluster.local",
				},
				ProviderConfig: *providerConfig,
			},
			Status: clusterv1.ClusterStatus{
				ProviderStatus: *providerStatus,
			},
		}
		cs, err := statefileutil.ReadStateFile()
		if err != nil {
			log.Fatal(err)
		}
		cs.Cluster = cluster
		cs.VIPConfiguration = sshClusterProviderConfig.VIPConfiguration
		cs.K8sVersion = common.K8S_VERSION
		fillCASecrets(&cs, cmd)
		fillSASecrets(&cs, cmd)
		if err := statefileutil.WriteStateFile(&cs); err != nil {
			log.Fatalf("error reading state: %v", err)
		}
		log.Println("Cluster created successfully.")
	},
}

func fillSASecrets(cs *common.ClusterState, cmd *cobra.Command) {
	saPrivateKeyFile := cmd.Flag("saPrivateKey").Value.String()
	saPublicKeyFile := cmd.Flag("saPublicKey").Value.String()
	if len(saPrivateKeyFile) == 0 && len(saPublicKeyFile) == 0 {
		key, err := certutil.NewPrivateKey()
		if err != nil {
			log.Fatalf("Failed to create a new key pair for service account with err %v\n", err)
		}
		err = common.WriteKey("/tmp", "sa", key)
		if err != nil {
			log.Fatalf("Failed to write key file with err %v\n", err)
		}
		err = common.WritePublicKey("/tmp", "sa", &key.PublicKey)
		if err != nil {
			log.Fatalf("Failed to write public key file with err %v\n", err)
		}
		saPrivateKeyFile = "/tmp/sa.key"
		saPublicKeyFile = "/tmp/sa.pub"
	} else if len(saPrivateKeyFile) == 0 || len(saPublicKeyFile) == 0 {
		log.Fatalf("Both saPrivateKey and saPublicKey need to specified")
	}
	privatekey, err := ioutil.ReadFile(saPrivateKeyFile)
	if err != nil {
		log.Fatalf("Failed to read file %s with error %v", saPrivateKeyFile, err)
	}
	publickey, err := ioutil.ReadFile(saPublicKeyFile)
	if err != nil {
		log.Fatalf("Failed to read file %s with error %v", saPublicKeyFile, err)
	}
	caSecret := corev1.Secret{}
	caSecret.Data = map[string][]byte{}
	caSecret.Data["privatekey"] = privatekey
	caSecret.Data["publickey"] = publickey
	cs.ServiceAccountKey = &caSecret
}

func fillCASecrets(cs *common.ClusterState, cmd *cobra.Command) {
	caKeyFile := cmd.Flag("cakey").Value.String()
	caCertFile := cmd.Flag("cacert").Value.String()
	if len(caKeyFile) == 0 && len(caCertFile) == 0 {
		cert, key, err := common.NewCertificateAuthority()
		if err != nil {
			log.Fatalf("Failed to create CA with err %v\n", err)
		}
		err = common.WriteCertAndKey("/tmp", "rootca", cert, key)
		if err != nil {
			log.Fatalf("Failed to write CA to disk with err %v\n", err)
		}
		caKeyFile = "/tmp/rootca.key"
		caCertFile = "/tmp/rootca.crt"
	} else if len(caKeyFile) == 0 || len(caCertFile) == 0 { //if only one of them is empty
		log.Fatalf("Both cacert and cakey need to specified")
	}
	tlscrt, err := ioutil.ReadFile(caCertFile)
	if err != nil {
		log.Fatalf("Failed to read file crt file %s with error %v", caCertFile, err)
	}
	tlskey, err := ioutil.ReadFile(caKeyFile)
	if err != nil {
		log.Fatalf("Failed to read key file %s with error %v", caKeyFile, err)
	}
	caSecret := corev1.Secret{}
	caSecret.Data = map[string][]byte{}
	caSecret.Data["tls.crt"] = tlscrt
	caSecret.Data["tls.key"] = tlskey
	cs.APIServerCA = &caSecret
	cs.FrontProxyCA = &caSecret
	cs.EtcdCA = &caSecret
}

var clusterCmdDelete = &cobra.Command{
	Use:   "cluster",
	Short: "Deletes a node to the cluster",
	Run: func(cmd *cobra.Command, args []string) {
		log.Println("Running cluster delete")
		cs, err := statefileutil.ReadStateFile()
		if err != nil {
			log.Fatalf("Failed to read cluster state file. Cannot delete cluster: %s", err)
		}
		if forceDelete {
			log.Println("Note: Forceful delete of cluster! Deleting cluster metadata")
			// Unset all objects created by the create cluster call
			cs.Cluster = clusterv1.Cluster{}
			cs.APIServerCA = nil
			cs.EtcdCA = nil
			cs.FrontProxyCA = nil
			cs.K8sVersion = ""
			cs.ServiceAccountKey = nil
			cs.VIPConfiguration = nil
			if err := statefileutil.WriteStateFile(&cs); err != nil {
				log.Fatalf("Unable to write cluster state file: %s", err)
			}
			return
		}
		if len(cs.Machines) > 0 {
			// There is alteast one machine present in the cluster. Don't continue delete
			var machineNames []string
			for _, machine := range cs.Machines {
				machineNames = append(machineNames, machine.ObjectMeta.Name)
			}
			log.Fatalf("Machines [%s] part of cluster. Please delete them before calling cluster delete.", machineNames)
		}
		// Unset all objects created by the create cluster call
		cs.Cluster = clusterv1.Cluster{}
		cs.APIServerCA = nil
		cs.EtcdCA = nil
		cs.FrontProxyCA = nil
		cs.K8sVersion = ""
		cs.ServiceAccountKey = nil
		cs.VIPConfiguration = nil
		if err := statefileutil.WriteStateFile(&cs); err != nil {
			log.Fatalf("Unable to write cluster state file: %s", err)
		}
		log.Println("Cluster deleted successfully")
	},
}

var clusterCmdGet = &cobra.Command{
	Use:   "cluster",
	Short: "Get the cluster details",
	Run: func(cmd *cobra.Command, args []string) {
		cs, err := statefileutil.ReadStateFile()
		if err != nil {
			log.Fatalf("Unable to read cluster spec file: %s", err)
		}
		switch outputFmt {
		case "yaml":
			// Flag yaml specificed. Print cluster spec as yaml
			bytes, err := yaml.Marshal(cs.Cluster)
			if err != nil {
				log.Fatalf("Unable to marshal cluster spec file to yaml: %s", err)
			}
			os.Stdout.Write(bytes)
		case "json":
			// Flag json specified. Print cluster spec as json
			bytes, err := json.Marshal(cs.Cluster)
			if err != nil {
				log.Fatalf("Unable to marshal cluster spec file to json: %s", err)
			}
			os.Stdout.Write(bytes)
		case "":
			// Pretty print cluster details
			t := template.Must(template.New("ClusterV1PrintTemplate").Parse(common.ClusterV1PrintTemplate))
			if err := t.Execute(os.Stdout, cs); err != nil {
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
	clusterCmdCreate.Flags().String("name", "example-cluster", "Name of the cluster")
	clusterCmdCreate.Flags().String("serviceNetwork", "10.1.0.0/16", "Network CIDR for services e.g. 10.1.0.0/16")
	clusterCmdCreate.Flags().String("podNetwork", "10.2.0.0/16", "Network CIDR for pods e.g. 10.2.0.0.16")
	clusterCmdCreate.Flags().String("vip", "192.168.10.5", "VIP ip to be used for multi master setup")
	clusterCmdCreate.Flags().String("routerID", "42", "Router ID for keepalived for multi master setup")
	clusterCmdCreate.Flags().String("cacert", "", "Location of file containing CA cert for components to trust")
	clusterCmdCreate.Flags().String("cakey", "", "Location of file containing CA key for signing certs")
	clusterCmdCreate.Flags().String("saPrivateKey", "", "Location of file containing private key used for sigining service account tokens")
	clusterCmdCreate.Flags().String("saPublicKey", "", "Location of file containing public key used for sigining service account tokens")

	//clusterCmdCreate.Flags().String("version", "1.10.2", "Kubernetes version")

	deleteCmd.AddCommand(clusterCmdDelete)
	deleteCmd.Flags().BoolVar(&forceDelete, "force", false, "Force delete a cluster")

	getCmd.AddCommand(clusterCmdGet)
	clusterCmdGet.Flags().StringVar(&outputFmt, "o", "", "output format json|yaml")

	upgradeCmd.AddCommand(clusterCmdUpgrade)
	recoverCmd.AddCommand(clusterCmdRecover)
	backupCmd.AddCommand(clusterCmdBackup)
}
