package cmd

import (
	"fmt"
	"log"
	"os"

	cctlstate "github.com/platform9/cctl/pkg/state/v1"

	spclientfake "github.com/platform9/ssh-provider/pkg/client/clientset_generated/clientset/fake"
	"github.com/spf13/cobra"
	kubeclientfake "k8s.io/client-go/kubernetes/fake"
	clusterclientfake "sigs.k8s.io/cluster-api/pkg/client/clientset_generated/clientset/fake"
)

var stateFilename string
var state *cctlstate.State

var rootCmd = &cobra.Command{
	Use: "cctl",
	PreRun: func(cmd *cobra.Command, args []string) {
		InitState()
	},
	Long: `Platform9 tool for Kubernetes cluster management.
This tool lets you create, scale, backup and restore
your on-premise Kubernetes cluster.`,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVar(&stateFilename, "state", "/etc/cctl-state.yaml", "state file")
}

func InitState() {
	kubeClient := kubeclientfake.NewSimpleClientset()
	clusterClient := clusterclientfake.NewSimpleClientset()
	spClient := spclientfake.NewSimpleClientset()
	state = cctlstate.NewFromFile(stateFilename, kubeClient, clusterClient, spClient)

	if err := state.PushToAPIs(); err != nil {
		log.Fatalf("Unable to sync on-disk state: %v", err)
	}
}
