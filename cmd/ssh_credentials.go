package cmd

import (
	"github.com/platform9/pf9-clusteradm/common"
	"github.com/platform9/pf9-clusteradm/statefileutil"
	"github.com/spf13/cobra"
	"log"
)

// nodeCmd represents the cluster command
var sshCredentialsCmd = &cobra.Command{
	Use:   "sshcredentials",
	Short: "Create new SSH credentails",
	Run: func(cmd *cobra.Command, args []string) {

		sshCredentials := common.SSHCredentials{
			User:       cmd.Flag("user").Value.String(),
			PrivateKey: cmd.Flag("privateKey").Value.String(),
		}

		sshSecret := common.SSHSecret{
			Name:           cmd.Flag("name").Value.String(),
			SSHCredentials: sshCredentials,
		}

		cs, err := statefileutil.ReadStateFile()
		if err != nil {
			log.Fatal(err)
		}

		cs.SSHSecret = append(cs.SSHSecret, sshSecret)
		statefileutil.WriteStateFile(&cs)
	},
}

func init() {
	createCmd.AddCommand(sshCredentialsCmd)
	sshCredentialsCmd.Flags().String("name", "sshCreds", "Label used to identify the credentials")
	sshCredentialsCmd.Flags().String("user", "root", "SSH username")
	sshCredentialsCmd.Flags().String("privateKey", "", "SSH privateKey")
	sshCredentialsCmd.MarkFlagRequired("privateKey")
}
