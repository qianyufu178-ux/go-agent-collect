#include "net.h"
#include <fstream>
#include <sstream>
#include <string>
#include <unistd.h>

namespace proc {

static unsigned long long prev_in = 0;
static unsigned long long prev_out = 0;
static bool first_read = true;

NetIO read_net_io() {
    NetIO result = {0.0, 0.0};
    std::ifstream file("/proc/net/dev");
    if (!file.is_open()) {
        return result;
    }

    std::string line;
    unsigned long long total_in = 0, total_out = 0;

    // Skip first two header lines
    std::getline(file, line);
    std::getline(file, line);

    while (std::getline(file, line)) {
        // Format: face |bytes    packets errs drop fifo frame compressed multicast|bytes ...
        std::istringstream ss(line);
        std::string iface;
        unsigned long long bytes_in, packets_in, errs_in, drop_in;
        unsigned long long bytes_out, packets_out, errs_out, drop_out;

        ss >> iface;
        // Remove trailing colon from interface name
        if (!iface.empty() && iface.back() == ':') {
            iface.pop_back();
        }

        // Skip loopback
        if (iface == "lo") {
            continue;
        }

        ss >> bytes_in >> packets_in >> errs_in >> drop_in;
        // Skip fifo, frame, compressed, multicast (4 fields)
        std::string skip;
        for (int i = 0; i < 4; i++) ss >> skip;
        ss >> bytes_out >> packets_out >> errs_out >> drop_out;

        total_in += bytes_in;
        total_out += bytes_out;
    }
    file.close();

    if (first_read) {
        prev_in = total_in;
        prev_out = total_out;
        first_read = false;
        usleep(100000);
        return read_net_io();
    }

    // Convert bytes to KB/s. With 100ms sleep: delta_bytes / 1024 / 0.1 = delta_bytes * 0.009765625
    double delta_in = static_cast<double>(total_in - prev_in) * 0.009765625;
    double delta_out = static_cast<double>(total_out - prev_out) * 0.009765625;

    prev_in = total_in;
    prev_out = total_out;

    result.in_kbps = delta_in;
    result.out_kbps = delta_out;
    return result;
}

} // namespace proc
