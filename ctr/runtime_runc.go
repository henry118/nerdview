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
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/containerd/containerd/api/services/tasks/v1"
	runcoptions "github.com/containerd/containerd/api/types/runc/options"
	"github.com/henry118/nerdview/logging"
	"google.golang.org/protobuf/proto"
)

type runcHelper struct {
	taskService tasks.TasksClient
}

func (r *runcHelper) Processes(ctx context.Context, containerID string) []processEntry {
	resp, err := r.taskService.ListPids(ctx, &tasks.ListPidsRequest{
		ContainerID: containerID,
	})
	if err != nil {
		return nil
	}
	var entries []processEntry
	for _, p := range resp.Processes {
		entry := processEntry{Pid: p.Pid}
		if p.Info != nil {
			var details runcoptions.ProcessDetails
			if err := proto.Unmarshal(p.Info.Value, &details); err == nil && details.ExecID != "" {
				entry.ID = details.ExecID
				entry.IsExec = true
			}
		}
		if entry.IsExec {
			entries = append(entries, entry)
		}
	}
	if len(entries) > 0 {
		logging.Debug("found %d exec(s) for container %s", len(entries), containerID)
	}
	return entries
}

func (r *runcHelper) BundlePath(stateDir, ns, containerID string) string {
	return filepath.Join(stateDir, "io.containerd.runtime.v2.task", ns, containerID)
}

func (r *runcHelper) Inspect(pid uint32) processDetail {
	if pid == 0 {
		return processDetail{}
	}
	return processDetail{
		Cmdline:    procCmdline(pid),
		StartedAt:  procStartTime(pid),
		Root:       procReadlink(pid, "root"),
		Cwd:        procReadlink(pid, "cwd"),
		Cgroups:    procCgroup(pid),
		Namespaces: procNamespaces(pid),
	}
}

func procReadlink(pid uint32, name string) string {
	target, err := os.Readlink(filepath.Join("/proc", fmt.Sprintf("%d", pid), name))
	if err != nil {
		return ""
	}
	return target
}

func procCmdline(pid uint32) string {
	data, err := os.ReadFile(filepath.Join("/proc", fmt.Sprintf("%d", pid), "cmdline"))
	if err != nil || len(data) == 0 {
		return ""
	}
	for i := range data {
		if data[i] == 0 {
			data[i] = ' '
		}
	}
	return strings.TrimSpace(string(data))
}

func procStartTime(pid uint32) string {
	fields, err := procStatFields(int(pid))
	if err != nil {
		return ""
	}
	if len(fields) <= 19 {
		return ""
	}
	starttime, err := strconv.ParseUint(fields[19], 10, 64)
	if err != nil {
		return ""
	}
	uptimeData, err := os.ReadFile("/proc/uptime")
	if err != nil {
		return ""
	}
	uptimeFields := strings.Fields(string(uptimeData))
	if len(uptimeFields) == 0 {
		return ""
	}
	systemUptime, err := strconv.ParseFloat(uptimeFields[0], 64)
	if err != nil {
		return ""
	}
	clkTck := uint64(100)
	procStartSec := float64(starttime) / float64(clkTck)
	bootTime := time.Now().Add(-time.Duration(systemUptime * float64(time.Second)))
	startedAt := bootTime.Add(time.Duration(procStartSec * float64(time.Second)))
	return startedAt.Format("2006-01-02 15:04:05")
}

func procStatFields(pid int) ([]string, error) {
	statData, err := os.ReadFile(filepath.Join("/proc", strconv.Itoa(pid), "stat"))
	if err != nil {
		return nil, err
	}
	s := string(statData)
	closeParenIdx := strings.LastIndex(s, ")")
	if closeParenIdx < 0 {
		return nil, fmt.Errorf("invalid /proc/stat format")
	}
	return strings.Fields(s[closeParenIdx+2:]), nil
}

func procCgroup(pid uint32) []string {
	data, err := os.ReadFile(filepath.Join("/proc", fmt.Sprintf("%d", pid), "cgroup"))
	if err != nil || len(data) == 0 {
		return nil
	}
	var lines []string
	for line := range strings.SplitSeq(strings.TrimSpace(string(data)), "\n") {
		if line != "" {
			lines = append(lines, line)
		}
	}
	return lines
}

func procNamespaces(pid uint32) map[string]string {
	dir := filepath.Join("/proc", fmt.Sprintf("%d", pid), "ns")
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil
	}
	ns := make(map[string]string)
	for _, entry := range entries {
		if strings.HasSuffix(entry.Name(), "_for_children") {
			continue
		}
		target, err := os.Readlink(filepath.Join(dir, entry.Name()))
		if err != nil {
			continue
		}
		ns[entry.Name()] = target
	}
	return ns
}
