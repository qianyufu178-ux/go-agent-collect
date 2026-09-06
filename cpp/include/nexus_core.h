#ifndef NEXUS_CORE_H
#define NEXUS_CORE_H

#include <cstdint>

#ifdef __cplusplus
extern "C" {
#endif

// ============================================================
// Rate Limiter Interfaces (CGO export)
// ============================================================

// Token Bucket
void* token_bucket_create(double rate, double burst);
int   token_bucket_allow(void* tb, double tokens);
void  token_bucket_destroy(void* tb);

// Sliding Window
void* sliding_window_create(int64_t window_ns, int64_t max_requests);
int   sliding_window_allow(void* sw);
void  sliding_window_destroy(void* sw);

// Max Concurrency
void* max_concurrency_create(int64_t max_conc);
int   max_concurrency_try_acquire(void* mc);
void  max_concurrency_release(void* mc);
void  max_concurrency_destroy(void* mc);

// ============================================================
// Proc Reader Interfaces (CGO export)
// ============================================================

// CPU usage (0.0 - 100.0)
double proc_read_cpu_usage(void);

// Memory usage percentage (0.0 - 100.0)
double proc_read_mem_usage(void);

// Disk IO (KB/s)
struct DiskIO {
    double read_kbps;
    double write_kbps;
};
struct DiskIO proc_read_disk_io(void);

// Network IO (KB/s)
struct NetIO {
    double in_kbps;
    double out_kbps;
};
struct NetIO proc_read_net_io(void);

// Disk usage percentage (0.0 - 100.0)
double proc_read_disk_usage(const char* mount_point);

#ifdef __cplusplus
}
#endif

#endif // NEXUS_CORE_H
