package cmd

import (
	"io/ioutil"
	"log"

	"encoding/base64"

	"github.com/platform9/pf9-clusteradm/common"
	"github.com/platform9/pf9-clusteradm/statefileutil"
	"github.com/spf13/cobra"
)

var credentialsCmdCreate = &cobra.Command{
	Use:   "credentials",
	Short: "Create new SSH credentials",
	Run: func(cmd *cobra.Command, args []string) {
		bytes, err := ioutil.ReadFile(cmd.Flag("privateKey").Value.String())
		if err != nil {
			log.Fatalf("Failed to read key file with err %v\n", err)
		}
		sshSecret := common.SSHSecret{
			User:       cmd.Flag("user").Value.String(),
			PrivateKey: base64.StdEncoding.EncodeToString(bytes),
		}
		cs, err := statefileutil.ReadStateFile()
		if err != nil {
			log.Fatal(err)
		}
		cs.SSHSecret = sshSecret
		statefileutil.WriteStateFile(&cs)
	},
}

func init() {
	createCmd.AddCommand(credentialsCmdCreate)
	credentialsCmdCreate.Flags().String("user", "root", "SSH username")
	credentialsCmdCreate.Flags().String("privateKey", "", "SSH privateKey file location")
	credentialsCmdCreate.MarkFlagRequired("privateKey")
}
