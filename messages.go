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
