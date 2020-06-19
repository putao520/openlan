#include <stdio.h>
#include "tcp_server.h"

int main(int argc, char *argv[]) {
    int port = 9090;
    if (argc > 1) {
        sscanf(argv[1], "%d", &port);
    }
    printf("Listen on %d!\n", port);
    start_tcp_server(port);
    return 0;
}