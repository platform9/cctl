package util_test

import (
	"io/ioutil"
	"log"
	"testing"

	"github.com/google/go-cmp/cmp"
	spclientfake "github.com/platform9/ssh-provider/pkg/client/clientset_generated/clientset/fake"
	sputil "github.com/platform9/ssh-provider/pkg/controller"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kubeclientfake "k8s.io/client-go/kubernetes/fake"
	clusterclientfake "sigs.k8s.io/cluster-api/pkg/client/clientset_generated/clientset/fake"

	"github.com/platform9/cctl/common"
	stateutil "github.com/platform9/cctl/pkg/state/util"
	"github.com/platform9/cctl/pkg/state/v0"
	"github.com/platform9/cctl/pkg/state/v1"
	"github.com/platform9/cctl/pkg/state/v2"
)

func TestStateV1FromStateV0(t *testing.T) {
	kubeClient := kubeclientfake.NewSimpleClientset()
	clusterClient := clusterclientfake.NewSimpleClientset()
	spClient := spclientfake.NewSimpleClientset()

	stateV0 := v0.NewWithFile("testdata/v0.yaml", kubeClient, clusterClient, spClient)
	if err := stateV0.PushToAPIs(); err != nil {
		t.Fatalf("Error reading from state: %v", err)
	}
	stateV1 := stateutil.StateV1FromStateV0(stateV0)

	// Test in-memory migration
	expectedSchemaVersion := v1.Version
	actualSchemaVersion := stateV1.SchemaVersion
	if expectedSchemaVersion != actualSchemaVersion {
		t.Fatalf("Expected SchemaVersion %d, found %d", expectedSchemaVersion, actualSchemaVersion)
	}

	// Test persistence after migration
	V1File, err := ioutil.TempFile("/tmp", "cctl-migrations-test")
	if err != nil {
		t.Fatalf("Error creating temp state file: %v", err)
	}
	defer V1File.Close()
	stateV1.Filename = V1File.Name()
	if err := stateV1.PullFromAPIs(); err != nil {
		t.Fatalf("Error calling PullFromAPIs after migration: %v", err)
	}
}

func TestStateV2FromStateV1(t *testing.T) {
	kubeClient := kubeclientfake.NewSimpleClientset()
	clusterClient := clusterclientfake.NewSimpleClientset()
	spClient := spclientfake.NewSimpleClientset()

	stateV1 := v1.NewWithFile("testdata/v1.yaml", kubeClient, clusterClient, spClient)
	if err := stateV1.PushToAPIs(); err != nil {
		t.Fatalf("Error reading from state: %v", err)
	}
	stateV2 := stateutil.StateV2FromStateV1(stateV1)

	// Test in-memory migration
	expectedSchemaVersion := v2.Version
	actualSchemaVersion := stateV2.SchemaVersion
	if expectedSchemaVersion != actualSchemaVersion {
		t.Fatalf("Expected SchemaVersion %d, found %d", expectedSchemaVersion, actualSchemaVersion)
	}
	expectedClusterConfig := stateutil.ClusterConfigForV0AndV1Cluster()
	cluster, err := stateV2.ClusterClient.ClusterV1alpha1().Clusters(common.DefaultNamespace).Get(common.DefaultClusterName, metav1.GetOptions{})
	if err != nil {
		log.Fatalf("unable to get cluster %s: %v", common.DefaultClusterName, err)
	}
	clusterSpec, err := sputil.GetClusterSpec(*cluster)
	if err != nil {
		log.Fatalf("unable to get cluster spec %s: %v", common.DefaultClusterName, err)
	}
	actualClusterConfig := clusterSpec.ClusterConfig
	if !cmp.Equal(expectedClusterConfig, actualClusterConfig) {
		t.Fatalf("Expected ClusterConfig %v, found %v", expectedClusterConfig, actualClusterConfig)
	}

	// Test persistence after migration
	V2File, err := ioutil.TempFile("/tmp", "cctl-migrations-test")
	if err != nil {
		t.Fatalf("Error creating temp state file: %v", err)
	}
	defer V2File.Close()
	stateV2.Filename = V2File.Name()
	if err := stateV2.PullFromAPIs(); err != nil {
		t.Fatalf("Error calling PullFromAPIs after migration: %v", err)
	}
}
