package cmd

import (
	"fmt"
	"log"

	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/platform9/pf9-clusteradm/common"
	"github.com/platform9/pf9-clusteradm/statefileutil"
	clusterv1 "sigs.k8s.io/cluster-api/pkg/apis/cluster/v1alpha1"
)

// clusterCmd represents the cluster command
var clusterCmdCreate = &cobra.Command{
	Use:   "cluster",
	Short: "Creates clusterspec in the current directory",
	Run: func(cmd *cobra.Command, args []string) {

		sshProviderConfig, err := common.CreateSSHClusterProviderConfig(cmd)
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
				ProviderConfig: *sshProviderConfig,
			},
		}
		cs, err := statefileutil.ReadStateFile()
		if err != nil {
			log.Fatal(err)
		}
		cs.Cluster = cluster
		cs.Extra.K8sVersion = cmd.Flag("version").Value.String()
		cs.Extra.Vip = cmd.Flag("vip").Value.String()
		statefileutil.WriteStateFile(&cs)
	},
}

var clusterCmdDelete = &cobra.Command{
	Use:   "cluster",
	Short: "Deletes a node to the cluster",
	Run: func(cmd *cobra.Command, args []string) {
		// Stub code
		fmt.Println("Running cluster delete")
	},
}

var clusterCmdGet = &cobra.Command{
	Use:   "cluster",
	Short: "Get the cluster details",
	Run: func(cmd *cobra.Command, args []string) {
		// Stub code
		fmt.Println("Running get cluster")
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
	clusterCmdCreate.Flags().String("cacert", "", "Base64 encoded CA cert for compoenents to trust")
	clusterCmdCreate.Flags().String("cakey", "", "Base64 encoded CA key for signing certs")
	clusterCmdCreate.Flags().String("version", "1.10.2", "Kubernetes version")

	deleteCmd.AddCommand(clusterCmdDelete)
	deleteCmd.Flags().String("force", "", "Force delete a cluster")

	getCmd.AddCommand(clusterCmdGet)

	upgradeCmd.AddCommand(clusterCmdUpgrade)

	recoverCmd.AddCommand(clusterCmdRecover)

	backupCmd.AddCommand(clusterCmdBackup)
}
