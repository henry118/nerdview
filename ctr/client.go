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
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/containerd/containerd/api/services/tasks/v1"
	tasktypes "github.com/containerd/containerd/api/types/task"
	containerd "github.com/containerd/containerd/v2/client"
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

const defaultStateDir = "/run/containerd"

// Client wraps a containerd gRPC client with convenience methods for
// fetching resources and subscribing to events.
type Client struct {
	inner   *containerd.Client
	address string
	eventCh <-chan *events.Envelope
	errCh   <-chan error
}

// New connects to the containerd daemon at the given socket address.
func New(address string) (*Client, error) {
	c, err := containerd.New(address, containerd.WithTimeout(connectTimeout))
	if err != nil {
		return nil, err
	}
	return &Client{inner: c, address: address}, nil
}

// Reconnect re-establishes the gRPC connection to the containerd daemon.
func (c *Client) Reconnect() error {
	return c.inner.Reconnect()
}

const (
	// Timeout for connection attempts and health checks.
	connectTimeout     = 2 * time.Second
	reconnectBaseDelay = 1 * time.Second
	reconnectMaxDelay  = 10 * time.Second
)

func (c *Client) reconnectAndSubscribe(ctx context.Context) error {
	delay := reconnectBaseDelay
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(delay):
		}
		if err := c.inner.Reconnect(); err != nil {
			delay = min(delay*2, reconnectMaxDelay)
			logging.Debug("reconnect failed, retrying in %s: %v", delay, err)
			continue
		}
		c.StartEventStream(ctx)
		logging.Info("reconnected to containerd")
		return nil
	}
}

// StateDir returns the containerd state directory.
// Derived from the socket path; falls back to the default if the address is not a unix path.
func (c *Client) StateDir() string {
	if strings.HasPrefix(c.address, "/") {
		dir := filepath.Dir(c.address)
		logging.Debug("state dir derived from socket: %s", dir)
		return dir
	}
	logging.Debug("state dir using default: %s", defaultStateDir)
	return defaultStateDir
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
	Spec      *specs.Spec
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
		if sandboxIDs[ctr.ID] || ctr.SandboxID == ctr.ID {
			info.IsSandbox = true
		}
		container, err := c.inner.LoadContainer(ctx, ctr.ID)
		if err == nil {
			spec, err := container.Spec(ctx)
			if err == nil {
				info.Spec = spec
			}
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
	Name     string
	Desc     ocispec.Descriptor
	Children []ImageTree
	// Chain ID referencing the snapshot tree root (from content labels).
	SnapshotKey string
	// Content store labels for this blob.
	Labels map[string]string
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
		if info, err := store.Info(ctx, img.Target.Digest); err == nil {
			tree.Labels = info.Labels
		}
		tree.Children = walkContent(ctx, store, snLabel, img.Target)
		trees = append(trees, tree)
	}
	return trees, nil
}

func isKnownDescriptor(desc ocispec.Descriptor) bool {
	if desc.Platform != nil && desc.Platform.OS == "unknown" {
		return false
	}
	mt := desc.MediaType
	if images.IsManifestType(mt) || images.IsIndexType(mt) || images.IsLayerType(mt) || images.IsConfigType(mt) {
		return true
	}
	// OCI index entries may omit MediaType; accept if they have a valid platform
	if mt == "" && desc.Platform != nil {
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
		if (images.IsManifestType(child.MediaType) || images.IsIndexType(child.MediaType)) && len(node.Children) == 0 {
			continue
		}
		// Read content info labels for snapshot cross-reference.
		info, err := store.Info(ctx, child.Digest)
		if err == nil {
			node.Labels = info.Labels
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

// TaskInfo pairs a task process with its bundle path and runtime-derived metadata.
type TaskInfo struct {
	ContainerID string
	Process     *tasktypes.Process
	ExecID      string
	BundlePath  string
	Cmdline     string
	StartedAt   string
	Root        string
	Cwd         string
	Cgroups     []string
	Namespaces  map[string]string
}

// Tasks returns all tasks in the namespace, each enriched with runtime-specific
// process metadata. Exec processes are included as separate entries.
func (c *Client) Tasks(ctx context.Context, ns string) ([]TaskInfo, error) {
	ctx = namespaces.WithNamespace(ctx, ns)
	resp, err := c.inner.TaskService().List(ctx, &tasks.ListTasksRequest{})
	if err != nil {
		logging.Error("failed to list tasks in ns=%s: %v", ns, err)
		return nil, err
	}
	logging.Debug("loaded %d tasks in ns=%s", len(resp.Tasks), ns)
	var result []TaskInfo
	for _, p := range resp.Tasks {
		runtimeName := ""
		container, err := c.inner.LoadContainer(ctx, p.ID)
		if err == nil {
			cInfo, err := container.Info(ctx)
			if err == nil {
				runtimeName = cInfo.Runtime.Name
			}
		}
		helper := newRuntimeHelper(runtimeName, c.inner.TaskService())

		detail := helper.Inspect(p.Pid)
		info := TaskInfo{
			ContainerID: p.ID,
			Process:     p,
			BundlePath:  helper.BundlePath(c.StateDir(), ns, p.ID),
			Cmdline:     detail.Cmdline,
			StartedAt:   detail.StartedAt,
			Root:        detail.Root,
			Cwd:         detail.Cwd,
			Cgroups:     detail.Cgroups,
			Namespaces:  detail.Namespaces,
		}
		result = append(result, info)

		for _, proc := range helper.Processes(ctx, p.ID) {
			if !proc.IsExec {
				continue
			}
			execDetail := helper.Inspect(proc.Pid)
			result = append(result, TaskInfo{
				ContainerID: p.ID,
				Process:     &tasktypes.Process{ID: p.ID, Pid: proc.Pid, Status: p.Status},
				ExecID:      proc.ID,
				Cmdline:     execDetail.Cmdline,
				StartedAt:   execDetail.StartedAt,
				Root:        execDetail.Root,
				Cwd:         execDetail.Cwd,
				Cgroups:     execDetail.Cgroups,
				Namespaces:  execDetail.Namespaces,
			})
		}
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

// DaemonStats holds resource usage metrics for the containerd daemon process.
type DaemonStats struct {
	PID int
	// Current CPU usage percentage (like htop, per-core).
	CPUPct float64
	// Virtual memory size in bytes.
	VMS uint64
	// Resident set size in bytes.
	RSS uint64
	// Number of threads.
	Threads int
	// Time since process start.
	Uptime time.Duration
}

var prevCPUSample struct {
	ticks uint64
	time  time.Time
}

// DaemonPID returns the containerd daemon's PID via the introspection API.
func (c *Client) DaemonPID(ctx context.Context) (int, error) {
	ctx, cancel := context.WithTimeout(ctx, connectTimeout)
	defer cancel()
	resp, err := c.inner.IntrospectionService().Server(ctx)
	if err != nil {
		logging.Error("failed to get daemon PID via introspection: %v", err)
		return 0, err
	}
	logging.Debug("daemon PID=%d", resp.Pid)
	return int(resp.Pid), nil
}

// ReadDaemonStats reads CPU, memory, thread, and uptime info from /proc for the given PID.
func ReadDaemonStats(pid int) (DaemonStats, error) {
	stats := DaemonStats{PID: pid}

	fields, err := procStatFields(pid)
	if err != nil {
		return stats, err
	}
	if len(fields) > 19 {
		utime, _ := strconv.ParseUint(fields[11], 10, 64)
		stime, _ := strconv.ParseUint(fields[12], 10, 64)
		numThreads, _ := strconv.Atoi(fields[17])
		starttime, _ := strconv.ParseUint(fields[19], 10, 64)

		stats.Threads = numThreads

		clkTck := uint64(100)
		totalTicks := utime + stime
		now := time.Now()

		if !prevCPUSample.time.IsZero() {
			elapsed := now.Sub(prevCPUSample.time).Seconds()
			if elapsed > 0 {
				deltaTicks := totalTicks - prevCPUSample.ticks
				stats.CPUPct = (float64(deltaTicks) / float64(clkTck)) / elapsed * 100.0
			}
		}
		prevCPUSample.ticks = totalTicks
		prevCPUSample.time = now

		uptimeData, err := os.ReadFile("/proc/uptime")
		if err == nil {
			uptimeFields := strings.Fields(string(uptimeData))
			if len(uptimeFields) > 0 {
				systemUptime, _ := strconv.ParseFloat(uptimeFields[0], 64)
				procStartSec := float64(starttime) / float64(clkTck)
				procUptime := systemUptime - procStartSec
				if procUptime > 0 {
					stats.Uptime = time.Duration(procUptime * float64(time.Second))
				}
			}
		}
	}

	statusData, err := os.ReadFile(filepath.Join("/proc", strconv.Itoa(pid), "status"))
	if err != nil {
		return stats, err
	}
	for line := range strings.SplitSeq(string(statusData), "\n") {
		if strings.HasPrefix(line, "VmSize:") {
			stats.VMS = parseKBValue(line)
		} else if strings.HasPrefix(line, "VmRSS:") {
			stats.RSS = parseKBValue(line)
		}
	}

	return stats, nil
}

func parseKBValue(line string) uint64 {
	fields := strings.Fields(line)
	if len(fields) >= 2 {
		val, _ := strconv.ParseUint(fields[1], 10, 64)
		return val * 1024
	}
	return 0
}
