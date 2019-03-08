/*
Copyright 2019 The cctl authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

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
