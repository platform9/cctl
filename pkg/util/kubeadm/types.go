package kubeadm

type MasterConfiguration struct {
	API API `json:"api,omitempty"`
}

type API struct {
	AdvertiseAddress     string `json:"advertiseAddress,omitempty"`
	BindPort             int32  `json:"bindPort,omitempty"`
	ControlPlaneEndpoint string `json:"controlPlaneEndpoint,omitempty"`
}
