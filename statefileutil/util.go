package statefileutil

import (
	"io/ioutil"
	"log"
	"os"

	"github.com/ghodss/yaml"
	"github.com/platform9/pf9-clusteradm/common"
	pm "github.com/platform9/ssh-provider/provisionedmachine"
	clusterv1 "sigs.k8s.io/cluster-api/pkg/apis/cluster/v1alpha1"
)

const (
	STATE_FILE_PATH = "/tmp/cluster-state.yaml"
)

func checkFileExists() (bool, error) {
	_, err := os.Stat(STATE_FILE_PATH)
	if err == nil {
		return true, nil
	}
	return false, err
}

func GetProvisionedMachine(cs common.ClusterState, ip string) *pm.ProvisionedMachine {
	for _, m := range cs.ProvisionedMachines {
		if m.SSHConfig.Host == ip {
			log.Printf("Found provisioned node for ip %s", ip)
			return &m
		}
	}
	return nil
}

func ReadStateFile() (common.ClusterState, error) {
	cs := new(common.ClusterState)
	ret, _ := checkFileExists()
	if ret == false {
		return *cs, nil
	}
	log.Print("Using the existing state file")
	d, err := ioutil.ReadFile(STATE_FILE_PATH)
	if err != nil {
		return *cs, err
	}
	yaml.Unmarshal(d, cs)
	return *cs, nil
}

func Upsert(key string, value string) {

}

func WriteStateFile(cs *common.ClusterState) error {
	if cs != nil {
		bytes, _ := yaml.Marshal(cs)
		ioutil.WriteFile(STATE_FILE_PATH, bytes, 0600)
	}
	return nil
}

func GetClusterSpec() (*clusterv1.Cluster, error) {
	cs, err := ReadStateFile()
	if err == nil {
		return &cs.Cluster, nil
	}
	return nil, err
}

func GetMachinesSpec() ([]clusterv1.Machine, error) {
	cs, err := ReadStateFile()
	if err == nil {
		return cs.Machines, nil
	}
	return nil, err
}
