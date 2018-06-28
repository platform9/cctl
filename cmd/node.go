package cmd

import (
	"fmt"

	"github.com/platform9/pf9-clusteradm/common"
	"github.com/platform9/pf9-clusteradm/statefileutil"
	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/apis/meta/v1"

	clusterv1 "sigs.k8s.io/cluster-api/pkg/apis/cluster/v1alpha1"

	"log"
)

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

		machine := clusterv1.Machine{
			TypeMeta: v1.TypeMeta{
				Kind:       "Machine",
				APIVersion: "cluster.k8s.io/v1alpha1",
			},
			ObjectMeta: v1.ObjectMeta{
				Name:              cmd.Flag("ip").Value.String(),
				CreationTimestamp: timestamp,
			},
			Spec: clusterv1.MachineSpec{
				ObjectMeta: v1.ObjectMeta{
					CreationTimestamp: timestamp,
				},
				ProviderConfig: *sshMachineProviderConfig,
			},
		}

		cs, err := statefileutil.ReadStateFile()
		if err != nil {
			log.Fatal(err)
		}

		machines := append(cs.Machines, machine)
		cs.Machines = machines

		//actuator, err := sshMachineActuator.NewActuator()
		//if err != nil {
		//	log.Fatal(err)
		//}

		//cluster := &cs.Cluster
		// call the actuator
		//err = actuator.Create(cluster, &machine)
		//if err != nil {
		//	log.Fatal(err)
		//}
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
