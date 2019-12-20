import os
import argparse

from . import cmd


def parse_args():
    parse = argparse.ArgumentParser()
    # global options
    parse.add_argument('--ol-token',
                       help='Token series',
                       default=os.environ.get("OL_TOKEN", ""))
    parse.add_argument('--ol-server',
                       help='Server address',
                       default=os.environ.get("OL_SERVER", "localhost:10000"))

    subparse = parse.add_subparsers()
    # list user
    list_user = subparse.add_parser('list-user', help="List all users")
    list_user.set_defaults(func=cmd.list_user)
    # add user
    add_user = subparse.add_parser('add-user', help="Add new user")
    add_user.add_argument("--username", help="Username")
    add_user.add_argument("--password", help="Password")
    add_user.set_defaults(func=cmd.add_user)
    # delete user
    del_user = subparse.add_parser('del-user', help="Delete a user")
    del_user.add_argument("--username", help="Username")
    del_user.set_defaults(func=cmd.del_user)

    return parse.parse_args()


if __name__ == '__main__':
    args = parse_args()
    args.func(args)
