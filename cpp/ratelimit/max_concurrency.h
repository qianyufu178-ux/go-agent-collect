#ifndef MAX_CONCURRENCY_H
#define MAX_CONCURRENCY_H

#include <atomic>
#include <cstdint>

class MaxConcurrency {
public:
    explicit MaxConcurrency(int64_t max_conc);
    ~MaxConcurrency() = default;

    bool try_acquire();
    void release();

private:
    int64_t max_conc_;
    std::atomic<int64_t> current_{0};
};

#endif // MAX_CONCURRENCY_H
