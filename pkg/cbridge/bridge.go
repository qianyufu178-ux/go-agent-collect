package cbridge

/*
#cgo LDFLAGS: -L${SRCDIR}/../../build/cpp -lnexus_core -lstdc++
#cgo CFLAGS: -I${SRCDIR}/../../cpp/include -I${SRCDIR}
#include "bridge.h"
*/
import "C"
import (
	"unsafe"
)

// ============================================================
// Token Bucket
// ============================================================

type TokenBucket struct {
	ptr unsafe.Pointer
}

func NewTokenBucket(rate, burst float64) *TokenBucket {
	ptr := C.cgo_token_bucket_create(C.double(rate), C.double(burst))
	return &TokenBucket{ptr: ptr}
}

func (tb *TokenBucket) Allow(tokens float64) bool {
	return C.cgo_token_bucket_allow(tb.ptr, C.double(tokens)) != 0
}

func (tb *TokenBucket) Destroy() {
	C.cgo_token_bucket_destroy(tb.ptr)
}

// ============================================================
// Sliding Window
// ============================================================

type SlidingWindow struct {
	ptr unsafe.Pointer
}

func NewSlidingWindow(windowNs int64, maxRequests int64) *SlidingWindow {
	ptr := C.cgo_sliding_window_create(C.int64_t(windowNs), C.int64_t(maxRequests))
	return &SlidingWindow{ptr: ptr}
}

func (sw *SlidingWindow) Allow() bool {
	return C.cgo_sliding_window_allow(sw.ptr) != 0
}

func (sw *SlidingWindow) Destroy() {
	C.cgo_sliding_window_destroy(sw.ptr)
}

// ============================================================
// Max Concurrency
// ============================================================

type MaxConcurrency struct {
	ptr unsafe.Pointer
}

func NewMaxConcurrency(maxConc int64) *MaxConcurrency {
	ptr := C.cgo_max_concurrency_create(C.int64_t(maxConc))
	return &MaxConcurrency{ptr: ptr}
}

func (mc *MaxConcurrency) TryAcquire() bool {
	return C.cgo_max_concurrency_try_acquire(mc.ptr) != 0
}

func (mc *MaxConcurrency) Release() {
	C.cgo_max_concurrency_release(mc.ptr)
}

func (mc *MaxConcurrency) Destroy() {
	C.cgo_max_concurrency_destroy(mc.ptr)
}

// ============================================================
// Proc Reader
// ============================================================

type DiskIO struct {
	ReadKbps  float64
	WriteKbps float64
}

type NetIO struct {
	InKbps  float64
	OutKbps float64
}

func ReadCPUUsage() float64 {
	return float64(C.cgo_read_cpu_usage())
}

func ReadMemUsage() float64 {
	return float64(C.cgo_read_mem_usage())
}

func ReadDiskIO() DiskIO {
	io := C.cgo_read_disk_io()
	return DiskIO{
		ReadKbps:  float64(io.read_kbps),
		WriteKbps: float64(io.write_kbps),
	}
}

func ReadNetIO() NetIO {
	io := C.cgo_read_net_io()
	return NetIO{
		InKbps:  float64(io.in_kbps),
		OutKbps: float64(io.out_kbps),
	}
}

func ReadDiskUsage(mountPoint string) float64 {
	mp := C.CString(mountPoint)
	defer C.free(unsafe.Pointer(mp))
	return float64(C.cgo_read_disk_usage(mp))
}
