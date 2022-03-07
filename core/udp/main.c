/*
 * Copyright (c) 2021-2022 OpenLAN Inc.
 *
 * This program is free software; you can redistribute it and/or modify
 * it under the terms of the GNU General Public License version 3 as
 * published by the Free Software Foundation.
 *
 */

#include <config.h>
#include <stdio.h>
#include <stdlib.h>
#include <errno.h>
#include <getopt.h>

#include "openvswitch/dynamic-string.h"
#include "openvswitch/poll-loop.h"
#include "openvswitch/vconn.h"
#include "openvswitch/vlog.h"

#include "command-line.h"
#include "confd-idl.h"
#include "daemon.h"
#include "udp.h"
#include "udp-idl.h"
#include "unixctl.h"
#include "ovs-thread.h"
#include "version.h"

VLOG_DEFINE_THIS_MODULE(main);

static char *db_remote;
static char *default_db_;
static char *udp_remote;
static int udp_port;

static inline const char *
run_dir()
{
    return "/var/openlan";
}

static inline const char *
unixctl_dir()
{
    return xasprintf("%s/%s.ctl", run_dir(), program_name);
}

static inline const char *
default_db(void)
{
    if (!default_db_) {
        default_db_ = xasprintf("unix:%s/confd.sock", run_dir());
    }
    return default_db_;
}

static inline const int
default_udp_port()
{
    return 4500;
}

static void
cancal_and_wait(pthread_t pid)
{
    if (pid > 0) {
        pthread_cancel(pid);
        pthread_join(pid, NULL);
    }
}

static void
usage(void)
{
    printf("\
%s: OpenLAN UDP Connection\n\
usage %s [OPTIONS]\n\
\n\
Options:\n\
  --port=PORT             connect to remote udp PORT\n\
                          (default: %d)\n\
  --remote=REMOTE         connect to server at REMOTE address\n\
  --db=DATABASE           connect to database at DATABASE\n\
                          (default: %s)\n\
  -h, --help              display this help message\n\
  -o, --options           list available options\n\
  -V, --version           display version information\n\
", program_name, program_name, default_udp_port(), default_db());
    vlog_usage();
    exit(EXIT_SUCCESS);
}

static void
parse_options(int argc, char *argv[])
{
    enum {
        VLOG_OPTION_ENUMS,
    };

    static struct option long_options[] = {
        {"port", required_argument, NULL, 'p'},
        {"remote", required_argument, NULL, 'r'},
        {"db", required_argument, NULL, 'd'},
        {"help", no_argument, NULL, 'h'},
        {"version", no_argument, NULL, 'V'},
        VLOG_LONG_OPTIONS,
        {NULL, 0, NULL, 0}
    };
    char *short_options = ovs_cmdl_long_options_to_short_options(long_options);

    for (;;) {
        int c;

        c = getopt_long(argc, argv, short_options, long_options, NULL);
        if (c == -1) {
            break;
        }

        switch (c) {
        case 'd':
            db_remote = xstrdup(optarg);
            break;
        case 'r':
            udp_remote = xstrdup(optarg);
            break;
        case 'p':
            udp_port = atoi(optarg);
            break;

        case 'h':
            usage();

        case 'V':
            ovs_print_version(OFP13_VERSION, OFP13_VERSION);
            exit(EXIT_SUCCESS);

        VLOG_OPTION_HANDLERS

        case '?':
            exit(EXIT_FAILURE);

        default:
            abort();
        }
    }
    free(short_options);

    if (!db_remote) {
        db_remote = xstrdup(default_db());
    }
    if (!udp_port) {
        udp_port = default_udp_port();
    }
}

static void
udp_exit(struct unixctl_conn *conn, int argc OVS_UNUSED,
        const char *argv[] OVS_UNUSED, void *exiting_)
{
    bool *exiting = exiting_;
    *exiting = true;

    unixctl_command_reply(conn, NULL);
}


void
udp_run(struct udp_idl *open_idl)
{
    if (!open_idl->idl_txn) {
        return;
    }
    const struct openrec_virtual_network *vn;

    /* Collects 'Virtual Network's. */
    OPENREC_VIRTUAL_NETWORK_FOR_EACH (vn, open_idl->idl) {
        VLOG_INFO("virtual_network: %s", vn->name);
    }
}


int
main(int argc, char *argv[])
{
    struct unixctl_server *unixctl;
    bool exiting = false;
    int retval;

    ovs_cmdl_proctitle_init(argc, argv);
    ovs_set_program_name(argv[0], CORE_PACKAGE_VERSION);
    service_start(&argc, &argv);
    parse_options(argc, argv);

    retval = unixctl_server_create(unixctl_dir(), &unixctl);
    if (retval) {
        exit(EXIT_FAILURE);
    }
    unixctl_command_register("exit", "", 0, 0, udp_exit, &exiting);

    /* Connect to OpenLAN database. */
    struct ovsdb_idl_loop open_idl_loop = OVSDB_IDL_LOOP_INITIALIZER(
        ovsdb_idl_create(db_remote, &openrec_idl_class, true, true));
    ovsdb_idl_get_initial_snapshot(open_idl_loop.idl);

    struct udp_server srv = {
        .port = udp_port,
        .socket = -1,
        .reply = true,
    };
    open_socket(&srv);
    if (configure_socket(&srv) < 0) {
        VLOG_ERR("configure_socket: %s\n", strerror(errno));
        return -1;
    }
    pthread_t send_t = 0;
    pthread_t recv_t = 0;
    if (udp_remote) {
        struct udp_connect conn = {
            .socket = srv.socket,
            .remote_port = srv.port,
            .remote_address = udp_remote,
            .spi = 0x11223344,
        };
        srv.reply = false;
        send_t = ovs_thread_create("send_ping", send_ping, (void *)&conn);
    }
    recv_t = ovs_thread_create("recv_ping", recv_ping, (void *)&srv);


    while(!exiting) {
        struct udp_idl open_idl = {
            .idl = open_idl_loop.idl,
            .idl_txn = ovsdb_idl_loop_run(&open_idl_loop),
        };
        udp_run(&open_idl);
        unixctl_server_run(unixctl);
        unixctl_server_wait(unixctl);
        if (exiting) {
            poll_immediate_wake();
        }
        ovsdb_idl_loop_commit_and_wait(&open_idl_loop);
        poll_block();
        if (should_service_stop()) {
            exiting = true;
        }
    }

    cancal_and_wait(recv_t);
    cancal_and_wait(send_t);

    unixctl_server_destroy(unixctl);
    ovsdb_idl_loop_destroy(&open_idl_loop);
    service_stop();

    free(db_remote);
    free(udp_remote);
    free(default_db_);

    exit(retval);
}
