package resource

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/table"
	tasktypes "github.com/containerd/containerd/api/types/task"
	"github.com/henry118/nerdtui/ctr"
)

var TaskKind = Kind{
	Name: "Tasks",
	Columns: []Column{
		{Title: "Container ID", MinWidth: 12, Flex: true},
		{Title: "PID", MinWidth: 8},
		{Title: "Status", MinWidth: 12},
	},
	ToRows: func(data any, folded map[string]bool) []table.Row {
		tasks, ok := data.([]ctr.TaskInfo)
		if !ok || len(tasks) == 0 {
			return nil
		}
		rows := make([]table.Row, len(tasks))
		for i, t := range tasks {
			rows[i] = table.Row{
				t.Process.ID,
				fmt.Sprintf("%d", t.Process.Pid),
				t.Process.Status.String(),
			}
		}
		return rows
	},
	ToDetail: func(data any, folded map[string]bool, index int) (string, string) {
		tasks, ok := data.([]ctr.TaskInfo)
		if !ok || index < 0 || index >= len(tasks) {
			return "", ""
		}
		t := tasks[index]
		p := t.Process
		var b strings.Builder
		fmt.Fprintf(&b, "Container ID: %s\n", p.ID)
		fmt.Fprintf(&b, "PID:          %d\n", p.Pid)
		fmt.Fprintf(&b, "Status:       %s\n", p.Status)

		if p.Status == tasktypes.Status_STOPPED {
			fmt.Fprintf(&b, "Exit Status:  %d\n", p.ExitStatus)
			if p.ExitedAt != nil {
				fmt.Fprintf(&b, "Exited At:    %s\n", p.ExitedAt.AsTime().Format("2006-01-02 15:04:05"))
			}
		}

		if t.BundlePath != "" {
			fmt.Fprintf(&b, "Bundle:       %s\n", t.BundlePath)
		}

		if t.Spec != nil {
			fmt.Fprintf(&b, "\n--- Runtime Spec ---\n")
			if t.Spec.Root != nil {
				fmt.Fprintf(&b, "RootFS:       %s\n", t.Spec.Root.Path)
				fmt.Fprintf(&b, "Readonly:     %t\n", t.Spec.Root.Readonly)
			}
			if t.Spec.Hostname != "" {
				fmt.Fprintf(&b, "Hostname:     %s\n", t.Spec.Hostname)
			}
			if t.Spec.Process != nil {
				proc := t.Spec.Process
				fmt.Fprintf(&b, "\n--- Process ---\n")
				if len(proc.Args) > 0 {
					fmt.Fprintf(&b, "Args:         %s\n", strings.Join(proc.Args, " "))
				}
				fmt.Fprintf(&b, "Cwd:          %s\n", proc.Cwd)
				fmt.Fprintf(&b, "Terminal:     %t\n", proc.Terminal)
				fmt.Fprintf(&b, "User:         uid=%d gid=%d\n", proc.User.UID, proc.User.GID)
				if len(proc.User.AdditionalGids) > 0 {
					fmt.Fprintf(&b, "Groups:       %v\n", proc.User.AdditionalGids)
				}
				fmt.Fprintf(&b, "NoNewPrivs:   %t\n", proc.NoNewPrivileges)
				if proc.ApparmorProfile != "" {
					fmt.Fprintf(&b, "AppArmor:     %s\n", proc.ApparmorProfile)
				}
				if proc.SelinuxLabel != "" {
					fmt.Fprintf(&b, "SELinux:      %s\n", proc.SelinuxLabel)
				}
				if proc.OOMScoreAdj != nil {
					fmt.Fprintf(&b, "OOMScoreAdj:  %d\n", *proc.OOMScoreAdj)
				}
				if proc.Capabilities != nil {
					caps := proc.Capabilities
					if len(caps.Bounding) > 0 {
						fmt.Fprintf(&b, "Capabilities (bounding):\n")
						for _, c := range caps.Bounding {
							fmt.Fprintf(&b, "  %s\n", c)
						}
					}
					if len(caps.Effective) > 0 {
						fmt.Fprintf(&b, "Capabilities (effective):\n")
						for _, c := range caps.Effective {
							fmt.Fprintf(&b, "  %s\n", c)
						}
					}
				}
				if len(proc.Env) > 0 {
					fmt.Fprintf(&b, "Env:\n")
					for _, e := range proc.Env {
						fmt.Fprintf(&b, "  %s\n", e)
					}
				}
				if len(proc.Rlimits) > 0 {
					fmt.Fprintf(&b, "Rlimits:\n")
					for _, r := range proc.Rlimits {
						fmt.Fprintf(&b, "  %s: soft=%d hard=%d\n", r.Type, r.Soft, r.Hard)
					}
				}
			}
			if t.Spec.Linux != nil {
				linux := t.Spec.Linux
				fmt.Fprintf(&b, "\n--- Linux ---\n")
				if linux.CgroupsPath != "" {
					fmt.Fprintf(&b, "CgroupsPath:  %s\n", linux.CgroupsPath)
				}
				if linux.RootfsPropagation != "" {
					fmt.Fprintf(&b, "RootfsPropag: %s\n", linux.RootfsPropagation)
				}
				if len(linux.Namespaces) > 0 {
					fmt.Fprintf(&b, "Namespaces:\n")
					for _, ns := range linux.Namespaces {
						if ns.Path != "" {
							fmt.Fprintf(&b, "  %s: %s\n", ns.Type, ns.Path)
						} else {
							fmt.Fprintf(&b, "  %s\n", ns.Type)
						}
					}
				}
				if linux.Resources != nil {
					res := linux.Resources
					if res.Memory != nil && res.Memory.Limit != nil {
						fmt.Fprintf(&b, "MemoryLimit:  %d\n", *res.Memory.Limit)
					}
					if res.CPU != nil {
						if res.CPU.Shares != nil {
							fmt.Fprintf(&b, "CPUShares:    %d\n", *res.CPU.Shares)
						}
						if res.CPU.Quota != nil {
							fmt.Fprintf(&b, "CPUQuota:     %d\n", *res.CPU.Quota)
						}
						if res.CPU.Period != nil {
							fmt.Fprintf(&b, "CPUPeriod:    %d\n", *res.CPU.Period)
						}
						if res.CPU.Cpus != "" {
							fmt.Fprintf(&b, "CPUs:         %s\n", res.CPU.Cpus)
						}
						if res.CPU.Mems != "" {
							fmt.Fprintf(&b, "Mems:         %s\n", res.CPU.Mems)
						}
					}
					if res.Pids != nil {
						fmt.Fprintf(&b, "PidsLimit:    %d\n", res.Pids.Limit)
					}
				}
				if len(linux.MaskedPaths) > 0 {
					fmt.Fprintf(&b, "MaskedPaths:\n")
					for _, p := range linux.MaskedPaths {
						fmt.Fprintf(&b, "  %s\n", p)
					}
				}
				if len(linux.ReadonlyPaths) > 0 {
					fmt.Fprintf(&b, "ReadonlyPaths:\n")
					for _, p := range linux.ReadonlyPaths {
						fmt.Fprintf(&b, "  %s\n", p)
					}
				}
			}
			if len(t.Spec.Mounts) > 0 {
				fmt.Fprintf(&b, "\n--- Mounts ---\n")
				for _, m := range t.Spec.Mounts {
					opts := ""
					if len(m.Options) > 0 {
						opts = " [" + strings.Join(m.Options, ",") + "]"
					}
					fmt.Fprintf(&b, "  %s -> %s (%s)%s\n", m.Source, m.Destination, m.Type, opts)
				}
			}
			if len(t.Spec.Annotations) > 0 {
				fmt.Fprintf(&b, "\n--- Annotations ---\n")
				for k, v := range t.Spec.Annotations {
					fmt.Fprintf(&b, "  %s: %s\n", k, v)
				}
			}
		}
		return p.ID, b.String()
	},
}
