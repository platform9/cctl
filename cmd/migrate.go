package cmd

import (
	"github.com/platform9/cctl/pkg/migrate"
	statePkg "github.com/platform9/cctl/pkg/state"
	"github.com/spf13/cobra"
	"io/ioutil"
	"log"
	"os"

	spclientfake "github.com/platform9/ssh-provider/pkg/client/clientset_generated/clientset/fake"
	kubeclientfake "k8s.io/client-go/kubernetes/fake"
	clusterclientfake "sigs.k8s.io/cluster-api/pkg/client/clientset_generated/clientset/fake"
)

// migrateCmd represents the migrate command
var migrateCmd = &cobra.Command{
	Use:   "migrate",
	Short: "Migrate state file to a newer schema",
	Run: func(cmd *cobra.Command, args []string) {
		Migrate()
	},
}

func init() {
	rootCmd.AddCommand(migrateCmd)
}

func Migrate() {
	log.Printf("Migrating state file to new schema")
	kubeClient := kubeclientfake.NewSimpleClientset()
	clusterClient := clusterclientfake.NewSimpleClientset()
	spClient := spclientfake.NewSimpleClientset()
	state = statePkg.NewWithFile(stateFilename, kubeClient, clusterClient, spClient)

	file, err := os.OpenFile(stateFilename, os.O_RDONLY|os.O_CREATE, statePkg.FileMode)
	if err != nil {
		log.Fatalf("Unable to open %q: %v\n", state.Filename, err)
	}

	defer file.Close()
	stateBytes, err := ioutil.ReadAll(file)

	migratedBytes, err := migrate.MigrateV0toV1(stateBytes)
	if err != nil {
		log.Fatal(err)
	}

	newState := migrate.DecodeMigratedState(migratedBytes)
	newState.KubeClient = state.KubeClient
	newState.ClusterClient = state.ClusterClient
	newState.SPClient = state.SPClient
	newState.Filename = stateFilename

	err = statePkg.CreateObjects(&newState)
	if err != nil {
		log.Fatalf("Unable to sync API objects: %v", err)
	}

	err = newState.PullFromAPIs()
	if err != nil {
		log.Fatalf("Unable to write to on-disk state: %v", err)
	}
	log.Printf("Finished migrating state file to new schema")

}
