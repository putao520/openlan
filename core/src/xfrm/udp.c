// This program will open udp port for espinudp.
//
// build deps:
//   kernel-headers
// compile it:
//   make udp

#include <stdio.h>
#include <stdlib.h>
#include <stdint.h>
#include <string.h>
#include <unistd.h>
#include <errno.h>
#include <sys/types.h>
#include <sys/socket.h>
#include <linux/udp.h>
#include <netinet/in.h>
#include <netinet/ip.h>
#include <linux/xfrm.h>
#include <linux/ipsec.h>
#include <linux/pfkeyv2.h>
#include <arpa/inet.h>
#include <pthread.h>

int fd = -1;
int re = -1;

void *send_ping(void *args) {
    int ret = 0;
    unsigned char buf[1024] = {0, 0, 0, 0, 1, 2, 3, 4};
    while (1) {
        struct sockaddr_in dst_addr = {
            .sin_family = AF_INET,
            .sin_port = htons(4500),
            .sin_addr = {
                .s_addr = inet_addr("117.89.130.90"),
            },
        };
        ret = sendto(fd, buf, 8, 0, (struct sockaddr*)&dst_addr, sizeof dst_addr );
        if (ret <= 0) {
            fprintf(stderr, "could not send data\n");
        }
        sleep(1);
    }
}

void *recv_ping(void *args) {
    struct sockaddr_in src_addr = {0};
    unsigned char buf[1024] = {0, 0, 0, 0, 1, 2, 3, 4};
    int ret = 0, len = sizeof src_addr;
    while (1) {
        if (re > 0) {
           sleep(2);
        }
        fprintf(stdout, "recvfrom: \n");
        ret = recvfrom(fd, buf, sizeof buf, 0, (struct sockaddr *)&src_addr, &len);
        if ( ret <= 0 ) {
            fprintf(stderr, "recvfrom: %s\n", strerror(errno));
            break;
        }
        printf("[%s:%d] %d bytes\n", inet_ntoa(src_addr.sin_addr), ntohs(src_addr.sin_port), ret);
        for (int i = 0; i < ret; i++ ) {
            fprintf(stdout, "%02x ", buf[i]);
        }
        printf("\n---\n");
        unsigned char buf[1024] = {0, 0, 0, 0, 2, 3, 4, 5, src_addr.sin_port & 0xff};
        struct sockaddr_in dst_addr = src_addr;
        ret = sendto(fd, buf, 9, 0, (struct sockaddr*)&dst_addr, sizeof dst_addr );
        if (ret <= 0) {
            fprintf(stderr, "could not send data\n");
        }
    }
}

int open_socket(int port) {
    int op = 1;
    struct sockaddr_in addr = {
        .sin_family = AF_INET,
        .sin_port = htons(port),
        .sin_addr = {
            .s_addr = INADDR_ANY,
        },
    };

    fd = socket(AF_INET, SOCK_DGRAM, IPPROTO_UDP);
    if (fd == -1) {
        return -1;
    }
    if (setsockopt(fd, SOL_SOCKET, SO_REUSEADDR, (void *)&op, sizeof op) < 0) {
        return -1;
    }
    if (bind(fd, (struct sockaddr *)&addr, sizeof addr) == -1) {
        return -1;
    }

    return fd;
}

int configure_socket() {
    int encap = UDP_ENCAP_ESPINUDP;
    struct xfrm_userpolicy_info pol;

    memset(&pol, 0, sizeof(pol));
    pol.action = XFRM_POLICY_ALLOW;
    pol.sel.family = AF_INET;

    pol.dir = XFRM_POLICY_OUT;
    if (setsockopt(fd, IPPROTO_IP, IP_XFRM_POLICY, &pol, sizeof pol) < 0) {
        return -1;
    }
    pol.dir = XFRM_POLICY_IN;
    if (setsockopt(fd, IPPROTO_IP, IP_XFRM_POLICY, &pol, sizeof pol) < 0) {
        return -1;
    }
    if (setsockopt(fd, IPPROTO_UDP, UDP_ENCAP, &encap, sizeof encap) < 0) {
        return -1;
    }
    return 0;
}

int main(int argc, char *argv[]) {
    int port = 4500;
    pthread_t send_t, recv_t;

    if (argc > 1) {
       port = atoi(argv[1]);
    }

    open_socket(port);
    if (configure_socket() < 0) {
        fprintf(stderr, "configure_socket: %s\n", strerror(errno));
        return -1;
    }

    if (argc > 2) {
        re = argc;
        pthread_create(&send_t, NULL, send_ping, NULL);
    }
    pthread_create(&recv_t, NULL, recv_ping, NULL);

    pthread_join(send_t, NULL);
    pthread_join(recv_t, NULL);

    return 0;
}
