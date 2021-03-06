// Copyright 2015 Google Inc. All Rights Reserved.
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

package cpuload

import (
	"fmt"
	"os"

	"github.com/golang/glog"
)

type CpuLoadReader struct {
	familyId uint16
	conn     *Connection
}

func New() (*CpuLoadReader, error) {
	conn, err := newConnection()
	if err != nil {
		return nil, fmt.Errorf("failed to create a new connection: %s", err)
	}

	id, err := getFamilyId(conn)
	if err != nil {
		return nil, fmt.Errorf("failed to get netlink family id for task stats: %s", err)
	}
	glog.V(2).Infof("Family id for taskstats: %d", id)
	return &CpuLoadReader{
		familyId: id,
		conn:     conn,
	}, nil
}

func (self *CpuLoadReader) Close() {
	if self.conn != nil {
		self.conn.Close()
	}
}

// This mirrors kernel internal structure.
type LoadStats struct {
	// Number of sleeping tasks.
	NrSleeping uint64 `json:"nr_sleeping"`

	// Number of running tasks.
	NrRunning uint64 `json:"nr_running"`

	// Number of tasks in stopped state
	NrStopped uint64 `json:"nr_stopped"`

	// Number of tasks in uninterruptible state
	NrUinterruptible uint64 `json:"nr_uninterruptible"`

	// Number of tasks waiting on IO
	NrIoWait uint64 `json:"nr_io_wait"`
}

// Returns instantaneous number of running tasks in a group.
// Caller can use historical data to calculate cpu load.
// path is an absolute filesystem path for a container under the CPU cgroup hierarchy.
// NOTE: non-hierarchical load is returned. It does not include load for subcontainers.
func (self *CpuLoadReader) GetCpuLoad(path string) (LoadStats, error) {
	if len(path) == 0 {
		return LoadStats{}, fmt.Errorf("cgroup path can not be empty!")
	}

	cfd, err := os.Open(path)
	if err != nil {
		return LoadStats{}, fmt.Errorf("failed to open cgroup path %s: %q", path, err)
	}

	stats, err := getLoadStats(self.familyId, cfd.Fd(), self.conn)
	if err != nil {
		return LoadStats{}, err
	}
	glog.V(1).Infof("Task stats for %q: %+v", path, stats)
	return stats, nil
}
