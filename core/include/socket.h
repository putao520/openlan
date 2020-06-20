//
// Created by albert on 6/19/2020.
//

#ifndef CORE_SOCKET_H
#define CORE_SOCKET_H

#include "types.h"

typedef struct {
    int socket_fd;
    int device_fd;
} peer_t;

int start_tcp_server(uint16_t port);
int start_tcp_client(const char *addr, uint16_t port);

#endif //CORE_SOCKET_H