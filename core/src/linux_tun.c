//
// Created by albert on 6/19/2020.
//

#include <stdio.h>
#include <assert.h>
#include <sys/types.h>
#include <sys/stat.h>
#include <fcntl.h>
#include <sys/ioctl.h>
#include <sys/socket.h>
#include <linux/if.h>
#include <linux/if_tun.h>
#include <string.h>

#include "include/linux_tun.h"

int create_tap(char *name) {
    struct ifreq ifr;
    int dev_fd = -1;
    int err = -1;
    const char *net_tun = "/dev/net/tun";

    assert(NULL != name);
    if((dev_fd = open(net_tun, O_RDWR)) < 0 ) {
        return -1;
    }
    memset(&ifr, 0, sizeof(ifr));
    ifr.ifr_flags = IFF_TAP;   /* IFF_TUN or IFF_TAP, plus maybe IFF_NO_PI */
    if (*name) {
        strncpy(ifr.ifr_name, name, IFNAMSIZ);
    }
    if((err = ioctl(dev_fd, TUNSETIFF, (void *) &ifr)) < 0) {
        close(dev_fd);
        return err;
    }
    strcpy(name, ifr.ifr_name);
    return dev_fd;
}

ssize_t read_tap(int fd, void *buf, ssize_t size) {
    assert(NULL != buf);
    return read(fd, buf, size);
}

ssize_t write_tap(int fd, const void *buf, ssize_t size) {
    assert(NULL != buf);
    return write(fd, buf, size);
}
