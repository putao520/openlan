//
// Created by albert on 6/19/2020.
//

#include <stdio.h>
#include <strings.h>
#include <memory.h>
#include <sys/types.h>
#include <sys/socket.h>
#include <arpa/inet.h>

#include "types.h"
#include "tun.h"
#include "tcp_server.h"

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

    int conn = 0;
    char dev_name[1024] = {0};
    int tap_fd = 0;
    conn = accept(server_fd, (struct sockaddr *)&conn_addr, &conn_addr_len);
    tap_fd = create_tap(dev_name);
    printf("open device on %s with %d\n", dev_name, tap_fd);
    while (1) {
        uint16 buf_size = 0;
        uint8 buf[4096];
        buf_size = recv(conn, buf, 4, 0);
        if (buf_size <= 0) {
            break;
        }
        int read_size = 0;

        memcpy(&read_size, buf + 2, 2);
        read_size = ntohs(read_size);
        printf("read %d\n", read_size);

        memset(buf, 0, sizeof buf);
        buf_size = recv(conn, buf, read_size, 0);
        if (buf_size != read_size) {
            printf("error on read %d != %d\n", read_size, buf_size);
            break;
        }
        int i = 0;
        for (i = 0; i < buf_size; i ++) {
            printf("%02x ", buf[i]);
        }
        printf("\n");
        write_tap(tap_fd, buf, read_size);
    }
    close(conn);
    close(server_fd);
}