//
// Created by albert on 6/19/2020.
//

#include <stdio.h>
#include <strings.h>
#include <memory.h>
#include <sys/types.h>
#include <sys/socket.h>
#include <arpa/inet.h>

int start_tcp_server(int port) {
    struct sockaddr_in server_addr;

    bzero(&server_addr, sizeof(struct sockaddr_in));
    server_addr.sin_family = AF_INET;
    server_addr.sin_addr.s_addr = htonl(INADDR_ANY);
    server_addr.sin_port = htons(port);

    int fd = socket(AF_INET, SOCK_STREAM, 0);
    if(bind(fd, (struct sockaddr*)&server_addr, sizeof(server_addr)) < 0) {
        printf("bind error\n");
        return -1;
    }
    if(listen(fd, 2) < 0) {
        printf("listen error\n");
        return -1;
    }

    struct sockaddr_in conn_addr;
    socklen_t conn_addr_len = sizeof(conn_addr);

    int conn = accept(fd, (struct sockaddr *)&conn_addr, &conn_addr_len);
    while (1) {
        int buf_size = 0;
        char buf[4096];
        buf_size = recv(conn, buf, 4, 0);
        if (buf_size == -1) {
            break;
        }
        int readSize = 0;
        memcpy(&readSize, buf, sizeof readSize);
        readSize = ntohl(readSize);
        buf_size = recv(conn, buf, readSize, 0);
        if (buf_size == -1) {
            break;
        }
        int i = 0;
        for (i = 0; i < buf_size; i ++) {
            printf("%x ", buf);
        }
        printf("\n");
    }
}