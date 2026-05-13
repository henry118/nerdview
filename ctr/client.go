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

	containerd "github.com/containerd/containerd/v2/client"
	"github.com/containerd/containerd/api/services/tasks/v1"
	tasktypes "github.com/containerd/containerd/api/types/task"
	"github.com/containerd/containerd/v2/core/containers"
	"github.com/containerd/containerd/v2/core/content"
	"github.com/containerd/containerd/v2/core/events"
	"github.com/containerd/containerd/v2/core/images"
	"github.com/containerd/containerd/v2/core/snapshots"
	"github.com/containerd/containerd/v2/pkg/namespaces"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	specs "github.com/opencontainers/runtime-spec/specs-go"
)

type Client struct {
	inner   *containerd.Client
	eventCh <-chan *events.Envelope
	errCh   <-chan error
}

func New(address string) (*Client, error) {
	c, err := containerd.New(address)
	if err != nil {
		return nil, err
	}
	return &Client{inner: c}, nil
}

func (c *Client) Close() error {
	return c.inner.Close()
}

func (c *Client) StartEventStream(ctx context.Context) {
	c.eventCh, c.errCh = c.inner.Subscribe(ctx)
}

func (c *Client) EventCh() <-chan *events.Envelope {
	return c.eventCh
}

func (c *Client) ErrCh() <-chan error {
	return c.errCh
}

func (c *Client) Namespaces(ctx context.Context) ([]string, error) {
	return c.inner.NamespaceService().List(ctx)
}

type ContainerInfo struct {
	Container containers.Container
	IsSandbox bool
}

func (c *Client) Containers(ctx context.Context, ns string) ([]ContainerInfo, error) {
	ctx = namespaces.WithNamespace(ctx, ns)
	ctrs, err := c.inner.ContainerService().List(ctx)
	if err != nil {
		return nil, err
	}

	// Get sandbox IDs from sandbox store
	sandboxIDs := make(map[string]bool)
	sandboxes, err := c.inner.SandboxStore().List(ctx)
	if err == nil {
		for _, sb := range sandboxes {
			sandboxIDs[sb.ID] = true
		}
	}

	var result []ContainerInfo
	for _, ctr := range ctrs {
		info := ContainerInfo{Container: ctr}
		// A container is a sandbox if it's registered in the sandbox store,
		// or if its SandboxID equals its own ID
		if sandboxIDs[ctr.ID] || ctr.SandboxID == ctr.ID {
			info.IsSandbox = true
		}
		result = append(result, info)
	}
	return result, nil
}

func (c *Client) Snapshots(ctx context.Context, ns string, snapshotter string) ([]snapshots.Info, error) {
	ctx = namespaces.WithNamespace(ctx, ns)
	sn := c.inner.SnapshotService(snapshotter)
	var result []snapshots.Info
	err := sn.Walk(ctx, func(_ context.Context, info snapshots.Info) error {
		result = append(result, info)
		return nil
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}

type ImageTree struct {
	Name     string
	Desc     ocispec.Descriptor
	Children []ImageTree
}

func (c *Client) ImageTrees(ctx context.Context, ns string) ([]ImageTree, error) {
	ctx = namespaces.WithNamespace(ctx, ns)
	imgList, err := c.inner.ImageService().List(ctx)
	if err != nil {
		return nil, err
	}
	store := c.inner.ContentStore()
	var trees []ImageTree
	for _, img := range imgList {
		tree := ImageTree{
			Name: img.Name,
			Desc: img.Target,
		}
		tree.Children = walkContent(ctx, store, img.Target)
		trees = append(trees, tree)
	}
	return trees, nil
}

func walkContent(ctx context.Context, store content.Store, desc ocispec.Descriptor) []ImageTree {
	children, err := images.Children(ctx, store, desc)
	if err != nil {
		return nil
	}
	var result []ImageTree
	for _, child := range children {
		node := ImageTree{
			Desc:     child,
			Children: walkContent(ctx, store, child),
		}
		result = append(result, node)
	}
	return result
}

type TaskInfo struct {
	Process    *tasktypes.Process
	Spec       *specs.Spec
	BundlePath string
}

func (c *Client) TasksWithSpec(ctx context.Context, ns string) ([]TaskInfo, error) {
	ctx = namespaces.WithNamespace(ctx, ns)
	resp, err := c.inner.TaskService().List(ctx, &tasks.ListTasksRequest{})
	if err != nil {
		return nil, err
	}
	actualNS := namespaces.Default
	if n, ok := namespaces.Namespace(ctx); ok {
		actualNS = n
	}
	var result []TaskInfo
	for _, p := range resp.Tasks {
		info := TaskInfo{Process: p}
		container, err := c.inner.LoadContainer(ctx, p.ID)
		if err == nil {
			cInfo, err := container.Info(ctx)
			if err == nil {
				info.BundlePath = "/run/containerd/" + cInfo.Runtime.Name + "/" + actualNS + "/" + p.ID
			}
			spec, err := container.Spec(ctx)
			if err == nil {
				info.Spec = spec
			}
		}
		result = append(result, info)
	}
	return result, nil
}



func (c *Client) Snapshotters(ctx context.Context) ([]string, error) {
	resp, err := c.inner.IntrospectionService().Plugins(ctx, "type==io.containerd.snapshotter.v1")
	if err != nil {
		return nil, err
	}
	var names []string
	for _, p := range resp.Plugins {
		names = append(names, p.ID)
	}
	return names, nil
}
