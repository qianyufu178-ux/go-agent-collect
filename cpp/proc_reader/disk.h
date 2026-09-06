#ifndef PROC_DISK_H
#define PROC_DISK_H

namespace proc {

struct DiskIO {
    double read_kbps;
    double write_kbps;
};

DiskIO read_disk_io();
double read_disk_usage(const char* mount_point);

} // namespace proc

#endif // PROC_DISK_H
