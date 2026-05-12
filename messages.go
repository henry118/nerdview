package main

import (
	"github.com/containerd/containerd/v2/core/containers"
	"github.com/containerd/containerd/v2/core/snapshots"
	"github.com/henry118/nerdtui/ctr"
)

type namespacesLoadedMsg struct {
	namespaces []string
}

type resourcesLoadedMsg struct {
	namespace  string
	images     []ctr.ImageTree
	containers []containers.Container
	tasks      []ctr.TaskInfo
	snapshots  []snapshots.Info
}

type snapshottersLoadedMsg struct {
	snapshotters []string
}

type errorMsg struct {
	err error
}

type tickMsg struct{}
