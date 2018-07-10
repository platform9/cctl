package cmd

import (
	"fmt"
	"strconv"

	"io/ioutil"
	"log"
	"strings"

	"github.com/platform9/pf9-clusteradm/common"
	"github.com/platform9/pf9-clusteradm/statefileutil"
	sshMachineActuator "github.com/platform9/ssh-provider/machine"
	pm "github.com/platform9/ssh-provider/provisionedmachine"
	sshproviderv1 "github.com/platform9/ssh-provider/sshproviderconfig/v1alpha1"
	"github.com/spf13/cobra"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	clustercommon "sigs.k8s.io/cluster-api/pkg/apis/cluster/common"
	clusterv1 "sigs.k8s.io/cluster-api/pkg/apis/cluster/v1alpha1"
)

func nodeAlreadyExists(ip string, cs common.ClusterState) bool {
	for _, machine := range cs.Machines {
		if machine.Name == ip {
			return true
		}
	}
	return false
}

// nodeCmd represents the node create command
var nodeCmdCreate = &cobra.Command{
	Use:   "node",
	Short: "Adds a node to the cluster",
	Run: func(cmd *cobra.Command, args []string) {
		sshMachineProviderConfig, err := common.CreateSSHMachineProviderConfig(cmd)
		if err != nil {
			log.Fatal(err)
		}

		timestamp := v1.Now()
		ip := cmd.Flag("ip").Value.String()

		cs, err := statefileutil.ReadStateFile()
		if err != nil {
			log.Fatal(err)
		}
		if nodeAlreadyExists(ip, cs) {
			log.Fatalf("Failed to add node, node already exists")
		}

		port, err := strconv.Atoi(cmd.Flag("port").Value.String())
		if err != nil {
			log.Fatalf("Invalid port %v", err)
		}
		publicKeyFiles := cmd.Flag("publicKeys").Value
		publicKeys := []string{}
		if publicKeyFiles != nil && len(publicKeyFiles.String()) > 0 {
			files := strings.Split(publicKeyFiles.String(), ",")
			publicKeys := []string{}
			for _, file := range files {
				bytes, err := ioutil.ReadFile(file)
				if err != nil {
					log.Fatalf("Failed to read file %s with error %v", file, err)
				}
				publicKeys = append(publicKeys, string(bytes))
			}
		}

		machine := clusterv1.Machine{
			TypeMeta: v1.TypeMeta{
				Kind:       "Machine",
				APIVersion: "cluster.k8s.io/v1alpha1",
			},
			ObjectMeta: v1.ObjectMeta{
				Name:              ip,
				CreationTimestamp: timestamp,
			},
			Spec: clusterv1.MachineSpec{
				ObjectMeta: v1.ObjectMeta{
					CreationTimestamp: timestamp,
				},
				Roles:          []clustercommon.MachineRole{clustercommon.MasterRole},
				ProviderConfig: *sshMachineProviderConfig,
			},
		}
		role := cmd.Flag("role").Value.String()
		if role == "master" {
			machine.Spec.Roles = []clustercommon.MachineRole{clustercommon.MasterRole}
		} else if role == "worker" {
			machine.Spec.Roles = []clustercommon.MachineRole{clustercommon.NodeRole}
		}

		provisionedMachine := pm.ProvisionedMachine{
			SSHConfig: &sshproviderv1.SSHConfig{
				Host:       ip,
				Port:       port,
				PublicKeys: publicKeys,
			},
		}

		cm := corev1.ConfigMap{}
		cm.Data = map[string]string{}
		provisionedMachine.ToConfigMap(&cm)
		cs.ProvisionedMachines = append(cs.ProvisionedMachines, provisionedMachine)

		machines := append(cs.Machines, machine)
		cs.Machines = machines
		var clusterToken *corev1.Secret
		if role == "worker" {
			clusterToken = getSecretFromAvailableMaster(cs)
			if clusterToken == nil {
				log.Fatalf("Failed to add worker no master node available")
			}
		}

		actuator, err := sshMachineActuator.NewActuator([]*corev1.ConfigMap{&cm},
			cs.SSHCredentials,
			cs.EtcdCA,
			cs.APIServerCA,
			cs.FrontProxyCA,
			cs.ServiceAccountKey,
			clusterToken,
		)

		if len(publicKeys) == 0 {
			actuator.InsecureIgnoreHostKey = true
		}
		if err != nil {
			log.Fatal(err)
		}

		cluster := &cs.Cluster
		err = actuator.Create(cluster, &machine)
		if err != nil {
			log.Fatal(err)
		}
		if role == "master" {
			clusterProviderStatus := common.DecodeSSHClusterProviderStatus(cluster.Status.ProviderStatus)
			if len(clusterProviderStatus.EtcdMembers) == 0 {
				clusterProviderStatus.EtcdMembers = []sshproviderv1.EtcdMember{}
			}
			machineProviderStatus := common.DecodeSSHMachineProviderStatus(machine.Status.ProviderStatus)
			clusterProviderStatus.EtcdMembers = append(clusterProviderStatus.EtcdMembers, *machineProviderStatus.EtcdMember)
			status, err := common.EncodeSSHClusterProviderStatus(clusterProviderStatus)
			if err != nil {
				log.Fatalf("Failed to encode clusterProvider status with error %v", err)
			}
			cs.Cluster.Status.ProviderStatus = *status
		}

		statefileutil.WriteStateFile(&cs)
	},
}

var nodeCmdDelete = &cobra.Command{
	Use:   "node",
	Short: "Deletes a node to the cluster",
	Run: func(cmd *cobra.Command, args []string) {
		// Stub code
		fmt.Println("Running node delete")
	},
}

var nodeCmdGet = &cobra.Command{
	Use:   "node",
	Short: "Get a node",
	Run: func(cmd *cobra.Command, args []string) {
		// Stub code
		fmt.Println("Running get node")
		// TODO: Implement node/nodes
	},
}

func getSecretFromAvailableMaster(cs common.ClusterState) *corev1.Secret {
	for _, m := range cs.Machines {
		if common.IsMaster(m) {
			log.Printf("Tyring to get cluster secret from Master %s", m.Name)
			pm := statefileutil.GetProvisionedMachine(cs, m.Name)
			if pm == nil {
				log.Printf("Failed to get machine with ip %s", m.Name)
				continue
			}
			client, err := common.SSHClient(pm, cs.SSHCredentials, true)
			session, err := client.NewSession()
			if err != nil {
				log.Printf("Failed to create ssh session with error %v", err)
				continue
			}
			output, err := session.CombinedOutput("/opt/bin/kubeadm token create --print-join-command")
			if err != nil {
				log.Printf("Could not get token from master %s", m.Name)
				session.Close()
				continue
			}
			values := strings.Split(string(output), " ")
			//Successful output would be of the type
			//kubeadm join <server:port> --token <token> --discovery-token-ca-cert-hash <sha>
			if len(values) != 7 { //TODO Needs a better way but seems good-enough for now
				log.Printf("Could not get token from master %s", m.Name)
				continue
			}
			secret := corev1.Secret{}
			server := []byte(values[2])
			token := []byte(values[4])
			cahash := []byte(values[6])
			secret.Data = map[string][]byte{}
			secret.Data["server"] = server
			secret.Data["token"] = token
			secret.Data["cahash"] = cahash
			return &secret
		}
	}
	return nil
}

func init() {
	createCmd.AddCommand(nodeCmdCreate)
	nodeCmdCreate.Flags().String("ip", "10.0.0.1", "IP of the machine")
	nodeCmdCreate.Flags().Int("port", 22, "SSH port")
	nodeCmdCreate.Flags().String("role", "worker", "Role of the node. Can be master/worker")
	nodeCmdCreate.Flags().String("publicKeys", "", "Comma separated list of public host keys for the machine")
	nodeCmdCreate.Flags().String("sshSecretName", "sshSecret", "Name of the sshSecret to use")

	deleteCmd.AddCommand(nodeCmdDelete)
	nodeCmdDelete.Flags().String("name", "", "Node name")
	nodeCmdDelete.Flags().String("force", "", "Force delete the node")

	getCmd.AddCommand(nodeCmdGet)
	nodeCmdGet.Flags().String("name", "", "Node name")
}
