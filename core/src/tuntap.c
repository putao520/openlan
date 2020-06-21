//
// Created by daniel on 6/19/2020.
//

#include <assert.h>
#include <fcntl.h>
#include <sys/ioctl.h>
#include <sys/socket.h>
#include <linux/if.h>
#include <linux/if_tun.h>
#include <string.h>

#include <tuntap.h>

int create_tap(char *name) {
    struct ifreq ifr;
    int fd = -1;
    int err = -1;

    assert(NULL != name);
    if((fd = open(DEV_NET_TUN, O_RDWR)) < 0 ) {
        return -1;
    }
    memset(&ifr, 0, sizeof(ifr));
    ifr.ifr_flags = IFF_TAP | IFF_NO_PI;   /* IFF_TUN or IFF_TAP, plus maybe IFF_NO_PI */
    if (*name) {
        strncpy(ifr.ifr_name, name, IFNAMSIZ);
    }
    if((err = ioctl(fd, TUNSETIFF, (void *) &ifr)) < 0) {
        close(fd);
        return err;
    }
    strcpy(name, ifr.ifr_name);
    return fd;
}
