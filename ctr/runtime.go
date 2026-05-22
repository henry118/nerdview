// Copyright Henry Wang
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package ctr

import (
	"context"
	"strings"

	"github.com/containerd/containerd/api/services/tasks/v1"
)

// processEntry describes a single process (init or exec) within a container task.
type processEntry struct {
	ID     string
	Pid    uint32
	IsExec bool
}

// processDetail holds runtime-derived metadata for a task process.
type processDetail struct {
	Cmdline    string
	StartedAt  string
	Root       string
	Cwd        string
	Cgroups    []string
	Namespaces map[string]string
}

// runtimeHelper abstracts runtime-specific task inspection. Each container
// runtime (runc, kata, etc.) may store process state differently. For runc,
// process info is available via host /proc; for VM-based runtimes the host
// PID is a shim and workload details require runtime-specific retrieval.
type runtimeHelper interface {
	// Processes returns all exec processes for the given container by
	// querying the containerd task service. Init processes are excluded.
	Processes(ctx context.Context, containerID string) []processEntry

	// BundlePath returns the on-disk OCI bundle path for a container task,
	// or "" if the path cannot be determined for this runtime.
	BundlePath(stateDir, ns, containerID string) string

	// Inspect returns detailed process metadata (cmdline, root, cwd,
	// cgroups, namespaces, start time) for the given host PID.
	Inspect(pid uint32) processDetail
}

// newRuntimeHelper returns the appropriate runtimeHelper for the given runtime name.
func newRuntimeHelper(runtimeName string, taskService tasks.TasksClient) runtimeHelper {
	for prefix, factory := range runtimeFactories {
		if strings.HasPrefix(runtimeName, prefix) {
			return factory(taskService)
		}
	}
	return &fallbackHelper{}
}

// runtimeFactories maps runtime name prefixes to their helper constructors.
// Prefix matching handles versioned names (e.g. "io.containerd.runc.v2").
var runtimeFactories = map[string]func(tasks.TasksClient) runtimeHelper{
	"io.containerd.runc": func(ts tasks.TasksClient) runtimeHelper {
		return &runcHelper{taskService: ts}
	},
}

// fallbackHelper is used for unknown runtimes. All methods return empty values,
// allowing the UI to gracefully display only containerd API-provided fields.
type fallbackHelper struct{}

func (f *fallbackHelper) Processes(_ context.Context, _ string) []processEntry {
	return nil
}

func (f *fallbackHelper) BundlePath(_, _, _ string) string {
	return ""
}

func (f *fallbackHelper) Inspect(_ uint32) processDetail {
	return processDetail{}
}
