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
	"testing"

	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
)

func TestStateDir(t *testing.T) {
	tests := []struct {
		address string
		want    string
	}{
		{"/run/containerd/containerd.sock", "/run/containerd"},
		{"/custom/path/containerd.sock", "/custom/path"},
		{"/var/run/containerd/containerd.sock", "/var/run/containerd"},
		{"localhost:1234", defaultStateDir},
		{"", defaultStateDir},
	}
	for _, tt := range tests {
		c := &Client{address: tt.address}
		if got := c.StateDir(); got != tt.want {
			t.Errorf("StateDir(%q) = %q, want %q", tt.address, got, tt.want)
		}
	}
}

func TestIsKnownDescriptor(t *testing.T) {
	tests := []struct {
		name string
		desc ocispec.Descriptor
		want bool
	}{
		{
			name: "known media type",
			desc: ocispec.Descriptor{MediaType: "application/vnd.oci.image.manifest.v1+json"},
			want: true,
		},
		{
			name: "unknown media type",
			desc: ocispec.Descriptor{MediaType: "unknown/unknown"},
			want: false,
		},
		{
			name: "empty media type without platform",
			desc: ocispec.Descriptor{MediaType: ""},
			want: false,
		},
		{
			name: "empty media type with platform",
			desc: ocispec.Descriptor{
				MediaType: "",
				Platform:  &ocispec.Platform{OS: "linux", Architecture: "amd64"},
			},
			want: true,
		},
		{
			name: "known media type with unknown platform (attestation)",
			desc: ocispec.Descriptor{
				MediaType: "application/vnd.oci.image.manifest.v1+json",
				Platform:  &ocispec.Platform{OS: "unknown", Architecture: "unknown"},
			},
			want: false,
		},
		{
			name: "empty media type with unknown platform",
			desc: ocispec.Descriptor{
				MediaType: "",
				Platform:  &ocispec.Platform{OS: "unknown", Architecture: "unknown"},
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isKnownDescriptor(tt.desc)
			if got != tt.want {
				t.Errorf("isKnownDescriptor(%v) = %v, want %v", tt.desc.MediaType, got, tt.want)
			}
		})
	}
}
