//
// Created by albert on 6/19/2020.
//

#include <stdio.h>
#include <memory.h>
#include <assert.h>

#include "../include/epoll.h"

int init_epoll() {
    int epool_fd = 0;

    if((epool_fd = epoll_create(1024)) < 0) {
        fprintf(stderr, "EPOLL CREATE failed");
        return -1;
    }
    return epool_fd;
}

int add_epoll(int epool_fd, int socket_fd, void *data) {
    struct epoll_event event = {0};

    memset(&event, 0, sizeof(event));
    event.events = EPOLLIN;
    event.data.fd = socket_fd;
    printf("%d:%d.....\n", socket_fd, event.data.fd);
    if(epoll_ctl(epool_fd, EPOLL_CTL_ADD, socket_fd, &event) < 0) {
        fprintf(stderr, "EPOLL CTL_ADD %d failed", socket_fd);
        return -1;
    }
    return socket_fd;
}

int del_epoll(int epool_fd, int socket_fd) {
    struct epoll_event event = {0};

    event.events  = EPOLLIN;
    event.data.fd = socket_fd;
    if(epoll_ctl(epool_fd, EPOLL_CTL_DEL, socket_fd, &event) < 0) {
        fprintf(stderr, "EPOLL CTL_DEL %d failed", socket_fd);
        return -1;
    }
    return socket_fd;
}

int wait_epoll(int epool_fd, struct epoll_event *events, int max_evs, int timeout) {
    assert(NULL != events);

    return epoll_wait(epool_fd, events, max_evs, timeout);
}
