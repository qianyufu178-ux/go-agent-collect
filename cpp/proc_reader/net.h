#ifndef PROC_NET_H
#define PROC_NET_H

namespace proc {

struct NetIO {
    double in_kbps;
    double out_kbps;
};

NetIO read_net_io();

} // namespace proc

#endif // PROC_NET_H
