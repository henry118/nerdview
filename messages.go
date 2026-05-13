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
	"github.com/henry118/nerdtui/ctr"
)

type namespacesLoadedMsg struct {
	namespaces []string
}

type resourcesLoadedMsg struct {
	namespace  string
	images     []ctr.ImageTree
	containers []ctr.ContainerInfo
	tasks      []ctr.TaskInfo
	snapshots  []snapshots.Info
}

type snapshottersLoadedMsg struct {
	snapshotters []string
}

type daemonStatsMsg struct {
	stats ctr.DaemonStats
}

type errorMsg struct {
	err error
}

type tickMsg struct{}
type statsTickMsg struct{}
