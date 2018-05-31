package cmd

type ClusterSpec struct {
	Name           string `yaml:"name"`
	ServiceNetwork string `yaml:"serviceNetwork"`
	PodNetwork     string `yaml:"podNetwork"`
	Vip            string `yaml:"vip"`
	Cacert         string `yaml:"cacert"`
	Cakey          string `yaml:"cakey"`
	Token          string `yaml:"token"`
	Version        string `yaml:"version"`
}
