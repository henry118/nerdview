package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/henry118/nerdtui/ctr"
)

func main() {
	namespace := flag.String("namespace", "default", "containerd namespace")
	flag.StringVar(namespace, "n", "default", "containerd namespace (shorthand)")
	flag.Parse()

	address := "/run/containerd/containerd.sock"
	if env := os.Getenv("CONTAINERD_ADDRESS"); env != "" {
		address = env
	}

	client, err := ctr.New(address)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to connect to containerd: %v\n", err)
		os.Exit(1)
	}
	defer client.Close()

	client.StartEventStream(context.Background())

	m := newModel(client, *namespace)
	p := tea.NewProgram(m, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}
