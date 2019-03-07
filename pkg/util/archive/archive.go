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

package archive

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
)

const (
	internalStateFile        = "state.yaml"
	internalEtcdSnapshotFile = "etcd.snapshot"
)

func Create(archivePath, statePath, etcdSnapshotPath string) error {
	tempDir, err := ioutil.TempDir(os.TempDir(), "cctl")
	if err != nil {
		return fmt.Errorf("unable to create temporary directory to create archive: %v", err)
	}
	defer os.RemoveAll(tempDir)
	tempArchivePath := filepath.Join(tempDir, "archive.tar")
	tempStatePath := filepath.Join(tempDir, internalStateFile)
	tempEtcdSnapshotPath := filepath.Join(tempDir, internalEtcdSnapshotFile)

	cmdCopyState := exec.Command("cp", statePath, tempStatePath)
	cmdCopyEtcdSnapshot := exec.Command("cp", etcdSnapshotPath, tempEtcdSnapshotPath)
	cmdAppendState := exec.Command("tar", "--file", tempArchivePath, "--directory", filepath.Dir(tempStatePath), "--append", filepath.Base(tempStatePath))
	cmdAppendEtcdSnapshot := exec.Command("tar", "--file", tempArchivePath, "--directory", filepath.Dir(tempEtcdSnapshotPath), "--append", filepath.Base(tempEtcdSnapshotPath))
	cmdCompressArchive := exec.Command("gzip", tempArchivePath)
	cmdMoveArchive := exec.Command("mv", fmt.Sprintf("%s.gz", tempArchivePath), archivePath)

	return runAllCommands([]*exec.Cmd{
		cmdCopyState,
		cmdCopyEtcdSnapshot,
		cmdAppendState,
		cmdAppendEtcdSnapshot,
		cmdCompressArchive,
		cmdMoveArchive,
	})
}

func Extract(archivePath, statePath, etcdSnapshotPath string) error {
	tempDir, err := ioutil.TempDir(os.TempDir(), "cctl")
	if err != nil {
		return fmt.Errorf("unable to create temporary directory to extract archive: %v", err)
	}
	defer os.RemoveAll(tempDir)
	tempStatePath := filepath.Join(tempDir, internalStateFile)
	tempEtcdSnapshotPath := filepath.Join(tempDir, internalEtcdSnapshotFile)

	cmdExtractFiles := exec.Command("tar", "--file", archivePath, "--directory", tempDir, "--extract", "--gzip")
	cmdMoveState := exec.Command("mv", tempStatePath, statePath)
	cmdMoveEtcdSnapshot := exec.Command("mv", tempEtcdSnapshotPath, etcdSnapshotPath)

	return runAllCommands([]*exec.Cmd{
		cmdExtractFiles,
		cmdMoveState,
		cmdMoveEtcdSnapshot,
	})
}

func runAllCommands(commands []*exec.Cmd) error {
	for _, c := range commands {
		err := c.Run()
		if err != nil {
			switch v := err.(type) {
			case *exec.Error:
				return fmt.Errorf("failed to run command %q: %s", v.Name, v.Err)
			case *exec.ExitError:
				return fmt.Errorf("command %q failed: %q", c.Path, v.Stderr)
			default:
				return err
			}
		}
	}
	return nil
}
