package cmd

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"text/template"
	"time"

	"log"
	"strings"

	"github.com/ghodss/yaml"
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
	clusterapiutil "sigs.k8s.io/cluster-api/pkg/util"
)

var (
	drainTimeout            time.Duration
	drainGracePeriodSeconds int
)

func machineAlreadyExists(ip string, cs *common.ClusterState) bool {
	for _, machine := range cs.Machines {
		if machine.Name == ip {
			return true
		}
	}
	return false
}

func clusterCreated(cs *common.ClusterState) bool {
	if cs.Cluster.ObjectMeta.GetName() == "" {
		return false
	}
	return true
}
func validateRequirementsForCreateOrDelete(cs *common.ClusterState) {
	if !clusterCreated(cs) {
		log.Fatalln("Cannot create machine without creating a cluster. Run create cluster first")
	}
	if cs.SSHCredentials == nil {
		log.Fatalln("Cannot create machine without ssh credentials. Run create credentials first")
	}
}
func validateRequirementsForCreate(cs *common.ClusterState, ip, role string) {
	validateRequirementsForCreateOrDelete(cs)
	if role != "node" && role != "master" {
		log.Fatalln("Failed to create machine, invalid role (use node/master)")
	}
	if machineAlreadyExists(ip, cs) {
		log.Fatalln("Failed to create machine, machine already exists")
	}
	if role == "node" {
		if statefileutil.GetMasterCount(cs) == 0 {
			log.Fatalln("Cannot create machine as node until master is created")
		}
	}
}

func validateRequirementsForDelete(cs *common.ClusterState, ip string) {
	validateRequirementsForCreateOrDelete(cs)
	machine := statefileutil.GetMachine(cs, ip)
	if machine == nil {
		log.Fatalf("Failed to delete machine, machine does not exist")
	}
	if clusterapiutil.IsMaster(machine) {
		//if it is the last master and there are still nodes in the cluster
		if statefileutil.GetMasterCount(cs) == 1 && statefileutil.GetNodeCount(cs) > 0 {
			log.Fatalln("Failed to delete machine, cannot delete all masters as cluster has nodes")
		}
	}
}

// machineCmdCreate represents the machine create command
var machineCmdCreate = &cobra.Command{
	Use:   "machine",
	Short: "Adds a machine to the cluster",
	Run: func(cmd *cobra.Command, args []string) {
		spv1Codec, err := sshproviderv1.NewCodec()
		if err != nil {
			log.Fatalf("Could not initialize codec for internal types: %v", err)
		}

		timestamp := v1.Now()
		ip := cmd.Flag("ip").Value.String()
		iface := cmd.Flag("iface").Value.String()
		role := cmd.Flag("role").Value.String()
		port, err := strconv.Atoi(cmd.Flag("port").Value.String())
		if err != nil {
			log.Fatalf("Invalid port %v", err)
		}

		cs, err := statefileutil.ReadStateFile()
		if err != nil {
			log.Fatal(err)
		}
		validateRequirementsForCreate(&cs, ip, role)

		publicKeyPaths := cmd.Flag("publicKeys").Value
		publicKeys := []string{}
		if publicKeyPaths != nil && len(publicKeyPaths.String()) > 0 {
			paths := strings.Split(publicKeyPaths.String(), ",")
			for _, path := range paths {
				file, err := os.Open(path)
				if err != nil {
					log.Fatalf("Failed to open public key file %q: %v", path, err)
				}
				defer file.Close()
				lineReader := bufio.NewReader(file)
				publicKeyBytes, isPrefix, err := lineReader.ReadLine()
				if err != nil {
					log.Fatalf("Failed to read public key from file %q: %v", path, err)
				}
				if isPrefix {
					log.Fatalf("Failed to read public key from file %q: first line exceeds buffer size", path)
				}
				publicKeys = append(publicKeys, string(publicKeyBytes))
			}
		}

		sshMachineProviderConfig := &sshproviderv1.SSHMachineProviderConfig{
			TypeMeta: v1.TypeMeta{
				APIVersion: "sshproviderconfig/v1alpha1",
				Kind:       "SSHMachineProviderConfig",
			},
		}
		providerConfig, err := spv1Codec.EncodeToProviderConfig(sshMachineProviderConfig)
		if err != nil {
			log.Fatal(err)
		}

		sshMachineProviderStatus := &sshproviderv1.SSHMachineProviderStatus{
			TypeMeta: v1.TypeMeta{
				APIVersion: "sshproviderconfig/v1alpha1",
				Kind:       "SSHMachineProviderStatus",
			},
		}
		providerStatus, err := spv1Codec.EncodeToProviderStatus(sshMachineProviderStatus)
		if err != nil {
			log.Fatal(err)
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
				Roles:          []clustercommon.MachineRole{clustercommon.MasterRole},
				ProviderConfig: *providerConfig,
			},
			Status: clusterv1.MachineStatus{
				ProviderStatus: *providerStatus,
			},
		}
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

		cm := &corev1.ConfigMap{
			Data: make(map[string]string),
		}
		if err := provisionedMachine.ToConfigMap(cm); err != nil {
			log.Fatalf("Error reading state: %v", err)
		}
		var clusterToken *corev1.Secret
		var kubeconfig []byte
		if role == "node" {
			clusterToken, kubeconfig = getSecretAndConfigFromMaster(cs)
			if clusterToken == nil {
				log.Fatalf("Failed to add . No master available")
			}
		}

		actuator, err := sshMachineActuator.NewActuator(
			cm,
			cs.SSHCredentials,
			cs.EtcdCA,
			cs.APIServerCA,
			cs.FrontProxyCA,
			cs.ServiceAccountKey,
			clusterToken,
		)
		if len(publicKeys) == 0 {
			actuator.InsecureIgnoreHostKey = true
			log.Printf("Not able to verify machine SSH identity: No public keys given. Continuing...")
		}
		if err != nil {
			log.Fatalf("Unable to create machine actuator: %s", err)
		}

		cluster := cs.Cluster
		err = actuator.Create(&cluster, &machine)
		if err != nil {
			log.Fatal(err)
		}
		if role == "master" {
			clusterProviderStatus := sshproviderv1.SSHClusterProviderStatus{}
			err := spv1Codec.DecodeFromProviderStatus(cluster.Status.ProviderStatus, &clusterProviderStatus)
			if err != nil {
				log.Fatalf("Failed to decode cluster ProviderStatus: %v", err)
			}

			clusterProviderConfig := sshproviderv1.SSHClusterProviderConfig{}
			err = spv1Codec.DecodeFromProviderConfig(cluster.Spec.ProviderConfig, &clusterProviderConfig)
			if err != nil {
				log.Fatalf("Failed to decode cluster ProviderConfig: %v", err)
			}
			var newEndpoint clusterv1.APIEndpoint
			if len(clusterProviderConfig.VIPConfiguration.IP) != 0 {
				newEndpoint = clusterv1.APIEndpoint{
					Host: clusterProviderConfig.VIPConfiguration.IP.String(),
					Port: common.DEFAULT_APISERVER_PORT,
				}
			} else {
				newEndpoint = clusterv1.APIEndpoint{
					Host: provisionedMachine.SSHConfig.Host,
					Port: common.DEFAULT_APISERVER_PORT,
				}
			}
			log.Println("Updating cluster API endpoints")
			statefileutil.UpsertAPIEndpoint(newEndpoint, &cluster)

			if len(clusterProviderStatus.EtcdMembers) == 0 {
				clusterProviderStatus.EtcdMembers = []sshproviderv1.EtcdMember{}
			}
			machineProviderStatus := sshproviderv1.SSHMachineProviderStatus{}
			err = spv1Codec.DecodeFromProviderStatus(machine.Status.ProviderStatus, &machineProviderStatus)
			if err != nil {
				log.Fatalf("Failed to decode machine ProviderStatus: %v", err)
			}
			clusterProviderStatus.EtcdMembers = append(clusterProviderStatus.EtcdMembers, *machineProviderStatus.EtcdMember)
			status, err := spv1Codec.EncodeToProviderStatus(&clusterProviderStatus)
			if err != nil {
				log.Fatalf("Failed to encode cluster provider status with error %v", err)
			}
			cs.Cluster.Status.ProviderStatus = *status
			status, err = spv1Codec.EncodeToProviderStatus(&machineProviderStatus)
			if err != nil {
				log.Fatalf("Failed to encode machine provider status with error %v", err)
			}
			machine.Status.ProviderStatus = *status
		}
		if role == "node" {
			client := getClientForMachine(&provisionedMachine, cs.SSHCredentials)
			if err := client.WriteFile("/etc/kubernetes/admin.conf", 0644, kubeconfig); err != nil {
				log.Fatalf("Failed to write admin kubeconfig to node: %v", err)
			}
		}
		cs.ProvisionedMachines = append(cs.ProvisionedMachines, provisionedMachine)
		cs.Machines = append(cs.Machines, machine)
		if err := statefileutil.WriteStateFile(&cs); err != nil {
			log.Fatalf("Error writing state: %v", err)
		}
	},
}

var machineCmdDelete = &cobra.Command{
	Use:   "machine",
	Short: "Deletes a machine from the cluster",
	Run: func(cmd *cobra.Command, args []string) {
		spv1Codec, err := sshproviderv1.NewCodec()
		if err != nil {
			log.Fatalf("Could not initialize codec for internal types: %v", err)
		}
		ip := cmd.Flag("ip").Value.String()
		cs, err := statefileutil.ReadStateFile()
		if err != nil {
			log.Fatal(err)
		}
		cluster := cs.Cluster
		provisionedMachine := statefileutil.GetProvisionedMachine(&cs, ip)
		cm := &corev1.ConfigMap{
			Data: make(map[string]string),
		}
		if err := provisionedMachine.ToConfigMap(cm); err != nil {
			log.Fatalf("Error reading state: %v", err)
		}
		clusterToken := &corev1.Secret{
			Data: make(map[string][]byte),
		}

		validateRequirementsForDelete(&cs, ip)

		machine := statefileutil.GetMachine(&cs, ip)

		actuator, err := sshMachineActuator.NewActuator(
			cm,
			cs.SSHCredentials,
			cs.EtcdCA,
			cs.APIServerCA,
			cs.FrontProxyCA,
			cs.ServiceAccountKey,
			clusterToken,
		)
		if len(provisionedMachine.SSHConfig.PublicKeys) == 0 {
			actuator.InsecureIgnoreHostKey = true
			log.Printf("Not able to verify machine SSH identity: No public keys given. Continuing...")
		}
		if err != nil {
			log.Fatalf("Unable to create machine actuator: %s", err)
		}

		masterPM := statefileutil.GetMaster(&cs)
		masterCM := &corev1.ConfigMap{
			Data: make(map[string]string),
		}
		if err := masterPM.ToConfigMap(masterCM); err != nil {
			log.Fatalf("Error reading state: %v", err)
		}
		masterClient := getClientForMachine(provisionedMachine, cs.SSHCredentials)

		// TODO(dlipovetsky) Handle /opt/bin/kubectl not found. Possibly infer
		// that the nodeadm reset ran at least as far as removing the kubectl
		// binary. nodeName includes the object kind, i.e.,

		// "node/the-name-of-the-node"
		stdOut, stdErr, err := masterClient.RunCommand("/opt/bin/kubectl --kubeconfig=/etc/kubernetes/admin.conf get node --selector kubernetes.io/hostname=$(hostname -f) -oname")
		if err != nil {
			log.Fatalf("Could not identify the cluster node for machine %s: %s (%s) (%s)", ip, err, string(stdOut), string(stdErr))
		}
		nodeName := strings.TrimSpace(string(stdOut))
		if len(nodeName) != 0 {
			log.Printf("Draining cluster node %s for machine", nodeName)
			// --ignore-daemonsets is used because critical components (kube-proxy, overlay network) run as daemonsets
			// --delete-local-data is NOT used; pods using emptyDir volumes must be removed by the user, since removing them causes the data to be lost
			// --force is NOT used; unmanaged pods must be removed by the user, since they won't be rescheduled to another node
			stdOut, stdErr, err = masterClient.RunCommand(fmt.Sprintf("/opt/bin/kubectl --kubeconfig=/etc/kubernetes/admin.conf drain --timeout=%v --grace-period=%v --ignore-daemonsets %v", drainTimeout, drainGracePeriodSeconds, nodeName))
			if err != nil {
				log.Fatalf("Could not drain pods from cluster node %s: %s (%s) (%s)", ip, err, string(stdOut), string(stdErr))
			}
			log.Println(string(stdOut))

			log.Printf("Deleting cluster node %s for machine", nodeName)
			stdOut, stdErr, err = masterClient.RunCommand(fmt.Sprintf("/opt/bin/kubectl --kubeconfig=/etc/kubernetes/admin.conf delete %v", nodeName))
			if err != nil {
				log.Fatalf("Could not delete cluster node %s: %s (%s) (%s)", ip, err, string(stdOut), string(stdErr))
			}
			log.Println(string(stdOut))
		}

		log.Printf("Deleting machine")
		err = actuator.Delete(&cluster, machine)
		if err != nil {
			log.Fatal(err)
		}

		log.Print("Updating state")
		if clusterapiutil.IsMaster(machine) {
			log.Printf("Removing etcd member from cluster state")
			statefileutil.DeleteEtcdMemberFromClusterState(&cs, machine)

			clusterProviderConfig := sshproviderv1.SSHClusterProviderConfig{}
			err = spv1Codec.DecodeFromProviderConfig(cluster.Spec.ProviderConfig, &clusterProviderConfig)
			if err != nil {
				log.Fatalf("Failed to decode cluster ProviderConfig: %v", err)
			}
			var targetEndpoint clusterv1.APIEndpoint
			if len(clusterProviderConfig.VIPConfiguration.IP) != 0 {
				// If there is 1 master, it is being deleted.
				if statefileutil.GetMasterCount(&cs) == 1 {
					targetEndpoint = clusterv1.APIEndpoint{
						Host: clusterProviderConfig.VIPConfiguration.IP.String(),
						Port: common.DEFAULT_APISERVER_PORT,
					}
				}
			} else {
				targetEndpoint = clusterv1.APIEndpoint{
					Host: provisionedMachine.SSHConfig.Host,
					Port: common.DEFAULT_APISERVER_PORT,
				}
			}
			log.Println("Updating cluster API endpoints")
			statefileutil.DeleteAPIEndpoint(targetEndpoint, &cluster)
		}
		statefileutil.DeleteMachine(&cs, ip)
		statefileutil.DeleteProvisionedMachine(&cs, ip)
		if err := statefileutil.WriteStateFile(&cs); err != nil {
			log.Fatalf("Error writing state: %v", err)
		}
		log.Printf("Successfully deleted machine %s", ip)
	},
}

var machineCmdGet = &cobra.Command{
	Use:   "machine",
	Short: "Get a machine",
	Run: func(cmd *cobra.Command, args []string) {
		cs, err := statefileutil.ReadStateFile()
		if err != nil {
			log.Fatalf("Unable to read state file: %s", err)
		}
		switch outputFmt {
		case "yaml":
			// Flag yaml specificed. Print cluster spec as yaml
			bytes, err := yaml.Marshal(cs.Machines)
			if err != nil {
				log.Fatalf("Unable to marshal cluster spec file to yaml: %s", err)
			}
			os.Stdout.Write(bytes)
		case "json":
			// Flag json specified. Print cluster spec as json
			bytes, err := json.Marshal(cs.Machines)
			if err != nil {
				log.Fatalf("Unable to marshal cluster spec file to json: %s", err)
			}
			os.Stdout.Write(bytes)
		case "":
			// Pretty print cluster details
			t := template.Must(template.New("MachineV1PrintTemplate").Parse(common.MachineV1PrintTemplate))
			if err := t.Execute(os.Stdout, cs.Machines); err != nil {
				log.Fatalf("Could not pretty print cluster details: %s", err)
			}
		default:
			log.Fatalf("Unsupported output format %q", outputFmt)
		}
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
	machineCmdDelete.Flags().String("ip", "", "IP of the machine")
	machineCmdDelete.Flags().String("force", "", "Force delete the machine")
	machineCmdDelete.Flags().DurationVar(&drainTimeout, "drain-timeout", common.DRAIN_TIMEOUT, "The length of time to wait before giving up, zero means infinite")
	machineCmdDelete.Flags().IntVar(&drainGracePeriodSeconds, "drain-graceperiod", common.DRAIN_GRACE_PERIOD_SECONDS, "Period of time in seconds given to each pod to terminate gracefully. If negative, the default value specified in the pod will be used.")

	getCmd.AddCommand(machineCmdGet)
}
