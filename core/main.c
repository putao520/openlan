#include <stdio.h>
#include "tcp_server.h"

int main(int argc, char *argv[]) {
    printf("Hello, World!\n");
    start_tcp_server(9090);
    return 0;
}