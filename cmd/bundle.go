/**
 *   Copyright 2018 Platform9 Systems, Inc.
 *
 *   Licensed under the Apache License, Version 2.0 (the "License");
 *   you may not use this file except in compliance with the License.
 *   You may obtain a copy of the License at
 *
 *       http://www.apache.org/licenses/LICENSE-2.0
 *
 *   Unless required by applicable law or agreed to in writing, software
 *   distributed under the License is distributed on an "AS IS" BASIS,
 *   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 *   See the License for the specific language governing permissions and
 *   limitations under the License.
 */
package cmd

import (
	"fmt"
	"time"

	"github.com/platform9/cctl/common"
	log "github.com/platform9/cctl/pkg/logrus"
	sputil "github.com/platform9/ssh-provider/pkg/controller"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func init() {
	rootCmd.AddCommand(bundleCmd)
	bundleCmd.Flags().String("output", "", "File path for support bundle tar file")
	bundleCmd.Flags().String("ip", "", "IP address of the machine")
	bundleCmd.MarkFlagRequired("ip")
}

var bundleCmd = &cobra.Command{
	Use:   "bundle",
	Short: "Create a support bundle for a node",
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		InitState()
		// PersistentPreRuns are not chained https://github.com/spf13/cobra/issues/216
		// Therefore LogLevel must be set in all the PersistentPreRuns
		if err := log.SetLogLevelUsingString(LogLevel); err != nil {
			log.Fatalf("Unable to parse log level %s", LogLevel)
		}
	},
	Run: func(cmd *cobra.Command, args []string) {
		ip := cmd.Flag("ip").Value.String()
		targetMachine, err := state.ClusterClient.ClusterV1alpha1().Machines(common.DefaultNamespace).Get(ip, metav1.GetOptions{})
		if err != nil {
			log.Fatalf("Unable to get machine %q: %v", ip, err)
		}
		targetMachineSpec, err := sputil.GetMachineSpec(*targetMachine)
		if err != nil {
			log.Fatalf("Unable to decode machine %q spec: %v", targetMachine.Name, err)
		}
		targetProvisionedMachine, err := state.SPClient.SshproviderV1alpha1().ProvisionedMachines(common.DefaultNamespace).Get(targetMachineSpec.ProvisionedMachineName, metav1.GetOptions{})
		if err != nil {
			log.Fatalf("Unable to get provisioned machine %q: %v", targetMachineSpec.ProvisionedMachineName, err)
		}
		targetMachineClient, err := sshMachineClientFromSSHConfig(targetProvisionedMachine.Spec.SSHConfig)
		if err != nil {
			log.Fatalf("unable to create machine client for machine %q: %v", targetMachine.Name, err)
		}
		t := time.Now()

		bundleFileBaseName := fmt.Sprintf("%s-%s-%s.tgz", common.SupportBundleFileNamePrefix, ip, t.Format(time.RFC3339))
		localPath := cmd.Flag("output").Value.String()
		if len(localPath) == 0 {
			localPath = bundleFileBaseName
		}
		remotePath := fmt.Sprintf("/tmp/%s", bundleFileBaseName)
		command := fmt.Sprintf("%s bundle --output %s", common.DashcamCommandPath, remotePath)
		log.Printf("Started creating support bundle for %s. This will take a few minutes.", ip)
		stdOut, stdErr, err := targetMachineClient.RunCommand(command)
		if err != nil {
			log.Fatalf("Failed to create support bundle %q: %v (stdout: %q, stderr: %q)", command, err, string(stdOut), string(stdErr))
		}
		defer targetMachineClient.RemoveFile(remotePath)
		if err = downloadRemoteFile(remotePath, localPath, targetMachineClient); err != nil {
			log.Fatalf("Failed to download support bundle: %v", err)
		}
		log.Infof("Support bundle downloaded to %s ", localPath)
	},
}
