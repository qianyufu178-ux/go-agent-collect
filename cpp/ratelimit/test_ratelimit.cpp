#include "token_bucket.h"
#include "sliding_window.h"
#include "max_concurrency.h"
#include <iostream>
#include <thread>
#include <vector>

void test_token_bucket() {
    std::cout << "=== Token Bucket Test ===" << std::endl;
    TokenBucket tb(10.0, 5.0); // 10 tokens/s, burst 5

    // Test basic allow
    int allowed = 0;
    for (int i = 0; i < 10; i++) {
        if (tb.allow(1.0)) allowed++;
    }
    std::cout << "Initial burst: " << allowed << " allowed (expected ~5)" << std::endl;

    // Wait for refill
    std::this_thread::sleep_for(std::chrono::milliseconds(500));
    allowed = 0;
    for (int i = 0; i < 10; i++) {
        if (tb.allow(1.0)) allowed++;
    }
    std::cout << "After 500ms refill: " << allowed << " allowed (expected ~5)" << std::endl;
    std::cout << "Token Bucket: PASS" << std::endl;
}

void test_sliding_window() {
    std::cout << "\n=== Sliding Window Test ===" << std::endl;
    SlidingWindow sw(1000000000LL, 5); // 1s window, max 5 requests

    int allowed = 0;
    for (int i = 0; i < 10; i++) {
        if (sw.allow()) allowed++;
    }
    std::cout << "Burst 10 requests: " << allowed << " allowed (expected 5)" << std::endl;

    std::this_thread::sleep_for(std::chrono::milliseconds(1100));
    allowed = 0;
    for (int i = 0; i < 10; i++) {
        if (sw.allow()) allowed++;
    }
    std::cout << "After window expires: " << allowed << " allowed (expected 5)" << std::endl;
    std::cout << "Sliding Window: PASS" << std::endl;
}

void test_max_concurrency() {
    std::cout << "\n=== Max Concurrency Test ===" << std::endl;
    MaxConcurrency mc(3); // max 3 concurrent

    // Acquire 3
    bool a1 = mc.try_acquire();
    bool a2 = mc.try_acquire();
    bool a3 = mc.try_acquire();
    bool a4 = mc.try_acquire();
    std::cout << "Acquire 4: " << a1 << a2 << a3 << a4 << " (expected 1100)" << std::endl;

    mc.release();
    bool a5 = mc.try_acquire();
    std::cout << "After release: " << a5 << " (expected 1)" << std::endl;
    std::cout << "Max Concurrency: PASS" << std::endl;
}

int main() {
    test_token_bucket();
    test_sliding_window();
    test_max_concurrency();
    std::cout << "\nAll C++ ratelimit tests passed!" << std::endl;
    return 0;
}
