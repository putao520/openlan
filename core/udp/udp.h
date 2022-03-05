#ifndef __OPENUDP_UDP_H
#define __OPENUDP_UDP_H  1

struct udp_server {
    int port;
    int socket;
    bool reply;
};

struct udp_connect {
    int socket;
    int remote_port;
    const char *remote_address;
    u_int32_t spi;
    u_int32_t seqno;
};

void *send_ping(void *args);
void *recv_ping(void *args);

int open_socket(struct udp_server *srv);
int configure_socket(struct udp_server *srv);

#endif
