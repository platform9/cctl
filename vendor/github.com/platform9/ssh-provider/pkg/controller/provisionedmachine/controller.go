/*
Copyright 2018 Platform 9 Systems, Inc.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package provisionedmachine

import (
	log "github.com/platform9/ssh-provider/pkg/logrus"

	"github.com/kubernetes-incubator/apiserver-builder/pkg/builders"

	"github.com/platform9/ssh-provider/pkg/apis/sshprovider/v1alpha1"
	listers "github.com/platform9/ssh-provider/pkg/client/listers_generated/sshprovider/v1alpha1"
	"github.com/platform9/ssh-provider/pkg/controller/sharedinformers"
)

// +controller:group=sshprovider,version=v1alpha1,kind=ProvisionedMachine,resource=provisionedmachines
type ProvisionedMachineControllerImpl struct {
	builders.DefaultControllerFns

	// lister indexes properties about ProvisionedMachine
	lister listers.ProvisionedMachineLister
}

// Init initializes the controller and is called by the generated code
// Register watches for additional resource types here.
func (c *ProvisionedMachineControllerImpl) Init(arguments sharedinformers.ControllerInitArguments) {
	// Use the lister for indexing provisionedmachines labels
	c.lister = arguments.GetSharedInformers().Factory.Sshprovider().V1alpha1().ProvisionedMachines().Lister()
}

// Reconcile handles enqueued messages
func (c *ProvisionedMachineControllerImpl) Reconcile(u *v1alpha1.ProvisionedMachine) error {
	// Implement controller logic here
	log.Printf("Running reconcile ProvisionedMachine for %s\n", u.Name)
	return nil
}

func (c *ProvisionedMachineControllerImpl) Get(namespace, name string) (*v1alpha1.ProvisionedMachine, error) {
	return c.lister.ProvisionedMachines(namespace).Get(name)
}
