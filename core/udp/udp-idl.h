#ifndef __OPENUDP_UDP_IDL_H
#define __OPENUDP_UDP_IDL_H  1

#include "ovsdb-data.h"
#include "ovsdb-idl-provider.h"

struct udp_idl {
    struct ovsdb_idl *idl;
    struct ovsdb_idl_txn *idl_txn;
};

#endif
