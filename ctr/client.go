package ctr

import (
	"context"

	containerd "github.com/containerd/containerd/v2/client"
	"github.com/containerd/containerd/api/services/tasks/v1"
	tasktypes "github.com/containerd/containerd/api/types/task"
	"github.com/containerd/containerd/v2/core/containers"
	"github.com/containerd/containerd/v2/core/events"
	"github.com/containerd/containerd/v2/core/images"
	"github.com/containerd/containerd/v2/core/snapshots"
	"github.com/containerd/containerd/v2/pkg/namespaces"
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

func (c *Client) Images(ctx context.Context, ns string) ([]images.Image, error) {
	ctx = namespaces.WithNamespace(ctx, ns)
	return c.inner.ImageService().List(ctx)
}

func (c *Client) Containers(ctx context.Context, ns string) ([]containers.Container, error) {
	ctx = namespaces.WithNamespace(ctx, ns)
	return c.inner.ContainerService().List(ctx)
}

func (c *Client) Tasks(ctx context.Context, ns string) ([]*tasktypes.Process, error) {
	ctx = namespaces.WithNamespace(ctx, ns)
	resp, err := c.inner.TaskService().List(ctx, &tasks.ListTasksRequest{})
	if err != nil {
		return nil, err
	}
	return resp.Tasks, nil
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
