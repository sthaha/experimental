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
	"io/ioutil"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/go-logr/logr"
	homedir "github.com/mitchellh/go-homedir"
	catalog "github.com/tektoncd/experimental/catalogs/pkg/api/v1alpha1"
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

func initRepo(log logr.Logger, spec FetchSpec) (Repo, error) {
	log = log.WithName("init").WithValues("url", spec.URL)

	clonePath := spec.clonePath()
	repo := Repo{Path: clonePath, Log: log}

	// if already cloned, cd to the cloned path
	if _, err := os.Stat(clonePath); err == nil {
		if err := os.Chdir(clonePath); err != nil {
			return Repo{}, fmt.Errorf("failed to change directory with path %s; err: %w", clonePath, err)
		}
		return repo, nil
	}

	if _, err := git(log, "", "init", clonePath); err != nil {
		return Repo{}, err
	}

	if err := os.Chdir(clonePath); err != nil {
		return Repo{}, fmt.Errorf("failed to change directory with path %s; err: %w", spec.Path, err)
	}

	if _, err := git(log, "", "remote", "add", "origin", spec.URL); err != nil {
		return Repo{}, err
	}

	if _, err := git(log, "", "config", "http.sslVerify", strconv.FormatBool(spec.SSLVerify)); err != nil {
		log.Error(err, "failed to set http.sslVerify in git configs")
		return Repo{}, err
	}
	return repo, nil
}

// Fetch fetches the specified git repository at the revision into path.
func Fetch(log logr.Logger, spec FetchSpec) (Repo, error) {
	spec.sanitize()

	if err := ensureHomeEnv(log); err != nil {
		return Repo{}, err
	}

	log.Info("clone to", "path", spec.clonePath())

	repo, err := initRepo(log, spec)

	if err != nil {
		os.RemoveAll(spec.clonePath())
		return Repo{}, err
	}

	fetchArgs := []string{
		"fetch",
		"--recurse-submodules=yes",
		"--depth=1",
		"origin", spec.Revision,
	}

	if _, err := git(log, "", fetchArgs...); err != nil {
		// Fetch can fail if an old commitid was used so try git pull, performing regardless of error
		// as no guarantee that the same error is returned by all git servers gitlab, github etc...
		if _, err := git(log, "", "pull", "--recurse-submodules=yes", "origin"); err != nil {
			log.Info("Failed to pull origin", "err", err)
		}
		if _, err := git(log, "", "checkout", spec.Revision); err != nil {
			return Repo{}, err
		}
	} else if _, err := git(log, "", "reset", "--hard", "FETCH_HEAD"); err != nil {
		return Repo{}, err
	}
	log.Info("successfully cloned", "url", spec.URL, "revision", spec.Revision, "path", repo.Path)

	return repo, nil
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

func git(log logr.Logger, dir string, args ...string) (string, error) {
	output, err := rawGit(dir, args...)

	if err != nil {
		log.Error(err, "git error", "args", args, "output", output)
		return "", err
	}
	return output, nil
}

func rawGit(dir string, args ...string) (string, error) {
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
		return "", err
	}
	return output.String(), nil
}

type Repo struct {
	Path string
	head string
	Log  logr.Logger
}

func (r Repo) Head() string {
	if r.head == "" {
		head, _ := rawGit("", "rev-parse", "HEAD")
		r.head = head
	}
	return r.head
}

func (r Repo) Tasks() ([]catalog.TaskInfo, error) {

	tasksPath := filepath.Join(r.Path, "tasks")
	tasks, err := ioutil.ReadDir(tasksPath)
	if err != nil {
		return nil, err
	}

	ret := []catalog.TaskInfo{}
	for _, t := range tasks {
		ret = append(ret, *taskInfo(r.Log, tasksPath, t))
	}

	return ret, nil
}

func taskInfo(log logr.Logger, tasksPath string, f os.FileInfo) *catalog.TaskInfo {
	if !f.IsDir() {
		return nil
	}

	log.Info(" ... checking ", "filename", f.Name())
	// path/<task>/<version>/*
	pattern := filepath.Join(tasksPath, f.Name(), "*", f.Name()+".yaml")

	matches, err := filepath.Glob(pattern)
	if err != nil {
		log.Error(err, "pattern  error")
		return nil
	}

	ti := &catalog.TaskInfo{Name: f.Name(), Versions: []string{}}
	for _, m := range matches {
		// TODO: validate if the task is valid by unmarshalling
		log.Info("      found:", "file", m)
		dir, _ := filepath.Split(m)
		ti.Versions = append(ti.Versions, filepath.Base(dir))
	}

	if len(ti.Versions) == 0 {
		return nil
	}
	return ti
}
