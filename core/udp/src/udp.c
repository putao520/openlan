#include <stdio.h>
#include <stdlib.h>
#include <stdint.h>
#include <stdbool.h>
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

#include "udp.h"

void *send_ping(void *args) {
    struct udp_connect *conn = (struct udp_connect *)args;
    int retval = 0;
    unsigned char buf[1024] = {0, 0, 0, 0, 1, 2, 3, 4};
    while (true) {
        struct sockaddr_in dst_addr = {
            .sin_family = AF_INET,
            .sin_port = htons(conn->remote_port),
            .sin_addr = {
                .s_addr = inet_addr(conn->remote_address),
            },
        };
        retval = sendto(conn->socket, buf, 8, 0, (struct sockaddr*)&dst_addr, sizeof dst_addr);
        if (retval <= 0) {
            fprintf(stderr, "could not send data\n");
        }
        sleep(1);
    }
}

void *recv_ping(void *args) {
    struct udp_server *srv = (struct udp_server *)args;
    struct sockaddr_in src_addr = {0};
    unsigned char buf[1024] = {0, 0, 0, 0, 1, 2, 3, 4};
    int i, retval = 0, len = sizeof src_addr;

    while (true) {
        retval = recvfrom(srv->socket, buf, sizeof buf, 0, (struct sockaddr *)&src_addr, &len);
        if ( retval <= 0 ) {
            fprintf(stderr, "recvfrom: %s\n", strerror(errno));
            break;
        }
        printf("recvfrom: [%s:%d] %d bytes\n", inet_ntoa(src_addr.sin_addr), ntohs(src_addr.sin_port), retval);
        for (i = 0; i < retval; i++ ) {
            fprintf(stdout, "%02x ", buf[i]);
        }
        printf("\n---\n");
        if (srv->reply) {
            unsigned char buf[1024] = {0, 0, 0, 0, 2, 3, 4, 5, src_addr.sin_port & 0xff};
            struct sockaddr_in dst_addr = src_addr;
            retval = sendto(srv->socket, buf, 9, 0, (struct sockaddr*)&dst_addr, sizeof dst_addr );
            if (retval <= 0) {
                fprintf(stderr, "could not send data\n");
            }
        }
    }
}

int open_socket(struct udp_server *srv) {
    int op = 1;
    struct sockaddr_in addr = {
        .sin_family = AF_INET,
        .sin_port = htons(srv->port),
        .sin_addr = {
            .s_addr = INADDR_ANY,
        },
    };

    srv->socket = socket(AF_INET, SOCK_DGRAM, IPPROTO_UDP);
    if (srv->socket == -1) {
        return -1;
    }
    if (setsockopt(srv->socket, SOL_SOCKET, SO_REUSEADDR, (void *)&op, sizeof op) < 0) {
        return -1;
    }
    if (bind(srv->socket, (struct sockaddr *)&addr, sizeof addr) == -1) {
        return -1;
    }

    return srv->socket;
}

int configure_socket(struct udp_server *srv) {
    int encap = UDP_ENCAP_ESPINUDP;
    struct xfrm_userpolicy_info pol;

    memset(&pol, 0, sizeof(pol));
    pol.action = XFRM_POLICY_ALLOW;
    pol.sel.family = AF_INET;

    pol.dir = XFRM_POLICY_OUT;
    if (setsockopt(srv->socket, IPPROTO_IP, IP_XFRM_POLICY, &pol, sizeof pol) < 0) {
        return -1;
    }
    pol.dir = XFRM_POLICY_IN;
    if (setsockopt(srv->socket, IPPROTO_IP, IP_XFRM_POLICY, &pol, sizeof pol) < 0) {
        return -1;
    }
    if (setsockopt(srv->socket, IPPROTO_UDP, UDP_ENCAP, &encap, sizeof encap) < 0) {
        return -1;
    }
    return srv->socket;
}
