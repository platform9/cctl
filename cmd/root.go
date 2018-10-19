package cmd

import (
	"fmt"
	"os"

	log "github.com/platform9/cctl/pkg/logrus"
	cctlstate "github.com/platform9/cctl/pkg/state/v2"
	"github.com/sirupsen/logrus"

	spclientfake "github.com/platform9/ssh-provider/pkg/client/clientset_generated/clientset/fake"
	"github.com/spf13/cobra"
	kubeclientfake "k8s.io/client-go/kubernetes/fake"
	clusterclientfake "sigs.k8s.io/cluster-api/pkg/client/clientset_generated/clientset/fake"
)

var stateFilename string
var state *cctlstate.State
var LogLevel string

var rootCmd = &cobra.Command{
	Use: "cctl",
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		logLevel, err := logrus.ParseLevel(LogLevel)
		if err != nil {
			log.Fatalf("Could not parse log level %v", logLevel)
		}
		log.SetLogLevel(logLevel)
	},
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
	rootCmd.PersistentFlags().StringVarP(&LogLevel, "log-level", "l", "info", "set log level for output, permitted values debug, info, warn, error, fatal and panic")
}

func InitState() {
	kubeClient := kubeclientfake.NewSimpleClientset()
	clusterClient := clusterclientfake.NewSimpleClientset()
	spClient := spclientfake.NewSimpleClientset()
	state = cctlstate.NewWithFile(stateFilename, kubeClient, clusterClient, spClient)

	if err := state.PushToAPIs(); err != nil {
		log.Fatalf("Unable to sync on-disk state: %v", err)
	}
}
