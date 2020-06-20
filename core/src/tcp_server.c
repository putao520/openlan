//
// Created by albert on 6/19/2020.
//

#include <stdio.h>
#include <stdlib.h>
#include <strings.h>
#include <memory.h>
#include <fcntl.h>
#include <sys/types.h>
#include <sys/socket.h>
#include <arpa/inet.h>
#include <assert.h>
#include <pthread.h>

#include "../include/types.h"
#include "../include/tuntap.h"
#include "../include/epoll.h"
#include "../include/tcp_server.h"

int non_blocking(int fd) {
    int flags = fcntl(fd, F_GETFL, 0);
    return fcntl(fd, F_SETFL, flags | O_NONBLOCK);
}

int recv_full(int fd, char *buf, ssize_t size) {
    ssize_t read_size = 0;

    for (; size > 0;) {
        read_size = recv(fd, buf, size, 0);
        if (read_size <= 0) return read_size;
        buf += read_size;
        size -= read_size;
    }
    return 0;
}

int send_full(int fd, char *buf, ssize_t size) {
    ssize_t write_size = 0;

    for (;size > 0;) {
        write_size = send(fd, buf, size, 0);
        if (write_size <= 0) return write_size;
        buf += write_size;
        size -= write_size;
    }
    return 0;
}

typedef struct {
    int socket_fd;
    int device_fd;
}connection_t;

void *read_client(void *argv) {
    uint16_t buf_size = 0;
    uint16_t read_size = 0;
    uint8_t buf[4096];
    connection_t *conn = NULL;

    assert(NULL != argv);
    conn = (connection_t *) argv;

    for(;;) {
        buf_size = recv_full(conn->socket_fd, buf, 4);
        if (buf_size != 0) {
            break;
        }
        read_size = ntohs(*(uint16_t *)(buf + 2));
//      printf("read %d\n", read_size);
//
        memset(buf, 0, sizeof buf);
        buf_size = recv_full(conn->socket_fd, buf, read_size);
        if (buf_size != 0) {
            printf("ERROR: on read %d != %d\n", read_size, buf_size);
            break;
        }
//        for (i = 0; i < buf_size; i ++) {
//            printf("%02x ", buf[i]);
//        }
//        printf("\n");
        write_tap(conn->device_fd, buf, read_size);
    }
}

void *read_device(void *argv) {
    uint16_t write_size = 0;
    uint16_t read_size = 0;
    uint8_t buf[4096];
    connection_t *conn = NULL;

    assert(NULL != argv);
    conn = (connection_t *) argv;

    for(;;) {
        read_size = read_tap(conn->device_fd, buf + 4, sizeof (buf));
        if (read_size <= 0) {
            continue;
        }
        *(uint16_t *)(buf + 2) = htons(read_size);
        read_size += 4;
        write_size = send_full(conn->socket_fd, buf, read_size);
        if (write_size != 0) {
            printf("ERROR: write to conn %d:%d", write_size, read_size);
            break;
        }
    }
}

int start_tcp_server(int port) {
    struct sockaddr_in server_addr;

    bzero(&server_addr, sizeof(struct sockaddr_in));
    server_addr.sin_family = AF_INET;
    server_addr.sin_addr.s_addr = htonl(INADDR_ANY);
    server_addr.sin_port = htons(port);

    int server_fd = 0;
    server_fd = socket(AF_INET, SOCK_STREAM, 0);
    if(bind(server_fd, (struct sockaddr*)&server_addr, sizeof(server_addr)) < 0) {
        printf("bind error\n");
        return -1;
    }
    if(listen(server_fd, 2) < 0) {
        printf("listen error\n");
        return -1;
    }

    struct sockaddr_in conn_addr;
    socklen_t conn_addr_len = sizeof(conn_addr);

    int conn_fd = 0;
    char dev_name[1024] = {0};
    int tap_fd = 0;

    conn_fd = accept(server_fd, (struct sockaddr *)&conn_addr, &conn_addr_len);
    printf("accept connection on %d\n", conn_fd);

    tap_fd = create_tap(dev_name);
    printf("open device on %s with %d\n", dev_name, tap_fd);

    connection_t conn = {
        socket_fd: conn_fd,
        device_fd: tap_fd,
    };
    pthread_t client_thread;
    pthread_t device_thread;

    if(pthread_create(&client_thread, NULL, read_client, &conn)) {
        fprintf(stderr, "Error creating thread\n");
        return 1;
    }
    if(pthread_create(&device_thread, NULL, read_device, &conn)) {
        fprintf(stderr, "Error creating thread\n");
        return 1;
    }

    if(pthread_join(client_thread, NULL)) {
        fprintf(stderr, "Error joining thread\n");
        return 2;
    }
    if(pthread_join(device_thread, NULL)) {
        fprintf(stderr, "Error joining thread\n");
        return 2;
    }

finish:
    close(conn_fd);
    close(server_fd);
    close(tap_fd);
    printf("exit from %d\n", server_fd);

    return 0;
}