/*
Copyright 2019 The cctl authors.

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

package ssh

import (
	"fmt"
	"io/ioutil"

	"golang.org/x/crypto/ssh"
)

func PublicKeyFromFile(file string) (ssh.PublicKey, error) {
	data, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, err
	}
	key, _, _, _, err := ssh.ParseAuthorizedKey(data)
	if err != nil {
		return nil, fmt.Errorf("error reading public key file %s: %v", file, err)
	}
	return key, nil
}
