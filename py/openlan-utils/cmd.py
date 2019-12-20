import json
import argparse
import os
from .client import Client

parse = argparse.ArgumentParser()
parse.add_argument('--ol-token',
                   help='Token series',
                   default=os.environ.get("OL_TOKEN", ""))
parse.add_argument('--ol-server',
                   help='Server address',
                   default=os.environ.get("OL_SERVER", "localhost:10000"))


def with_client(func):
    def decorate(opt, *args, **kws):
        c = Client(opt.ol_server, opt.ol_token)
        return func(c, opt, *args, **kws)

    return decorate


subparsers = parse.add_subparsers()


def cmd(sub_parser):
    def decorate(func):
        sub_parser.set_defaults(func=func)
        return func

    return decorate


subpar = subparsers.add_parser('list-user', help="List all users")


@cmd(subpar)
@with_client
def cmd_list_user(client, opt):
    resp = client.request("user", "GET")
    print json.dumps(resp.json(), indent=4)


subpar = subparsers.add_parser('add-user', help="Add new user")
subpar.add_argument("--username", help="Username")
subpar.add_argument("--password", help="Password")


@cmd(subpar)
@with_client
def cmd_add_user(client, opt):
    if opt.username is None or opt.password is None:
        return

    data = {
        'name': opt.username,
        'password': opt.password
    }
    resp = client.request("user", "POST", data)
    print json.dumps(resp.json(), indent=4)


subpar = subparsers.add_parser('del-user', help="Delete a user")
subpar.add_argument("--username", help="Username")


@cmd(subpar)
@with_client
def cmd_del_user(client, opt):
    if opt.username is None:
        return

    data = {
        'name': opt.username
    }
    resp = client.request("user", "DELETE", data)

    print resp.text


def parse_args():
    return parse.parse_args()


