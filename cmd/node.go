package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"golang.org/x/crypto/ssh"
)

// nodeCmd represents the cluster command
var nodeCmd = &cobra.Command{
	Use:   "node",
	Short: "Adds a node to the cluster",
	Run: func(cmd *cobra.Command, args []string) {
		ip := cmd.Flag("ip").Value.String()
		role := cmd.Flag("role").Value.String()
		sshConfig := &ssh.ClientConfig{
			User: "root",
			Auth: []ssh.AuthMethod{
				ssh.Password(""),
			},
		}
		sshConfig.HostKeyCallback = ssh.InsecureIgnoreHostKey()
		connection, err := ssh.Dial("tcp", ip+":22", sshConfig)
		if err != nil {
			fmt.Println("Failed to dial: %s", err)
			return
		}
		session, err := connection.NewSession()
		if err != nil {
			fmt.Println("Failed to create session: %s", err)
			return
		}
		out, err := session.CombinedOutput("ls -al")
		if err != nil {
			panic(err)
		}
		fmt.Println(string(out))
		connection.Close()
	},
}

func init() {
	addCmd.AddCommand(nodeCmd)
	nodeCmd.Flags().String("ip", "10.0.0.1", "IP of the node")
	nodeCmd.Flags().String("role", "worker", "Role of the node. Can be master/worker")
}
