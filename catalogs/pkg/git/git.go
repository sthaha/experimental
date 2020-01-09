/*
Copyright 2020 The Tekton Authors

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
package git

import (
	"bytes"
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/go-logr/logr"
	homedir "github.com/mitchellh/go-homedir"
)

// FetchSpec describes how to initialize and fetch from a Git repository.
type FetchSpec struct {
	URL       string
	Revision  string
	Path      string
	Depth     uint
	SSLVerify bool
}

func (f *FetchSpec) sanitize() {
	f.URL = strings.TrimSpace(f.URL)
	f.Path = strings.TrimSpace(f.Path)
	f.Revision = strings.TrimSpace(f.Revision)
	if f.Revision == "" {
		f.Revision = "master"
	}
}

func (f *FetchSpec) clonePath() string {
	f.sanitize()
	u, _ := url.Parse(f.URL)
	return filepath.Join(f.Path, u.Host, u.Path+"@"+f.Revision)
}

func initRepo(log logr.Logger, spec FetchSpec) error {

	clonePath := spec.clonePath()

	if _, err := os.Stat(clonePath); err == nil {
		return nil
	}

	if _, err := run(log, "", "init", clonePath); err != nil {
		return err
	}

	if err := os.Chdir(clonePath); err != nil {
		return fmt.Errorf("failed to change directory with path %s; err: %w", spec.Path, err)
	}

	if _, err := run(log, "", "remote", "add", "origin", spec.URL); err != nil {
		return err
	}

	if _, err := run(log, "", "config", "http.sslVerify", strconv.FormatBool(spec.SSLVerify)); err != nil {
		log.Error(err, "failed to set http.sslVerify in git configs")
		return err
	}
	return nil
}

// Fetch fetches the specified git repository at the revision into path.
func Fetch(log logr.Logger, spec FetchSpec) error {
	spec.sanitize()

	if err := ensureHomeEnv(log); err != nil {
		return err
	}

	log.Info("clone to", "path", spec.clonePath())

	if err := initRepo(log, spec); err != nil {
		os.RemoveAll(spec.clonePath())
		return err
	}

	fetchArgs := []string{
		"fetch",
		"--recurse-submodules=yes",
		"--depth=1",
		"origin", spec.Revision,
	}

	if _, err := run(log, "", fetchArgs...); err != nil {
		// Fetch can fail if an old commitid was used so try git pull, performing regardless of error
		// as no guarantee that the same error is returned by all git servers gitlab, github etc...
		if _, err := run(log, "", "pull", "--recurse-submodules=yes", "origin"); err != nil {
			log.Info("Failed to pull origin", "err", err)
		}
		if _, err := run(log, "", "checkout", spec.Revision); err != nil {
			return err
		}
	} else if _, err := run(log, "", "reset", "--hard", "FETCH_HEAD"); err != nil {
		return err
	}
	log.Info("Successfully cloned", "url", spec.URL, "revision", spec.Revision, "path", spec.clonePath())
	return nil
}

func Commit(log logr.Logger, revision, path string) (string, error) {
	output, err := run(log, path, "rev-parse", "HEAD")
	if err != nil {
		return "", err
	}
	return strings.TrimSuffix(output, "\n"), nil
}

func SubmoduleFetch(log logr.Logger, path string) error {
	if err := ensureHomeEnv(log); err != nil {
		return err
	}

	if path != "" {
		if err := os.Chdir(path); err != nil {
			return fmt.Errorf("failed to change directory with path %s; err: %w", path, err)
		}
	}
	if _, err := run(log, "", "submodule", "init"); err != nil {
		return err
	}
	if _, err := run(log, "", "submodule", "update", "--recursive"); err != nil {
		return err
	}
	log.Info("Successfully initialized and updated submodules in path %s", path)
	return nil
}

func ensureHomeEnv(log logr.Logger) error {
	// HACK: This is to get git+ssh to work since ssh doesn't respect the HOME
	// env variable.
	homepath, err := homedir.Dir()
	if err != nil {
		log.Error(err, "Unexpected error: getting the user home directory")
		return err
	}
	homeenv := os.Getenv("HOME")
	euid := os.Geteuid()
	// Special case the root user/directory
	if euid == 0 {
		if err := os.Symlink(homeenv+"/.ssh", "/root/.ssh"); err != nil {
			// Only do a warning, in case we don't have a real home
			// directory writable in our image
			log.Error(err, "Unexpected error: creating symlink")
		}
	} else if homeenv != "" && homeenv != homepath {
		if _, err := os.Stat(homepath + "/.ssh"); os.IsNotExist(err) {
			if err := os.Symlink(homeenv+"/.ssh", homepath+"/.ssh"); err != nil {
				// Only do a warning, in case we don't have a real home
				// directory writable in our image
				log.Error(err, "Unexpected error: creating symlink: %v", err)
			}
		}
	}
	return nil
}

func run(log logr.Logger, dir string, args ...string) (string, error) {
	c := exec.Command("git", args...)
	var output bytes.Buffer
	c.Stderr = &output
	c.Stdout = &output
	// This is the optional working directory. If not set, it defaults to the current
	// working directory of the process.
	if dir != "" {
		c.Dir = dir
	}
	if err := c.Run(); err != nil {
		log.Error(err, "git error", "args", args, "output", output.String())
		return "", err
	}
	return output.String(), nil
}
