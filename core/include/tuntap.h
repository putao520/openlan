//
// Created by daniel on 6/19/2020.
//

#ifndef CORE_TUNTAP_H
#define CORE_TUNTAP_H

#include <unistd.h>

#define DEV_NET_TUN "/dev/net/tun"

int create_tap(char *name);

#endif //CORE_TUNTAP_H