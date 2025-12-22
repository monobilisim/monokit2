//go:build osHealth

package main

type ProcUsage struct {
	Pid  int32
	Name string
	CPU  float64
	RAM  float32
}

type DiskInfo struct {
	Device     string
	Mountpoint string
	Used       string
	Total      string
	UsedPct    float64
	Fstype     string
}

type ZFSPoolHealth struct {
	Name   string
	Health string
}

type ZFSPoolCapacity struct {
	Name     string
	Capacity int
}
