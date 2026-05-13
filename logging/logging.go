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

package logging

import (
	"fmt"
	"os"
	"time"
)

var (
	file *os.File
	pid  int
)

func Init(enabled bool) error {
	if !enabled {
		return nil
	}
	pid = os.Getpid()
	logPath := fmt.Sprintf("/var/log/nerdtui-%d.log", pid)
	f, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return fmt.Errorf("failed to open log file %s: %w", logPath, err)
	}
	file = f
	Info("nerdtui started, log=%s", logPath)
	return nil
}

func write(level, format string, args ...any) {
	if file == nil {
		return
	}
	ts := time.Now().UTC().Format("2006-01-02T15:04:05.000000000Z")
	msg := fmt.Sprintf(format, args...)
	fmt.Fprintf(file, "%s level=%s msg=%s\n", ts, level, msg)
}

func Info(format string, args ...any) {
	write("info", format, args...)
}

func Error(format string, args ...any) {
	write("error", format, args...)
}

func Debug(format string, args ...any) {
	write("debug", format, args...)
}
