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
	"os"
	"testing"
)

func TestNewRuntimeHelper_Runc(t *testing.T) {
	tests := []string{
		"io.containerd.runc.v2",
		"io.containerd.runc.v1",
		"io.containerd.runc",
	}
	for _, name := range tests {
		helper := newRuntimeHelper(name, nil)
		if _, ok := helper.(*runcHelper); !ok {
			t.Errorf("newRuntimeHelper(%q) = %T, want *runcHelper", name, helper)
		}
	}
}

func TestNewRuntimeHelper_Fallback(t *testing.T) {
	tests := []string{
		"io.containerd.kata.v2",
		"io.containerd.runhcs.v1",
		"unknown-runtime",
		"",
	}
	for _, name := range tests {
		helper := newRuntimeHelper(name, nil)
		if _, ok := helper.(*fallbackHelper); !ok {
			t.Errorf("newRuntimeHelper(%q) = %T, want *fallbackHelper", name, helper)
		}
	}
}

func TestFallbackHelper_ReturnsEmpty(t *testing.T) {
	h := &fallbackHelper{}

	if procs := h.Processes(context.Background(), "test"); procs != nil {
		t.Errorf("fallback Processes() = %v, want nil", procs)
	}
	if path := h.BundlePath("/run/containerd", "default", "ctr1"); path != "" {
		t.Errorf("fallback BundlePath() = %q, want empty", path)
	}
	detail := h.Inspect(1234)
	if detail.Cmdline != "" || detail.StartedAt != "" || detail.Root != "" || detail.Cwd != "" {
		t.Errorf("fallback Inspect() should return empty processDetail, got %+v", detail)
	}
	if detail.Cgroups != nil || detail.Namespaces != nil {
		t.Errorf("fallback Inspect() slices/maps should be nil, got cgroups=%v ns=%v", detail.Cgroups, detail.Namespaces)
	}
}

func TestRuncHelper_BundlePath(t *testing.T) {
	h := &runcHelper{}
	tests := []struct {
		stateDir    string
		ns          string
		containerID string
		want        string
	}{
		{"/run/containerd", "default", "abc123", "/run/containerd/io.containerd.runtime.v2.task/default/abc123"},
		{"/custom/state", "k8s.io", "pod-xyz", "/custom/state/io.containerd.runtime.v2.task/k8s.io/pod-xyz"},
	}
	for _, tt := range tests {
		got := h.BundlePath(tt.stateDir, tt.ns, tt.containerID)
		if got != tt.want {
			t.Errorf("BundlePath(%q, %q, %q) = %q, want %q", tt.stateDir, tt.ns, tt.containerID, got, tt.want)
		}
	}
}

func TestRuncHelper_Inspect_ZeroPid(t *testing.T) {
	h := &runcHelper{}
	detail := h.Inspect(0)
	if detail.Cmdline != "" || detail.Root != "" {
		t.Errorf("Inspect(0) should return empty, got %+v", detail)
	}
}

func TestRuncHelper_Inspect_Self(t *testing.T) {
	h := &runcHelper{}
	pid := uint32(os.Getpid())
	detail := h.Inspect(pid)

	if detail.Cmdline == "" {
		t.Error("Inspect(self) should return non-empty Cmdline")
	}
	if detail.Root == "" {
		t.Error("Inspect(self) should return non-empty Root")
	}
	if detail.Cwd == "" {
		t.Error("Inspect(self) should return non-empty Cwd")
	}
	if len(detail.Cgroups) == 0 {
		t.Error("Inspect(self) should return non-empty Cgroups")
	}
	if len(detail.Namespaces) == 0 {
		t.Error("Inspect(self) should return non-empty Namespaces")
	}
	if detail.StartedAt == "" {
		t.Error("Inspect(self) should return non-empty StartedAt")
	}
}
