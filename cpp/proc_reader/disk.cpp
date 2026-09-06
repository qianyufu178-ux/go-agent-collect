#include "disk.h"
#include <fstream>
#include <sstream>
#include <string>
#include <vector>
#include <cstring>
#include <sys/statvfs.h>
#include <unistd.h>

namespace proc {

static unsigned long long prev_read = 0;
static unsigned long long prev_write = 0;
static bool first_read = true;

DiskIO read_disk_io() {
    DiskIO result = {0.0, 0.0};
    std::ifstream file("/proc/diskstats");
    if (!file.is_open()) {
        return result;
    }

    unsigned long long total_read = 0, total_write = 0;
    std::string line;

    while (std::getline(file, line)) {
        // Format: major minor name rio rmerge rsect ruse wio wmerge wsect wuse running use aveq
        std::istringstream ss(line);
        int major, minor;
        std::string name;
        unsigned long long rio, rmerge, rsect, ruse, wio, wmerge, wsect, wuse;
        int running, use, aveq;

        ss >> major >> minor >> name >> rio >> rmerge >> rsect >> ruse
           >> wio >> wmerge >> wsect >> wuse >> running >> use >> aveq;

        // Skip partitions (e.g., sda1, nvme0n1p1) - only count whole devices
        if (name.find('p') != std::string::npos || (name.length() > 2 && isdigit(name.back()))) {
            continue;
        }

        total_read += rsect;
        total_write += wsect;
    }
    file.close();

    if (first_read) {
        prev_read = total_read;
        prev_write = total_write;
        first_read = false;
        usleep(100000);
        return read_disk_io();
    }

    // Sectors are 512 bytes. Convert to KB.
    // delta_sectors * 512 / 1024 / time_sec = KB/s
    // With 100ms sleep: delta_sectors * 512 / 1024 / 0.1 = delta_sectors * 5
    double delta_read = static_cast<double>(total_read - prev_read) * 5.0;
    double delta_write = static_cast<double>(total_write - prev_write) * 5.0;

    prev_read = total_read;
    prev_write = total_write;

    result.read_kbps = delta_read;
    result.write_kbps = delta_write;
    return result;
}

double read_disk_usage(const char* mount_point) {
    struct statvfs stat;
    if (statvfs(mount_point, &stat) != 0) {
        return -1.0;
    }

    unsigned long long total = stat.f_blocks * stat.f_frsize;
    unsigned long long available = stat.f_bavail * stat.f_frsize;

    if (total == 0) {
        return -1.0;
    }

    double usage = (1.0 - static_cast<double>(available) / static_cast<double>(total)) * 100.0;
    return usage;
}

} // namespace proc
