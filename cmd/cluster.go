package cmd

import (
	"log"
	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/apis/meta/v1"

	clusterv1 "sigs.k8s.io/cluster-api/pkg/apis/cluster/v1alpha1"
	"github.com/platform9/pf9-clusteradm/statefileutil"
	"github.com/platform9/pf9-clusteradm/common"
)

// clusterCmd represents the cluster command
var clusterCmd = &cobra.Command{
	Use:   "cluster",
	Short: "Creates clusterspec in the current directory",
	Run: func(cmd *cobra.Command, args []string) {

		sshProviderConfig, err := common.CreateSSHClusterProviderConfig(cmd)
		if err != nil {
			log.Fatal(err)
		}

		cluster := clusterv1.Cluster{
			TypeMeta: v1.TypeMeta{
				Kind: "Cluster",
			},
			ObjectMeta: v1.ObjectMeta{
				Name: cmd.Flag("name").Value.String(),
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
