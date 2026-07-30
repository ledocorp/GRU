//go:build !windows

package main

import "runtime"

type processMetrics struct {
	CPUPercent float64
	WorkingMB  float64
	HeapMB     float64
}

type processSampler struct{}

func newProcessSampler() *processSampler { return &processSampler{} }

func (s *processSampler) Sample() processMetrics {
	var mem runtime.MemStats
	runtime.ReadMemStats(&mem)
	return processMetrics{
		HeapMB: float64(mem.Alloc) / (1024 * 1024),
	}
}
