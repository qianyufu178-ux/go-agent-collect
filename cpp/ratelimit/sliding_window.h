#ifndef SLIDING_WINDOW_H
#define SLIDING_WINDOW_H

#include <chrono>
#include <deque>
#include <mutex>

class SlidingWindow {
public:
    SlidingWindow(int64_t window_ns, int64_t max_requests);
    ~SlidingWindow() = default;

    bool allow();

private:
    void cleanup();

    int64_t window_ns_;
    int64_t max_requests_;
    std::deque<std::chrono::steady_clock::time_point> requests_;
    std::mutex mtx_;
};

#endif // SLIDING_WINDOW_H
