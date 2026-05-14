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

// Package ctr wraps the containerd client SDK, providing methods to fetch
// images, containers, tasks, snapshots, and namespaces scoped by namespace.
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
	"github.com/henry118/nerdview/logging"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	specs "github.com/opencontainers/runtime-spec/specs-go"
)

// Client wraps a containerd gRPC client with convenience methods for
// fetching resources and subscribing to events.
type Client struct {
	inner   *containerd.Client
	eventCh <-chan *events.Envelope
	errCh   <-chan error
}

// New connects to the containerd daemon at the given socket address.
func New(address string) (*Client, error) {
	c, err := containerd.New(address)
	if err != nil {
		return nil, err
	}
	return &Client{inner: c}, nil
}

// Close closes the underlying gRPC connection.
func (c *Client) Close() error {
	return c.inner.Close()
}

// StartEventStream subscribes to containerd events across all namespaces.
func (c *Client) StartEventStream(ctx context.Context) {
	c.eventCh, c.errCh = c.inner.Subscribe(ctx)
}

// EventCh returns the channel receiving containerd event envelopes.
func (c *Client) EventCh() <-chan *events.Envelope {
	return c.eventCh
}

// ErrCh returns the channel receiving event subscription errors.
func (c *Client) ErrCh() <-chan error {
	return c.errCh
}

// Namespaces returns all namespace names from containerd.
func (c *Client) Namespaces(ctx context.Context) ([]string, error) {
	ns, err := c.inner.NamespaceService().List(ctx)
	if err != nil {
		logging.Error("failed to list namespaces: %v", err)
		return nil, err
	}
	logging.Debug("loaded %d namespaces", len(ns))
	return ns, nil
}

// ContainerInfo pairs a container with its sandbox classification.
type ContainerInfo struct {
	Container containers.Container
	IsSandbox bool
}

// Containers returns all containers in the namespace with sandbox detection.
func (c *Client) Containers(ctx context.Context, ns string) ([]ContainerInfo, error) {
	ctx = namespaces.WithNamespace(ctx, ns)
	ctrs, err := c.inner.ContainerService().List(ctx)
	if err != nil {
		logging.Error("failed to list containers in ns=%s: %v", ns, err)
		return nil, err
	}
	logging.Debug("loaded %d containers in ns=%s", len(ctrs), ns)

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

// Snapshots walks the named snapshotter and returns all snapshot metadata.
func (c *Client) Snapshots(ctx context.Context, ns string, snapshotter string) ([]snapshots.Info, error) {
	ctx = namespaces.WithNamespace(ctx, ns)
	sn := c.inner.SnapshotService(snapshotter)
	var result []snapshots.Info
	err := sn.Walk(ctx, func(_ context.Context, info snapshots.Info) error {
		result = append(result, info)
		return nil
	})
	if err != nil {
		logging.Error("failed to list snapshots in ns=%s snapshotter=%s: %v", ns, snapshotter, err)
		return nil, err
	}
	logging.Debug("loaded %d snapshots in ns=%s snapshotter=%s", len(result), ns, snapshotter)
	return result, nil
}

// ImageTree represents an image and its content hierarchy (manifests, configs, layers).
type ImageTree struct {
	Name        string
	Desc        ocispec.Descriptor
	Children    []ImageTree
	SnapshotKey string // Chain ID referencing the snapshot tree root (from content labels).
}

// ImageTrees returns all images in the namespace as trees, walking the content
// store to resolve manifests and layers. Unknown media types and unpulled
// manifests are filtered out. The snapshotter name is used to resolve
// snapshot cross-references from content labels.
func (c *Client) ImageTrees(ctx context.Context, ns, snapshotter string) ([]ImageTree, error) {
	ctx = namespaces.WithNamespace(ctx, ns)
	imgList, err := c.inner.ImageService().List(ctx)
	if err != nil {
		logging.Error("failed to list images in ns=%s: %v", ns, err)
		return nil, err
	}
	logging.Debug("loaded %d images in ns=%s", len(imgList), ns)
	store := c.inner.ContentStore()
	snLabel := "containerd.io/gc.ref.snapshot." + snapshotter
	var trees []ImageTree
	for _, img := range imgList {
		if !isKnownDescriptor(img.Target) {
			continue
		}
		tree := ImageTree{
			Name: img.Name,
			Desc: img.Target,
		}
		tree.Children = walkContent(ctx, store, snLabel, img.Target)
		trees = append(trees, tree)
	}
	return trees, nil
}

var knownMediaTypes = map[string]bool{
	"application/vnd.oci.image.index.v1+json":                  true,
	"application/vnd.oci.image.manifest.v1+json":               true,
	"application/vnd.oci.image.config.v1+json":                 true,
	"application/vnd.oci.image.layer.v1.tar":                   true,
	"application/vnd.oci.image.layer.v1.tar+gzip":              true,
	"application/vnd.oci.image.layer.v1.tar+zstd":              true,
	"application/vnd.oci.image.layer.nondistributable.v1.tar":  true,
	"application/vnd.oci.image.layer.nondistributable.v1.tar+gzip": true,
	"application/vnd.docker.distribution.manifest.v2+json":     true,
	"application/vnd.docker.distribution.manifest.list.v2+json": true,
	"application/vnd.docker.container.image.v1+json":           true,
	"application/vnd.docker.image.rootfs.diff.tar.gzip":        true,
}

func isKnownDescriptor(desc ocispec.Descriptor) bool {
	if desc.Platform != nil && desc.Platform.OS == "unknown" {
		return false
	}
	if knownMediaTypes[desc.MediaType] {
		return true
	}
	// OCI index entries may omit MediaType; accept if they have a valid platform
	if desc.MediaType == "" && desc.Platform != nil {
		return true
	}
	return false
}

func isManifestType(mediaType string) bool {
	switch mediaType {
	case "application/vnd.oci.image.index.v1+json",
		"application/vnd.oci.image.manifest.v1+json",
		"application/vnd.docker.distribution.manifest.v2+json",
		"application/vnd.docker.distribution.manifest.list.v2+json":
		return true
	}
	return false
}

func walkContent(ctx context.Context, store content.Store, snLabel string, desc ocispec.Descriptor) []ImageTree {
	children, err := images.Children(ctx, store, desc)
	if err != nil {
		return nil
	}
	var result []ImageTree
	for _, child := range children {
		if !isKnownDescriptor(child) {
			continue
		}
		node := ImageTree{
			Desc:     child,
			Children: walkContent(ctx, store, snLabel, child),
		}
		// Skip manifests that have no children (content not downloaded)
		if isManifestType(child.MediaType) && len(node.Children) == 0 {
			continue
		}
		// Read content info labels for snapshot cross-reference.
		info, err := store.Info(ctx, child.Digest)
		if err == nil {
			if key, ok := info.Labels[snLabel]; ok {
				node.SnapshotKey = key
			}
		}
		// Propagate snapshot key from children (config blob) up to manifest.
		if node.SnapshotKey == "" {
			for _, c := range node.Children {
				if c.SnapshotKey != "" {
					node.SnapshotKey = c.SnapshotKey
					break
				}
			}
		}
		result = append(result, node)
	}
	return result
}

// TaskInfo pairs a task process with its container's OCI runtime spec and bundle path.
type TaskInfo struct {
	Process    *tasktypes.Process
	Spec       *specs.Spec
	BundlePath string
}

// TasksWithSpec returns all tasks in the namespace, each enriched with the
// container's OCI runtime spec and computed bundle path.
func (c *Client) TasksWithSpec(ctx context.Context, ns string) ([]TaskInfo, error) {
	ctx = namespaces.WithNamespace(ctx, ns)
	resp, err := c.inner.TaskService().List(ctx, &tasks.ListTasksRequest{})
	if err != nil {
		logging.Error("failed to list tasks in ns=%s: %v", ns, err)
		return nil, err
	}
	logging.Debug("loaded %d tasks in ns=%s", len(resp.Tasks), ns)
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



// Snapshotters returns the names of all available snapshotter plugins.
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
