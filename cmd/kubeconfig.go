package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var kubeconfigCmdGet = &cobra.Command{
	Use:   "kubeconfig",
	Short: "Get kubeconfig for cluster",
	Run: func(cmd *cobra.Command, args []string) {
		// Stub code
		fmt.Println("Running get kubeconfig")
	},
}

func init() {
	getCmd.AddCommand(kubeconfigCmdGet)
	kubeconfigCmdGet.Flags().String("file", "", "Specify the file to write kubeconfig to. If not specified, output on stdout")
}
