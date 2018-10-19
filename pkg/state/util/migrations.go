package util

import (
	log "github.com/platform9/cctl/pkg/logrus"

	spv1 "github.com/platform9/ssh-provider/pkg/apis/sshprovider/v1alpha1"
	sputil "github.com/platform9/ssh-provider/pkg/controller"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/platform9/cctl/common"
	"github.com/platform9/cctl/pkg/state/v0"
	"github.com/platform9/cctl/pkg/state/v1"
	"github.com/platform9/cctl/pkg/state/v2"
)

func StateV1FromStateV0(stateV0 *v0.State) *v1.State {
	stateV1 := v1.State{
		SchemaVersion: v1.Version,
		Filename:      stateV0.Filename,
		ClusterClient: stateV0.ClusterClient,
		KubeClient:    stateV0.KubeClient,
		SPClient:      stateV0.SPClient,
	}
	return &stateV1
}

func ClusterConfigForV0AndV1Cluster() *spv1.ClusterConfig {
	failSwapOn := false
	kubeAPIQPS := int32(20)

	cc := spv1.ClusterConfig{
		KubeAPIServer: map[string]string{
			"allow-privileged":        "true",
			"service-node-port-range": "80-32767",
		},
		KubeControllerManager: map[string]string{
			"pod-eviction-timeout": "20s",
		},
		Kubelet: &spv1.KubeletConfiguration{
			FailSwapOn:   &failSwapOn,
			KubeAPIBurst: 40,
			KubeAPIQPS:   &kubeAPIQPS,
			MaxPods:      500,
		},
	}
	return &cc
}

func StateV2FromStateV1(stateV1 *v1.State) *v2.State {
	stateV2 := v2.State{
		SchemaVersion: v2.Version,
		Filename:      stateV1.Filename,
		ClusterClient: stateV1.ClusterClient,
		KubeClient:    stateV1.KubeClient,
		SPClient:      stateV1.SPClient,
	}
	cluster, err := stateV2.ClusterClient.ClusterV1alpha1().Clusters(common.DefaultNamespace).Get(common.DefaultClusterName, metav1.GetOptions{})
	if err != nil {
		log.Fatalf("unable to get cluster %s: %v", common.DefaultClusterName, err)
	}
	clusterSpec, err := sputil.GetClusterSpec(*cluster)
	if err != nil {
		log.Fatalf("unable to get cluster spec %s: %v", common.DefaultClusterName, err)
	}
	clusterSpec.ClusterConfig = ClusterConfigForV0AndV1Cluster()
	if err := sputil.PutClusterSpec(*clusterSpec, cluster); err != nil {
		log.Fatalf("Unable to update cluster spec %s: %v", common.DefaultClusterName, err)
	}
	if _, err = stateV2.ClusterClient.ClusterV1alpha1().Clusters(common.DefaultNamespace).Update(cluster); err != nil {
		log.Fatalf("unable to update cluster spec %s: %v", common.DefaultClusterName, err)
	}
	return &stateV2
}
