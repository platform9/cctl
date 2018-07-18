package api

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"

	sp_install "github.com/platform9/ssh-provider/pkg/apis/sshprovider/install"

	"github.com/kubernetes-incubator/apiserver-builder/pkg/builders"
	ca_install "sigs.k8s.io/cluster-api/pkg/apis/cluster/install"
)

var (
	groupFactoryRegistry = builders.GroupFactoryRegistry
	// Registry is an instance of an API registry.
	Registry = builders.Registry
	// Scheme for API object types
	Scheme = builders.Scheme
	// ParameterCodec handles versioning of objects that are converted to query parameters.
	ParameterCodec = runtime.NewParameterCodec(Scheme)
	// Codecs for creating a server config
	Codecs = builders.Codecs
)

func init() {
	sp_install.Install(groupFactoryRegistry, Registry, Scheme)
	ca_install.Install(groupFactoryRegistry, Registry, Scheme)

	metav1.AddToGroupVersion(Scheme, schema.GroupVersion{Version: "v1"})

	unversioned := schema.GroupVersion{Group: "", Version: "v1"}
	Scheme.AddUnversionedTypes(unversioned,
		&metav1.Status{},
		&metav1.APIVersions{},
		&metav1.APIGroupList{},
		&metav1.APIGroup{},
		&metav1.APIResourceList{},
	)
}
