#include "cpu.h"
#include "mem.h"
#include "disk.h"
#include "net.h"
#include <iostream>

int main() {
    std::cout << "=== Proc Reader Tests ===" << std::endl;

    // CPU usage (takes ~100ms for first sample)
    double cpu = proc::read_cpu_usage();
    std::cout << "CPU Usage: " << cpu << "%" << std::endl;

    // Memory usage
    double mem = proc::read_mem_usage();
    std::cout << "Memory Usage: " << mem << "%" << std::endl;

    // Disk IO
    auto disk_io = proc::read_disk_io();
    std::cout << "Disk IO - Read: " << disk_io.read_kbps << " KB/s, Write: " << disk_io.write_kbps << " KB/s" << std::endl;

    // Disk usage
    double disk_usage = proc::read_disk_usage("/");
    std::cout << "Disk Usage (/): " << disk_usage << "%" << std::endl;

    // Network IO
    auto net_io = proc::read_net_io();
    std::cout << "Net IO - In: " << net_io.in_kbps << " KB/s, Out: " << net_io.out_kbps << " KB/s" << std::endl;

    std::cout << "\nAll proc reader tests passed!" << std::endl;
    return 0;
}
