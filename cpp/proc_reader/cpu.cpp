#include "cpu.h"
#include <fstream>
#include <sstream>
#include <string>
#include <unistd.h>

namespace proc {

static unsigned long long prev_idle = 0;
static unsigned long long prev_total = 0;
static bool first_read = true;

double read_cpu_usage() {
    std::ifstream file("/proc/stat");
    if (!file.is_open()) {
        return -1.0;
    }

    std::string line;
    std::getline(file, line);
    file.close();

    if (line.substr(0, 4) != "cpu ") {
        return -1.0;
    }

    std::istringstream ss(line);
    std::string cpu_label;
    unsigned long long user, nice, system, idle, iowait, irq, softirq, steal;

    ss >> cpu_label >> user >> nice >> system >> idle >> iowait >> irq >> softirq >> steal;

    unsigned long long idle_ticks = idle + iowait;
    unsigned long long total_ticks = user + nice + system + idle + iowait + irq + softirq + steal;

    if (first_read) {
        prev_idle = idle_ticks;
        prev_total = total_ticks;
        first_read = false;
        // Need at least two samples for a delta
        usleep(100000); // sleep 100ms
        return read_cpu_usage(); // recursive call for second sample
    }

    unsigned long long delta_idle = idle_ticks - prev_idle;
    unsigned long long delta_total = total_ticks - prev_total;

    prev_idle = idle_ticks;
    prev_total = total_ticks;

    if (delta_total == 0) {
        return 0.0;
    }

    double usage = (1.0 - static_cast<double>(delta_idle) / static_cast<double>(delta_total)) * 100.0;
    return usage;
}

} // namespace proc
