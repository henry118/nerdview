package main

import (
	tasktypes "github.com/containerd/containerd/api/types/task"
	"github.com/containerd/containerd/v2/core/containers"
	"github.com/containerd/containerd/v2/core/images"
)

type namespacesLoadedMsg struct {
	namespaces []string
}

type resourcesLoadedMsg struct {
	namespace  string
	images     []images.Image
	containers []containers.Container
	tasks      []*tasktypes.Process
}

type errorMsg struct {
	err error
}

type tickMsg struct{}
