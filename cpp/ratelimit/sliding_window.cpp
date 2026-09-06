#include "sliding_window.h"

SlidingWindow::SlidingWindow(int64_t window_ns, int64_t max_requests)
    : window_ns_(window_ns)
    , max_requests_(max_requests) {
}

void SlidingWindow::cleanup() {
    auto now = std::chrono::steady_clock::now();
    auto cutoff = now - std::chrono::nanoseconds(window_ns_);
    while (!requests_.empty() && requests_.front() < cutoff) {
        requests_.pop_front();
    }
}

bool SlidingWindow::allow() {
    std::lock_guard<std::mutex> lock(mtx_);
    cleanup();
    if (static_cast<int64_t>(requests_.size()) < max_requests_) {
        requests_.push_back(std::chrono::steady_clock::now());
        return true;
    }
    return false;
}
