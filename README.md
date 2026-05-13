# nerdtui

A terminal UI for [containerd](https://containerd.io/). Browse images, containers, tasks, snapshots, and events in real time.

![screenshot](doc/images/screenshot.png)

## Features

- **Images** — tree view showing index, platform manifests, and layers (foldable)
- **Containers** — sandbox/container hierarchy with type column
- **Tasks** — process info with full OCI runtime spec detail (rootfs, namespaces, cgroups, capabilities, mounts)
- **Snapshots** — parent-child tree (foldable), snapshotter selection
- **Events** — live stream of containerd events
- **Live updates** — subscribes to containerd events, refreshes automatically
- **Namespace switching** — discover and switch namespaces at runtime
- **Daemon stats** — real-time CPU, memory, threads, and uptime in the title bar

## Install

```bash
git clone https://github.com/henry118/nerdtui.git
cd nerdtui
make build
```

Requires Go 1.25+.

## Usage

```bash
sudo ./nerdtui                        # default namespace
sudo ./nerdtui -n k8s.io              # specify namespace
CONTAINERD_ADDRESS=/path/to/sock ./nerdtui  # custom socket
```

The binary needs access to the containerd socket (default: `/run/containerd/containerd.sock`).

## Key Bindings

| Key | Action |
|-----|--------|
| `Tab` / `←` / `→` | Switch resource tab |
| `j` / `↓` | Move cursor down |
| `k` / `↑` | Move cursor up |
| `Space` | Fold/unfold tree node |
| `Enter` | Open detail dialog |
| `n` | Select namespace |
| `s` | Select snapshotter |
| `Esc` | Close dialog / quit |
| `Ctrl+C` | Force quit |

## License

See [LICENSE](LICENSE).
