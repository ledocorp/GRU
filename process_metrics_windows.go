//go:build windows

package main

import (
	"runtime"
	"syscall"
	"time"
	"unsafe"
)

type processMetrics struct {
	CPUPercent float64
	WorkingMB  float64
	HeapMB     float64
}

type processSampler struct {
	lastWall time.Time
	lastCPU  uint64
}

func newProcessSampler() *processSampler {
	return &processSampler{lastWall: time.Now(), lastCPU: processCPUTime100ns()}
}

func (s *processSampler) Sample() processMetrics {
	var mem runtime.MemStats
	runtime.ReadMemStats(&mem)

	now := time.Now()
	cpuNow := processCPUTime100ns()
	elapsed := now.Sub(s.lastWall).Seconds()
	cpuPct := 0.0
	if elapsed > 0 && cpuNow >= s.lastCPU {
		// FILETIME is 100ns units. Normalize by logical CPU count so 100% means
		// the process consumed one full machine worth of CPU capacity.
		cpuSec := float64(cpuNow-s.lastCPU) / 10_000_000.0
		cpuPct = cpuSec / (elapsed * float64(runtime.NumCPU())) * 100.0
	}
	s.lastWall = now
	s.lastCPU = cpuNow

	return processMetrics{
		CPUPercent: cpuPct,
		WorkingMB:  float64(processWorkingSetBytes()) / (1024 * 1024),
		HeapMB:     float64(mem.Alloc) / (1024 * 1024),
	}
}

var (
	kernel32                 = syscall.NewLazyDLL("kernel32.dll")
	psapi                    = syscall.NewLazyDLL("psapi.dll")
	procGetCurrentProcess    = kernel32.NewProc("GetCurrentProcess")
	procGetProcessTimes      = kernel32.NewProc("GetProcessTimes")
	procGetProcessMemoryInfo = psapi.NewProc("GetProcessMemoryInfo")
)

type filetime struct {
	LowDateTime  uint32
	HighDateTime uint32
}

func (ft filetime) uint64() uint64 {
	return uint64(ft.HighDateTime)<<32 | uint64(ft.LowDateTime)
}

type processMemoryCountersEx struct {
	CB                         uint32
	PageFaultCount             uint32
	PeakWorkingSetSize         uintptr
	WorkingSetSize             uintptr
	QuotaPeakPagedPoolUsage    uintptr
	QuotaPagedPoolUsage        uintptr
	QuotaPeakNonPagedPoolUsage uintptr
	QuotaNonPagedPoolUsage     uintptr
	PagefileUsage              uintptr
	PeakPagefileUsage          uintptr
	PrivateUsage               uintptr
}

func currentProcessHandle() uintptr {
	h, _, _ := procGetCurrentProcess.Call()
	return h
}

func processCPUTime100ns() uint64 {
	var creation, exit, kernel, user filetime
	ret, _, _ := procGetProcessTimes.Call(
		currentProcessHandle(),
		uintptr(unsafe.Pointer(&creation)),
		uintptr(unsafe.Pointer(&exit)),
		uintptr(unsafe.Pointer(&kernel)),
		uintptr(unsafe.Pointer(&user)),
	)
	if ret == 0 {
		return 0
	}
	return kernel.uint64() + user.uint64()
}

func processWorkingSetBytes() uint64 {
	var counters processMemoryCountersEx
	counters.CB = uint32(unsafe.Sizeof(counters))
	ret, _, _ := procGetProcessMemoryInfo.Call(
		currentProcessHandle(),
		uintptr(unsafe.Pointer(&counters)),
		uintptr(counters.CB),
	)
	if ret == 0 {
		return 0
	}
	return uint64(counters.WorkingSetSize)
}
