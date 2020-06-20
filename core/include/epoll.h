//
// Created by albert on 6/19/2020.
//

#ifndef CORE_EPOLL_H
#define CORE_EPOLL_H

#include <sys/epoll.h>

int init_epoll();
int add_epoll(int epool_fd, int socket_fd, void *data);
int del_epoll(int epool_fd, int socket_fd);
int wait_epoll(int epool_fd, struct epoll_event *events, int max_evs, int timeout);

#endif //CORE_EPOLL_H
