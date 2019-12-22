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
parse.add_argument('--debug',
                   help='Enable verbose',
                   default=False)


def with_client(func):
    def decorate(opt, *args, **kws):
        c = Client(opt.ol_server, opt.ol_token, debug=opt.debug)
        return func(c, opt, *args, **kws)

    return decorate


subparsers = parse.add_subparsers()


def cmd(sub_parser):
    def decorate(func):
        sub_parser.set_defaults(func=func)
        return func

    return decorate


def Output(resp, format='json'):
    if format == 'json':
        print json.dumps(resp.json(), indent=2)
    elif format == 'yaml':
        print 'TODO'
    else:
        print resp.text


subpar = subparsers.add_parser('list-user', help="List all users")


@cmd(subpar)
@with_client
def cmd_list_user(client, opt):
    resp = client.request("user", "GET")
    Output(resp)


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
    resp = client.request("user/{}".format(opt.username), "POST", data)
    Output(resp)


subpar = subparsers.add_parser('del-user', help="Delete a user")
subpar.add_argument("--username", help="Username")


@cmd(subpar)
@with_client
def cmd_del_user(client, opt):
    if opt.username is None:
        return

    resp = client.request("user/{}".format(opt.username), "DELETE")
    Output(resp)


subpar = subparsers.add_parser('get-user', help="Get a user")
subpar.add_argument("--username", help="Username")


@cmd(subpar)
@with_client
def cmd_get_user(client, opt):
    if opt.username is None:
        return

    resp = client.request("user/{}".format(opt.username), "GET")
    Output(resp)


def parse_args():
    return parse.parse_args()


