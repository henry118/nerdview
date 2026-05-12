# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

nerdtui is a terminal UI for containerd. It connects to the containerd daemon via gRPC socket (`/run/containerd/containerd.sock`) and presents container resources (images, containers, etc.) in interactive terminal tables.

## Build & Run

```bash
go build -o nerdtui .        # build
go run .                     # run (requires containerd socket access)
go test ./...                # run all tests
go test -run TestName ./pkg  # run a single test
```

Requires Go 1.24+. The binary needs access to a running containerd instance (typically requires root or membership in the appropriate socket group).

## Features & Behavior

**Resources displayed:** images, containers, tasks — grouped by containerd namespace.

**Default view:** table showing brief info per resource (id, name, created time). Resources are organized under their namespace.

**Live updates:** the UI subscribes to containerd events and updates the display dynamically as resources are created, modified, or removed.

**Detail dialog:** user navigates resources with cursor keys. Pressing Enter opens a dialog overlay showing detailed info about the highlighted resource. Pressing Esc dismisses the dialog.

## Architecture

The app follows the Elm Architecture via [Bubble Tea](https://github.com/charmbracelet/bubbletea):

- **Model** — application state (tables per namespace/resource type, detail dialog state)
- **Update** — handles key messages, containerd event messages, dispatches commands
- **View** — renders the model to a string using lipgloss for styling

Data flow: containerd client SDK → fetch resources → populate table rows → Bubble Tea render loop. Containerd events arrive as Bubble Tea messages and trigger incremental state updates.

## Key Dependencies

| Package | Role |
|---------|------|
| `charmbracelet/bubbletea` | TUI framework (Elm Architecture) |
| `charmbracelet/bubbles` | Reusable UI components (table, list, etc.) |
| `charmbracelet/lipgloss` | Terminal styling/layout |
| `containerd/containerd/v2` | containerd client SDK (gRPC) |
