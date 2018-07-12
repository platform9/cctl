package cmd

import (
	"fmt"
	"strconv"

	"io/ioutil"
	"log"
	"strings"

	"github.com/platform9/pf9-clusteradm/common"
	"github.com/platform9/pf9-clusteradm/machine"
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

func machineAlreadyExists(ip string, cs common.ClusterState) bool {
	for _, machine := range cs.Machines {
		if machine.Name == ip {
			return true
		}
	}
	return false
}

// machineCmdCreate represents the machine create command
var machineCmdCreate = &cobra.Command{
	Use:   "machine",
	Short: "Adds a machine to the cluster",
	Run: func(cmd *cobra.Command, args []string) {
		sshMachineProviderConfig, err := common.CreateSSHMachineProviderConfig(cmd)
		if err != nil {
			log.Fatal(err)
		}

		timestamp := v1.Now()
		ip := cmd.Flag("ip").Value.String()
		iface := cmd.Flag("iface").Value.String()

		cs, err := statefileutil.ReadStateFile()
		if err != nil {
			log.Fatal(err)
		}
		if machineAlreadyExists(ip, cs) {
			log.Fatalf("Failed to add machine, machine already exists")
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
		} else if role == "node" {
			machine.Spec.Roles = []clustercommon.MachineRole{clustercommon.NodeRole}
		}

		provisionedMachine := pm.ProvisionedMachine{
			SSHConfig: &sshproviderv1.SSHConfig{
				Host:       ip,
				Port:       port,
				PublicKeys: publicKeys,
			},
			VIPNetworkInterface: iface,
		}

		cm := corev1.ConfigMap{}
		cm.Data = map[string]string{}
		provisionedMachine.ToConfigMap(&cm)
		cs.ProvisionedMachines = append(cs.ProvisionedMachines, provisionedMachine)

		machines := append(cs.Machines, machine)
		cs.Machines = machines
		var clusterToken *corev1.Secret
		var kubeconfig []byte
		if role == "node" {
			clusterToken, kubeconfig = getSecretAndConfigFromMaster(cs)
			if clusterToken == nil {
				log.Fatalf("Failed to add node. No master available")
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
			endPoint := clusterv1.APIEndpoint{}
			endPoint.Host = provisionedMachine.SSHConfig.Host
			endPoint.Port = common.DEFAULT_APISERVER_PORT
			cluster.Status.APIEndpoints = append(cluster.Status.APIEndpoints, endPoint)
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
		if role == "node" {
			client := getClientForMachine(&provisionedMachine, cs.SSHCredentials)
			client.WriteFile("/etc/kubernetes/admin.conf", 0644, kubeconfig)
		}
		statefileutil.WriteStateFile(&cs)
	},
}

var machineCmdDelete = &cobra.Command{
	Use:   "machine",
	Short: "Deletes a machine from the cluster",
	Run: func(cmd *cobra.Command, args []string) {
		// Stub code
		fmt.Println("Running machine delete")
	},
}

var machineCmdGet = &cobra.Command{
	Use:   "machine",
	Short: "Get a machine",
	Run: func(cmd *cobra.Command, args []string) {
		// Stub code
		fmt.Println("Running get machine")
		// TODO: Implement machine/machines
	},
}

func getClientForMachine(pm *pm.ProvisionedMachine, credentials *corev1.Secret) machine.Client {
	client, err := machine.NewClient(pm.SSHConfig.Host,
		pm.SSHConfig.Port,
		string(credentials.Data["username"]),
		string(credentials.Data["ssh-privatekey"]),
		pm.SSHConfig.PublicKeys, true)
	if err != nil {
		log.Fatalf("Failed to get client for machine %s", pm.SSHConfig.Host)
	}
	return client
}

func getSecretAndConfigFromMaster(cs common.ClusterState) (*corev1.Secret, []byte) {
	pm := statefileutil.GetMaster(&cs)
	client := getClientForMachine(pm, cs.SSHCredentials)
	output, errOutput, err := client.RunCommand("/opt/bin/kubeadm token create --print-join-command")
	if err != nil {
		log.Fatalf("Could not get token from master %s", pm.SSHConfig.Host)
		log.Fatalf(string(errOutput))
	}
	values := strings.Split(string(output), " ")
	//Successful output would be of the type
	//kubeadm join <server:port> --token <token> --discovery-token-ca-cert-hash <sha>
	if len(values) != 7 { //TODO Needs a better way but seems good-enough for now
		log.Fatalf("Could not get token from master %s", pm.SSHConfig.Host)
	}
	secret := corev1.Secret{}
	token := []byte(values[4])
	cahash := []byte(values[6])
	secret.Data = map[string][]byte{}
	secret.Data["token"] = token
	secret.Data["cahash"] = cahash
	bytes, err := client.ReadFile("/etc/kubernetes/admin.conf")
	if err != nil {
		log.Fatalf("Failed to get kubeconfig with err %v", err)
	}
	return &secret, bytes
}

func init() {
	createCmd.AddCommand(machineCmdCreate)
	machineCmdCreate.Flags().String("ip", "10.0.0.1", "IP of the machine")
	machineCmdCreate.Flags().Int("port", 22, "SSH port")
	machineCmdCreate.Flags().String("role", "node", "Role of the machine. Can be master/node")
	machineCmdCreate.Flags().String("publicKeys", "", "Comma separated list of public host keys for the machine")
	machineCmdCreate.Flags().String("sshSecretName", "sshSecret", "Name of the sshSecret to use")
	machineCmdCreate.Flags().String("iface", "eth0", "Interface that keepalived will bind to in case of master")

	deleteCmd.AddCommand(machineCmdDelete)
	machineCmdDelete.Flags().String("name", "", "Machine name")
	machineCmdDelete.Flags().String("force", "", "Force delete the machine")

	getCmd.AddCommand(machineCmdGet)
	machineCmdGet.Flags().String("name", "", "Machine name")
}
