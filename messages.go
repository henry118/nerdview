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

package main

import (
	"github.com/containerd/containerd/v2/core/snapshots"
	"github.com/henry118/nerdview/ctr"
)

// namespacesLoadedMsg is sent when the namespace list is fetched from the daemon.
type namespacesLoadedMsg struct {
	namespaces []string
}

// resourcesLoadedMsg is sent when all resources for a namespace are loaded.
type resourcesLoadedMsg struct {
	namespace  string
	images     []ctr.ImageTree
	containers []ctr.ContainerInfo
	tasks      []ctr.TaskInfo
	snapshots  []snapshots.Info
}

// snapshottersLoadedMsg is sent when available snapshotters are fetched.
type snapshottersLoadedMsg struct {
	snapshotters []string
}

// daemonStatsMsg carries updated daemon resource usage metrics.
type daemonStatsMsg struct {
	stats ctr.DaemonStats
}

// errorMsg carries an error to display in the status bar.
type errorMsg struct {
	err error
}

// tickMsg triggers a periodic full data refresh.
type tickMsg struct{}

// statsTickMsg triggers a periodic daemon stats refresh.
type statsTickMsg struct{}

// debounceMsg fires after the debounce interval to trigger pending refreshes.
type debounceMsg struct {
	gen int // Generation counter to ignore stale timers.
}

// imagesRefreshedMsg is sent when images are reloaded due to a containerd event.
type imagesRefreshedMsg []ctr.ImageTree

// snapshotsRefreshedMsg is sent when snapshots are reloaded due to a containerd event.
type snapshotsRefreshedMsg []snapshots.Info

// containersRefreshedMsg is sent when containers are reloaded due to a containerd event.
type containersRefreshedMsg []ctr.ContainerInfo

// tasksRefreshedMsg is sent when tasks are reloaded due to a containerd event.
type tasksRefreshedMsg []ctr.TaskInfo
