package resource

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/table"
	"github.com/containerd/containerd/v2/core/images"
)

var ImageKind = Kind{
	Name: "Images",
	Columns: []Column{
		{Title: "Name", MinWidth: 20, Flex: true},
		{Title: "Digest", MinWidth: 20},
		{Title: "Created", MinWidth: 20},
	},
	ToRows: func(data any) []table.Row {
		imgs, ok := data.([]images.Image)
		if !ok || len(imgs) == 0 {
			return nil
		}
		rows := make([]table.Row, len(imgs))
		for i, img := range imgs {
			digest := img.Target.Digest.String()
			if len(digest) > 19 {
				digest = digest[:19]
			}
			rows[i] = table.Row{
				img.Name,
				digest,
				img.CreatedAt.Format("2006-01-02 15:04:05"),
			}
		}
		return rows
	},
	ToDetail: func(data any, index int) (string, string) {
		imgs, ok := data.([]images.Image)
		if !ok || index < 0 || index >= len(imgs) {
			return "", ""
		}
		img := imgs[index]
		var b strings.Builder
		fmt.Fprintf(&b, "Name:       %s\n", img.Name)
		fmt.Fprintf(&b, "Digest:     %s\n", img.Target.Digest)
		fmt.Fprintf(&b, "MediaType:  %s\n", img.Target.MediaType)
		fmt.Fprintf(&b, "Size:       %d\n", img.Target.Size)
		fmt.Fprintf(&b, "Created:    %s\n", img.CreatedAt.Format("2006-01-02 15:04:05"))
		fmt.Fprintf(&b, "Updated:    %s\n", img.UpdatedAt.Format("2006-01-02 15:04:05"))
		if len(img.Labels) > 0 {
			fmt.Fprintf(&b, "Labels:\n")
			for k, v := range img.Labels {
				fmt.Fprintf(&b, "  %s: %s\n", k, v)
			}
		}
		return img.Name, b.String()
	},
}
