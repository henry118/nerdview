# AGENT.md

This file provides guidance to AI coding agents when working with code in this repository.

## Project Overview

nerdview is a terminal UI for containerd. It connects to the containerd daemon via gRPC socket (`/run/containerd/containerd.sock`) and presents container resources in interactive terminal tables with a compact htop-like style.

## Build & Run

```bash
make build                   # build (stripped binary)
make vet                     # run go vet
make test                    # run all tests
make clean                   # remove binary
go test -run TestName ./pkg  # run a single test
```

```bash
sudo ./nerdview               # run with default namespace
sudo ./nerdview -n k8s.io     # run with specific namespace
CONTAINERD_ADDRESS=/path/to/sock ./nerdview  # custom socket
```

Requires Go 1.25+. The binary needs access to a running containerd instance (typically requires root or membership in the appropriate socket group).

## Features & Behavior

**Resource tabs:** Images, Containers, Tasks, Snapshots, Events — switch with Tab/Shift+Tab or left/right arrows.

**Images:** Tree view showing image index → platform manifests → config/layers. Foldable at image and platform level (Space to toggle). Shows full digest, media type, and size. Folded by default.

**Containers:** Tree view showing sandbox → member containers. Type column distinguishes sandbox vs container. Sandbox detection uses containerd's sandbox store. Unfoldable with Space.

**Tasks:** Shows container ID, PID, status. Detail view (Enter) includes full OCI runtime spec: rootfs, process args/env/user/capabilities, Linux namespaces, cgroups, resource limits, mounts, annotations. Bundle path shown. Exit info only displayed for stopped tasks.

**Snapshots:** Tree view showing parent-child relationships. Foldable at root level (Space). Supports snapshotter selection (press `s`).

**Events:** Live stream of containerd events with timestamp, namespace, topic. Newest first, capped at 200 entries.

**Namespace switching:** Press `n` to open a selector popup listing all discovered namespaces. Default namespace is "default", overridable with `-n` flag.

**Live updates:** Subscribes to containerd events, refreshes affected resource on change. 30s periodic refresh as safety net.

**Detail dialog:** Enter opens auto-sized scrollable overlay. Esc closes. j/k to scroll.

**Key bindings:**
- `Tab`/`→` / `Shift+Tab`/`←` — switch resource tabs
- `j`/`↓` / `k`/`↑` — navigate rows
- `Space` — fold/unfold tree node
- `Enter` — open detail dialog
- `n` — namespace selector
- `s` — snapshotter selector
- `Esc` — close overlay or quit app
- `Ctrl+C` — force quit

## Architecture

The app follows the Elm Architecture via [Bubble Tea](https://github.com/charmbracelet/bubbletea):

```
nerdview/
├── main.go              # entry point: parse flags, create client, run tea.Program
├── model.go             # root model (Init/Update/View), overlay state, key routing
├── messages.go          # custom tea.Msg types
├── keymap.go            # key bindings
├── styles.go            # lipgloss styles (header, tabs, selectors)
├── ctr/
│   ├── client.go        # containerd client wrapper (namespaces, images, containers, tasks, snapshots)
│   └── events.go        # event subscription → tea.Cmd (one-event-per-Cmd pattern)
├── resource/
│   ├── resource.go      # Kind/Column/Tab abstraction, fold support, dynamic column sizing
│   ├── images.go        # ImageKind: tree with fold, content walking
│   ├── containers.go    # ContainerKind: sandbox/container tree with fold
│   ├── tasks.go         # TaskKind: process + runtime spec detail
│   ├── snapshots.go     # SnapshotKind: parent-child tree with fold
│   └── events.go        # EventKind: live event stream
└── ui/
    ├── dialog.go        # auto-sizing scrollable detail overlay
    └── help.go          # bottom help bar
```

**Data flow:** containerd client SDK → fetch resources → Kind.ToRows builds table rows (with fold state) → Bubble Tea render loop. Containerd events arrive via one-event-per-Cmd subscription and trigger targeted resource refresh.

**Extensibility:** To add a new resource type: create a file in `resource/` defining a `Kind` (columns, ToRows, ToDetail, optionally RowID/InitFolded), add a fetch method to `ctr/Client`, register the tab in `model.go`.

**Fold system:** `Kind.RowID` identifies foldable rows, `Kind.InitFolded` sets default fold state, `Tab.Folded` map tracks current state. `ToRows` receives the fold map and skips children of folded nodes.

**Column sizing:** Columns auto-size to fit content. Flex columns can shrink when total exceeds terminal width. No right-padding to fill screen.

## Key Dependencies

| Package | Role |
|---------|------|
| `charmbracelet/bubbletea` | TUI framework (Elm Architecture) |
| `charmbracelet/bubbles` | Table, viewport components |
| `charmbracelet/lipgloss` | Terminal styling/layout |
| `containerd/containerd/v2` | containerd client SDK (gRPC) |
| `opencontainers/image-spec` | OCI image types (Descriptor, Platform) |
| `opencontainers/runtime-spec` | OCI runtime spec types (for task detail) |
