package cmd

import (
	"github.com/platform9/pf9-clusteradm/common"
	"github.com/platform9/pf9-clusteradm/statefileutil"
	"github.com/spf13/cobra"
	"log"
)

var credentialsCmdCreate = &cobra.Command{
	Use:   "credentials",
	Short: "Create new SSH credentials",
	Run: func(cmd *cobra.Command, args []string) {

		sshSecret := common.SSHSecret{
			User:       cmd.Flag("user").Value.String(),
			PrivateKey: cmd.Flag("privateKey").Value.String(),
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
	credentialsCmdCreate.Flags().String("privateKey", "", "SSH privateKey")
	credentialsCmdCreate.MarkFlagRequired("privateKey")
}
