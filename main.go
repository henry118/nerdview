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

package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/henry118/nerdtui/ctr"
	"github.com/henry118/nerdtui/logging"
)

func main() {
	namespace := flag.String("namespace", "default", "containerd namespace")
	flag.StringVar(namespace, "n", "default", "containerd namespace (shorthand)")
	debug := flag.Bool("debug", false, "enable debug logging to /var/log/nerdtui-<pid>.log")
	flag.Parse()

	if err := logging.Init(*debug); err != nil {
		fmt.Fprintf(os.Stderr, "warning: logging disabled: %v\n", err)
	}

	address := "/run/containerd/containerd.sock"
	if env := os.Getenv("CONTAINERD_ADDRESS"); env != "" {
		address = env
	}

	logging.Info("connecting to containerd at %s, namespace=%s", address, *namespace)

	client, err := ctr.New(address)
	if err != nil {
		logging.Error("failed to connect to containerd: %v", err)
		fmt.Fprintf(os.Stderr, "failed to connect to containerd: %v\n", err)
		os.Exit(1)
	}
	defer func() { _ = client.Close() }()

	client.StartEventStream(context.Background())
	logging.Info("event stream started")

	m := newModel(client, *namespace)
	p := tea.NewProgram(m, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		logging.Error("program exited with error: %v", err)
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
	logging.Info("nerdtui exited normally")
}
