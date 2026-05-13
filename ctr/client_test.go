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

func TestKnownMediaTypes(t *testing.T) {
	known := []string{
		"application/vnd.oci.image.index.v1+json",
		"application/vnd.oci.image.manifest.v1+json",
		"application/vnd.oci.image.config.v1+json",
		"application/vnd.oci.image.layer.v1.tar+gzip",
		"application/vnd.oci.image.layer.v1.tar+zstd",
		"application/vnd.oci.image.layer.v1.tar",
		"application/vnd.docker.distribution.manifest.v2+json",
		"application/vnd.docker.distribution.manifest.list.v2+json",
		"application/vnd.docker.container.image.v1+json",
		"application/vnd.docker.image.rootfs.diff.tar.gzip",
	}
	for _, mt := range known {
		if !knownMediaTypes[mt] {
			t.Errorf("expected %q to be a known media type", mt)
		}
	}

	unknown := []string{
		"unknown/unknown",
		"application/vnd.oci.artifact.manifest.v1+json",
		"application/octet-stream",
	}
	for _, mt := range unknown {
		if knownMediaTypes[mt] {
			t.Errorf("expected %q to NOT be a known media type", mt)
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
