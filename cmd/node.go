package cmd

import (
	"github.com/platform9/pf9-clusteradm/common"
	"github.com/platform9/pf9-clusteradm/statefileutil"
	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/apis/meta/v1"

	clusterv1 "sigs.k8s.io/cluster-api/pkg/apis/cluster/v1alpha1"

	"log"
)

// nodeCmd represents the cluster command
var nodeCmd = &cobra.Command{
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
		// call the actuator
		statefileutil.WriteStateFile(&cs)
	},
}

func init() {
	addCmd.AddCommand(nodeCmd)
	nodeCmd.Flags().String("ip", "10.0.0.1", "IP of the machine")
	nodeCmd.Flags().Int("port", 22, "SSH port")
	nodeCmd.Flags().String("role", "worker", "Role of the node. Can be master/worker")
	nodeCmd.Flags().String("publicKeys", "", "Comma separated list of public host keys for the machine")
	nodeCmd.Flags().String("sshSecretName", "sshSecret", "Name of the sshSecret to use")
}
