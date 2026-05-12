package resource

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/table"
	"github.com/containerd/containerd/v2/core/containers"
)


var ContainerKind = Kind{
	Name: "Containers",
	Columns: []Column{
		{Title: "ID", MinWidth: 12, Flex: true},
		{Title: "Image", MinWidth: 20, Flex: true},
		{Title: "Runtime", MinWidth: 16},
		{Title: "Created", MinWidth: 20},
	},
	ToRows: func(data any) []table.Row {
		ctrs, ok := data.([]containers.Container)
		if !ok || len(ctrs) == 0 {
			return nil
		}
		rows := make([]table.Row, len(ctrs))
		for i, c := range ctrs {
			id := c.ID
			if len(id) > 19 {
				id = id[:19]
			}
			rows[i] = table.Row{
				id,
				c.Image,
				c.Runtime.Name,
				c.CreatedAt.Format("2006-01-02 15:04:05"),
			}
		}
		return rows
	},
	ToDetail: func(data any, index int) (string, string) {
		ctrs, ok := data.([]containers.Container)
		if !ok || index < 0 || index >= len(ctrs) {
			return "", ""
		}
		c := ctrs[index]
		var b strings.Builder
		fmt.Fprintf(&b, "ID:          %s\n", c.ID)
		fmt.Fprintf(&b, "Image:       %s\n", c.Image)
		fmt.Fprintf(&b, "Runtime:     %s\n", c.Runtime.Name)
		fmt.Fprintf(&b, "Snapshotter: %s\n", c.Snapshotter)
		fmt.Fprintf(&b, "SnapshotKey: %s\n", c.SnapshotKey)
		fmt.Fprintf(&b, "Created:     %s\n", c.CreatedAt.Format("2006-01-02 15:04:05"))
		fmt.Fprintf(&b, "Updated:     %s\n", c.UpdatedAt.Format("2006-01-02 15:04:05"))
		if c.SandboxID != "" {
			fmt.Fprintf(&b, "SandboxID:   %s\n", c.SandboxID)
		}
		if len(c.Labels) > 0 {
			fmt.Fprintf(&b, "Labels:\n")
			for k, v := range c.Labels {
				fmt.Fprintf(&b, "  %s: %s\n", k, v)
			}
		}
		return c.ID, b.String()
	},
}
