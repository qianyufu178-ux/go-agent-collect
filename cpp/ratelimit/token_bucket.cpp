#include "token_bucket.h"
#include <algorithm>

TokenBucket::TokenBucket(double rate, double burst)
    : rate_(rate)
    , burst_(burst)
    , tokens_(burst)
    , last_refill_(std::chrono::steady_clock::now()) {
}

void TokenBucket::refill() {
    auto now = std::chrono::steady_clock::now();
    auto elapsed = std::chrono::duration_cast<std::chrono::nanoseconds>(now - last_refill_).count();
    double tokens_to_add = (static_cast<double>(elapsed) / 1e9) * rate_;
    if (tokens_to_add > 0) {
        tokens_ = std::min(tokens_ + tokens_to_add, burst_);
        last_refill_ = now;
    }
}

bool TokenBucket::allow(double tokens) {
    std::lock_guard<std::mutex> lock(mtx_);
    refill();
    if (tokens_ >= tokens) {
        tokens_ -= tokens;
        return true;
    }
    return false;
}
