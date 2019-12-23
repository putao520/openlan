import os
import sys
import json
import argparse
import ruamel.yaml
from .client import Client


class Cli(object):

    def __init__(self, parser):
        self._parser = parser
        self._subparsers = {}

    def parser(self, name, **kws):
        def decorate(func):
            sub = self._subparsers.get(name)
            if sub is None:
                sub = self._parser.add_parser(name, **kws)
                self._subparsers[name] = sub
            sub.set_defaults(func=func)
            return func

        return decorate

    def argument(self, name, argument, **kws):
        def decorate(func):
            sub = self._subparsers.get(name)
            if sub is not None:
                sub.add_argument(argument, **kws)
            return func

        return decorate

    def output(self, func):
        def decorate(opt, *args, **kws):
            resp = func(opt, *args, **kws)
            if opt.format == 'json':
                json.dump(resp.json(), sys.stdout, indent=2)
            elif opt.format == 'yaml':
                yaml = ruamel.yaml.YAML()
                yaml.indent(sequence=2)
                yaml.dump(resp.json(), sys.stdout)
            else:
                sys.stdout.write(resp.text)
            return resp
        return decorate


def with_client(func):
    def decorate(opt, *args, **kws):
        c = Client(opt.ol_server, opt.ol_token, debug=opt.debug)
        return func(c, opt, *args, **kws)

    return decorate


parse = argparse.ArgumentParser()
parse.add_argument('--ol-token',
                   help='set token series',
                   default=os.environ.get("OL_TOKEN", ""))
parse.add_argument('--ol-server',
                   help='set server address',
                   default=os.environ.get("OL_SERVER", "localhost:10000"))
parse.add_argument('--debug',
                   help='print verbose log',
                   action='store_true', default=False)
parse.add_argument('--format',
                   help='set format such as json, yaml',
                   default='json')


def parse_args():
    return parse.parse_args()


cli = Cli(parse.add_subparsers())


@cli.parser('list-user', help="display all users")
@cli.output
@with_client
def cmd_list_user(client, opt):
    resp = client.request("user", "GET")
    return resp


@cli.argument('add-user', "--username", help="set username")
@cli.argument('add-user', "--password", help="set password")
@cli.parser('add-user', help="add new user")
@cli.output
@with_client
def cmd_add_user(client, opt):
    if opt.username is None or opt.password is None:
        return

    data = {
        'name': opt.username,
        'password': opt.password
    }
    resp = client.request("user/{}".format(opt.username), "POST", data)
    return resp


@cli.argument('del-user', "--username", help="set username")
@cli.parser('del-user', help="delete one user")
@cli.output
@with_client
def cmd_del_user(client, opt):
    if opt.username is None:
        return

    resp = client.request("user/{}".format(opt.username), "DELETE")
    return resp


@cli.argument('get-user', "--username", help="set username")
@cli.parser('get-user', help="get one user")
@cli.output
@with_client
def cmd_get_user(client, opt):
    if opt.username is None:
        return

    resp = client.request("user/{}".format(opt.username), "GET")
    return resp


@cli.parser('list-network', help="display all network")
@cli.output
@with_client
def cmd_list_network(client, opt):
    resp = client.request("network", "GET")
    return resp


@cli.parser('list-point', help="display all point")
@cli.output
@with_client
def cmd_list_point(client, opt):
    resp = client.request("point", "GET")
    return resp


@cli.parser('list-link', help="display all link")
@cli.output
@with_client
def cmd_list_link(client, opt):
    resp = client.request("link", "GET")
    return resp


@cli.parser('list-neighbor', help="display all neighbors")
@cli.output
@with_client
def cmd_list_neighbor(client, opt):
    resp = client.request("neighbor", "GET")
    return resp

