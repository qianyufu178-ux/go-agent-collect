#include "mem.h"
#include <fstream>
#include <sstream>
#include <string>

namespace proc {

double read_mem_usage() {
    std::ifstream file("/proc/meminfo");
    if (!file.is_open()) {
        return -1.0;
    }

    std::string line;
    unsigned long long total = 0, available = 0;

    while (std::getline(file, line)) {
        if (line.substr(0, 9) == "MemTotal:") {
            std::istringstream ss(line.substr(9));
            ss >> total;
        } else if (line.substr(0, 14) == "MemAvailable:") {
            std::istringstream ss(line.substr(14));
            ss >> available;
        }
    }
    file.close();

    if (total == 0) {
        return -1.0;
    }

    double usage = (1.0 - static_cast<double>(available) / static_cast<double>(total)) * 100.0;
    return usage;
}

} // namespace proc
