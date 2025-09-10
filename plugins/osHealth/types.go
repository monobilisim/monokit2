//go:build osHealth
// +build osHealth

package main

type ProcUsage struct {
	Pid  int32
	Name string
	CPU  float64
	RAM  float32
}
