#ifndef NEXUS_CGO_BRIDGE_H
#define NEXUS_CGO_BRIDGE_H

#include <cstdint>

#ifdef __cplusplus
extern "C" {
#endif

// ============================================================
// Rate Limiter CGO exports
// ============================================================

void* cgo_token_bucket_create(double rate, double burst);
int   cgo_token_bucket_allow(void* tb, double tokens);
void  cgo_token_bucket_destroy(void* tb);

void* cgo_sliding_window_create(int64_t window_ns, int64_t max_requests);
int   cgo_sliding_window_allow(void* sw);
void  cgo_sliding_window_destroy(void* sw);

void* cgo_max_concurrency_create(int64_t max_conc);
int   cgo_max_concurrency_try_acquire(void* mc);
void  cgo_max_concurrency_release(void* mc);
void  cgo_max_concurrency_destroy(void* mc);

// ============================================================
// Proc Reader CGO exports
// ============================================================

double cgo_read_cpu_usage(void);
double cgo_read_mem_usage(void);

struct cgo_disk_io {
    double read_kbps;
    double write_kbps;
};
struct cgo_disk_io cgo_read_disk_io(void);

struct cgo_net_io {
    double in_kbps;
    double out_kbps;
};
struct cgo_net_io cgo_read_net_io(void);

double cgo_read_disk_usage(const char* mount_point);

#ifdef __cplusplus
}
#endif

#endif // NEXUS_CGO_BRIDGE_H
