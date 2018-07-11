package cmd

import (
	"io/ioutil"
	"log"

	"github.com/platform9/pf9-clusteradm/statefileutil"
	"github.com/spf13/cobra"
	v1 "k8s.io/api/core/v1"
)

var credentialsCmdCreate = &cobra.Command{
	Use:   "credentials",
	Short: "Create new SSH credentials",
	Run: func(cmd *cobra.Command, args []string) {
		bytes, err := ioutil.ReadFile(cmd.Flag("privateKey").Value.String())
		if err != nil {
			log.Fatalf("Failed to read key file with err %v\n", err)
		}
		sshSecret := v1.Secret{}
		sshSecret.Data = map[string][]byte{}
		sshSecret.Data["username"] = []byte(cmd.Flag("user").Value.String())
		sshSecret.Data["ssh-privatekey"] = bytes
		cs, err := statefileutil.ReadStateFile()
		if err != nil {
			log.Fatal(err)
		}
		cs.SSHCredentials = &sshSecret
		if err := statefileutil.WriteStateFile(&cs); err != nil {
			log.Fatalf("error reading state: %v", err)
		}
	},
}

func init() {
	createCmd.AddCommand(credentialsCmdCreate)
	credentialsCmdCreate.Flags().String("user", "root", "SSH username")
	credentialsCmdCreate.Flags().String("privateKey", "", "SSH privateKey file location")
	credentialsCmdCreate.MarkFlagRequired("privateKey")
}
