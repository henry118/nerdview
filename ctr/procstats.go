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
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

// DaemonStats holds resource usage metrics for the containerd daemon process.
type DaemonStats struct {
	PID     int
	CPUPct  float64       // Average CPU usage over process lifetime (like ps).
	VMS     uint64        // Virtual memory size in bytes.
	RSS     uint64        // Resident set size in bytes.
	Threads int           // Number of threads.
	Uptime  time.Duration // Time since process start.
}

// DaemonPID discovers the containerd daemon's PID from the pidfile or /proc.
func (c *Client) DaemonPID() (int, error) {
	// Try standard pidfile locations
	for _, path := range []string{
		"/run/containerd/containerd.pid",
		"/var/run/containerd/containerd.pid",
	} {
		data, err := os.ReadFile(path)
		if err == nil {
			pid, err := strconv.Atoi(strings.TrimSpace(string(data)))
			if err == nil {
				return pid, nil
			}
		}
	}
	// Fall back: find containerd process in /proc
	entries, err := os.ReadDir("/proc")
	if err != nil {
		return 0, err
	}
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		pid, err := strconv.Atoi(entry.Name())
		if err != nil {
			continue
		}
		comm, err := os.ReadFile(fmt.Sprintf("/proc/%d/comm", pid))
		if err != nil {
			continue
		}
		if strings.TrimSpace(string(comm)) == "containerd" {
			return pid, nil
		}
	}
	return 0, fmt.Errorf("containerd process not found")
}

// ReadDaemonStats reads CPU, memory, thread, and uptime info from /proc for the given PID.
func ReadDaemonStats(pid int) (DaemonStats, error) {
	stats := DaemonStats{PID: pid}

	// Read /proc/<pid>/stat for CPU and threads
	statData, err := os.ReadFile(fmt.Sprintf("/proc/%d/stat", pid))
	if err != nil {
		return stats, err
	}
	s := string(statData)
	closeParenIdx := strings.LastIndex(s, ")")
	if closeParenIdx < 0 {
		return stats, fmt.Errorf("invalid /proc/stat format")
	}
	fields := strings.Fields(s[closeParenIdx+2:])
	// fields[0]=state, [1]=ppid, ..., [11]=utime, [12]=stime, ..., [17]=num_threads, ..., [19]=starttime
	if len(fields) > 19 {
		utime, _ := strconv.ParseUint(fields[11], 10, 64)
		stime, _ := strconv.ParseUint(fields[12], 10, 64)
		numThreads, _ := strconv.Atoi(fields[17])
		starttime, _ := strconv.ParseUint(fields[19], 10, 64)

		stats.Threads = numThreads

		clkTck := uint64(100)
		totalTicks := utime + stime

		uptimeData, err := os.ReadFile("/proc/uptime")
		if err == nil {
			uptimeFields := strings.Fields(string(uptimeData))
			if len(uptimeFields) > 0 {
				systemUptime, _ := strconv.ParseFloat(uptimeFields[0], 64)
				procStartSec := float64(starttime) / float64(clkTck)
				procUptime := systemUptime - procStartSec
				if procUptime > 0 {
					stats.Uptime = time.Duration(procUptime * float64(time.Second))
					stats.CPUPct = (float64(totalTicks) / float64(clkTck)) / procUptime * 100.0
				}
			}
		}
	}

	// Read /proc/<pid>/status for VmSize and VmRSS
	statusData, err := os.ReadFile(fmt.Sprintf("/proc/%d/status", pid))
	if err != nil {
		return stats, nil
	}
	for _, line := range strings.Split(string(statusData), "\n") {
		if strings.HasPrefix(line, "VmSize:") {
			stats.VMS = parseKBValue(line)
		} else if strings.HasPrefix(line, "VmRSS:") {
			stats.RSS = parseKBValue(line)
		}
	}

	return stats, nil
}

func parseKBValue(line string) uint64 {
	fields := strings.Fields(line)
	if len(fields) >= 2 {
		val, _ := strconv.ParseUint(fields[1], 10, 64)
		return val * 1024
	}
	return 0
}
