#ifndef TOKEN_BUCKET_H
#define TOKEN_BUCKET_H

#include <chrono>
#include <mutex>

class TokenBucket {
public:
    TokenBucket(double rate, double burst);
    ~TokenBucket() = default;

    bool allow(double tokens = 1.0);

private:
    void refill();

    double rate_;       // tokens per second
    double burst_;      // max tokens
    double tokens_;     // current tokens
    std::chrono::steady_clock::time_point last_refill_;
    std::mutex mtx_;
};

#endif // TOKEN_BUCKET_H
