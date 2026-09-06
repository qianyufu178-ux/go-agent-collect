#include "max_concurrency.h"

MaxConcurrency::MaxConcurrency(int64_t max_conc)
    : max_conc_(max_conc) {
}

bool MaxConcurrency::try_acquire() {
    int64_t cur = current_.load(std::memory_order_relaxed);
    while (cur < max_conc_) {
        if (current_.compare_exchange_weak(cur, cur + 1, std::memory_order_acq_rel)) {
            return true;
        }
    }
    return false;
}

void MaxConcurrency::release() {
    current_.fetch_sub(1, std::memory_order_acq_rel);
}
