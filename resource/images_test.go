package resource

import (
	"testing"

	"github.com/henry118/nerdtui/ctr"
	digest "github.com/opencontainers/go-digest"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
)

func testImageTrees() []ctr.ImageTree {
	return []ctr.ImageTree{
		{
			Name: "docker.io/library/nginx:latest",
			Desc: ocispec.Descriptor{
				MediaType: "application/vnd.oci.image.index.v1+json",
				Digest:    digest.FromString("nginx-index"),
				Size:      1024,
			},
			Children: []ctr.ImageTree{
				{
					Desc: ocispec.Descriptor{
						MediaType: "application/vnd.oci.image.manifest.v1+json",
						Digest:    digest.FromString("nginx-linux-amd64"),
						Size:      512,
						Platform:  &ocispec.Platform{OS: "linux", Architecture: "amd64"},
					},
					Children: []ctr.ImageTree{
						{
							Desc: ocispec.Descriptor{
								MediaType: "application/vnd.oci.image.config.v1+json",
								Digest:    digest.FromString("nginx-config"),
								Size:      2048,
							},
						},
						{
							Desc: ocispec.Descriptor{
								MediaType: "application/vnd.oci.image.layer.v1.tar+gzip",
								Digest:    digest.FromString("nginx-layer1"),
								Size:      30 * 1024 * 1024,
							},
						},
					},
				},
				{
					Desc: ocispec.Descriptor{
						MediaType: "application/vnd.oci.image.manifest.v1+json",
						Digest:    digest.FromString("nginx-linux-arm64"),
						Size:      480,
						Platform:  &ocispec.Platform{OS: "linux", Architecture: "arm64"},
					},
				},
			},
		},
		{
			Name: "docker.io/library/alpine:3.19",
			Desc: ocispec.Descriptor{
				MediaType: "application/vnd.oci.image.manifest.v1+json",
				Digest:    digest.FromString("alpine-manifest"),
				Size:      256,
			},
		},
	}
}

func TestImageKindToRows_Unfolded(t *testing.T) {
	data := testImageTrees()
	rows := ImageKind.ToRows(data, nil)

	// nginx (1) + linux/amd64 (1) + config (1) + layer (1) + linux/arm64 (1) + alpine (1) = 6
	if len(rows) != 6 {
		t.Fatalf("Expected 6 rows unfolded, got %d", len(rows))
	}

	// First row should be the image name with fold icon
	if rows[0][0] != "▾ docker.io/library/nginx:latest" {
		t.Errorf("First row name = %q, want %q", rows[0][0], "▾ docker.io/library/nginx:latest")
	}

	// Second row should be a child with tree connector
	if rows[1][0] != "├─ ▾ linux/amd64" {
		t.Errorf("Second row name = %q, want %q", rows[1][0], "├─ ▾ linux/amd64")
	}

	// Alpine has no children, no fold icon
	if rows[5][0] != "docker.io/library/alpine:3.19" {
		t.Errorf("Last row name = %q, want %q", rows[5][0], "docker.io/library/alpine:3.19")
	}
}

func TestImageKindToRows_Folded(t *testing.T) {
	data := testImageTrees()
	nginxDigest := data[0].Desc.Digest.String()
	folded := map[string]bool{nginxDigest: true}

	rows := ImageKind.ToRows(data, folded)

	// nginx folded (1) + alpine (1) = 2
	if len(rows) != 2 {
		t.Fatalf("Expected 2 rows with nginx folded, got %d", len(rows))
	}

	if rows[0][0] != "▸ docker.io/library/nginx:latest" {
		t.Errorf("Folded row = %q, want %q", rows[0][0], "▸ docker.io/library/nginx:latest")
	}
}

func TestImageKindInitFolded(t *testing.T) {
	data := testImageTrees()
	folded := ImageKind.InitFolded(data)

	nginxDigest := data[0].Desc.Digest.String()
	if !folded[nginxDigest] {
		t.Error("nginx index should be folded (has children)")
	}

	alpineDigest := data[1].Desc.Digest.String()
	if folded[alpineDigest] {
		t.Error("alpine should not be folded (no children)")
	}
}

func TestImageKindRowID(t *testing.T) {
	data := testImageTrees()
	folded := map[string]bool{}

	// Index 0 is nginx (has children) — should return its digest
	id := ImageKind.RowID(data, folded, 0)
	if id == "" {
		t.Error("RowID for nginx (index 0) should be non-empty")
	}

	// Find alpine's index (last row = 5 when unfolded)
	id = ImageKind.RowID(data, folded, 5)
	if id != "" {
		t.Errorf("RowID for alpine (no children) should be empty, got %q", id)
	}
}

func TestShortMediaType(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"application/vnd.oci.image.index.v1+json", "index"},
		{"application/vnd.oci.image.manifest.v1+json", "manifest"},
		{"application/vnd.oci.image.config.v1+json", "config"},
		{"application/vnd.oci.image.layer.v1.tar+gzip", "layer/gzip"},
		{"application/vnd.oci.image.layer.v1.tar+zstd", "layer/zstd"},
		{"application/vnd.docker.distribution.manifest.v2+json", "manifest"},
		{"application/vnd.docker.image.rootfs.diff.tar.gzip", "layer/gzip"},
		{"something.unknown", "unknown"},
	}
	for _, tt := range tests {
		got := shortMediaType(tt.input)
		if got != tt.want {
			t.Errorf("shortMediaType(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestFormatSize(t *testing.T) {
	tests := []struct {
		input int64
		want  string
	}{
		{0, "0B"},
		{512, "512B"},
		{1024, "1.0K"},
		{1536, "1.5K"},
		{1048576, "1.0M"},
		{1073741824, "1.0G"},
	}
	for _, tt := range tests {
		got := formatSize(tt.input)
		if got != tt.want {
			t.Errorf("formatSize(%d) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestDescLabel(t *testing.T) {
	node := ctr.ImageTree{
		Desc: ocispec.Descriptor{
			Platform: &ocispec.Platform{OS: "linux", Architecture: "amd64"},
		},
	}
	if got := descLabel(node); got != "linux/amd64" {
		t.Errorf("descLabel with platform = %q, want %q", got, "linux/amd64")
	}

	nodeWithVariant := ctr.ImageTree{
		Desc: ocispec.Descriptor{
			Platform: &ocispec.Platform{OS: "linux", Architecture: "arm", Variant: "v7"},
		},
	}
	if got := descLabel(nodeWithVariant); got != "linux/arm/v7" {
		t.Errorf("descLabel with variant = %q, want %q", got, "linux/arm/v7")
	}

	nodeNoPlatform := ctr.ImageTree{
		Desc: ocispec.Descriptor{
			MediaType: "application/vnd.oci.image.config.v1+json",
		},
	}
	if got := descLabel(nodeNoPlatform); got != "config" {
		t.Errorf("descLabel no platform = %q, want %q", got, "config")
	}
}
