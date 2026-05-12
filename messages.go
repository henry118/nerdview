package main

import (
	tasktypes "github.com/containerd/containerd/api/types/task"
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
	tasks      []*tasktypes.Process
	snapshots  []snapshots.Info
}

type snapshottersLoadedMsg struct {
	snapshotters []string
}

type errorMsg struct {
	err error
}

type tickMsg struct{}
