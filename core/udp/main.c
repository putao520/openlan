#include <stdio.h>
#include <stdlib.h>
#include <errno.h>

#include "ovs-thread.h"
#include "udp.h"

int main(int argc, char *argv[]) {
    struct udp_server srv = {
        .port = 4500,
        .socket = -1,
    };
    pthread_t send_t = 0;
    pthread_t recv_t = 0;

    if (argc > 1) {
       srv.port = atoi(argv[1]);
    }

    open_socket(&srv);
    if (configure_socket(&srv) < 0) {
        fprintf(stderr, "configure_socket: %s\n", strerror(errno));
        return -1;
    }

    struct udp_connect conn = {
        .socket = srv.socket,
        .remote_port = 4500,
        .remote_address = "117.89.130.90",
    };
    if (argc > 2) {
        srv.reply = false;
        send_t = ovs_thread_create("send_ping", send_ping, (void *)&conn);
    }
    recv_t = ovs_thread_create("recv_ping", recv_ping, (void *)&srv);
 
    pause();

    fprintf(stdout, "exit.\n");

    if (send_t > 0) {
        pthread_cancel(send_t);
        pthread_join(send_t, NULL);
    }
    pthread_cancel(recv_t);
    pthread_join(recv_t, NULL);
    return 0;
}
