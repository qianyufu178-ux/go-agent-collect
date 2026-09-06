#include "bridge.h"
#include "../cpp/ratelimit/token_bucket.h"
#include "../cpp/ratelimit/sliding_window.h"
#include "../cpp/ratelimit/max_concurrency.h"
#include "../cpp/proc_reader/cpu.h"
#include "../cpp/proc_reader/mem.h"
#include "../cpp/proc_reader/disk.h"
#include "../cpp/proc_reader/net.h"

// ============================================================
// Token Bucket CGO bridge
// ============================================================

void* cgo_token_bucket_create(double rate, double burst) {
    auto* tb = new TokenBucket(rate, burst);
    return static_cast<void*>(tb);
}

int cgo_token_bucket_allow(void* tb, double tokens) {
    auto* bucket = static_cast<TokenBucket*>(tb);
    return bucket->allow(tokens) ? 1 : 0;
}

void cgo_token_bucket_destroy(void* tb) {
    delete static_cast<TokenBucket*>(tb);
}

// ============================================================
// Sliding Window CGO bridge
// ============================================================

void* cgo_sliding_window_create(int64_t window_ns, int64_t max_requests) {
    auto* sw = new SlidingWindow(window_ns, max_requests);
    return static_cast<void*>(sw);
}

int cgo_sliding_window_allow(void* sw) {
    auto* window = static_cast<SlidingWindow*>(sw);
    return window->allow() ? 1 : 0;
}

void cgo_sliding_window_destroy(void* sw) {
    delete static_cast<SlidingWindow*>(sw);
}

// ============================================================
// Max Concurrency CGO bridge
// ============================================================

void* cgo_max_concurrency_create(int64_t max_conc) {
    auto* mc = new MaxConcurrency(max_conc);
    return static_cast<void*>(mc);
}

int cgo_max_concurrency_try_acquire(void* mc) {
    auto* concurrency = static_cast<MaxConcurrency*>(mc);
    return concurrency->try_acquire() ? 1 : 0;
}

void cgo_max_concurrency_release(void* mc) {
    auto* concurrency = static_cast<MaxConcurrency*>(mc);
    concurrency->release();
}

void cgo_max_concurrency_destroy(void* mc) {
    delete static_cast<MaxConcurrency*>(mc);
}

// ============================================================
// Proc Reader CGO bridge
// ============================================================

double cgo_read_cpu_usage() {
    return proc::read_cpu_usage();
}

double cgo_read_mem_usage() {
    return proc::read_mem_usage();
}

struct cgo_disk_io cgo_read_disk_io() {
    auto io = proc::read_disk_io();
    return {io.read_kbps, io.write_kbps};
}

struct cgo_net_io cgo_read_net_io() {
    auto io = proc::read_net_io();
    return {io.in_kbps, io.out_kbps};
}

double cgo_read_disk_usage(const char* mount_point) {
    return proc::read_disk_usage(mount_point);
}
